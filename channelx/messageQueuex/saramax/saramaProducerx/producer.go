package saramaProducerx

import (
	"context"
	"encoding/json"
	"fmt"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/messageQueuex"
	"github.com/IBM/sarama"
)

type SaramaProducerStr[ProducerTyp any] struct {
	SyncProducer  sarama.SyncProducer
	AsyncProducer sarama.AsyncProducer
	Config        *sarama.Config
	ProducerTyp   uint // 0-同步，1-异步
	// 保存 cancelFunc
	cancelFunc context.CancelFunc
}

// NewSaramaProducerStr 创建一个SaramaProducerStr实现【Sync单挑消息，Async异步才支持批量】
//   - 同步发送消息，就注入同步发送消息的配置监听success通道配置为true
//   - 异步批量发送消息，需注入异步批量发送消息的配置Producer.Flush.Messages Producer.Flush.Frequency以及监听success和error通道配置为true
//   - 如果项目中既有同步发送消息，也有异步发送消息，那么在wire构造注入时，单独wire.NewSet同步实现和wire.NewSet异步实现
func NewSaramaProducerStr[ProducerTyp any, /*ProducerTyp: sarama.SyncProducer & sarama.AsyncProducer*/
](Producer ProducerTyp /*sarama.SyncProducer & sarama.AsyncProducer*/, config *sarama.Config) messageQueuex.ProducerIn[ProducerTyp] {

	h := &SaramaProducerStr[ProducerTyp]{}
	h.Config = config

	switch producer := any(Producer).(type) {
	case sarama.SyncProducer:
		h.ProducerTyp = 0
		h.SyncProducer = producer
		if h.Config.Producer.Return.Successes != true {
			h.Config.Producer.Return.Successes = true
		}
	case sarama.AsyncProducer:
		h.ProducerTyp = 1
		if h.Config.Producer.Return.Successes != true || h.Config.Producer.Return.Errors != true {
			h.Config.Producer.Return.Successes = true
			h.Config.Producer.Return.Errors = true
		}
		h.AsyncProducer = producer
		// 是否为异步批量发送消息，需持续监听success和error通道
		if h.Config.Producer.Flush.Messages > 0 || h.Config.Producer.Flush.Frequency > 0 {
			ctx, cancel := context.WithCancel(context.Background())
			// 将 cancelFunc 赋值给结构体
			h.cancelFunc = cancel
			go h.handleAsyncResults(ctx, h.cancelFunc)
		}
	default:
		panic("kafka Producer Invalid producer type, kafka Producer无效的卡夫卡的生产者类型【非sarama.SyncProducer/sarama.AsyncProducer】")
	}
	return h
}

//func NewSaramaProducerStrV1[ProducerTyp any, /*ProducerTyp: sarama.SyncProducer sarama.AsyncProducer*/
//	Val any](Producer ProducerTyp /*sarama.SyncProducer sarama.AsyncProducer*/, config *sarama.Config) *SaramaProducerStr[ProducerTyp, Val] {
//
//	h := &SaramaProducerStr[ProducerTyp, Val]{}
//	h.Config = config
//
//	switch producer := any(Producer).(type) {
//	case sarama.SyncProducer:
//		h.ProducerTyp = 0
//		h.SyncProducer = producer
//		if h.Config.Producer.Return.Successes != true {
//			h.Config.Producer.Return.Successes = true
//		}
//	case sarama.AsyncProducer:
//		h.ProducerTyp = 1
//		if h.Config.Producer.Return.Successes != true || h.Config.Producer.Return.Errors != true {
//			h.Config.Producer.Return.Successes = true
//			h.Config.Producer.Return.Errors = true
//		}
//		h.AsyncProducer = producer
//		// 是否为异步批量发送消息，需持续监听success和error通道
//		if h.Config.Producer.Flush.Messages > 0 || h.Config.Producer.Flush.Frequency > 0 {
//			ctx, cancel := context.WithCancel(context.Background())
//			// 将 cancelFunc 赋值给结构体
//			h.cancelFunc = cancel
//			go h.handleAsyncResults(ctx, h.cancelFunc)
//		}
//	default:
//		panic("kafka Producer Invalid producer type, kafka Producer无效的卡夫卡的生产者类型【非sarama.SyncProducer/sarama.AsyncProducer】")
//	}
//	return h
//}

// SendMessage Producer生产者发送消息
func (s *SaramaProducerStr[ProducerTyp]) SendMessage(ctx context.Context, keyOrTopic messageQueuex.Tp, value []byte) error {
	switch s.ProducerTyp {
	case 0:
		return s.sendMessageSync(ctx, keyOrTopic, value)
	case 1:
		return s.sendMessageAsync(ctx, keyOrTopic, value)
	default:
		return fmt.Errorf("kafka Producer Invalid producer type, kafka Producer无效的卡夫卡的生产者类型【使用的非sarama.SyncProducer/sarama.AsyncProducer】")
	}
}

