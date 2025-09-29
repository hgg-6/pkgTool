package redsyncx

import (
	"context"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"log"
	"testing"
	"time"
)

/*
	单节点redis实现分布式锁
*/

func TestRedsync1(t *testing.T) {
	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	client1 := redis.NewClient(&redis.Options{
		Addr: "localhost:6380",
	})
	client2 := redis.NewClient(&redis.Options{
		Addr: "localhost:6380",
	})

	// 测试 Redis 连接
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatal("Redis连接失败:", err)
	}
	// 测试 Redis 连接
	if err := client1.Ping(context.Background()).Err(); err != nil {
		log.Fatal("Redis1连接失败:", err)
	}
	// 测试 Redis 连接
	if err := client2.Ping(context.Background()).Err(); err != nil {
		log.Fatal("Redis1连接失败:", err)
	}

	// 创建 RedSync 池
	pool := goredis.NewPool(client)
	pool1 := goredis.NewPool(client1)
	pool2 := goredis.NewPool(client2)

	// 创建 RedSync 实例
	rs := redsync.New(pool, pool1, pool2)

	// 创建互斥锁
	mutexName := "my-distributed-lock"
	mutex := rs.NewMutex(mutexName,
		redsync.WithExpiry(13*time.Second),           // 锁过期时间
		redsync.WithTries(3),                         // 最大重试次数
		redsync.WithRetryDelay(500*time.Millisecond), // 重试延迟
	)

	// 尝试获取锁
	fmt.Println("尝试获取分布式锁...")
	if err := mutex.Lock(); err != nil {
		log.Fatal("获取锁失败:", err)
	}
	fmt.Println("成功获取锁!")

	// 执行业务逻辑
	doBusinessLogic1()

	// 释放锁
	if ok, err := mutex.Unlock(); !ok || err != nil {
		log.Fatal("释放锁失败:", err)
	}
	fmt.Println("锁已释放")

}
func doBusinessLogic1() {
	fmt.Println("执行业务逻辑中...")
	time.Sleep(10 * time.Second)
	fmt.Println("业务逻辑执行完成")
}
