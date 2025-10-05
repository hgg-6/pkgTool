package cron

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestCron(t *testing.T) {
	expr := cron.New(cron.WithSeconds()) // 秒级
	// https://help.aliyun.com/document_detail/133509.html可以参考这个，任务调度cron表达式
	id, err := expr.AddFunc("@every 5s", func() { // 5秒一次定时任务
		t.Log("5秒一次定时任务")
	})
	assert.NoError(t, err)
	t.Log("任务id: ", id)

	// 获取所有任务条目
	entries := expr.Entries()
	fmt.Printf("当前有 %d 个任务\n", len(entries))
	for k, v := range entries {
		t.Logf("第%d个任务，任务详情:%v", k+1, v)
	}

	expr.Start() // 启动定时器

	// 运行1分钟，5秒一次，任务需持续运行的话实际也可在main控制stop退出
	time.Sleep(time.Minute)

	ctx := expr.Stop() // 暂停定时器，不调度新任务执行了，正在执行的继续执行
	t.Log("发出停止信号")
	<-ctx.Done() // 彻底停止定时器
	t.Log("彻底停止，没有任务执行了")
}

func TestCronTicker(t *testing.T) {
	ctx, cancelc := context.WithTimeout(context.Background(), time.Minute)
	defer cancelc()

	// 每10分钟执行一次入库
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ticker.C:
			// 时间到了，可以执行任务了
			// 【限制任务总时间的话，eg: 5秒运行一次任务，总计1分钟，运行12次，那么for外部创建1分钟的context.WithTimeout】
			log.Println("每10秒执行一次任务")
		case <-ctx.Done():
			log.Println("任务持续总时长一分钟，任务结束")
			//return
			break loop // break不能中断当前for循环，【但是可以使用golang中标签，break loop就能跳出指定循环位置】
		}
	}

	t.Log("跳出循环，任务结束")
}

func TestCronTickerV1(t *testing.T) {
	ctx, cancelc := context.WithTimeout(context.Background(), time.Minute)
	defer cancelc()

	// 每10分钟执行一次入库
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 时间到了，可以执行任务了
			// 【限制任务总时间的话，eg: 5秒运行一次任务，总计1分钟，运行12次，那么for外部创建1分钟的context.WithTimeout】
			log.Println("每10秒执行一次任务")
		case <-ctx.Done():
			log.Println("任务持续总时长一分钟，任务结束")
			goto end // 跳出当前for循环【最好不用goto ，可读性差】
		}
	}

end:
	t.Log("跳出循环，任务结束")
}
