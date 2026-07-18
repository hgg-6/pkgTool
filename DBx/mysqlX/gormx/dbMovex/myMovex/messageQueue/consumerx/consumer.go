package consumerx

import (
	"context"
	"encoding/json"
	events2 "github.com/hgg-6/pkgTool/v2/DBx/mysqlX/gormx/dbMovex/myMovex/events"
	"github.com/hgg-6/pkgTool/v2/channelx/mqX"
	"github.com/hgg-6/pkgTool/v2/channelx/mqX/kafkaX/saramaX/consumerX"
	"github.com/hgg-6/pkgTool/v2/logx"
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

type Consumer[T events2.InconsistentEvent] struct {
	ConsumerConfig ConsumerConf
	DbConfig       DbConf
	//Fn             func(msg *sarama.ConsumerMessage, event events.InconsistentEvent) error
	L logx.Loggerx
	//ConsumerIn     messageQueuex.ConsumerIn
	Consumer mqX.Consumer
}

func NewConsumer[T events2.InconsistentEvent](consumerConf ConsumerConf, dbConfig DbConf, l logx.Loggerx) *Consumer[T] {
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
	ov, err := NewOverrideFixer[events2.TestUser](f.db.SrcDb, f.db.DstDb)
	if err != nil {
		return err
	}
	var event events2.InconsistentEvent
	err = json.Unmarshal(msg.Value, &event)
	if err != nil {
		return err
	}
	// P0-23: 修复数据。旧实现忽略 Fix 错误直接 return nil，导致修复失败的消息也被 ACK，
	// 不一致数据被永久丢失（消息已提交）。改为返回 error，让上游不 ACK 以便重试。
	if err := ov.Fix(context.Background(), event.ID); err != nil {
		f.l.Error("修复不一致数据失败", logx.Int64("id", event.ID), logx.Error(err))
		return err
	}
	f.l.Info("receive message success, 消费不一致数据", logx.Int64("value_id: ", event.ID), logx.Any("event: ", event))
	return nil
}

func (f *fn) HandleBatch(ctx context.Context, msgs []*mqX.Message) (success bool, err error) {
	return true, nil
}
