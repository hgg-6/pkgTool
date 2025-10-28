package mqX

// messageQueuex/mq.go

import "context"

// Message 通用消息结构
type Message struct {
	Topic string
	Key   []byte
	Value []byte
}

// Producer 生产者抽象接口
//
//go:generate mockgen -source=./types.go -package=Producermocks -destination=mocks/Producermocks/mqX.Producermock.go mqX
type Producer interface {
	Send(ctx context.Context, msg *Message) error
	SendBatch(ctx context.Context, msgs []*Message) error
	Close() error
}

// ConsumerHandler 消费者处理接口
//
//go:generate mockgen -source=./types.go -package=ConsumerHandlermocks -destination=mocks/ConsumerHandler/mqX.ConsumerHandlermock.go mqX
type ConsumerHandler interface {
	Handle(ctx context.Context, msg *Message) error
}

// BatchConsumerHandler 批量消费者处理接口
//
//go:generate mockgen -source=./types.go -package=BatchConsumerHandlermocks -destination=mocks/BatchConsumerHandler/mqX.BatchConsumerHandlermock.go mqX
type BatchConsumerHandler interface {
	HandleBatch(ctx context.Context, msgs []*Message) (success bool, err error)
}

// Consumer 消费者抽象接口
//
//go:generate mockgen -source=./types.go -package=Consumermocks -destination=mocks/Consumer/mqX.Consumerrmock.go mqX
type Consumer interface {
	Subscribe(ctx context.Context, topics []string, handler ConsumerHandler) error
}

type UserEventTest struct {
	UserId int64
	Name   string
}
