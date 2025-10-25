package rankingServiceCronX

import (
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/serviceLogicX/cronX"
)

// RankingServiceCron 定时任务服务封装
type RankingServiceCron struct {
	cron *cronX.CronX
	logx logx.Loggerx
}

// NewRankingServiceCron 创建定时任务服务
//   - 需优先调用函数设置任务表达式 SetExpr
//   - 需优先调用函数设置任务执行业务逻辑 SetCmd
func NewRankingServiceCron(logx logx.Loggerx, cronx *cronX.CronX) *RankingServiceCron {
	return &RankingServiceCron{logx: logx, cron: cronx}
}
