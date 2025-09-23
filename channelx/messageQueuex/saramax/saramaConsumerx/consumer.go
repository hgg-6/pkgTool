package saramaConsumerx

import (
	"context"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/messageQueuex"
	"github.com/IBM/sarama"
)

type ConsumerGroup struct {
	// ConsumerGroup sararma的消费者组接口逻辑
	ConsumerGroup sarama.ConsumerGroup
	// ConsumerGroupHandlers 消费者组处理逻辑
	ConsumerGroupHandlers sarama.ConsumerGroupHandler
}

// NewConsumerIn
//   - consumerGroup New一个sarama.NewConsumerGroup
//   - ConsumerGroupHandlers 可自信封装消费者组处理逻辑【也可使用hgg的ConsumerGroupHandlerx包下默认封装的逻辑】
func NewConsumerIn(consumerGroup sarama.ConsumerGroup, ConsumerGroupHandlers sarama.ConsumerGroupHandler) messageQueuex.ConsumerIn {
	c := &ConsumerGroup{
		ConsumerGroup:         consumerGroup,
		ConsumerGroupHandlers: ConsumerGroupHandlers,
	}
	return c
}

func (c *ConsumerGroup) ReceiveMessage(ctx context.Context, keyOrTopic []messageQueuex.KeyOrTopic) error {
	var topic []string
	for _, v := range keyOrTopic {
		topic = append(topic, v.Topic)
	}
	return c.ConsumerGroup.Consume(ctx, topic, c.ConsumerGroupHandlers)
}
