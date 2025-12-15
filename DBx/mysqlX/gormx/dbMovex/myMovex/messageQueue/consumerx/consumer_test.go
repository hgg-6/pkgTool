package consumerx

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"gitee.com/hgg_test/pkg_tool/v2/DBx/mysqlX/gormx/dbMovex/myMovex/events"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/logx/zerologx"
	"github.com/IBM/sarama"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

/*
	======================
	测试不一致消息上报kafka后，消费者消费不一致消息
	======================
*/

func setupTestSrcDB() (*gorm.DB, error) {
	srcdb, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13306)/src_db?parseTime=true"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect source database: %w", err)
	}

	// 自动迁移表结构
	err = srcdb.AutoMigrate(&events.TestUser{})
	if err != nil {
		return nil, fmt.Errorf("failed to auto migrate source database: %w", err)
	}
	return srcdb, nil
}

func setupTestDstDB() (*gorm.DB, error) {
	dstdb, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13306)/dst_db?parseTime=true"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect destination database: %w", err)
	}

	// 自动迁移表结构
	err = dstdb.AutoMigrate(&events.TestUser{})
	if err != nil {
		return nil, fmt.Errorf("failed to auto migrate destination database: %w", err)
	}
	return dstdb, nil
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
	// 设置测试数据库连接
	srcDb, err := setupTestSrcDB()
	if err != nil {
		t.Skipf("跳过测试：无法连接源数据库: %v", err)
		return
	}

	dstDb, err := setupTestDstDB()
	if err != nil {
		t.Skipf("跳过测试：无法连接目标数据库: %v", err)
		return
	}

	cfg := sarama.NewConfig()
	// 创建消费者配置
	cm := NewConsumer(ConsumerConf{
		Addr:       addr,
		GroupId:    "test_group",
		SaramaConf: cfg,
	}, DbConf{
		SrcDb: srcDb,
		DstDb: dstDb,
	}, InitLog())

	// 初始化消费者
	err = cm.InitConsumer(context.Background(), "dbMove")
	if err != nil {
		// 检查是否为连接错误
		if isConnectionError(err) {
			t.Skipf("跳过测试：无法连接Kafka或初始化消费者: %v", err)
			return
		}
		assert.NoError(t, err)
	}
}

// isConnectionError 检查错误是否为连接错误
func isConnectionError(err error) bool {
	errStr := err.Error()
	// 检查常见的连接错误消息
	connectionErrors := []string{
		"connection refused",
		"connectex",
		"no such host",
		"timeout",
		"dial tcp",
		"network is unreachable",
		"brokers not available",
		"cannot connect",
	}

	for _, connErr := range connectionErrors {
		if containsIgnoreCase(errStr, connErr) {
			return true
		}
	}
	return false
}

// containsIgnoreCase 检查字符串是否包含子字符串（忽略大小写）
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
