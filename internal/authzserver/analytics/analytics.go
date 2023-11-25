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
	store                      storage.AnalyticsHandler // 接口类型，提供了 Connect 和 AppendToSetPipelined 函数，分别用来连接和上报数据给储存系统，本项目使用 redis
	poolSize                   int                      // 指定开启 worker 开启的个数，也就是开启多少个 Go 协程来消费 recordsChan
	recordsChan                chan *AnalyticsRecord    // 授权日志会缓存在到该通道中
	workerBufferSize           uint64                   // 批量投递给下游系统的消息数，批量投递可以进一步提高消费能力、减少CPU消耗
	recordsBufferFlushInterval uint64                   // 设置最迟多久投递一次，也就是投递数据的超时时间
	shouldStop                 uint32                   // 是否停止
	poolWg                     sync.WaitGroup
}

var analytics *Analytics

// NewAnalytics returns a new analytics instance.
// storage 缓存处理程序
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

// Start 启动数据上报服务，改服务会启动多个 worker 来监听 recordsChan 中的消息，然后将数据上报 给 redis
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

// RecordHit will storage an AnalyticsRecord in Redis.
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

// recordWorker 函数会将接收到的授权日志保存在 recordsBuffer 切片中，
// 当数组内元素个数为 workerBufferSize， 或者距离上次投递的时间间隔 recordsBufferFlushInterval 时，
// 就会将 recordsBuffer 数组中的数据上报给目标系统（redis）
// 设计技巧：
// 使用 msgpack 序列化消息，比 json 更快、更小
// 支持 Batch Windows：当 worker 的消息数达到指定阀值时，会批量投递消息给 Redis
// 超时投递：为了避免因为产生消息太慢，一直到不到阀值，从而出现无法投递消息的情况，投递逻辑也支持超时投递
// 支持优雅关停：当 recordsChan 关闭时，将 recordsBuffer 中的消息批量投递给 redis，之后退出worker协程
func (r *Analytics) recordWorker() {
	defer r.poolWg.Done()

	// 这是向 Redis 发送一个流水线命令的缓冲区
	// 使用 r.recordsBufferSize 作为数据采集的上游以减少切片重新分配
	recordsBuffer := make([][]byte, 0, r.workerBufferSize)

	// read records from channel and process
	lastSentTS := time.Now()

	for {
		var readyToSend bool

		select {
		case record, ok := <-r.recordsChan: // 监听记录通道
			if !ok {
				// 当通道关闭时，发送缓冲区的剩余内容 到 redis 中
				r.store.AppendToSetPipelined(analyticsKeyName, recordsBuffer)
				return
			}

			// 有新记录 - 就添加到缓冲区
			if encoded, err := msgpack.Marshal(record); err != nil {
				log.Errorf("Error encoding analytics data: %s", err.Error())
			} else {
				recordsBuffer = append(recordsBuffer, encoded)
			}

			// 确保 recordsBuffer 处于发送就绪状态
			readyToSend = uint64(len(recordsBuffer)) == r.workerBufferSize

		case <-time.After(time.Duration(r.recordsBufferFlushInterval) * time.Millisecond):
			// 在指定的时间过后，强制将缓存中的日志保存在 Redis 中
			readyToSend = true
		}

		// 将日志保存在 redis 中，并重置 recordsBuffer
		if len(recordsBuffer) > 0 && (readyToSend || time.Since(lastSentTS) >= recordsBufferForcedFlushInterval) {
			r.store.AppendToSetPipelined(analyticsKeyName, recordsBuffer)

			// 投递完成后重置 recordsBuffer 和计时器，否则会重复投递数据
			recordsBuffer = recordsBuffer[:0]
			lastSentTS = time.Now() // 重置处理时间
		}
	}
}

// DurationToMillisecond convert time duration type to float64.
func DurationToMillisecond(d time.Duration) float64 {
	return float64(d) / 1e6
}