// CloseProducer 关闭生产者
//   - 请在main函数最顶层defer住生产者的Producer.Close()，优雅关闭防止goroutine泄露
func (s *SaramaProducerStr[ProducerTyp]) CloseProducer() error {
	var err error
	switch s.ProducerTyp {
	case 1:
		// 取消 context，通知后台 goroutine 退出
		if s.cancelFunc != nil {
			s.cancelFunc()
		}
		// 关闭 AsyncProducer，这会自动关闭 Successes() 和 Errors() 通道
		if s.AsyncProducer != nil {
			err = s.AsyncProducer.Close()
		}
	case 0:
		//  关闭 SyncProducer（如果存在）
		if s.SyncProducer != nil {
			if closeErr := s.SyncProducer.Close(); closeErr != nil {
				err = closeErr
			}
		}
	default:
		err = fmt.Errorf("close kafka Producer Invalid producer type, 关闭kafka Producer时无效的卡夫卡的生产者类型【使用的非sarama.SyncProducer/sarama.AsyncProducer】")
	}
	return err
}

// ========================内部逻辑=======================

// sendMessageSync 同步发送消息
func (s *SaramaProducerStr[ProducerTyp]) sendMessageSync(ctx context.Context, keyOrTopic messageQueuex.Tp, value []byte) error {
	//if s.ProducerTyp != 0 {
	//	return fmt.Errorf("kafka Producer Invalid producer type, kafka Producer无效的卡夫卡的生产者类型【使用的非sarama.SyncProducer】")
	//}
	v, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, _, err = s.SyncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: keyOrTopic.Topic,
		Key:   sarama.StringEncoder(keyOrTopic.Key),
		Value: sarama.StringEncoder(v),
	})
	return err
}

// sendMessageAsync 异步发送消息
//   - 异步批量发送，必须配置sarama.NewConfig的批量发送配置
//   - cfg.Producer.Flush.Frequency = 5 * time.Second // 5秒刷新一次，不管有没有达到批量发送数量条件【只有提交了才会写入broker，success/errors通道才会有消息】
//   - cfg.Producer.Flush.Messages = 5	// 触发刷新所需的最大消息数,5条消息刷新批量发送一次
func (s *SaramaProducerStr[ProducerTyp]) sendMessageAsync(ctx context.Context, keyOrTopic messageQueuex.Tp, value []byte) error {
	if s.ProducerTyp != 1 {
		return fmt.Errorf("kafka Producer Invalid producer type, kafka Producer无效的卡夫卡的生产者类型【使用的非sarama.AsyncProducer】")
	}
	// 异步批量发送消息，需走new方法构造的后台for持续监听success和error通道
	if s.Config.Producer.Flush.Messages > 0 || s.Config.Producer.Flush.Frequency > 0 {
		msg := &sarama.ProducerMessage{
			Topic: keyOrTopic.Topic,
			Key:   sarama.StringEncoder(keyOrTopic.Key),
			Value: sarama.StringEncoder(value),
		}
		select {
		case s.AsyncProducer.Input() <- msg:
			return nil // 入队成功，立即返回
		case <-ctx.Done():
			return ctx.Err() // 超时
		}
	} else {
		// 异步发送单个消息
		s.AsyncProducer.Input() <- &sarama.ProducerMessage{
			Topic: keyOrTopic.Topic,
			Key:   sarama.StringEncoder(keyOrTopic.Key),
			Value: sarama.StringEncoder(value),
		}
		select {
		case <-s.AsyncProducer.Successes():
			return nil // 入队成功
		case er := <-s.AsyncProducer.Errors():
			return fmt.Errorf("send kafka error【监听kafka的errors通道出现错误】%v", er)
		case <-ctx.Done():
			return ctx.Err() // 超时
		}
	}
}

// handleAsyncResults 监听异步生产结果
func (s *SaramaProducerStr[ProducerTyp]) handleAsyncResults(ctx context.Context, cancel context.CancelFunc) {
	defer func() {
		<-s.AsyncProducer.Successes() // 等待所有Successes处理完
		<-s.AsyncProducer.Errors()    // 等待所有Errors处理完
		cancel()
	}()
	for {
		select {
		case success, ok := <-s.AsyncProducer.Successes():
			//log.Printf("✅ 消息发送成功: Topic=%s, Partition=%d, Offset=%d",
			//	success.Topic, success.Partition, success.Offset)
			if !ok {
				return
			}
			// 处理成功消息
			_ = success
		case err, ok := <-s.AsyncProducer.Errors():
			//log.Printf("❌ 消息发送失败: %v", err)
			if !ok {
				return
			}
			// 处理错误
			_ = err
		case <-ctx.Done(): // 退出
			return
		}
	}
}
