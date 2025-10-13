package consumerx

import (
	"context"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/gormx/dbMovex/myMovex/events"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/messageQueuex"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/messageQueuex/saramax/saramaConsumerx"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/messageQueuex/saramax/saramaConsumerx/ConsumerGroupHandlerx"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/messageQueuex/saramax/saramaConsumerx/serviceLogic"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"github.com/IBM/sarama"
	"gorm.io/gorm"
)

/*
	==============================
	利用kafka消息队列，削峰、解耦，消费者处理不一致数据
	==============================
*/

type ConsumerConf struct {
	Addr       []string
	GroupId    string
	SaramaConf *sarama.Config
}

type DbConf struct {
	SrcDb *gorm.DB
	DstDb *gorm.DB
}

type Consumer[T events.InconsistentEvent] struct {
	ConsumerConfig ConsumerConf
	DbConfig       DbConf
	Fn             func(msg *sarama.ConsumerMessage, event events.InconsistentEvent) error
	L              logx.Loggerx
	ConsumerIn     messageQueuex.ConsumerIn
}

func NewConsumer[T events.InconsistentEvent](consumerConf ConsumerConf, dbConfig DbConf, l logx.Loggerx) *Consumer[T] {
	c := &Consumer[T]{
		ConsumerConfig: consumerConf,
		DbConfig:       dbConfig,
		L:              l,
	}
	c.SetFn(c.fn()) // 业务逻辑函数初始化
	c.ConsumerIn = saramaConsumerx.NewConsumerIn(c.newConsumerGroup(), c.newConsumerGroupHandler())
	return c
}

func (c *Consumer[T]) newConsumerGroup() sarama.ConsumerGroup {
	cg, err := sarama.NewConsumerGroup(c.ConsumerConfig.Addr, c.ConsumerConfig.GroupId, c.ConsumerConfig.SaramaConf)
	if err != nil {
		panic(err)
	}
	return cg
}
func (c *Consumer[T]) newConsumerGroupHandler() sarama.ConsumerGroupHandler {
	consumerMsg := serviceLogic.NewSaramaConsumerGroupMessage[events.InconsistentEvent](c.L, c.Fn, nil)
	return ConsumerGroupHandlerx.NewConsumerGroupHandler[events.InconsistentEvent](consumerMsg)
}

func (c *Consumer[T]) InitConsumer(ctx context.Context, topic string) error {
	return c.ConsumerIn.ReceiveMessage(ctx, []messageQueuex.Tp{{Topic: topic}})
}

func (c *Consumer[T]) SetFn(fn func(msg *sarama.ConsumerMessage, event events.InconsistentEvent) error) {
	c.Fn = fn
}

func (c *Consumer[T]) fn() func(msg *sarama.ConsumerMessage, event events.InconsistentEvent) error {
	return func(msg *sarama.ConsumerMessage, event events.InconsistentEvent) error {
		ov, err := NewOverrideFixer[events.TestUser](c.DbConfig.SrcDb, c.DbConfig.DstDb)
		if err != nil {
			panic(err)
		}
		// 修复数据
		err = ov.Fix(context.Background(), event.ID)
		c.L.Info("receive message success, 消费不一致数据", logx.Int64("value_id: ", event.ID), logx.Any("event: ", event))
		return nil
	}
}
