package consumerX

import (
	"context"
	"encoding/json"
	"log"
	"testing"
	"time"

	"gitee.com/hgg_test/pkg_tool/v2/channelx/mqX"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
)

var addr []string = []string{"localhost:9094"}

// // ============消费数据后业务逻辑=================
type MyHandler struct{}

// IsBatch 业务处理逻辑是否批量
func (h *MyHandler) IsBatch() bool {
	return false
}

func (h *MyHandler) HandleBatch(ctx context.Context, msgs []*mqX.Message) (success bool, err error) {
	//var events []mqX.UserEventTest
	//for _, v := range msgs {
	//	var event mqX.UserEventTest
	//	if er := json.Unmarshal(v.Value, &event); er != nil {
	//		return false, er
	//	}
	//	events = append(events, event)
	//}
	//log.Println("Received events:", events)
	return true, nil
}

func (h *MyHandler) Handle(ctx context.Context, msg *mqX.Message) error {
	var event mqX.UserEventTest
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return err
	}
	// 处理业务逻辑
	log.Println("Received event:", event)
	return nil
}

// ============消费者=================
// 测试的消费者
func TestNewKafkaConsumer(t *testing.T) {
	cfg := sarama.NewConfig()
	saramaCG, err := sarama.NewConsumerGroup(addr, "test_group", cfg)
	if err != nil {
		t.Skipf("无法连接 Kafka: %v", err)
		return
	}
	defer saramaCG.Close()

	// 创建你的封装消费者
	kafkaConsumer := NewKafkaConsumer(saramaCG, &ConsumerConfig{
		BatchSize:    20,              // 批量大小
		BatchTimeout: 3 * time.Second, // 批量超时时间
	})

	// 调用你的通用接口方法，消费消息
	err = kafkaConsumer.Subscribe(context.Background(), []string{"user-events"}, &MyHandler{})
	assert.NoError(t, err)

}
