package analytics

import (
	"github.com/nico612/iam-demo/pkg/log"
	"github.com/nico612/iam-demo/pkg/storage"
	"github.com/vmihailenco/msgpack/v5"
	"sync"
	"sync/atomic"
	"time"
)

const analyticsKeyName = "iam-system-analytics"

const (
	// 缓冲区强制刷新时间间隔
	recordsBufferForcedFlushInterval = 1 * time.Second
)

// AnalyticsRecord encodes the details of a authorization request.
type AnalyticsRecord struct {
	TimeStamp  int64     `json:"timestamp"`  // 记录时间
	Username   string    `json:"username"`   // 用户名
	Effect     string    `json:"effect"`     // 作用效果
	Conclusion string    `json:"conclusion"` // 结果
	Request    string    `json:"request"`    // 请求
	Policies   string    `json:"policies"`   // 策略
	Deciders   string    `json:"deciders"`
	ExpireAt   time.Time `json:"expireAt"   bson:"expireAt"` // 缓存过期时间
}

// SetExpiry set expiration time to a key.
func (a *AnalyticsRecord) SetExpiry(expiresInSeconds int64) {
	expiry := time.Duration(expiresInSeconds) * time.Second
	if expiresInSeconds == 0 {
		// Expiry is set to 100 years
		expiry = 24 * 365 * 100 * time.Hour
	}

	t := time.Now()
	t2 := t.Add(expiry)
	a.ExpireAt = t2
}

// Analytics will record analytics data to a redis back end as defined in the Config object.
type Analytics struct {
	store                      storage.AnalyticsHandler // 缓存出来，authz项目使用的时redis来缓存
	poolSize                   int                      // 协程 大小
	recordsChan                chan *AnalyticsRecord    // 记录通道
	workerBufferSize           uint64                   // 每个协程 处理缓存的大小
	recordsBufferFlushInterval uint64                   // 缓存刷新时间间隔, 单位 毫秒
	shouldStop                 uint32                   // 是否停止
	poolWg                     sync.WaitGroup
}

var analytics *Analytics

// NewAnalytics returns a new analytics instance.
// store 缓存处理程序
func NewAnalytics(options *AnalyticsOptions, store storage.AnalyticsHandler) *Analytics {
	ps := options.PoolSize
	recordsBufferSize := options.RecordsBufferSize
	workerBufferSize := recordsBufferSize / uint64(ps) // 每个协程处理的缓存大小
	log.Debug("Analytics pool worker buffer size", log.Uint64("workerBufferSize", workerBufferSize))

	recordsChan := make(chan *AnalyticsRecord, recordsBufferSize)

	analytics = &Analytics{
		store:                      store,
		poolSize:                   ps,
		recordsChan:                recordsChan,
		workerBufferSize:           workerBufferSize,
		recordsBufferFlushInterval: options.FlushInterval,
	}

	return analytics
}

// GetAnalytics returns the existed analytics instance.
// Need to initialize `analytics` instance before calling GetAnalytics.
func GetAnalytics() *Analytics {
	return analytics
}

// Start the analytics service.
func (r *Analytics) Start() {
	r.store.Connect() // 连接 redis

	// start worker pool
	// 原子操作，将r.shouldStop 值设置为0，线程安全
	atomic.SwapUint32(&r.shouldStop, 0)

	// 开启多个协程，处理记录
	for i := 0; i < r.poolSize; i++ {
		r.poolWg.Add(1)
		go r.recordWorker()
	}
}

// Stop the analytics service.
func (r *Analytics) Stop() {

	// 标记停止发送 records 到 redis
	atomic.SwapUint32(&r.shouldStop, 1)

	// 关闭通道, 关闭通道后，每个协程会将剩余未处理的 records 发送到 redis 并退出协程
	close(r.recordsChan)

	// 等待每个协程处理完成
	r.poolWg.Wait()
}

// RecordHit will store an AnalyticsRecord in Redis.
func (r *Analytics) RecordHit(record *AnalyticsRecord) error {
	// 检查通道是否停止，线程安全操作
	if atomic.LoadUint32(&r.shouldStop) > 0 {
		return nil
	}

	// just send record to channel consumed by pool of workers
	// leave all data crunching and Redis I/O work for pool workers
	r.recordsChan <- record

	return nil
}

// 记录操作，监听记录通道，当缓冲区满或者强制刷新时间到时，将缓存 发送到 redis
func (r *Analytics) recordWorker() {
	defer r.poolWg.Done()

	// 这是向 redis 发送一个管道命令的缓冲区
	// 使用 r.recordsBufferSize 作为上限来减少切片重新分配
	recordsBuffer := make([][]byte, 0, r.workerBufferSize)

	// read records from channel and process
	lastSentTS := time.Now()

	for {
		var readyToSend bool

		select {
		case record, ok := <-r.recordsChan: // 监听记录通道
			if !ok {
				// 当通道关闭时，发送缓冲区的剩余内容
				r.store.AppendToSetPipelined(analyticsKeyName, recordsBuffer)
				return
			}

			// 有新记录 - 就添加到缓冲区
			if encoded, err := msgpack.Marshal(record); err != nil {
				log.Errorf("Error encoding analytics data: %s", err.Error())
			} else {
				recordsBuffer = append(recordsBuffer, encoded)
			}

			// 标记是否准本好发送
			readyToSend = uint64(len(recordsBuffer)) == r.workerBufferSize

		case <-time.After(time.Duration(r.recordsBufferFlushInterval) * time.Millisecond):
			readyToSend = true
		}

		if len(recordsBuffer) > 0 && (readyToSend || time.Since(lastSentTS) >= recordsBufferForcedFlushInterval) {
			r.store.AppendToSetPipelined(analyticsKeyName, recordsBuffer)
			recordsBuffer = recordsBuffer[:0]
			lastSentTS = time.Now() // 重置处理时间
		}
	}
}

// DurationToMillisecond convert time duration type to float64.
func DurationToMillisecond(d time.Duration) float64 {
	return float64(d) / 1e6
}
