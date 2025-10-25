package cronX

import (
	"fmt"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/logx/zerologx"
	"gitee.com/hgg_test/pkg_tool/v2/serviceLogicX/rankingListX/rankingServiceX"
	"gitee.com/hgg_test/pkg_tool/v2/systemLoad/gopsutilx"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand/v2"
	"os"
	"testing"
	"time"
)

// 测试定时任务，且判断系统负载进行自动暂停/继续任务
func TestNewRankingServiceCron(t *testing.T) {
	r := NewCronX(initLog(), initSyatemLoad())
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

	time.Sleep(3 * time.Second)   // 模拟十秒后，暂定任务
	r.PauseCrons()                // 暂停所有任务
	r.DeleteCron("任务1_1111222") // 删除任务1
	r.DeleteCron("任务2_1111333") // 删除任务2
	time.Sleep(time.Second)       // 模拟一秒后，恢复任务3
	r.ResumeCron("任务3_1111444") // 恢复任务3

	r.AddCronTask(CronXCmdConfig{ // 添加任务4
		CronKeys: "任务4_1111555",
		CronName: "任务4",
		CronId:   1111555,
		CronExpr: "*/1 * * * * ?",
		CronCmd: func() {
			log.Println("任务4执行中...")
		},
	})

	time.Sleep(time.Minute) // 模拟堵塞1分钟

}

func initLog() logx.Loggerx {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// Level日志级别【可以考虑作为参数传】，测试传zerolog.InfoLevel/NoLevel不打印
	// 模块化: Str("module", "userService模块")
	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().CallerWithSkipFrameCount(4).Str("module", "userService模块").Timestamp().Logger()

	return zerologx.NewZeroLogger(&logger)
}

func initSyatemLoad() *gopsutilx.SystemLoad {
	return gopsutilx.NewSystemLoad()
}

// 定时任务执行逻辑
func cmd() {
	offset := 0
	total := 100000
	batchSize := 100

	s := rankingServiceX.NewRankingServiceBatch(10, UserScoreProvider{}, initLog())
	s.SetBatchSize(batchSize)
	s.SetSource(func(batchSize int) ([]UserScore, bool) {
		var batch []UserScore
		for i := 0; i < total/100; i++ { // 模拟分批次获取数据
			if offset >= total {
				// 数据已经获取完毕
				return nil, false
			}
			end := offset + batchSize
			if end > total {
				end = total
			}
			batch = make([]UserScore, end-offset) // 创建一个切片,用来存储数据
			for k, _ := range batch {
				batch[k] = UserScore{
					UserID: fmt.Sprintf("user_%d", offset+k+1), // 生成用户ID
					Score:  rand.Float64() * 10000,             // 随机生成一个分数
				}
			}
			offset = end
		}
		return batch, offset < total
	})

	// 获取 Top-10 列表【此处可接入自己业务逻辑，榜单后存入本地缓存、redis、数据库等】
	top10 := s.GetTopN() // 获取 Top-10 列表
	fmt.Println("Top 10 Users:")
	for i, u := range top10 {
		fmt.Printf("#%d | UserID: %s | Score: %.2f\n", i+1, u.UserID, u.Score)
	}
}

// 定义业务结构体
type UserScore struct {
	UserID string
	Score  float64
}

// 实现 ScoreProvider 接口
type UserScoreProvider struct{}

func (p UserScoreProvider) Score(item UserScore) float64 {
	return item.Score
}

func systemLoad() (uint, error) {
	s := gopsutilx.NewSystemLoad()
	return s.SystemLoad()
}
