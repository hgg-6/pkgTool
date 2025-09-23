// Package messageQueuex 消息队列抽象接口
package messageQueuex

import "context"

type KeyOrTopic struct {
	Key   []byte
	Topic string
	// .......
}

// ProducerIn 生产者抽象接口
//   - 当使用 sarama.NewSyncProducer() 创建生产者时，请使用 NewSaramaProducerStr()
//   - 请在main函数最顶层defer住生产者的Producer.Close()，优雅关闭防止goroutine泄露
type ProducerIn[ProducerTyp any] interface {
	SendMessage(ctx context.Context, keyOrTopic KeyOrTopic, value []byte) error
	// CloseProducer 关闭生产者Producer，请在main函数最顶层defer住生产者的Producer.Close()，优雅关闭防止goroutine泄露
	CloseProducer() error
}

// ConsumerIn 消费者抽象接口
type ConsumerIn interface {
	ReceiveMessage(ctx context.Context, keyOrTopic []KeyOrTopic) error
}
