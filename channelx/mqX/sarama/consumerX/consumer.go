package consumerX

// saramaConsumerx/kafka_consumer.go

import (
	"context"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/mqX"
	"github.com/IBM/sarama"
	"time"
)

type KafkaConsumer struct {
	consumerGroup sarama.ConsumerGroup
	config        *ConsumerConfig
}

// NewKafkaConsumer 创建 Kafka 消费者
// config 可为 nil，将使用默认值
func NewKafkaConsumer(cg sarama.ConsumerGroup, config *ConsumerConfig) *KafkaConsumer {
	if config == nil {
		config = DefaultConsumerConfig()
	}
	config.Validate() // 自动修正无效值
	return &KafkaConsumer{
		consumerGroup: cg,
		config:        config,
	}
}

func (kc *KafkaConsumer) Subscribe(ctx context.Context, topics []string, handler mqX.ConsumerHandler) error {
	adapter := &ConsumerGroupHandlerAdapter{handler: handler}
	return kc.consumerGroup.Consume(ctx, topics, adapter)
}

// ==================
// ==================
// ==================

type ConsumerGroupHandlerAdapter struct {
	handler mqX.ConsumerHandler
	config  *ConsumerConfig
}

func NewConsumerGroupHandlerAdapter(handler mqX.ConsumerHandler, config *ConsumerConfig) *ConsumerGroupHandlerAdapter {
	if config == nil {
		config = DefaultConsumerConfig()
	}
	config.Validate()
	return &ConsumerGroupHandlerAdapter{
		handler: handler,
		config:  config,
	}
}

func (a *ConsumerGroupHandlerAdapter) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (a *ConsumerGroupHandlerAdapter) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (a *ConsumerGroupHandlerAdapter) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	batchHandler, isBatch := a.handler.(mqX.BatchConsumerHandler)

	// 非批量模式：走单条逻辑
	if !isBatch {
		for msg := range claim.Messages() {
			genericMsg := &mqX.Message{
				Topic: msg.Topic,
				Key:   msg.Key,
				Value: msg.Value,
			}
			if err := a.handler.Handle(context.Background(), genericMsg); err != nil {
				return err
			}
			sess.MarkMessage(msg, "")
		}
		return nil
	}

	// === 批量模式（带超时）===
	batchSize := a.config.BatchSize
	batchTimeout := a.config.BatchTimeout

	msgBuffer := make([]*mqX.Message, 0, batchSize)
	saramaMsgBuffer := make([]*sarama.ConsumerMessage, 0, batchSize)

	var timer *time.Timer
	var timerC <-chan time.Time

	// flush 批量缓冲区
	flushBatch := func() error {
		if len(msgBuffer) == 0 {
			return nil
		}

		success, err := batchHandler.HandleBatch(context.Background(), msgBuffer)
		if err != nil {
			return err
		}
		if success {
			sess.MarkMessage(saramaMsgBuffer[len(saramaMsgBuffer)-1], "")
		}

		// 重置
		msgBuffer = msgBuffer[:0]
		saramaMsgBuffer = saramaMsgBuffer[:0]

		// 停止并重置 timer
		if timer != nil {
			timer.Stop()
			timer = nil
			timerC = nil
		}
		return nil
	}

	defer func() {
		// 退出前 flush 剩余消息
		_ = flushBatch()
		if timer != nil {
			timer.Stop()
		}
	}()

	for msg := range claim.Messages() {
		genericMsg := &mqX.Message{
			Topic: msg.Topic,
			Key:   msg.Key,
			Value: msg.Value,
		}

		// 首条消息：启动 timer
		if len(msgBuffer) == 0 && batchTimeout > 0 {
			timer = time.NewTimer(batchTimeout)
			timerC = timer.C
		}

		msgBuffer = append(msgBuffer, genericMsg)
		saramaMsgBuffer = append(saramaMsgBuffer, msg)

		// 触发条件1：达到批大小
		if len(msgBuffer) >= batchSize {
			if err := flushBatch(); err != nil {
				return err
			}
			continue
		}

		// 触发条件2：超时（非阻塞检查）
		select {
		case <-timerC:
			// 超时触发 flush
			if err := flushBatch(); err != nil {
				return err
			}
		default:
			// 未超时，继续收消息
		}
	}

	return nil
}
