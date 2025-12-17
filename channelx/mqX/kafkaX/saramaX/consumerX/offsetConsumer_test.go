package consumerX

import (
	"context"
	"encoding/json"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/mqX"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

type MyHandlers struct{}

func (h *MyHandlers) HandleBatch(ctx context.Context, msgs []*mqX.Message) (success bool, err error) {
	var events []mqX.UserEventTest
	for _, v := range msgs {
		var event mqX.UserEventTest
		if err := json.Unmarshal(v.Value, &event); err != nil {
			return false, err
		}
		events = append(events, event)
	}
	// 处理业务逻辑
	log.Println("Received event:", events)
	return true, nil
}

func TestExampleOffsetConsumer(t *testing.T) {
	consumer, err := NewOffsetConsumer(addr, &OffsetConsumerConfig{
		BatchSize:    50,
		BatchTimeout: 2 * time.Second,
		AutoCommit:   false, // 由 handler 控制
	})
	if err != nil {
		t.Skipf("跳过测试：无法连接 Kafka: %v", err)
		return
	}
	defer consumer.Close()

	handlers := &MyHandlers{}

	// 生产可在方法传一个ctx，但是具体ctx返回至main函数defer住
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 从 offset=1000 开始消费 partition 0
	err = consumer.ConsumeFrom(ctx, "user-events", 0, 10, handlers)
	assert.NoError(t, err)
}
