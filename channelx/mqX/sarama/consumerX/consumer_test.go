package consumerX

import (
	"context"
	"encoding/json"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/mqX"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

var addr []string = []string{"localhost:9094"}

type MyHandler struct{}

func (h *MyHandler) Handle(ctx context.Context, msg *mqX.Message) error {
	var event mqX.UserEventTest
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return err
	}
	// 处理业务逻辑
	log.Println("Received event:", event)
	return nil
}

func TestNewKafkaConsumer(t *testing.T) {
	cfg := sarama.NewConfig()
	saramaCG, err := sarama.NewConsumerGroup(addr, "test_group", cfg)
	assert.NoError(t, err)
	defer saramaCG.Close()

	// 创建你的封装消费者
	kafkaConsumer := NewKafkaConsumer(saramaCG, &ConsumerConfig{
		BatchSize:    20,              // 批量大小
		BatchTimeout: 3 * time.Second, // 批量超时时间
	})

	// 调用你的通用接口方法
	err = kafkaConsumer.Subscribe(context.Background(), []string{"user-events"}, &MyHandler{})
	assert.NoError(t, err)

}
