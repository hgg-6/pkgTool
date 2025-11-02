package consumerx

import (
	"context"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/mysqlX/gormx/dbMovex/myMovex/events"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/logx/zerologx"
	"github.com/IBM/sarama"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"testing"
)

/*
	======================
	测试不一致消息上报kafka后，消费者消费不一致消息
	======================
*/

func setupTestSrcDB() *gorm.DB {
	srcdb, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13306)/src_db?parseTime=true"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移表结构
	srcdb.AutoMigrate(&events.TestUser{})
	return srcdb
}
func setupTestDstDB() *gorm.DB {
	dstdb, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13306)/dst_db?parseTime=true"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移表结构
	dstdb.AutoMigrate(&events.TestUser{})
	return dstdb
}

var addr []string = []string{"localhost:9094"}

func InitLog() logx.Loggerx {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// Level日志级别【可以考虑作为参数传】，测试传zerolog.InfoLevel/NoLevel不打印
	// 模块化: Str("module", "userService模块")
	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().CallerWithSkipFrameCount(4).Caller().Timestamp().Logger()
	return zerologx.NewZeroLogger(&logger)
}

func TestNewConsumerGroupHandler1(t *testing.T) {
	cfg := sarama.NewConfig()
	//consumer, err := sarama.NewConsumerGroup(addr, "test_group", cfg)
	//assert.NoError(t, err)

	// 开始构造log和消费者消费后的业务逻辑
	// 构造logx
	//l := InitLog()
	//// 模拟业务逻辑处理，入库、缓存等等。。。
	//fn := func(msg *sarama.ConsumerMessage, event events.InconsistentEvent) error {
	//	srcDb := setupTestSrcDB()
	//	dstDb := setupTestDstDB()
	//	ov, er := NewOverrideFixer[myMovex.TestUser](srcDb, dstDb)
	//	assert.NoError(t, er)
	//	er = ov.Fix(context.Background(), event.ID)
	//	assert.NoError(t, er)
	//
	//	l.Info("receive message success", logx.Int64("value_id: ", event.ID), logx.Any("event: ", event))
	//	return nil
	//}
	//consumerMsg := serviceLogic.NewSaramaConsumerGroupMessage[events.InconsistentEvent](l, fn, nil)
	//// 构造消费者组处理逻辑【也可自行实现sarama.ConsumerGroupHandler接口】
	//consumerGroupHandlers := ConsumerGroupHandlerx.NewConsumerGroupHandler[events.InconsistentEvent](consumerMsg)
	//
	//csr := saramaConsumerx.NewConsumerIn(consumer, consumerGroupHandlers)
	//
	//ctx, cancel := context.WithCancel(context.Background()) // 持续消费
	//defer cancel()
	//err = csr.ReceiveMessage(ctx, []messageQueuex.Tp{{Topic: "dbMove"}})
	//assert.NoError(t, err)

	cm := NewConsumer(ConsumerConf{Addr: addr, GroupId: "test_group", SaramaConf: cfg}, DbConf{SrcDb: setupTestSrcDB(), DstDb: setupTestDstDB()}, InitLog())
	err := cm.InitConsumer(context.Background(), "dbMove")
	assert.NoError(t, err)
}
