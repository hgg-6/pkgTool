package cronX

import (
	"fmt"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/syncX"
	"gitee.com/hgg_test/pkg_tool/v2/systemLoad/gopsutilx"
	"github.com/robfig/cron/v3"
	"sync"
	"sync/atomic"
	"time"
)

type CronCmd func()

// CronX 定时任务服务封装，支持定时任务执行时检查系统负载等功能，如果过高则自动暂停任务从而避免任务执行压垮系统
type CronX struct {
	mu                    sync.Mutex
	cron                  *cron.Cron
	adminCron             *syncX.Map[string, CronXConfig]    // 管理定时任务，用户删除任务等
	exprOrCmd             *syncX.Map[string, CronXCmdConfig] // cron 表达式【任务多久执行一次/每天xxx执行一次】，可参考https://help.aliyun.com/document_detail/133509.html
	atc                   atomic.Int32                       // 0: 默认值，表示任务正常执行；1: 表示任务被暂停；
	logx                  logx.Loggerx
	IsSystemInfo          bool                  // 是否启用自动刷新系统负载信息,默认false
	systemInfo            *gopsutilx.SystemLoad // 系统负载信息
	refreshSystemInfoTime time.Duration         // 刷新系统负载间隔, 默认5秒
}

// NewCronX 创建定时任务服务【任务添加后, 启动定时器后，任务默认暂停状态，需显式调用ResumeCron/ResumeCrons启动】
//   - 需优先调用函数设置任务表达式和任务逻辑 SetExprOrCmd
//
// func NewCronX(logx logx.Loggerx, systemInfo *gopsutilx.SystemLoad) *CronX {
func NewCronX(logx logx.Loggerx) *CronX {
	//c := &CronX{logx: logx, atc: atomic.Int32{}, systemInfo: systemInfo, refreshSystemInfo: time.Second * 5}
	c := &CronX{logx: logx, refreshSystemInfoTime: time.Second * 5, IsSystemInfo: false}
	c.atc.Store(cronPause)
	c.exprOrCmd = &syncX.Map[string, CronXCmdConfig]{}
	c.adminCron = &syncX.Map[string, CronXConfig]{}
	c.cron = cron.New(cron.WithSeconds())
	return c
}

func (r *CronX) Start() error {
	// 启动定时器时可没有没有任务，可以启动定时器后热添加任务
	//ok := r.exprOrCmd.IsEmpty()
	//if ok {
	//	r.logx.Error("【定时任务表达式或任务执行逻辑未配置】请先调用SetExprOrCmd，设置定时任务表达式和任务执行逻辑", logx.Error(fmt.Errorf("定时任务表达式或任务执行逻辑未配置")))
	//	return fmt.Errorf("【定时任务表达式或任务执行逻辑未配置】请先调用SetExprOrCmd，设置定时任务表达式和任务执行逻辑")
	//}

	err := r.addTask(r.cron) // 添加任务
	if err != nil {
		r.logx.Error("添加定时任务失败", logx.Error(err))
		return err
	}

	r.cron.Start() // 启动任务
	if r.IsSystemInfo && r.systemInfo != nil && r.refreshSystemInfoTime != 0 {
		go r.refreshSystemLoad() // 异步刷新系统负载，默认5秒刷新一次，用于控制当前系统健康是否适合进行定时任务执行
	}
	return nil
}

func (r *CronX) addTask(expr *cron.Cron) error {
	var err error
	r.exprOrCmd.Range(func(key string, value CronXCmdConfig) bool {
		r.mu.Lock()
		defer r.mu.Unlock()
		EntryID, err := expr.AddFunc(value.CronExpr, r.makeJobFunc(key, value))
		if err != nil {
			r.logx.Error("添加定时任务失败", logx.Error(err), logx.String("任务map keys", key), logx.String("cronName", value.CronName), logx.Int64("任务ID", value.CronId))
		}
		r.adminCron.Store(key, CronXConfig{EntryID: EntryID, cronStatus: 1}) // 存储cron的任务ID
		return true
	})
	return err
}

// makeJobFunc 抽离 job 函数生成逻辑，避免重复
func (r *CronX) makeJobFunc(key string, value CronXCmdConfig) cron.FuncJob {
	return func() {
		val, ok := r.adminCron.Load(key)
		if !ok {
			r.logx.Info("任务执行前查询任务状态失败", logx.String("任务map keys", key), logx.String("cronName", value.CronName), logx.Int64("任务ID", value.CronId))
		}
		switch val.cronStatus {
		case 0:
			r.logx.Info("任务开始执行", logx.String("任务map keys", key), logx.String("cronName", value.CronName), logx.Int64("任务ID", value.CronId))
			value.CronCmd()
			return
		case 1:
			r.logx.Info("任务已暂停", logx.String("任务map keys", key), logx.String("cronName", value.CronName), logx.Int64("任务ID", value.CronId))
			return
		case 2:
			r.logx.Info("任务不存在", logx.String("任务map keys", key), logx.String("cronName", value.CronName), logx.Int64("任务ID", value.CronId))
			return
		default:
			r.logx.Info("任务状态未知", logx.String("任务map keys", key), logx.String("cronName", value.CronName), logx.Int64("任务ID", value.CronId))
			return
		}
	}
}

