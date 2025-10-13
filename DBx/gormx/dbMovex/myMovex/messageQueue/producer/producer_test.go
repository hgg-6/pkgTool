package producer

import (
	"context"
	"encoding/json"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/gormx/dbMovex/myMovex/doubleWritePoolx"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/gormx/dbMovex/myMovex/events"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/messageQueuex"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/messageQueuex/saramax/saramaProducerx"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var addr []string = []string{"localhost:9094"}

// 测试同步发送
func TestNewSaramaProducerStrSync(t *testing.T) {
	cfg := sarama.NewConfig()
	//========同步发送==========
	cfg.Producer.Return.Successes = true

	syncPro, err := sarama.NewSyncProducer(addr, cfg)
	assert.NoError(t, err)
	pro := saramaProducerx.NewSaramaProducerStr[sarama.SyncProducer](syncPro, cfg)
	// CloseProducer 关闭生产者Producer，请在main函数最顶层defer住生产者的Producer.Close()，优雅关闭防止goroutine泄露
	defer pro.CloseProducer()

	value, err := json.Marshal(events.InconsistentEvent{
		ID:        10,
		Direction: "SRC",
		Type:      doubleWritePoolx.PatternSrcFirst,
	})
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// 同步发送
	err = pro.SendMessage(ctx, messageQueuex.Tp{Topic: "dbMove"}, value)
	assert.NoError(t, err)
	cancel()

}
