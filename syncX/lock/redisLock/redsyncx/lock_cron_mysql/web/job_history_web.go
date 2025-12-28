package web

import (
	"errors"
	"strconv"

	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/service"
	"github.com/gin-gonic/gin"
)

// JobHistoryWeb 任务执行历史Web处理器
type JobHistoryWeb struct {
	historySvc service.JobHistoryService
	l          logx.Loggerx
}

// NewJobHistoryWeb 创建JobHistoryWeb实例
func NewJobHistoryWeb(historySvc service.JobHistoryService, l logx.Loggerx) *JobHistoryWeb {
	return &JobHistoryWeb{
		historySvc: historySvc,
		l:          l,
	}
}

// Register 注册路由
func (h *JobHistoryWeb) Register(server interface{}) {
	var g *gin.RouterGroup
	
	// 支持接受 gin.Engine 或 gin.RouterGroup
	switch s := server.(type) {
	case *gin.Engine:
		g = s.Group("/job-history")
	case *gin.RouterGroup:
		g = s
	default:
		return
	}
	
	{
		// 查询单条历史记录
		g.GET("/detail/:id", h.GetHistory)
		// 查询任务的执行历史列表（分页）
		g.GET("/list/:cron_id", h.GetHistoryList)
		// 根据状态查询执行历史列表（分页）
		g.GET("/list-by-status", h.GetHistoryListByStatus)
		// 根据时间范围查询执行历史列表
		g.GET("/list-by-time", h.GetHistoryListByTimeRange)
		// 获取任务的最新执行历史
		g.GET("/latest/:cron_id", h.GetLatestHistory)
		// 获取任务的执行统计信息
		g.GET("/statistics/:cron_id", h.GetStatistics)
		// 删除单条历史记录
		g.DELETE("/delete/:id", h.DeleteHistory)
		// 删除任务的所有历史记录
		g.DELETE("/delete-by-job/:cron_id", h.DeleteHistoryByCronId)
		// 清理旧的历史记录
		g.DELETE("/cleanup", h.CleanupOldHistory)
	}
}

// GetHistory 获取单条执行历史
func (h *JobHistoryWeb) GetHistory(ctx *gin.Context) {
	id := ctx.Param("id")
	historyId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		h.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	history, err := h.historySvc.GetHistory(ctx.Request.Context(), historyId)
	switch err {
	case service.ErrDataRecordNotFound:
		h.l.Error("没有该数据", logx.Int64("history_id", historyId), logx.Error(err))
		ctx.JSON(404, gin.H{"error": "记录不存在"})
		return
	case nil:
		h.l.Info("查询执行历史成功", logx.Int64("history_id", historyId))
		ctx.JSON(200, gin.H{
			"code": 200,
			"msg":  "success",
			"data": history,
		})
	default:
		h.l.Error("查询执行历史失败", logx.Int64("history_id", historyId), logx.Error(err))
		ctx.JSON(500, gin.H{"error": "查询失败"})
	}
}

