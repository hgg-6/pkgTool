package redsyncx

import (
	"context"
	"os"
	"testing"
	"time"

	"gitee.com/hgg_test/pkg_tool/v2/logx/zerologx"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewResSyncStr12(t *testing.T) {
	var clis []*redis.Client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	clis = append(clis, client)
	client1 := redis.NewClient(&redis.Options{
		Addr: "localhost:6380",
	})
	clis = append(clis, client1)
	client2 := redis.NewClient(&redis.Options{
		Addr: "localhost:6381",
	})
	clis = append(clis, client2)

	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("Redis连接失败: %v", err)
	}
	if err := client1.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("Redis1连接失败: %v", err)
	}
	if err := client2.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("Redis2连接失败: %v", err)
	}

	// =========================
	// =========================
	// =========================

	// 创建日志器
	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
	zlog := zerologx.NewZeroLogger(&logger)

	// 创建分布式锁配置
	config := Config{
		LockName:   "test-lock",
		Expiry:     10 * time.Second,
		RetryDelay: 1 * time.Second,
		MaxRetries: 2,
	}

	// 创建分布式锁实例【可以根据gopsutilx包下的获取系统负载，然后判断选择负载较低的实例，负载信息可以存储redis/mysql】
	dl := NewLockRedsync(clis, zlog, config)
	defer dl.Stop() // 停止锁并释放资源
	dl.Start()      // 启动锁获取和续约

	time.Sleep(time.Second)

	// ============方式1===============
	// ============方式1===============
	// 监听锁状态，定时任务测试
	// 1. 生成一个cron表达式
	expr := cron.New(cron.WithSeconds())          // 秒级
	id, err := expr.AddFunc("@every 5s", func() { // 5秒一次定时任务
		if dl.IsLocked() {
			logicService(t)
		}
	})
	id1, err1 := expr.AddFunc("@every 5s", func() { // 5秒一次定时任务
		if dl.IsLocked() {
			logicService11(t)
		}
	})
	assert.NoError(t, err)
	assert.NoError(t, err1)
	t.Log("任务id: ", id)
	t.Log("任务id: ", id1)

	expr.Start() // 启动定时器

	// 模拟定时任务总时间20秒，20秒后停止定时器【实际业务可以expr := cron.New后返回expr，由main控制退出】
	time.Sleep(time.Second * 30)

	ctx := expr.Stop() // 暂停定时器，不调度新任务执行了，正在执行的继续执行
	t.Log("发出停止信号")
	<-ctx.Done() // 彻底停止定时器
	t.Log("彻底停止，没有任务执行了")

	// ==============方式2=============
	// ==============方式2=============
	// 监听锁状态，定时任务测试, 方式2
	//ticker := time.NewTicker(time.Second * 5)
	//for {
	//	select {
	//	case <-ticker.C:
	//		// 锁已获取，执行业务逻辑
	//		if dl.IsLocked() {
	//			logicService(t)
	//		}
	//		continue
	//	}
	//}
}

func logicService(t *testing.T) {
	t.Log(time.Now().Format(time.DateTime), "开始执行业务逻辑1")
	time.Sleep(time.Second * 2)
	t.Log(time.Now().Format(time.DateTime), "done logicService")
}

func logicService11(t *testing.T) {
	t.Log(time.Now().Format(time.DateTime), "开始执行业务逻辑2")
	time.Sleep(time.Second * 2)
	t.Log(time.Now().Format(time.DateTime), "done logicService")
}
