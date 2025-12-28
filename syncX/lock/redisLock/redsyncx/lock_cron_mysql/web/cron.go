package web

import (
	"errors"
	"strconv"
	"strings"

	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/service"
	"github.com/gin-gonic/gin"
)

type CronWeb struct {
	cronSvc service.CronService
	l       logx.Loggerx
}

// NewCronWeb 创建CronWeb实例
func NewCronWeb(cronSvc service.CronService, l logx.Loggerx) *CronWeb {
	return &CronWeb{cronSvc: cronSvc, l: l}
}

func (c *CronWeb) Register(server *gin.Engine) {
	g := server.Group("/cron")
	{
		g.GET("/find/:cron_id", c.FindId)      // 获取单个任务 eg: GET /find/123【g.GET("/find", c.FindId)，请求示例：GET /find?cron_id=123】
		g.GET("/profile", c.FindAll)           // 获取所有任务
		g.POST("/add", c.Add)                  // 添加任务
		g.POST("/adds", c.Adds)                // 批量添加任务
		g.DELETE("/delete/:cron_id", c.Delete) // 删除单个任务，删除任务后续可实现管理员删除的用户权限控制
		g.DELETE("/deletes/", c.Deletes)       // 批量删除任务
		// 状态管理接口
		g.PUT("/start/:cron_id", c.StartJob)   // 启动任务
		g.PUT("/pause/:cron_id", c.PauseJob)   // 暂停任务
		g.PUT("/resume/:cron_id", c.ResumeJob) // 恢复任务
	}
}

func (c *CronWeb) FindId(ctx *gin.Context) {
	id := ctx.Param("cron_id")
	cronId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": errors.New("查找失败")})
		return
	}
	cron, err := c.cronSvc.GetCronJob(ctx.Request.Context(), cronId)
	switch err {
	case service.ErrDataRecordNotFound:
		c.l.Error("没有该数据", logx.Int64("cronId", cronId), logx.Error(err))
		ctx.JSON(404, gin.H{"error": errors.New("查找失败")})
		return
	case nil:
		c.l.Info("查找成功", logx.Int64("cronId", cronId), logx.Any("data", cron))
		ctx.JSON(200, gin.H{
			"code": 200,
			"msg":  "success",
			"data": cron,
		})
	default:
		c.l.Error("查找失败", logx.Int64("cronId", cronId), logx.Error(err))
		ctx.JSON(500, gin.H{"error": errors.New("查找失败")})
	}

}

func (c *CronWeb) FindAll(ctx *gin.Context) {
	crons, err := c.cronSvc.GetCronJobs(ctx.Request.Context())
	switch err {
	case service.ErrDataRecordNotFound:
		c.l.Error("没有该数据", logx.Error(err))
		ctx.JSON(404, gin.H{"error": errors.New("查找失败")})
		return
	case nil:
		c.l.Info("查找成功", logx.Any("data", crons))
		ctx.JSON(200, gin.H{
			"code": 200,
			"msg":  "success",
			"data": crons,
		})
	default:
		c.l.Error("查找失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": errors.New("查找失败")})
	}
}

func (c *CronWeb) Add(ctx *gin.Context) {
	var cronJob domain.CronJob
	err := ctx.Bind(&cronJob)
	if err != nil {
		ctx.JSON(400, gin.H{"error": errors.New("参数错误")})
	}
	err = c.cronSvc.AddCronJob(ctx.Request.Context(), cronJob)
	switch err {
	case service.ErrDuplicateData:
		c.l.Error("数据已存在, 添加失败", logx.Int64("cronId", cronJob.ID), logx.String("cronName", cronJob.Name), logx.Error(err))
		ctx.JSON(400, gin.H{"error": errors.New("添加失败")})
	case nil:
		c.l.Info("添加成功", logx.Any("data", cronJob))
		ctx.JSON(200, gin.H{
			"code": 200,
			"msg":  "success",
			"data": "add ok!",
		})
	default:
		c.l.Error("添加失败", logx.Int64("cronId", cronJob.ID), logx.String("cronName", cronJob.Name), logx.Error(err))
		ctx.JSON(500, gin.H{"error": errors.New("添加失败")})
	}
}

func (c *CronWeb) Adds(ctx *gin.Context) {
	var cronJobs []domain.CronJob
	err := ctx.Bind(&cronJobs)
	if err != nil {
		c.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": errors.New("参数错误")})
	}
	err = c.cronSvc.AddCronJobs(ctx.Request.Context(), cronJobs)
	switch err {
	case service.ErrDuplicateData:
		c.l.Error("数据已存在, 添加失败", logx.Error(err))
		ctx.JSON(400, gin.H{"error": errors.New("添加失败")})
	case nil:
		c.l.Info("添加成功", logx.Any("data", cronJobs))
		ctx.JSON(200, gin.H{
			"code": 200,
			"msg":  "success",
			"data": "add ok!",
		})
	default:
		c.l.Error("添加失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": errors.New("添加失败")})
	}
}