// CronXCmdConfig 定时任务配置
//   - CronKeys: 任务map keys，一般使用【任务名+任务ID】组成，防止任务名重复，覆盖其他任务
//   - CronName: 任务名称
//   - CronId: 任务ID
//   - CronExpr: 定时任务执行表达式，cron表达式可参考https://help.aliyun.com/document_detail/133509.html
//   - CronCmd: 任务执行逻辑
type CronXCmdConfig struct {
	CronKeys string  // 定时任务存储map时的key，一般可以使用【任务名+任务ID】组成，防止任务名重复，覆盖其他任务
	CronName string  // 定时任务名称
	CronId   int64   // 定时任务ID
	CronExpr string  // 定时任务执行表达式，cron表达式可参考https://help.aliyun.com/document_detail/133509.html
	CronCmd  CronCmd // 任务执行逻辑
}

// CronXConfig 定时任务
//   - EntryID: 定时任务ID
//   - cronStatus: 任务状态，0: 默认值，表示任务正常执行；1: 表示任务被暂停；2: 表示任务不存在
type CronXConfig struct {
	EntryID    cron.EntryID // 定时任务ID
	cronStatus int          // 任务状态，0: 默认值，表示任务正常执行；1: 表示任务被暂停；2: 表示任务不存在
}

// SetExprOrCmd 设置任务执行表达式和任务执行逻辑
func (r *CronX) SetExprOrCmd(exprOrCmd ...CronXCmdConfig) {
	//r.cmd = cmd
	for _, v := range exprOrCmd {
		_, ok := r.exprOrCmd.Load(v.CronKeys) // 检查任务是否存在，防止任务名重复，覆盖其他任务
		//_, oke := r.expr.Load(v.CronName) // 检查任务是否存在，防止任务名重复，覆盖其他任务
		if ok {
			r.logx.Error("设置任务失败，任务名重复，已存在", logx.Error(fmt.Errorf("任务名重复，已存在")), logx.String("任务名", v.CronName), logx.Int64("任务ID", v.CronId))
			return
		}
		if v.CronKeys == "" || v.CronName == "" || v.CronId == 0 || v.CronExpr == "" || v.CronCmd == nil {
			r.logx.Error("设置任务失败，任务参数错误，参数不可为空", logx.Error(fmt.Errorf("任务参数错误")), logx.String("任务名", v.CronName), logx.Int64("任务ID", v.CronId))
			return
		}
		r.exprOrCmd.Store(v.CronKeys, CronXCmdConfig{
			CronName: v.CronName,
			CronId:   v.CronId,
			CronExpr: v.CronExpr,
			CronCmd:  v.CronCmd,
		})
	}
}

// PauseCrons 暂停所有任务
func (r *CronX) PauseCrons() {
	r.logx.Info("暂停任务")
	r.adminCron.Range(func(key string, value CronXConfig) bool {
		r.adminCron.Store(key, CronXConfig{EntryID: value.EntryID, cronStatus: 1})
		return true
	})
}

// PauseCron 暂停指定任务
func (r *CronX) PauseCron(keys string) error {
	r.logx.Info("暂停任务", logx.String("任务map keys", keys))
	val, ok := r.adminCron.Load(keys)
	if !ok {
		r.logx.Warn("任务暂停失败，任务不存在", logx.String("任务map keys", keys))
		return fmt.Errorf("任务暂停失败，任务不存在")
	}
	r.adminCron.Store(keys, CronXConfig{EntryID: val.EntryID, cronStatus: 1})
	return nil
}

// ResumeCrons 恢复所有任务
func (r *CronX) ResumeCrons() {
	r.logx.Info("恢复任务")
	r.adminCron.Range(func(key string, value CronXConfig) bool {
		r.adminCron.Store(key, CronXConfig{EntryID: value.EntryID, cronStatus: 0})
		return true
	})
}

// ResumeCron 恢复指定任务
func (r *CronX) ResumeCron(keys string) error {
	r.logx.Info("恢复任务", logx.String("任务map keys", keys))
	val, ok := r.adminCron.Load(keys)
	if !ok {
		r.logx.Warn("任务恢复失败，任务不存在", logx.String("任务map keys", keys))
		return fmt.Errorf("任务恢复失败，任务不存在")
	}
	r.adminCron.Store(keys, CronXConfig{EntryID: val.EntryID, cronStatus: 0})
	return nil
}

// DeleteCron 销毁删除定时任务服务
func (r *CronX) DeleteCron(keys string) bool {
	r.logx.Info("销毁定时任务服务", logx.String("任务map keys", keys))
	_, ok := r.adminCron.Load(keys)
	if !ok {
		r.logx.Warn("任务删除失败，任务不存在", logx.String("任务map keys", keys))
		return ok
	}
	r.exprOrCmd.Delete(keys)
	val, ok := r.adminCron.LoadAndDelete(keys)
	if !ok {
		r.logx.Warn("任务删除失败，任务不存在", logx.String("任务map keys", keys))
		return ok
	}
	r.cron.Remove(val.EntryID)
	return true
}

