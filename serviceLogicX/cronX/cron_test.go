package cronX

import (
	"context"
	"fmt"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/logx/zerologx"
	"gitee.com/hgg_test/pkg_tool/v2/serviceLogicX/rankingListX/rankingServiceX"
	"gitee.com/hgg_test/pkg_tool/v2/serviceLogicX/rankingListX/rankingServiceX/buildGormX"
	"gitee.com/hgg_test/pkg_tool/v2/serviceLogicX/rankingListX/rankingServiceX/types"
	"gitee.com/hgg_test/pkg_tool/v2/systemLoad/gopsutilx"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"math/rand/v2"
	"os"
	"strconv"
	"testing"
	"time"
)

// 测试定时任务，且判断系统负载进行自动暂停/继续任务
func TestNewRankingServiceCron(t *testing.T) {
	r := NewCronX(InitLog(), initSyatemLoad())
	// 设置任务表达式和任务执行逻辑
	r.SetExprOrCmd(
		CronXCmdConfig{
			CronKeys: "任务1_1111222", // 任务map keys，一般使用【任务名+任务ID】组成，防止任务名重复，覆盖其他任务
			CronName: "任务1",
			CronId:   int64(1111222),
			CronExpr: "*/5 * * * * ?",
			CronCmd:  cmd,
		},
		CronXCmdConfig{
			CronKeys: "任务2_1111333",
			CronName: "任务2",
			CronId:   int64(1111333),
			CronExpr: "*/1 * * * * ?",
			CronCmd: func() {
				log.Println("任务2执行中...")
			},
		},
		CronXCmdConfig{
			CronKeys: "任务3_1111444",
			CronName: "任务3",
			CronId:   int64(1111444),
			CronExpr: "*/2 * * * * ?",
			CronCmd: func() {
				log.Println("任务3执行中...")
			},
		},
	)
	err := r.Start() // 启动定时任务
	assert.NoError(t, err)

	time.Sleep(3 * time.Second) // 模拟十秒后，暂定任务
	r.PauseCrons()              // 暂停所有任务
	r.DeleteCron("任务1_1111222") // 删除任务1
	r.DeleteCron("任务2_1111333") // 删除任务2
	time.Sleep(time.Second)     // 模拟一秒后，恢复任务3
	r.ResumeCron("任务3_1111444") // 恢复任务3

	r.AddCronTask(CronXCmdConfig{ // 任务启动后，再添加任务4
		CronKeys: "任务4_1111555",
		CronName: "任务4",
		CronId:   1111555,
		CronExpr: "*/1 * * * * ?",
		CronCmd: func() {
			log.Println("任务4执行中...")
		},
	})
	r.ResumeCron("任务4_1111555") // 恢复任务4

	time.Sleep(time.Minute) // 模拟堵塞1分钟

}

func initSyatemLoad() *gopsutilx.SystemLoad {
	return gopsutilx.NewSystemLoad()
}

// 定时任务执行逻辑
func cmd() {
	batchSize := 100 // 数据库获取每批数据源大小
	s := rankingServiceX.NewRankingServiceBatch(10, types.HotScoreProvider{}, InitLog())
	s.SetBatchSize(batchSize)
	s.SetSource(buildGormX.BuildDataSource[TestInteractive](
		context.Background(),
		InitDb(),
		func(db *gorm.DB) *gorm.DB {
			return db.Where("biz = ?", "test_biz")
		},
		func(interactive TestInteractive) types.HotScore {
			return interactive.dataOrigin()
		}, InitLog()),
	)

	// 获取 Top-10 列表【此处可接入自己业务逻辑，榜单后存入本地缓存、redis、数据库等】
	top10 := s.GetTopN() // 获取 Top-10 列表
	fmt.Println("Top 10 Users:")
	for i, u := range top10 {
		fmt.Printf("#%d | Biz_BizId: %s | Score: %.2f\n", i+1, u.Biz+u.BizID, u.Score)
	}
}

func systemLoad() (uint, error) {
	s := gopsutilx.NewSystemLoad()
	return s.SystemLoad()
}

// 定义业务结构体，数据源
type TestInteractive struct {
	Id int64 `gorm:"primaryKey, autoIncrement"` //主键

	// <bizid biz>，联合索引，bizId和id建立一个联合索引biz_type_id
	BizId int64  `gorm:"uniqueIndex:biz_type_id"`                   //业务id
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"` //业务类型

	ReadCnt int64 //阅读次数
	//LikeCnt    int64 //点赞次数
	//CollectCnt int64 //收藏次数
	Utime int64 //更新时间
	Ctime int64 //创建时间
}

func (u TestInteractive) dataOrigin() types.HotScore {
	return types.HotScore{
		Biz:   u.Biz,
		BizID: strconv.FormatInt(u.BizId, 10),
		Score: rand.Float64()*1000 + float64(u.ReadCnt), // 随机生成一个分数, 实际可使用其他得分算法，根据点赞、收藏计算得分
		//Score: float64(u.ReadCnt), // 随机生成一个分数, 实际可使用其他得分算法，根据点赞、收藏计算得分
		Title: u.Biz + strconv.FormatInt(u.BizId, 10),
	}
}

func InitLog() logx.Loggerx {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// Level日志级别【可以考虑作为参数传】，测试传zerolog.InfoLevel/NoLevel不打印
	// 模块化: Str("module", "userService模块")
	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().CallerWithSkipFrameCount(4).Timestamp().Logger()

	return zerologx.NewZeroLogger(&logger)
}

func InitDb() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(127.0.0.1:13306)/src_db?charset=utf8mb4&parseTime=True&loc=Local"))
	if err != nil {
		log.Println("数据库连接失败", err)
	}
	err = db.AutoMigrate(&TestInteractive{})
	if err != nil {
		return nil
	}
	return db
}
