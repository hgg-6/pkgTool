package rankingServiceCronX

import (
	"fmt"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/logx/zerologx"
	"gitee.com/hgg_test/pkg_tool/v2/serviceLogicX/rankingListX/rankingServiceX"
	"gitee.com/hgg_test/pkg_tool/v2/systemLoad/gopsutilx"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"math/rand/v2"
	"os"
	"testing"
	"time"
)

// 测试定时任务，且判断系统负载进行暂停/继续
func TestNewRankingServiceCron(t *testing.T) {
	r := NewRankingServiceCron(initLog(), initSyatemLoad())
	r.SetExpr("*/5 * * * * ?") // 每5秒执行一次
	r.SetCmd(cmd)
	err := r.start() // 启动定时任务
	assert.NoError(t, err)

	time.Sleep(10 * time.Second) // 模拟十秒后，暂定任务
	go r.Pause()
	//time.Sleep(10 * time.Second) // 模拟十秒后，继续任务
	//r.Resume()
	time.Sleep(10 * time.Minute) // 模拟十分钟后，结束任务

	//ticker := time.NewTicker(time.Second * 5)
	//defer ticker.Stop()
	//for {
	//	select {
	//	case <-ticker.C: // 定时刷新系统负载
	//		sid, err := systemLoad()
	//		assert.NoError(t, err)
	//		switch sid {
	//		case uint(1), uint(2):
	//			if r.atc == int32(1) { //	判断任务状态
	//				r.Resume() // 恢复继续任务
	//			}
	//		case uint(0), uint(3):
	//			if r.atc == int32(0) { //	判断任务状态
	//				r.Pause() // 暂停任务
	//			}
	//		}
	//		//default:
	//		//	return
	//	}
	//}

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

func cmd() {
	offset := 0
	total := 100000
	batchSize := 100

	s := rankingServiceX.NewRankingServiceBatch(100, UserScoreProvider{}, initLog())
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
