package ConsumerGroupHandlerx

import (
	"context"
	"encoding/json"
	"fmt"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/messageQueuex/saramax/saramaConsumerx/serviceLogic"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"github.com/IBM/sarama"
	"time"
)

/*
	====================================
	均在serviceLogic.NewSaramaConsumerGroupMessage中new配置
	1、普通的利用context超时控制的消费者处理逻辑
	2、指定偏移量处理逻辑
	3、批量消费消息处理逻辑
	====================================
*/

type ConsumerGroupHandler[EvenT any] struct {
	serviceLogic *serviceLogic.SaramaConsumerGroupMessage[EvenT]
}

func NewConsumerGroupHandler[EvenT any](serviceLogic *serviceLogic.SaramaConsumerGroupMessage[EvenT]) sarama.ConsumerGroupHandler {
	return &ConsumerGroupHandler[EvenT]{
		serviceLogic: serviceLogic,
	}
}

func (c *ConsumerGroupHandler[EvenT]) Setup(session sarama.ConsumerGroupSession) error {
	c.serviceLogic.L.Info("消费消息: Setup")
	var offset int64

	// 未配置偏移量
	if !c.serviceLogic.IsOffset {
		//offset = sarama.OffsetNewest // 【等同-1】默认从最新的消息开始消费
		//session.ResetOffset(topic, v, offset, "")
		return nil
	} else {
		c.serviceLogic.L.Info("消费消息: Setup: 配置了偏移量开始读取")
		// 遍历所有分配的主题和分区
		for topic, partitions := range session.Claims() {
			switch c.serviceLogic.OffsetTopic == topic {
			case true:
				for _, partition := range partitions {
					// 配置了偏移量，按照配置的指定分区的偏移量都设置为指定的偏移量
					offset = c.serviceLogic.Offset
					session.ResetOffset(topic, partition, offset, "")
					c.serviceLogic.IsOffset = false
					return nil
				}
			case false:
				c.serviceLogic.L.Error("消费者配置的偏移量开始读取， 配置的偏移量和指定topic在kafka中没有找到对应的topic, topic set not found", logx.String("set topic", topic))
				c.serviceLogic.IsOffset = false
				return fmt.Errorf("消费者配置的偏移量开始读取， 配置的偏移量和指定topic在kafka中没有找到对应的topic: %s, topic set not found", c.serviceLogic.OffsetTopic)
			}
		}
	}
	return nil
}

func (c *ConsumerGroupHandler[EvenT]) Cleanup(session sarama.ConsumerGroupSession) error {
	c.serviceLogic.L.Info("消费消息: Cleanup")
	return nil
}

func (c *ConsumerGroupHandler[EvenT]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages() // 获取消息通道
	var err error
	// 单条消息处理逻辑
	if !c.serviceLogic.IsBatch {
		err = c.oneMessageLogic(msgs, session)
	} else {
		// 批量消息处理逻辑
		err = c.manyMessagesLogic(msgs, session)
	}
	return err
}

func (c *ConsumerGroupHandler[EvenT]) oneMessageLogic(msgs <-chan *sarama.ConsumerMessage, session sarama.ConsumerGroupSession) error {
	if c.serviceLogic.SvcLogicFn == nil {
		c.serviceLogic.L.Error("没有设置消费消息的业务逻辑函数 no SaramaConsumerGroupMessage SvcLogicFn")
		return fmt.Errorf("没有设置消费消息的业务逻辑函数 no SaramaConsumerGroupMessage SvcLogicFn")
	}
	// 单条消息处理逻辑
	c.serviceLogic.L.Info("触发消费消息logic: oneMessageLogic")
	for msg := range msgs {
		// 处理消息
		var t EvenT
		err := json.Unmarshal(msg.Value, &t) // 反序列化消息
		if err != nil {
			// 也可以也引入重试的逻辑
			c.serviceLogic.L.Error("json.Unmarshal fail【反序列化消息失败】", logx.String("topic", msg.Topic),
				logx.Int32("Partition", msg.Partition), logx.Int64("offset", msg.Offset), logx.Error(err))
			//continue // 跳过这条消息
			return err
		}
		err = c.serviceLogic.SvcLogicFn(msg, t) // 处理消息
		if err != nil {
			// 也可以也引入重试的逻辑
			c.serviceLogic.L.Error("处理消息失败", logx.String("topic", msg.Topic),
				logx.Int32("Partition", msg.Partition), logx.Int64("offset", msg.Offset), logx.Error(err))
			//continue // 跳过这条消息
			return err
		}
		session.MarkMessage(msg, "") // 标记消息为已消费
	}
	return nil
}

func (c *ConsumerGroupHandler[EvenT]) manyMessagesLogic(msgs <-chan *sarama.ConsumerMessage, session sarama.ConsumerGroupSession) error {
	if c.serviceLogic.SvcLogicFns == nil {
		c.serviceLogic.L.Error("没有设置批量消费消息的业务逻辑函数 no SaramaConsumerGroupMessage SvcLogicFns")
		return fmt.Errorf("没有设置批量消费消息的业务逻辑函数 no SaramaConsumerGroupMessage SvcLogicFns")
	}
	c.serviceLogic.L.Info("触发批量消费消息logic: manyMessagesLogic")
	ts := make([]EvenT, 0, c.serviceLogic.BatchSize)
	for {
		batch := make([]*sarama.ConsumerMessage, 0, c.serviceLogic.BatchSize)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		var done bool = false
		for i := 0; i < int(c.serviceLogic.BatchSize) && !done; i++ {
			select {
			case <-ctx.Done():
				// 超时
				done = true
			case msg, ok := <-msgs:
				if !ok {
					cancel()
					return nil
				}
				var t EvenT
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					c.serviceLogic.L.Error("json.Unmarshal fail【反序列化消息失败】", logx.String("topic", msg.Topic),
						logx.Int32("Partition", msg.Partition), logx.Int64("offset", msg.Offset), logx.Error(err))
					continue // 跳过这条消息
				}
				batch = append(batch, msg)
				ts = append(ts, t)
			}
		}
		// 关闭上下文
		cancel()
		// 凑够一批消息了
		err := c.serviceLogic.SvcLogicFns(batch, ts)
		if err != nil {
			// 可以整个msgs都记录下来
			c.serviceLogic.L.Error("kafka批量消费消息时，业务逻辑处理失败", logx.Any("*sarama.ConsumerMessage_batch", batch), logx.Error(err))
			//continue // 跳过这条消息
		}
		for _, msg := range batch {
			session.MarkMessage(msg, "") // 标记消息为已消费
		}
	}
}
