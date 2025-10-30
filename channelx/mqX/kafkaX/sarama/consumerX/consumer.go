package consumerX

// saramaConsumerx/kafka_consumer.go

import (
	"context"
	"fmt"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/mqX"
	"github.com/IBM/sarama"
	"time"
)

type KafkaConsumer struct {
	consumerGroup   sarama.ConsumerGroup
	config          *ConsumerConfig
	consumerHandler *consumerGroupHandlerAdapter
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

func (kc *KafkaConsumer) Subscribe(ctx context.Context, topics []string, handler mqX.ConsumerHandlerType) error {
	kc.consumerHandler = newConsumerGroupHandlerAdapter(handler, kc.config)
	//kc.consumerHandler = &consumerGroupHandlerAdapter{handler: handler}
	return kc.consumerGroup.Consume(ctx, topics, kc.consumerHandler)
}

// ==================
// ==================
// ==================

type consumerGroupHandlerAdapter struct {
	handler mqX.ConsumerHandlerType
	config  *ConsumerConfig
}

func newConsumerGroupHandlerAdapter(handler mqX.ConsumerHandlerType, config *ConsumerConfig) *consumerGroupHandlerAdapter {
	if config == nil {
		config = DefaultConsumerConfig()
	}
	config.Validate()
	return &consumerGroupHandlerAdapter{
		handler: handler,
		config:  config,
	}
}

func (a *consumerGroupHandlerAdapter) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (a *consumerGroupHandlerAdapter) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (a *consumerGroupHandlerAdapter) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	//batchHandler, isBatch := a.handler.(mqX.BatchConsumerHandler)

	//if !isBatch {
	var errs error
	switch a.handler.IsBatch() {
	case false:
		// 非批量模式：走单条逻辑
		for msg := range claim.Messages() {
			genericMsg := &mqX.Message{
				Topic: msg.Topic,
				Key:   msg.Key,
				Value: msg.Value,
			}
			if err := a.handler.Handle(context.Background(), genericMsg); err != nil {
				errs = err
				return err
			}
			sess.MarkMessage(msg, "")
		}
		return errs
	case true:
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

			success, err := a.handler.HandleBatch(context.Background(), msgBuffer)
			if err != nil {
				errs = err
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
		return errs
	}
	return fmt.Errorf("unknown batch mode, 未知的批量模式/单条模式， 未实现IsBatch()接口")
}
