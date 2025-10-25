package rankingServiceX

import (
	"fmt"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/logx/zerologx"
	"github.com/rs/zerolog"
	"math/rand"
	"os"
	"testing"
)

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

func TestNewRankingServiceBatch(t *testing.T) {
	offset := 0
	total := 100000
	batchSize := 100

	s := NewRankingServiceBatch(100, UserScoreProvider{}, initLog())
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

func initLog() logx.Loggerx {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// Level日志级别【可以考虑作为参数传】，测试传zerolog.InfoLevel/NoLevel不打印
	// 模块化: Str("module", "userService模块")
	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().CallerWithSkipFrameCount(4).Str("module", "userService模块").Timestamp().Logger()

	return zerologx.NewZeroLogger(&logger)
}
