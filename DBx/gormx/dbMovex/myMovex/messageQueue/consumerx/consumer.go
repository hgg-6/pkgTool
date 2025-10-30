package consumerx

import (
	"context"
	"encoding/json"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/gormx/dbMovex/myMovex/events"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/mqX"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/mqX/kafkaX/sarama/consumerX"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"github.com/IBM/sarama"
	"gorm.io/gorm"
	"log"
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
	//Fn             func(msg *sarama.ConsumerMessage, event events.InconsistentEvent) error
	L logx.Loggerx
	//ConsumerIn     messageQueuex.ConsumerIn
	Consumer mqX.Consumer
}

func NewConsumer[T events.InconsistentEvent](consumerConf ConsumerConf, dbConfig DbConf, l logx.Loggerx) *Consumer[T] {
	c := &Consumer[T]{
		ConsumerConfig: consumerConf,
		DbConfig:       dbConfig,
		L:              l,
	}
	//c.SetFn(c.fn()) // 业务逻辑函数初始化
	//c.ConsumerIn = saramaConsumerx.NewConsumerIn(c.newConsumerGroup(), c.newConsumerGroupHandler())
	consumerCg, err := sarama.NewConsumerGroup(c.ConsumerConfig.Addr, c.ConsumerConfig.GroupId, c.ConsumerConfig.SaramaConf)
	if err != nil {
		c.L.Error("new consumer group error", logx.Error(err))
	}
	c.Consumer = consumerX.NewKafkaConsumer(consumerCg, &consumerX.ConsumerConfig{
		BatchSize:    0,
		BatchTimeout: 0,
	})
	return c
}

func (c *Consumer[T]) InitConsumer(ctx context.Context, topic string) error {
	//return c.ConsumerIn.ReceiveMessage(ctx, []messageQueuex.Tp{{Topic: topic}})
	return c.Consumer.Subscribe(ctx, []string{topic}, newFn(&c.DbConfig, c.L))
}

type fn struct {
	db *DbConf
	l  logx.Loggerx
}

func newFn(db *DbConf, l logx.Loggerx) *fn {
	return &fn{db: db, l: l}
}

func (f *fn) IsBatch() bool {
	return false
}

func (f *fn) Handle(ctx context.Context, msg *mqX.Message) error {
	log.Println("receive message")
	f.l.Info("receive message", logx.Any("msg: ", msg))
	ov, err := NewOverrideFixer[events.TestUser](f.db.SrcDb, f.db.DstDb)
	if err != nil {
		//panic(err)
		return err
	}
	var event events.InconsistentEvent
	err = json.Unmarshal(msg.Value, &event)
	if err != nil {
		return err
	}
	// 修复数据
	err = ov.Fix(context.Background(), event.ID)
	f.l.Info("receive message success, 消费不一致数据", logx.Int64("value_id: ", event.ID), logx.Any("event: ", event))
	return nil
}

func (f *fn) HandleBatch(ctx context.Context, msgs []*mqX.Message) (success bool, err error) {
	return true, nil
}
