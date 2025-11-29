package producerX

import (
	"context"
	"encoding/json"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/mqX"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"strconv"
	"testing"
	"time"
)

var addr []string = []string{"localhost:9094"}

func TestNewKafkaProducer(t *testing.T) {

	// =========创建异步批量生产者=========
	pro, err := NewKafkaProducer(addr, &ProducerConfig{
		Async:        true,            // true默认为异步批量发送
		BatchSize:    200,             // 批量发送消息数量
		BatchTimeout: 3 * time.Second, // 批量发送消息时间间隔
	})
	assert.NoError(t, err)

	// =========创建同步生产者=========
	pros, err := NewKafkaProducer(addr, &ProducerConfig{
		Async: false, // false为同步
	})
	assert.NoError(t, err)

	defer func() {
		pro.Close()
		pros.Close()
	}()

	user := mqX.UserEventTest{
		UserId: 1,
		Name:   "hggTest",
	}
	val, err := json.Marshal(&user)
	assert.NoError(t, err)

	// 并发执行按单个发送和批量发送生产者
	var eg errgroup.Group
	for i := 0; i < 5; i++ {
		eg.Go(func() error {
			return pro.Send(context.Background(), &mqX.Message{
				Topic: "user-events",
				Key:   []byte("user-123"),
				Value: val,
			})
		})
	}
	time.Sleep(time.Second)
	eg.Go(func() error {
		var ms []*mqX.Message
		var er error
		for i := 0; i < 20; i++ {
			use := mqX.UserEventTest{
				UserId: int64(i + 1),
				Name:   "hggTest" + strconv.Itoa(i+1),
			}
			va, e := json.Marshal(&use)
			assert.NoError(t, e)
			ms = append(ms, &mqX.Message{Topic: "user-events", Value: va})
		}
		err = pros.SendBatch(context.Background(), ms)
		assert.NoError(t, err)
		return er
	})
	err = eg.Wait()
	assert.NoError(t, err)

}