func (c *CronWeb) Delete(ctx *gin.Context) {
	id := ctx.Param("cron_id")
	cronId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": errors.New("查找失败")})
		return
	}
	err = c.cronSvc.DelCronJob(ctx.Request.Context(), cronId)
	if err != nil {
		c.l.Error("删除失败", logx.Int64("cronId", cronId), logx.Error(err))
		ctx.JSON(500, gin.H{"error": errors.New("删除失败")})
	}
	c.l.Info("删除成功", logx.Int64("cronId", cronId))
	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": "delete ok!",
	})
}

func (c *CronWeb) Deletes(ctx *gin.Context) {
	// 获取查询参数ids
	idsStr := ctx.Query("cron_ids") // 返回"1,2,3"这样的字符串

	if idsStr == "" {
		ctx.JSON(400, gin.H{"error": "ids parameter is required"})
		return
	}

	// 解析ID字符串
	idStrings := strings.Split(idsStr, ",")
	var ids []int64
	for _, v := range idStrings {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.l.Error("参数错误, 批量删除时，cron_id参数解析失败", logx.Error(err))
			ctx.JSON(400, gin.H{"error": "参数错误"})
		}
		ids = append(ids, id)
	}

	err := c.cronSvc.DelCronJobs(ctx.Request.Context(), ids)
	if err != nil {
		c.l.Error("批量删除失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": errors.New("删除失败")})
	}
	c.l.Info("批量删除成功")
	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": "delete ok!",
	})
}

// StartJob 启动任务
func (c *CronWeb) StartJob(ctx *gin.Context) {
	id := ctx.Param("cron_id")
	cronId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": errors.New("参数错误")})
		return
	}

	err = c.cronSvc.StartJob(ctx.Request.Context(), cronId)
	switch err {
	case service.ErrDataRecordNotFound:
		c.l.Error("任务不存在", logx.Int64("cronId", cronId), logx.Error(err))
		ctx.JSON(404, gin.H{"error": errors.New("任务不存在")})
	case service.ErrInvalidStatusChange:
		c.l.Error("无效的状态变更", logx.Int64("cronId", cronId), logx.Error(err))
		ctx.JSON(400, gin.H{"error": errors.New("任务当前状态不允许启动")})
	case nil:
		c.l.Info("启动任务成功", logx.Int64("cronId", cronId))
		ctx.JSON(200, gin.H{
			"code": 200,
			"msg":  "success",
			"data": "task started",
		})
	default:
		c.l.Error("启动任务失败", logx.Int64("cronId", cronId), logx.Error(err))
		ctx.JSON(500, gin.H{"error": errors.New("启动任务失败")})
	}
}

// PauseJob 暂停任务
func (c *CronWeb) PauseJob(ctx *gin.Context) {
	id := ctx.Param("cron_id")
	cronId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": errors.New("参数错误")})
		return
	}

	err = c.cronSvc.PauseJob(ctx.Request.Context(), cronId)
	switch err {
	case service.ErrDataRecordNotFound:
		c.l.Error("任务不存在", logx.Int64("cronId", cronId), logx.Error(err))
		ctx.JSON(404, gin.H{"error": errors.New("任务不存在")})
	case service.ErrInvalidStatusChange:
		c.l.Error("无效的状态变更", logx.Int64("cronId", cronId), logx.Error(err))
		ctx.JSON(400, gin.H{"error": errors.New("任务当前状态不允许暂停")})
	case nil:
		c.l.Info("暂停任务成功", logx.Int64("cronId", cronId))
		ctx.JSON(200, gin.H{
			"code": 200,
			"msg":  "success",
			"data": "task paused",
		})
	default:
		c.l.Error("暂停任务失败", logx.Int64("cronId", cronId), logx.Error(err))
		ctx.JSON(500, gin.H{"error": errors.New("暂停任务失败")})
	}
}

// ResumeJob 恢复任务
func (c *CronWeb) ResumeJob(ctx *gin.Context) {
	id := ctx.Param("cron_id")
	cronId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": errors.New("参数错误")})
		return
	}

	err = c.cronSvc.ResumeJob(ctx.Request.Context(), cronId)
	switch err {
	case service.ErrDataRecordNotFound:
		c.l.Error("任务不存在", logx.Int64("cronId", cronId), logx.Error(err))
		ctx.JSON(404, gin.H{"error": errors.New("任务不存在")})
	case service.ErrInvalidStatusChange:
		c.l.Error("无效的状态变更", logx.Int64("cronId", cronId), logx.Error(err))
		ctx.JSON(400, gin.H{"error": errors.New("任务当前状态不允许恢复")})
	case nil:
		c.l.Info("恢复任务成功", logx.Int64("cronId", cronId))
		ctx.JSON(200, gin.H{
			"code": 200,
			"msg":  "success",
			"data": "task resumed",
		})
	default:
		c.l.Error("恢复任务失败", logx.Int64("cronId", cronId), logx.Error(err))
		ctx.JSON(500, gin.H{"error": errors.New("恢复任务失败")})
	}
}
