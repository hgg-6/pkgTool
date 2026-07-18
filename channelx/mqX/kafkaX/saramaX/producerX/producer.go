package producerX

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hgg-6/pkgTool/v2/channelx/mqX"
	"github.com/IBM/sarama"
)

// KafkaProducer saramaProducerx/kafka_producer.go
type KafkaProducer struct {
	config *ProducerConfig

	// 同步生产者（仅当 Async=false 时使用）
	syncProducer sarama.SyncProducer

	// 异步生产者（仅当 Async=true 时使用）
	asyncProducer sarama.AsyncProducer
	msgBuffer     []*mqX.Message
	saramaBuffer  []*sarama.ProducerMessage
	timer         *time.Timer
	timerC        <-chan time.Time
	mu            sync.Mutex
	closed        bool

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewKafkaProducer 创建 Kafka 生产者
func NewKafkaProducer(addrs []string, config *ProducerConfig) (mqX.Producer, error) {
	if config == nil {
		config = DefaultProducerConfig()
	}
	config.Validate()

	saramaCfg := sarama.NewConfig()
	saramaCfg.Version = sarama.V2_8_0_0
	saramaCfg.Producer.RequiredAcks = sarama.WaitForAll
	saramaCfg.Producer.Return.Successes = true
	saramaCfg.Producer.Return.Errors = true

	var kp KafkaProducer
	kp.config = config
	kp.ctx, kp.cancel = context.WithCancel(context.Background())

	if config.Async {
		// === 异步模式：启用批量 + 超时 ===
		producer, err := sarama.NewAsyncProducer(addrs, saramaCfg)
		if err != nil {
			return nil, fmt.Errorf("create async producer: %w", err)
		}
		kp.asyncProducer = producer

		kp.wg.Add(1)
		go kp.handleAsyncResults()

		// P0-20: 后台 flush goroutine。旧实现只在 SendBatch 调用时用非阻塞 select
		// 顺带检查 timer，调用方停止发送后 buffer 消息永远不会被 flush（直到 Close）。
		if config.BatchTimeout > 0 {
			kp.wg.Add(1)
			go kp.batchTimeoutLoop()
		}
	} else {
		// === 同步模式：无缓冲，立即发送 ===
		producer, err := sarama.NewSyncProducer(addrs, saramaCfg)
		if err != nil {
			return nil, fmt.Errorf("create sync producer: %w", err)
		}
		kp.syncProducer = producer
	}

	return &kp, nil
}

// Send 发送单条消息
func (kp *KafkaProducer) Send(ctx context.Context, msg *mqX.Message) error {
	return kp.SendBatch(ctx, []*mqX.Message{msg})
}

// SendBatch 发送批量消息
func (kp *KafkaProducer) SendBatch(ctx context.Context, msgs []*mqX.Message) error {
	if len(msgs) == 0 {
		return nil
	}

	if !kp.config.Async {
		// === 同步模式：立即发送，不缓冲 ===
		var lastErr error
		for _, m := range msgs {
			_, _, err := kp.syncProducer.SendMessage(&sarama.ProducerMessage{
				Topic: m.Topic,
				Key:   sarama.ByteEncoder(m.Key),
				Value: sarama.ByteEncoder(m.Value),
			})
			if err != nil {
				lastErr = err
			}
		}
		return lastErr
	}

	// === 异步模式：走缓冲 + 批量逻辑 ===
	kp.mu.Lock()
	defer kp.mu.Unlock()

	if kp.closed {
		return fmt.Errorf("producer closed")
	}

	// 添加到缓冲区
	for _, m := range msgs {
		kp.msgBuffer = append(kp.msgBuffer, m)
		kp.saramaBuffer = append(kp.saramaBuffer, &sarama.ProducerMessage{
			Topic: m.Topic,
			Key:   sarama.ByteEncoder(m.Key),
			Value: sarama.ByteEncoder(m.Value),
		})
	}

	// 首条消息：启动 timer（由后台 batchTimeoutLoop 监听并 flush）
	if len(kp.msgBuffer) == 1 && kp.config.BatchTimeout > 0 {
		kp.timer = time.NewTimer(kp.config.BatchTimeout)
		kp.timerC = kp.timer.C
	}

	// 达到批大小立即 flush；超时由后台 batchTimeoutLoop 负责（P0-20）。
	if len(kp.msgBuffer) >= kp.config.BatchSize {
		return kp.flushLocked()
	}

	return nil
}

// flushLocked 异步模式下 flush 缓冲区（必须持有锁）
func (kp *KafkaProducer) flushLocked() error {
	if len(kp.saramaBuffer) == 0 {
		return nil
	}

	for _, msg := range kp.saramaBuffer {
		select {
		case kp.asyncProducer.Input() <- msg:
		case <-kp.ctx.Done():
			return fmt.Errorf("producer closed during flush")
		}
	}

	// 重置缓冲区（保留容量）
	kp.msgBuffer = kp.msgBuffer[:0]
	kp.saramaBuffer = kp.saramaBuffer[:0]

	// 重置 timer
	if kp.timer != nil {
		kp.timer.Stop()
		kp.timer = nil
		kp.timerC = nil
	}

	return nil
}

// batchTimeoutLoop 后台监听 batchTimeout timer，到点 flush 缓冲区（P0-20）。
func (kp *KafkaProducer) batchTimeoutLoop() {
	defer kp.wg.Done()
	for {
		kp.mu.Lock()
		timerC := kp.timerC
		kp.mu.Unlock()

		if timerC == nil {
			select {
			case <-kp.ctx.Done():
				return
			case <-time.After(10 * time.Millisecond):
				continue
			}
		}

		select {
		case <-kp.ctx.Done():
			return
		case <-timerC:
			kp.mu.Lock()
			// 再次校验 timerC 仍是触发的那一个（避免已被 flush 重置后误 flush）。
			if kp.timerC != nil {
				_ = kp.flushLocked()
			}
			kp.mu.Unlock()
		}
	}
}

// handleAsyncResults 处理异步结果
func (kp *KafkaProducer) handleAsyncResults() {
	defer kp.wg.Done()
	for {
		select {
		case <-kp.asyncProducer.Successes():
			// 可记录成功（如 metrics）
		case err := <-kp.asyncProducer.Errors():
			// 调用错误处理回调
			if kp.config.OnError != nil {
				kp.config.OnError(err)
			}
		case <-kp.ctx.Done():
			return
		}
	}
}

// Close 关闭生产者
func (kp *KafkaProducer) Close() error {
	kp.mu.Lock()
	if kp.closed {
		kp.mu.Unlock()
		return nil
	}
	kp.closed = true
	kp.cancel()
	kp.mu.Unlock()

	if kp.config.Async {
		// 异步：flush 剩余 + 等待 goroutine
		kp.mu.Lock()
		_ = kp.flushLocked()
		kp.mu.Unlock()
		kp.wg.Wait()
		return kp.asyncProducer.Close()
	} else {
		// 同步：直接关闭
		return kp.syncProducer.Close()
	}
}
