package rankingServiceX

import (
	"context"
	"fmt"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/logx/zerologx"
	"gitee.com/hgg_test/pkg_tool/v2/serviceLogicX/rankingListX/rankingServiceX/buildGormX"
	"gitee.com/hgg_test/pkg_tool/v2/serviceLogicX/rankingListX/rankingServiceX/types"
	"github.com/rs/zerolog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"os"
	"strconv"
	"testing"
)

func TestNewRankingServiceBatch(t *testing.T) {
	//offset := 0
	//total := 100000
	batchSize := 100

	// 创建一个排行榜服务
	s := NewRankingServiceBatch(10, types.HotScoreProvider{}, InitLog())
	// 设置批量数据源的批量大小
	s.SetBatchSize(batchSize)

	db := InitDb()
	// 构造数据库查询数据
	baseQuery := func(db *gorm.DB) *gorm.DB { // 构造数据库查询数据，Select/Where/Order，注意不要加分页offset/limit！
		// 查询时有可能会有status等字段需要过滤，自定义构建数据源的status等，所以暴漏外部来控制数据源查询条件
		return db.Select("id", "biz_id", "biz", "read_cnt").Where("biz = ?", "test_biz")
	}
	BuildDataSource := buildGormX.BuildDataSource[TestInteractive](
		context.Background(),
		db,
		baseQuery,
		func(interactive TestInteractive) types.HotScore {
			return interactive.dataOrigin()
		},
		InitLog(),
	)

	// 设置数据源
	s.SetSource(BuildDataSource)

	//	s.SetSource(func(offset, limit int) ([]UserScore, error) {
	//		var batch []UserScore
	//		if offset >= total {
	//			// 数据已经获取完毕
	//			return batch, nil
	//		}
	//		end := offset + batchSize
	//		if end > total {
	//			end = total
	//		}
	//		if offset >= total {
	//			return batch, nil
	//		}
	//		batch = make([]UserScore, end-offset) // 创建一个切片,用来存储数据
	//		for k, _ := range batch {
	//			batch[k] = UserScore{
	//				Biz:   "test_biz",
	//				BizID: fmt.Sprintf("id_%d", offset+k+1), // 生成用户ID
	//				Score: rand.Float64() * 10000,           // 随机生成一个分数
	//			}
	//		}
	//		offset = end
	//		return batch, nil
	//	})

	//	// 获取 Top-10 列表【此处可接入自己业务逻辑，榜单后存入本地缓存、redis、数据库等】
	top10 := s.GetTopN() // 获取 Top-10 列表
	fmt.Println("Top 10 Users:")
	for i, u := range top10 {
		fmt.Printf("#%d | UserID: %s | Score: %.2f\n", i+1, u.BizID, u.Score)
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
