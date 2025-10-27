package rankingServiceCronX

import (
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/serviceLogicX/cronX"
	"gitee.com/hgg_test/pkg_tool/v2/serviceLogicX/rankingListX"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx"
)

// RankingServiceCron 分布式锁+定时任务服务封装
type RankingServiceCron[T any] struct {
	redSync      *redsyncx.LockRedsync       // 分布式锁
	cron         *cronX.CronX                // 定时任务服务
	rankingListX rankingListX.RankingTopN[T] // 榜单服务
	logx         logx.Loggerx
}

// NewRankingServiceCron 创建定时任务服务
//   - 需优先调用函数设置任务表达式 SetExpr
//   - 需优先调用函数设置任务执行业务逻辑 SetCmd
func NewRankingServiceCron[T any](redSync *redsyncx.LockRedsync, cron *cronX.CronX, logx logx.Loggerx) *RankingServiceCron[T] {
	return &RankingServiceCron[T]{
		redSync: redSync,
		cron:    cron,
		logx:    logx,
	}
}

func (r *RankingServiceCron[T]) Start() error {
	// 启动分布式锁
	r.redSync.Start()
	// 启动定时任务服务
	return r.cron.Start()
}

func (r *RankingServiceCron[T]) Stop() {
	r.cron.StopCron()
	r.redSync.Stop()
}
