package serviceLogic

import (
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"github.com/IBM/sarama"
)

/*
	====================================
	普通的利用context超时控制的消费者处理逻辑
	====================================
*/

// SvcLogicFn 消费消息后的，自身业务逻辑处理eg：入库、缓存...
type SvcLogicFn[EvenT any] func(msg *sarama.ConsumerMessage, event EvenT) error

// SvcLogicFns 批量消费后的，自身业务逻辑处理eg：入库、缓存...
type SvcLogicFns[EvenT any] func(msgs []*sarama.ConsumerMessage, event []EvenT) error

// SaramaConsumerGroupMessage 消费者组消息处理【消费消息后的，自身业务逻辑处理eg：入库、缓存...】
type SaramaConsumerGroupMessage[EvenT any] struct {
	// 日志注入
	L logx.Loggerx
	// 消费消息后的，自身业务逻辑处理eg：入库、缓存...
	SvcLogicFn[EvenT]
	// 【批量】消费后的，自身业务逻辑处理eg：入库、缓存...
	SvcLogicFns[EvenT]

	// Offset配置指定偏移量消费【消费历史消息、或者从某个消息开始消费】
	IsOffset    bool
	OffsetTopic string
	Offset      int64

	// 批量消费配置
	IsBatch   bool
	BatchSize int64
}

// NewSaramaConsumerGroupMessage 创建消费者组消息处理【消费消息后的，自身业务逻辑处理eg：入库、缓存...】
//   - fn 消费消息后的，自身业务逻辑处理eg：入库、缓存...
//   - SvcLogicFns 批量消费后的，自身业务逻辑处理eg：入库、缓存...
//   - 使用fn的话，fns传nil即可，同理反之
func NewSaramaConsumerGroupMessage[EvenT any](l logx.Loggerx, fn SvcLogicFn[EvenT], fns SvcLogicFns[EvenT]) *SaramaConsumerGroupMessage[EvenT] {
	return &SaramaConsumerGroupMessage[EvenT]{
		L:           l,
		SvcLogicFn:  fn,
		SvcLogicFns: fns,
		IsOffset:    false,
		Offset:      0,
		IsBatch:     false,
		BatchSize:   0,
	}
}

// SetOffset 设置偏移量消费【消费历史消息、或者从某个消息开始消费】
func (s *SaramaConsumerGroupMessage[EvenT]) SetOffset(IsOffset bool, OffsetTopic string, Offset int64) {
	s.IsOffset = IsOffset
	s.OffsetTopic = OffsetTopic
	s.Offset = Offset
}

// SetBatch 批量消费配置
func (s *SaramaConsumerGroupMessage[EvenT]) SetBatch(IsBatch bool, BatchSize int64) {
	s.IsBatch = IsBatch
	s.BatchSize = BatchSize
}
