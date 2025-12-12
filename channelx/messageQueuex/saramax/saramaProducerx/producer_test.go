package saramaProducerx

import (
	"context"
	"encoding/json"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/messageQueuex"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"log"
	"strconv"
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
	if err != nil {
		t.Skipf("无法连接 Kafka: %v", err)
		return
	}
	pro := NewSaramaProducerStr[sarama.SyncProducer](syncPro, cfg)
	// CloseProducer 关闭生产者Producer，请在main函数最顶层defer住生产者的Producer.Close()，优雅关闭防止goroutine泄露
	defer pro.CloseProducer()

	value, err := json.Marshal(ValTest{
		Name: "test-同步发送",
		Age:  18,
	})
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// 同步发送
	err = pro.SendMessage(ctx, messageQueuex.Tp{Key: []byte("hgg-18"), Topic: "test_topic"}, value)
	assert.NoError(t, err)
	cancel()

}

// 异步发送
func TestNewSaramaProducerStrAsync(t *testing.T) {
	cfg := sarama.NewConfig()
	//========同步发送==========
	cfg.Producer.Return.Successes = true
	// =========异步发送==========新增
	cfg.Producer.Return.Errors = true

	asyncPro, err := sarama.NewAsyncProducer(addr, cfg)
	if err != nil {
		t.Skipf("无法连接 Kafka: %v", err)
		return
	}
	pro := NewSaramaProducerStr[sarama.AsyncProducer](asyncPro, cfg)
	// CloseProducer 关闭生产者Producer，请在main函数最顶层defer住生产者的Producer.Close()，优雅关闭防止goroutine泄露
	defer pro.CloseProducer()

	value, err := json.Marshal(ValTest{
		Name: "test-异步发送",
		Age:  18,
	})
	assert.NoError(t, err)
	log.Println("这是 Marshal(ValTest: 【调试信息】Marshal(ValTest: ", string(value))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// 异步发送
	err = pro.SendMessage(ctx, messageQueuex.Tp{Key: []byte("hgg-18"), Topic: "test_topic111"}, value)
	cancel()
	assert.NoError(t, err)
}

// 异步批量发送+多个实现使用【需为创建不同的Producer实现以及配置】
func TestNewSaramaProducerStrAsyncs(t *testing.T) {
	cfg := sarama.NewConfig()  // 同步发送配置
	cfgs := sarama.NewConfig() // 异步发送配置
	//========同步发送==========
	cfg.Producer.Return.Successes = true
	cfgs.Producer.Return.Successes = true
	// =========异步发送==========新增
	cfgs.Producer.Return.Errors = true
	// =========批量发送=========新增
	cfgs.Producer.Flush.Frequency = 5 * time.Second // 5秒刷新一次，不管有没有达到批量发送数量条件【只有提交了才会写入broker，success/errors通道才会有消息】
	cfgs.Producer.Flush.Messages = 5                // 触发刷新所需的最大消息数,5条消息刷新批量发送一次

	// =========创建同步生产者=========
	syncPro, err := sarama.NewSyncProducer(addr, cfg)
	if err != nil {
		t.Skipf("无法连接 Kafka (同步生产者): %v", err)
		return
	}
	pro := NewSaramaProducerStr[sarama.SyncProducer](syncPro, cfg)

	// =========创建异步生产者=========
	asyncPros, err := sarama.NewAsyncProducer(addr, cfgs)
	if err != nil {
		pro.CloseProducer()
		t.Skipf("无法连接 Kafka (异步生产者): %v", err)
		return
	}
	pros := NewSaramaProducerStr[sarama.AsyncProducer](asyncPros, cfgs)

	// 发送数据
	value, err := json.Marshal(ValTest{
		Name: "test-异步批量发送",
		Age:  18,
	})
	assert.NoError(t, err)

	// 模拟业务中多个地方同时发送数据，单条或者异步批量发送
	var eg errgroup.Group

	// 同步单条发送【并发执行】
	eg.Go(func() error {
		err = pro.SendMessage(context.Background(), messageQueuex.Tp{Key: []byte("hgg-同步发送"), Topic: "test_topic"}, value)
		return err
	})

	// 异步批量发送
	eg.Go(func() error {
		var errs error
		for i := 0; i < 20; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			er := pros.SendMessage(ctx, messageQueuex.Tp{Key: []byte("hgg-异步批量发送-" + strconv.Itoa(i+1)), Topic: "test_topic"}, value)
			cancel()
			time.Sleep(time.Millisecond * 100)
			errs = er
		}
		return errs
	})

	// 等待所有消息处理完【并发执行】
	err = eg.Wait()
	assert.NoError(t, err)

	// CloseProducer 关闭生产者Producer，请在main函数最顶层defer住生产者的Producer.Close()，优雅关闭防止goroutine泄露
	err = pro.CloseProducer()
	assert.NoError(t, err)
	// CloseProducer 关闭生产者Producer，请在main函数最顶层defer住生产者的Producer.Close()，优雅关闭防止goroutine泄露
	err = pros.CloseProducer()
	assert.NoError(t, err)
}
