package cacheCountServiceX

import (
	"context"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/localCahceX/cacheLocalRistrettox"
	"github.com/dgraph-io/ristretto/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"testing"
	"time"
)

// 【此包可计算时间内的热榜，业务逻辑需处理定时写入数据库比如：10分钟一次的热榜数据，就算计算当日热榜，只需统计当日每十分钟的热榜然后取当日热榜】
// 【理论此方式也可计算总办但是会影响数据最终一致性，所以计算总时间榜单，需业务自行处理，数据库查出后可redis等计算排序存储】

const UserArticle string = "userArticle"

var Rank []RankItem

func TestCount(t *testing.T) {
	// 创建redis缓存
	redisCache := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	// 创建ristretto缓存
	ristrettoCache, err := ristretto.NewCache(&ristretto.Config[string, string]{
		NumCounters: 1e7,     // 按键跟踪次数为（10M）。
		MaxCost:     1 << 30, // 最大缓存成本（1GB）“位左移运算符”。
		BufferItems: 64,      // 每个Get缓冲区的键数。
	})
	assert.NoError(t, err)

	// 创建基于ristretto本地缓存
	localCache := cacheLocalRistrettox.NewCacheLocalRistrettoStr(ristrettoCache)
	// 缓存关闭，勿忘
	defer localCache.Close()

	// 创建计数服务
	countCache := NewCount[string, string](redisCache, localCache)

	// 初始化数据库
	db := inDB()

	// ==============================初始化完成==============================

	// 模拟文章Biz点赞增加计数，文章id【bizId】：1、2、3、4、5、6、7、8、9。。。。不同文章的不同点赞数量
	for i := 1; i <= 100; i++ {
		for w := 1; w <= i; w++ {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
			err = countCache.SetCnt(ctx, UserArticle, int64(i)).ResErr()
			assert.NoError(t, err)
			cancel()
		}
	}

	cal, _ := countCache.getCnt(context.Background(), UserArticle, 80)
	log.Println("当前id:80的点赞数为: ", cal)
	time.Sleep(time.Minute * 5)
	//ctx, cancelc := context.WithTimeout(context.Background(), time.Second)
	//defer cancelc()

	// 每10分钟执行一次入库
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// 时间到了，可以执行任务了
			// 任务限制执行时间5秒钟，5秒内入库完成，否则就超时
			// 【限制任务总时间的话，eg: 10秒运行一次任务，总计1分钟，运行6次，那么for外部创建1分钟的context.WithTimeout】
			ctx, cancelc := context.WithTimeout(context.Background(), time.Second*5)
			setDb(ctx, db, redisCache, countCache)
			cancelc()
		}
	}
}

// 每10分钟持久化热榜，十分钟写一次数据到数据库，redis缓存cnt数据为11分钟
func setDb(ctx context.Context, db *gorm.DB, rdb redis.Cmdable, cnt *Count[string, string]) {
	type TenMinuteRank struct {
		Biz          string
		TimeSlot     time.Time
		BizID        int64
		Score        int64
		RankPosition int64
	}

	// 获取当前10分钟时间片
	timeSlot := getCurrentTimeSlot()

	// 从Redis获取当前热榜
	rankKey, err := cnt.GetCntRank(ctx, UserArticle, GetCntType{Offset: 0, Limit: 10})
	if err != nil {
		log.Println("获取当前热榜失败: ", err)
	}
	log.Println("当前热榜: ", rankKey)

	// 批量写入数据库
	rankRecord := make([]TenMinuteRank, len(rankKey))
	for k, v := range rankKey {
		rankRecord[k].BizID = v.BizID
		rankRecord[k].Score = v.Score
		rankRecord[k].RankPosition = v.Rank
		rankRecord[k].TimeSlot = timeSlot
		rankRecord[k].Biz = UserArticle
	}
	// 分批插入，每批5条
	result := db.CreateInBatches(&rankRecord, 5)
	if result.Error != nil {
		log.Println("批量写入数据库失败: ", result.Error)
	}
}

// 获取当前10分钟时间片
func getCurrentTimeSlot() time.Time {
	now := time.Now()
	// 向下取整到10分钟
	minutes := (now.Minute() / 10) * 10
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), minutes, 0, 0, now.Location())
}

func inDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13306)/hgg"))
	if err != nil {
		panic(err)
	}
	return db
}
