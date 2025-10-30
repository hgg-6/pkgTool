package mqX

// messageQueuex/mq.go

import "context"

// Message 通用消息结构，如果需要在处理程序调用之后保留它们，【请复制】。
//   - represents a Kafka message.
//   - Note: Key and Value may share memory with internal buffers.
//   - Make a copy if you need to retain them beyond the handler call.
//   - 表示Kafka消息。
//   - 注意：Key和Value可以与内部缓冲区共享内存。
//   - 如果需要在处理程序调用之后保留它们，【请复制】。
//   - 例如，可能与主 goroutine 冲突（如果主 goroutine 复用 buffer）
//   - go func() {
//     log.Println(string(msg.Value))
//     }()
//   - 更危险：修改 msg
//
// msg.Value[0] = 'X' // 破坏原始数据，且可能影响底层 buffer（见下文）
type Message struct {
	Topic string
	Key   []byte // read-only in handlers; copy if retained, 在处理程序中只读;如果保留则复制
	Value []byte // read-only in handlers; copy if retained, 在处理程序中只读;如果保留则复制
}

// Producer 生产者抽象接口
//
//go:generate mockgen -source=./types.go -package=Producermocks -destination=mocks/Producermocks/mqX.Producermock.go mqX
type Producer interface {
	Send(ctx context.Context, msg *Message) error
	SendBatch(ctx context.Context, msgs []*Message) error
	Close() error
}

// ConsumerHandlerType 消费者处理接口类型
//   - IsBatch() bool: 是否批量处理, true: 批量处理需实现BatchConsumerHandler, false: 单条处理需实现ConsumerHandler
//
//go:generate mockgen -source=./types.go -package=ConsumerHandlerTypemocks -destination=mocks/ConsumerHandlerType/mqX.ConsumerHandlerTypemock.go mqX
type ConsumerHandlerType interface {
	IsBatch() bool // 是否批量处理, true: 批量处理需实现BatchConsumerHandler, false: 单条处理需实现ConsumerHandler
	ConsumerHandler
	BatchConsumerHandler
}

// ConsumerHandler 单条消费者处理接口
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
	Subscribe(ctx context.Context, topics []string, handler ConsumerHandlerType) error
}

type UserEventTest struct {
	UserId int64
	Name   string
}