// GetHistoryList 获取任务的执行历史列表
func (h *JobHistoryWeb) GetHistoryList(ctx *gin.Context) {
	cronIdStr := ctx.Param("cron_id")
	cronId, err := strconv.ParseInt(cronIdStr, 10, 64)
	if err != nil {
		h.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

	histories, total, err := h.historySvc.GetHistoryList(ctx.Request.Context(), cronId, page, pageSize)
	if err != nil {
		h.l.Error("查询执行历史列表失败", logx.Int64("cron_id", cronId), logx.Error(err))
		ctx.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	h.l.Info("查询执行历史列表成功", logx.Int64("cron_id", cronId), logx.Int("count", len(histories)))
	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"list":      histories,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetHistoryListByStatus 根据状态查询执行历史列表
func (h *JobHistoryWeb) GetHistoryListByStatus(ctx *gin.Context) {
	statusStr := ctx.Query("status")
	if statusStr == "" {
		ctx.JSON(400, gin.H{"error": "status参数不能为空"})
		return
	}

	status := domain.ExecutionStatus(statusStr)

	// 获取分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

	histories, total, err := h.historySvc.GetHistoryListByStatus(ctx.Request.Context(), status, page, pageSize)
	if err != nil {
		h.l.Error("根据状态查询执行历史列表失败", logx.String("status", string(status)), logx.Error(err))
		ctx.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	h.l.Info("根据状态查询执行历史列表成功", logx.String("status", string(status)), logx.Int("count", len(histories)))
	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"list":      histories,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetHistoryListByTimeRange 根据时间范围查询执行历史列表
func (h *JobHistoryWeb) GetHistoryListByTimeRange(ctx *gin.Context) {
	startTimeStr := ctx.Query("start_time")
	endTimeStr := ctx.Query("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		ctx.JSON(400, gin.H{"error": "start_time和end_time参数不能为空"})
		return
	}

	startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "start_time参数格式错误"})
		return
	}

	endTime, err := strconv.ParseInt(endTimeStr, 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "end_time参数格式错误"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

	histories, err := h.historySvc.GetHistoryListByTimeRange(ctx.Request.Context(), startTime, endTime, page, pageSize)
	if err != nil {
		h.l.Error("根据时间范围查询执行历史列表失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	h.l.Info("根据时间范围查询执行历史列表成功", logx.Int("count", len(histories)))
	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"list":      histories,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetLatestHistory 获取任务的最新执行历史
func (h *JobHistoryWeb) GetLatestHistory(ctx *gin.Context) {
	cronIdStr := ctx.Param("cron_id")
	cronId, err := strconv.ParseInt(cronIdStr, 10, 64)
	if err != nil {
		h.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	history, err := h.historySvc.GetLatestHistory(ctx.Request.Context(), cronId)
	switch err {
	case service.ErrDataRecordNotFound:
		h.l.Warn("该任务暂无执行历史", logx.Int64("cron_id", cronId))
		ctx.JSON(404, gin.H{"error": "暂无执行历史"})
		return
	case nil:
		h.l.Info("获取最新执行历史成功", logx.Int64("cron_id", cronId))
		ctx.JSON(200, gin.H{
			"code": 200,
			"msg":  "success",
			"data": history,
		})
	default:
		h.l.Error("获取最新执行历史失败", logx.Int64("cron_id", cronId), logx.Error(err))
		ctx.JSON(500, gin.H{"error": "查询失败"})
	}
}

// GetStatistics 获取任务的执行统计信息
func (h *JobHistoryWeb) GetStatistics(ctx *gin.Context) {
	cronIdStr := ctx.Param("cron_id")
	cronId, err := strconv.ParseInt(cronIdStr, 10, 64)
	if err != nil {
		h.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	// 获取天数参数，默认7天
	days, _ := strconv.Atoi(ctx.DefaultQuery("days", "7"))

	stats, err := h.historySvc.GetStatistics(ctx.Request.Context(), cronId, days)
	if err != nil {
		h.l.Error("获取执行统计信息失败", logx.Int64("cron_id", cronId), logx.Error(err))
		ctx.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	h.l.Info("获取执行统计信息成功", logx.Int64("cron_id", cronId), logx.Int("days", days))
	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": stats,
	})
}

// DeleteHistory 删除单条历史记录
func (h *JobHistoryWeb) DeleteHistory(ctx *gin.Context) {
	id := ctx.Param("id")
	historyId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		h.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	err = h.historySvc.DeleteHistory(ctx.Request.Context(), historyId)
	if err != nil {
		h.l.Error("删除执行历史失败", logx.Int64("history_id", historyId), logx.Error(err))
		ctx.JSON(500, gin.H{"error": "删除失败"})
		return
	}

	h.l.Info("删除执行历史成功", logx.Int64("history_id", historyId))
	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": "delete ok!",
	})
}

// DeleteHistoryByCronId 删除任务的所有历史记录
func (h *JobHistoryWeb) DeleteHistoryByCronId(ctx *gin.Context) {
	cronIdStr := ctx.Param("cron_id")
	cronId, err := strconv.ParseInt(cronIdStr, 10, 64)
	if err != nil {
		h.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	err = h.historySvc.DeleteHistoryByCronId(ctx.Request.Context(), cronId)
	if err != nil {
		h.l.Error("删除任务的所有执行历史失败", logx.Int64("cron_id", cronId), logx.Error(err))
		ctx.JSON(500, gin.H{"error": "删除失败"})
		return
	}

	h.l.Info("删除任务的所有执行历史成功", logx.Int64("cron_id", cronId))
	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": "delete ok!",
	})
}

// CleanupOldHistory 清理旧的历史记录
func (h *JobHistoryWeb) CleanupOldHistory(ctx *gin.Context) {
	// 获取天数参数，默认30天
	days, _ := strconv.Atoi(ctx.DefaultQuery("days", "30"))

	err := h.historySvc.CleanupOldHistory(ctx.Request.Context(), days)
	if err != nil {
		h.l.Error("清理旧的执行历史失败", logx.Int("days", days), logx.Error(err))
		ctx.JSON(500, gin.H{"error": errors.New("清理失败")})
		return
	}

	h.l.Info("清理旧的执行历史成功", logx.Int("days", days))
	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": "cleanup ok!",
	})
}
