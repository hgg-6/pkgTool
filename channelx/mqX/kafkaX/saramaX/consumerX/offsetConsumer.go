package consumerX

import (
	"context"
	"fmt"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/mqX"
	"github.com/IBM/sarama"
	"sync"
	"time"
)

// OffsetConsumerConfig 配置
type OffsetConsumerConfig struct {
	BatchSize    int
	BatchTimeout time.Duration
	// 是否自动提交 offset（默认 false，由 handler 决定是否提交）
	AutoCommit bool
}

func DefaultOffsetConsumerConfig() *OffsetConsumerConfig {
	return &OffsetConsumerConfig{
		BatchSize:    100,
		BatchTimeout: 5 * time.Second,
		AutoCommit:   false,
	}
}

type OffsetConsumer struct {
	brokers []string
	config  *OffsetConsumerConfig

	client   sarama.Client
	consumer sarama.Consumer

	mu     sync.Mutex
	closed bool
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

// NewOffsetConsumer 创建 offset 消费者
func NewOffsetConsumer(brokers []string, config *OffsetConsumerConfig) (*OffsetConsumer, error) {
	if config == nil {
		config = DefaultOffsetConsumerConfig()
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 100
	}
	if config.BatchTimeout <= 0 {
		config.BatchTimeout = 5 * time.Second
	}

	saramaCfg := sarama.NewConfig()
	saramaCfg.Version = sarama.V2_8_0_0

	client, err := sarama.NewClient(brokers, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("create consumer: %w", err)
	}

	return &OffsetConsumer{
		brokers:  brokers,
		config:   config,
		client:   client,
		consumer: consumer,
	}, nil
}

// ConsumeFrom 从指定 topic/partition/offset 开始消费
func (oc *OffsetConsumer) ConsumeFrom(ctx context.Context, topic string,
	partition int32, startOffset int64,
	handler mqX.BatchConsumerHandler) error {
	oc.mu.Lock()
	if oc.closed {
		oc.mu.Unlock()
		return fmt.Errorf("consumer closed")
	}
	oc.mu.Unlock()

	// 创建带取消的上下文
	ctx, oc.cancel = context.WithCancel(ctx)

	// 创建 partition consumer
	pc, err := oc.consumer.ConsumePartition(topic, partition, startOffset)
	if err != nil {
		return fmt.Errorf("consume partition: %w", err)
	}
	defer pc.Close()

	oc.wg.Add(1)
	defer oc.wg.Done()

	var (
		msgBuffer = make([]*mqX.Message, 0, oc.config.BatchSize)
		timer     *time.Timer
		timerC    <-chan time.Time
	)

	flushBatch := func() error {
		if len(msgBuffer) == 0 {
			return nil
		}

		commit, err := handler.HandleBatch(ctx, msgBuffer)
		if err != nil {
			return fmt.Errorf("handle batch: %w", err)
		}

		// 手动提交 offset（如果 handler 要求）
		if commit || oc.config.AutoCommit {
			lastMsg := msgBuffer[len(msgBuffer)-1]
			// 注意：Sarama 的 PartitionConsumer 不支持自动提交，
			// 但你可以记录 offset 到外部存储（如 DB）
			// 这里仅打印，实际应持久化
			_ = lastMsg // 避免 unused
			// fmt.Printf("Would commit offset: %d for partition %d\n", lastMsg.Offset, partition)
		}

		msgBuffer = msgBuffer[:0] // reset

		if timer != nil {
			timer.Stop()
			timer = nil
			timerC = nil
		}
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			_ = flushBatch()
			return ctx.Err()

		case msg, ok := <-pc.Messages():
			if !ok {
				// channel closed
				_ = flushBatch()
				return nil
			}

			// 转换为通用 Message
			kafkaMsg := &mqX.Message{
				Topic: msg.Topic,
				Key:   msg.Key,
				Value: msg.Value,
			}

			msgBuffer = append(msgBuffer, kafkaMsg)

			// 首条消息：启动 timer
			if len(msgBuffer) == 1 && oc.config.BatchTimeout > 0 {
				timer = time.NewTimer(oc.config.BatchTimeout)
				timerC = timer.C
			}

			// 触发条件1：达到批大小
			if len(msgBuffer) >= oc.config.BatchSize {
				if err := flushBatch(); err != nil {
					return err
				}
			}

		case <-timerC:
			// 触发条件2：超时
			if err := flushBatch(); err != nil {
				return err
			}
		}
	}
}

// Close 关闭消费者
func (oc *OffsetConsumer) Close() error {
	oc.mu.Lock()
	if oc.closed {
		oc.mu.Unlock()
		return nil
	}
	oc.closed = true
	if oc.cancel != nil {
		oc.cancel()
	}
	oc.mu.Unlock()

	oc.wg.Wait()

	var errs []error
	if err := oc.consumer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close consumer: %w", err))
	}
	if err := oc.client.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close client: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors on close: %v", errs)
	}
	return nil
}
