package load

import (
	"context"
	"github.com/marmotedu/errors"
	"github.com/nico612/iam-demo/pkg/log"
	"github.com/nico612/iam-demo/pkg/storage"
	"sync"
	"time"
)

type Loader interface {
	Reload() error
}

type Load struct {
	ctx    context.Context
	lock   *sync.RWMutex
	loader Loader
}

// NewLoader return a loader with a loader implement.
func NewLoader(ctx context.Context, loader Loader) *Load {
	return &Load{
		ctx:    ctx,
		lock:   new(sync.RWMutex),
		loader: loader,
	}

}

// Start a loop service.
func (l *Load) Start() {
	go startPubSubLoop()   // 订阅 redis 消息，并写入刷新队列
	go l.reloadQueueLoop() // 读取刷新队列

	// 1s is the minimum amount of time between hot reloads. The
	// interval counts from the start of one reload to the next.
	go l.reloadLoop()

}

// 订阅 redis 消息
func startPubSubLoop() {
	cacheStore := storage.RedisCluster{}
	cacheStore.Connect()

	// On message, Synchronize
	for {
		// 订阅 redis 通道, 并将回调的消息 写入 reloadQueue 队列中
		err := cacheStore.StartPubSubHandler(RedisPubSubChannel, func(v interface{}) {
			handleRedisEvent(v, nil, nil)
		})

		if err != nil {
			if !errors.Is(err, storage.ErrRedisIsDown) {
				log.Errorf("Connection to Redis failed, reconnect in 10s: %s", err.Error())
			}
		}

		time.Sleep(10 * time.Second)
		log.Warnf("Reconnecting: %s", err.Error())
	}
}

// shouldReload returns true if we should perform any reload. Reloads happens if
// we have reload callback queued.
func shouldReload() ([]func(), bool) {
	requeueLock.Lock()
	defer requeueLock.Unlock()

	if len(requeue) == 0 {
		return nil, false
	}

	n := requeue
	requeue = []func(){}

	return n, true
}

func (l *Load) reloadLoop(complete ...func()) {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-l.ctx.Done():
			return
		// We don't check for reload right away as the gateway peroms this on the
		// startup sequence. We expect to start checking on the first tick after the
		// gateway is up and running.
		case <-ticker.C:
			cb, ok := shouldReload()
			if !ok {
				continue
			}
			start := time.Now()
			l.DoReload()
			for _, c := range cb {
				// most of the callbacks are nil, we don't want to execute nil functions to
				// avoid panics.
				if c != nil {
					c() // 处理 requeue 中的回调函数
				}
			}
			if len(complete) != 0 {
				complete[0]()
			}
			log.Infof("reload: cycle completed in %v", time.Since(start))
		}
	}
}

// reloadQueue used to queue a reload. It's not
// buffered, as reloadQueueLoop should pick these up immediately.
var reloadQueue = make(chan func())

var requeueLock sync.Mutex

// This is a list of callbacks to execute on the next reload. It is protected by
// requeueLock for concurrent use.
var requeue []func()

func (l *Load) reloadQueueLoop(cb ...func()) {
	for {
		select {
		case <-l.ctx.Done():
			return
		case fn := <-reloadQueue: // 从刷新队列中读取数据
			requeueLock.Lock()
			requeue = append(requeue, fn) // 储存requeue函数
			requeueLock.Unlock()
			log.Info("Reload queued")
			if len(cb) != 0 {
				cb[0]()
			}
		}
	}
}

// DoReload reload secrets and policies.
func (l *Load) DoReload() {
	l.lock.Lock()
	defer l.lock.Unlock()

	// 刷新缓存
	if err := l.loader.Reload(); err != nil {
		log.Errorf("faild to refresh target storage: %s", err.Error())
	}

	log.Debug("refresh target storage succ")
}
