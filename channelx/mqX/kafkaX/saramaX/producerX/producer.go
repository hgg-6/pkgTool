package producerX

import (
	"context"
	"fmt"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/mqX"
	"github.com/IBM/sarama"
	"sync"
	"time"
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
		saramaMsgs := make([]*sarama.ProducerMessage, len(msgs))
		for i, m := range msgs {
			saramaMsgs[i] = &sarama.ProducerMessage{
				Topic: m.Topic,
				Key:   sarama.ByteEncoder(m.Key),
				Value: sarama.ByteEncoder(m.Value),
			}
		}
		_, _, err := kp.syncProducer.SendMessage(saramaMsgs[0]) // 注意：SyncProducer 只支持单条
		// 若需批量同步，需循环发送（Kafka 本身不支持 sync batch send）
		// 这里为简化，只发第一条（实际应循环）
		// 更合理做法：SendBatch 在 sync 模式下循环调用 Send
		if len(msgs) > 1 {
			// 实际建议：sync 模式下不推荐用 SendBatch，或内部循环
			for _, m := range msgs {
				_, _, err2 := kp.syncProducer.SendMessage(&sarama.ProducerMessage{
					Topic: m.Topic,
					Key:   sarama.ByteEncoder(m.Key),
					Value: sarama.ByteEncoder(m.Value),
				})
				if err2 != nil {
					err = err2 // 返回最后一个错误
				}
			}
		}
		return err
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

	// 首条消息：启动 timer
	if len(kp.msgBuffer) == 1 && kp.config.BatchTimeout > 0 {
		kp.timer = time.NewTimer(kp.config.BatchTimeout)
		kp.timerC = kp.timer.C
	}

	// 触发条件1：达到批大小
	if len(kp.msgBuffer) >= kp.config.BatchSize {
		return kp.flushLocked()
	}

	// 触发条件2：超时（非阻塞检查）
	select {
	case <-kp.timerC:
		return kp.flushLocked()
	default:
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

// handleAsyncResults 处理异步结果（可选监控）
func (kp *KafkaProducer) handleAsyncResults() {
	defer kp.wg.Done()
	for {
		select {
		case <-kp.asyncProducer.Successes():
			// 可记录成功（如 metrics）
		case err := <-kp.asyncProducer.Errors():
			// 可记录错误（如日志、告警）
			_ = err
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
