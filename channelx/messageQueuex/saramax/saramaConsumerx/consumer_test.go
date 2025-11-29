package saramaConsumerx

import (
	"context"
	"encoding/json"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/messageQueuex"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/messageQueuex/saramax/saramaConsumerx/ConsumerGroupHandlerx"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/messageQueuex/saramax/saramaConsumerx/serviceLogic"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/messageQueuex/saramax/saramaProducerx"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/logx/zerologx"
	"github.com/IBM/sarama"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"
)

/*
	saramaX.NewConsumerGroup的参数groupID核心总结对比表
		特性:			相同 Group ID (竞争消费者模式)			不同 Group ID (发布-订阅模式)
		核心目的			负载均衡，横向扩展处理能力				广播，多个独立系统处理相同数据
		消息传递语义		队列模式								发布-订阅模式
		消息处理			一条消息仅被组内一个消费者处理			一条消息被每一个消费者组处理一次
		偏移量管理		组内共享同一份偏移量					每个组维护自己独立的偏移量
		典型应用			同一服务的多个实例						不同的微服务（如通知服务、分析服务）
*/

var addr []string = []string{"localhost:9094"}

// 测试消费者组处理逻辑, 利用ctx控制消费退出
func TestNewConsumerGroupHandler(t *testing.T) {
	cfg := sarama.NewConfig()
	consumer, err := sarama.NewConsumerGroup(addr, "test_group", cfg)
	assert.NoError(t, err)

	// 开始构造log和消费者消费后的业务逻辑
	// 构造logx
	l := InitLog()
	// 模拟业务逻辑处理，入库、缓存等等。。。
	fn := func(msg *sarama.ConsumerMessage, event saramaProducerx.ValTest) error {
		l.Info("receive message success", logx.String("value_name: ", event.Name))
		return nil
	}
	consumerMsg := serviceLogic.NewSaramaConsumerGroupMessage[saramaProducerx.ValTest](l, fn, nil)
	// 构造消费者组处理逻辑【也可自行实现sarama.ConsumerGroupHandler接口】
	consumerGroupHandlers := ConsumerGroupHandlerx.NewConsumerGroupHandler[saramaProducerx.ValTest](consumerMsg)

	csr := NewConsumerIn(consumer, consumerGroupHandlers)

	//ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	ctx, cancel := context.WithCancel(context.Background()) // 持续消费
	defer cancel()
	err = csr.ReceiveMessage(ctx, []messageQueuex.Tp{{Topic: "test_topic111"}})
	assert.NoError(t, err)
}

// 测试偏移量消费+ctx
func TestNewConsumerGroupHandler_Offset(t *testing.T) {
	cfg := sarama.NewConfig()
	consumer, err := sarama.NewConsumerGroup(addr, "test_group", cfg)
	assert.NoError(t, err)

	// 开始构造log和消费者消费后的业务逻辑
	// 构造logx
	l := InitLog()
	// 模拟业务逻辑处理，入库、缓存等等。。。
	fn := func(msg *sarama.ConsumerMessage, event saramaProducerx.ValTest) error {
		v, er := json.Marshal(event)
		assert.NoError(t, er)
		l.Info("receive message success", logx.String("value_name: ", string(v)), logx.Int64("offset", msg.Offset))
		return nil
	}
	consumerMsg := serviceLogic.NewSaramaConsumerGroupMessage[saramaProducerx.ValTest](l, fn, nil)
	consumerMsg.SetOffset(true, "test_topic111", 30) // 设置偏移量

	// 构造消费者组处理逻辑【也可自行实现sarama.ConsumerGroupHandler接口】
	consumerGroupHandlers := ConsumerGroupHandlerx.NewConsumerGroupHandler[saramaProducerx.ValTest](consumerMsg)

	csr := NewConsumerIn(consumer, consumerGroupHandlers)

	//ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	log.Println("第一次消费")
	ctx, cancel := context.WithCancel(context.Background()) // 持续消费
	err = csr.ReceiveMessage(ctx, []messageQueuex.Tp{{Topic: "test_topic111"}})
	cancel()
	assert.NoError(t, err)
}

// 测试批量消费+批量偏移量+ctx
func TestNewConsumerGroupHandler_batch(t *testing.T) {
	cfg := sarama.NewConfig()
	consumer, err := sarama.NewConsumerGroup(addr, "test_group", cfg)
	assert.NoError(t, err)

	// 开始构造log和消费者消费后的业务逻辑
	// 构造logx
	l := InitLog()
	// 模拟业务逻辑处理，入库、缓存等等。。。
	fns := func(msgs []*sarama.ConsumerMessage, event []saramaProducerx.ValTest) error {
		for k, msg := range msgs {
			l.Info("receive message success", logx.String("value_name: ", event[k].Name), logx.Int64("offset", msg.Offset))
		}
		time.Sleep(time.Second * 2)
		return nil
	}
	consumerMsg := serviceLogic.NewSaramaConsumerGroupMessage[saramaProducerx.ValTest](l, nil, fns)
	consumerMsg.SetOffset(true, "test_topic111", 30) // 设置偏移量
	consumerMsg.SetBatch(true, 5)                    // 设置批量消费

	// 构造消费者组处理逻辑【也可自行实现sarama.ConsumerGroupHandler接口】
	consumerGroupHandlers := ConsumerGroupHandlerx.NewConsumerGroupHandler[saramaProducerx.ValTest](consumerMsg)

	csr := NewConsumerIn(consumer, consumerGroupHandlers)

	//ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	log.Println("第一次消费")
	ctx, cancel := context.WithCancel(context.Background()) // 持续消费
	err = csr.ReceiveMessage(ctx, []messageQueuex.Tp{{Topic: "test_topic111"}})
	cancel()
	assert.NoError(t, err)
}

func InitLog() logx.Loggerx {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// Level日志级别【可以考虑作为参数传】，测试传zerolog.InfoLevel/NoLevel不打印
	// 模块化: Str("module", "userService模块")
	logger := zerolog.New(os.Stderr).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	return zerologx.NewZeroLogger(&logger)
}
