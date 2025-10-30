// Package messageQueuex 消息队列抽象接口
package messageQueuex

import "context"

// Deprecated: messageQueuex此包弃用，此方法将在未来版本中删除，请使用mqX包
type KeyOrTopic struct {
	Key   []byte
	Topic string
	// .......
}

// Deprecated: messageQueuex此包弃用，此方法将在未来版本中删除，请使用mqX包
type Tp KeyOrTopic

// Deprecated: messageQueuex此包弃用，此方法将在未来版本中删除，请使用mqX包
// ProducerIn 生产者抽象接口
//   - 当使用 sarama.NewSyncProducer() 创建生产者时，请使用 NewSaramaProducerStr()
//   - 请在main函数最顶层defer住生产者的Producer.Close()，优雅关闭防止goroutine泄露
type ProducerIn[ProducerTyp any] interface {
	SendMessage(ctx context.Context, keyOrTopic Tp, value []byte) error
	// CloseProducer 关闭生产者Producer，请在main函数最顶层defer住生产者的Producer.Close()，优雅关闭防止goroutine泄露
	CloseProducer() error
}

// Deprecated: messageQueuex此包弃用，此方法将在未来版本中删除，请使用mqX包
// ConsumerIn 消费者抽象接口
type ConsumerIn interface {
	ReceiveMessage(ctx context.Context, keyOrTopic []Tp) error
}