// AddCronTask 动态添加一个定时任务（支持 cron 已启动后调用）
func (r *CronX) AddCronTask(config CronXCmdConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if config.CronKeys == "" || config.CronName == "" || config.CronId == 0 || config.CronExpr == "" || config.CronCmd == nil {
		return fmt.Errorf("任务参数错误，不可为空")
	}

	// 检查是否已存在
	if _, ok := r.exprOrCmd.Load(config.CronKeys); ok {
		return fmt.Errorf("任务已存在，key: %s", config.CronKeys)
	}

	// 先保存配置
	r.exprOrCmd.Store(config.CronKeys, config)

	// 立即注册到 cron 调度器（即使已 Start）
	entryID, err := r.cron.AddFunc(config.CronExpr, func() {
		val, ok := r.adminCron.Load(config.CronKeys)
		if !ok {
			r.logx.Info("任务执行前查询任务状态失败", logx.String("任务map keys", config.CronKeys), logx.String("cronName", config.CronName), logx.Int64("任务ID", config.CronId))
			return
		}
		switch val.cronStatus {
		case 0:
			r.logx.Info("任务开始执行", logx.String("任务map keys", config.CronKeys), logx.String("cronName", config.CronName), logx.Int64("任务ID", config.CronId))
			config.CronCmd()
		case 1:
			r.logx.Info("任务已暂停", logx.String("任务map keys", config.CronKeys), logx.String("cronName", config.CronName), logx.Int64("任务ID", config.CronId))
		case 2:
			r.logx.Info("任务不存在", logx.String("任务map keys", config.CronKeys), logx.String("cronName", config.CronName), logx.Int64("任务ID", config.CronId))
		default:
			r.logx.Info("任务状态未知", logx.String("任务map keys", config.CronKeys), logx.String("cronName", config.CronName), logx.Int64("任务ID", config.CronId))
		}
	})
	if err != nil {
		r.exprOrCmd.Delete(config.CronKeys) // 回滚
		return fmt.Errorf("添加定时任务失败: %w", err)
	}

	// 保存管理信息
	r.adminCron.Store(config.CronKeys, CronXConfig{EntryID: entryID, cronStatus: 1})
	r.logx.Info("动态添加定时任务成功", logx.String("任务map keys", config.CronKeys))
	return nil
}

// RefreshSystemLoad 刷新系统负载
func (r *CronX) refreshSystemLoad() {
	ticker := time.NewTicker(r.refreshSystemInfoTime)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			r.logx.Info("刷新系统负载")
			sid, err := r.systemInfo.SystemLoad() // 0: 获取失败，1: 良好【cpu&men<%70】，2: 负载警告【%90>cpu&men>%70】，3: 负载过高且内存使用率过高【cpu/men>%90】
			if err != nil {
				r.logx.Error("刷新系统负载失败", logx.Error(err))
				return
			}
			switch sid {
			case uint(1), uint(2):
				//if r.atc == int32(1) { //	判断任务状态
				if r.atc.Load() == int32(cronSuccess) { //	判断任务状态
					r.logx.Info("系统负载恢复正常, 即将恢复定时任务", logx.String("系统负载", fmt.Sprintf("%d", sid)))
					r.ResumeCrons() // 恢复继续任务
				}
			case uint(0), uint(3):
				//if r.atc == int32(0) { //	判断任务状态
				if r.atc.Load() == int32(cronPause) { //	判断任务状态
					r.logx.Info("当前系统负载异常, 即将暂停定时任务", logx.String("系统负载", fmt.Sprintf("%d", sid)))
					r.PauseCrons() // 暂停任务
				}
			default:
				r.logx.Error("获取系统负载失败", logx.Uint("系统负载", sid), logx.Error(err))
			}
		}
	}
}

// SetSystemInfo 设置/启用刷新系统负载配置
//   - systemInfo: 系统负载对象
//   - isSystemInfo: 是否启用刷新系统负载配置
//   - refreshTime: 刷新系统负载间隔时间
func (r *CronX) SetSystemInfo(systemInfo *gopsutilx.SystemLoad, isSystemInfo bool, refreshTime time.Duration) {
	r.systemInfo = systemInfo
	r.IsSystemInfo = isSystemInfo
	r.refreshSystemInfoTime = refreshTime
}

// StopCron 停止定时器任务服务
func (r *CronX) StopCron() {
	ctx := r.cron.Stop() // 暂停定时器，不调度新任务执行了，正在执行的继续执行
	r.logx.Warn("正在停止定时任务调度器")
	<-ctx.Done() // 彻底停止定时器
	r.logx.Warn("定时任务调度器已停止")
}

const (
	// 定时任务状态正常
	cronSuccess = iota
	// 定时任务状态暂停
	cronPause
	// 定时任务不存在
	cronNotExist
)
