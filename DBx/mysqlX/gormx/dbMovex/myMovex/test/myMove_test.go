package test

import (
	"bytes"
	"context"
	"encoding/json"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/mysqlX/gormx/dbMovex/myMovex/doubleWritePoolx"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/mysqlX/gormx/dbMovex/myMovex/events"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/mysqlX/gormx/dbMovex/myMovex/messageQueue/consumerx"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/mysqlX/gormx/dbMovex/myMovex/scheduler"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/mqX"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/mqX/kafkaX/saramaX/producerX"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/logx/zerologx"
	"gitee.com/hgg_test/pkg_tool/v2/webx/ginx"
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
)

type MoveTest struct {
	suite.Suite
	srcDb    *gorm.DB
	dstDb    *gorm.DB
	server   *gin.Engine
	doubleDb *doubleWritePoolx.DoubleWritePool // 双写连接池
	db       *gorm.DB                          // 双写使用的连接池
	//producer messageQueuex.ProducerIn[saramaX.SyncProducer]
	producer mqX.Producer
}

// 测试套件，测试用例测试前，会执行
func (m *MoveTest) SetupSuite() {
	var us events.TestUser
	srcDb := initDb("root:root@tcp(127.0.0.1:13306)/src_db")
	m.srcDb = srcDb
	dstDb := initDb("root:root@tcp(127.0.0.1:13306)/dst_db")
	m.dstDb = dstDb

	// 创建表结构
	err := srcDb.AutoMigrate(&us)
	assert.NoError(m.T(), err)
	err = dstDb.AutoMigrate(&us)
	assert.NoError(m.T(), err)

	m.clearTableTest()

	var dstUs events.TestUser
	var srcUs events.TestUser
	// 插入起始数据，src20条，dst10条
	for i := 1; i <= 10; i++ {
		dstUs = events.TestUser{
			Name:      "",
			Email:     "test" + strconv.Itoa(i) + "@test.com",
			UpdatedAt: 0,
			Ctime:     0,
			Utime:     0,
		}
		err = dstDb.Model(&dstUs).Create(&dstUs).Error
		assert.NoError(m.T(), err)
	}
	for i := 1; i <= 20; i++ {
		srcUs = events.TestUser{
			Name:      "",
			Email:     "test" + strconv.Itoa(i) + "@test.com",
			UpdatedAt: 0,
			Ctime:     0,
			Utime:     0,
		}
		err = srcDb.Model(&srcUs).Create(&srcUs).Error
		assert.NoError(m.T(), err)
	}

	server := gin.Default()
	m.server = server

	m.doubleDb = m.initDouble()

	m.db = m.initDbDouble()

	m.producer = newProducer()

	m.server = initGinServer()
}

// 测试套件，测试用例执行后，会执行
func (m *MoveTest) TearDownTest() {
	// 清空表数据
	m.clearTableTest()

	// 关闭生产者
	//m.producer.CloseProducer()
	m.producer.Close()
}

// 清空表数据，用于每次测试完清空，测试之间互不影响
func (m *MoveTest) clearTableTest() {
	err := m.srcDb.Exec("truncate table test_users").Error
	assert.NoError(m.T(), err)
	err = m.dstDb.Exec("truncate table test_users").Error
	assert.NoError(m.T(), err)
}

// ======================================
// ======================================
// 执行测试开始
// ======================================
// ======================================

// TestDoubleWritePool 测试双写连接池
func (m *MoveTest) TestDoubleWritePool() {
	t := m.T()

	testCases := []struct {
		name         string
		before       func(t *testing.T) // 测试前的准备
		beforeInsert int                // 测试时插入数据数量
		after        func(t *testing.T) // 测试后的验证

		wanPattern string // 双写模式
		wanInt     int    // src、dst两库表数据相差/不存在数据数量
	}{
		{
			name: "只源库读写",
			before: func(t *testing.T) {
				err := m.doubleDb.UpdatePattern(doubleWritePoolx.PatternSrcOnly)
				assert.NoError(t, err)
				var uc events.TestUser
				// 使用双写连接池插入数据
				for i := 0; i < 10; i++ {
					uc = events.TestUser{
						Name:      "",
						Email:     "testInsert" + strconv.Itoa(i+1) + "@testInsert.com",
						UpdatedAt: 0,
						Ctime:     0,
						Utime:     0,
					}
					err = m.db.Model(&uc).WithContext(context.Background()).Create(&uc).Error
					assert.NoError(t, err)
				}
			},
			beforeInsert: 10,
			after: func(t *testing.T) {
				var srcUc []events.TestUser
				err := m.srcDb.Model(&events.TestUser{}).WithContext(context.Background()).Where("email like ?", "testInsert%").Find(&srcUc).Error
				assert.NoError(t, err)
				assert.True(t, len(srcUc) == 10)

				var dstUc []events.TestUser
				err = m.dstDb.Model(&events.TestUser{}).WithContext(context.Background()).Where("email like ?", "testInsert%").Find(&dstUc).Error
				assert.NoError(t, err)
				assert.Equal(t, len(dstUc), 0)
				m.clearTableTest() // 清空数据
			},
			wanPattern: doubleWritePoolx.PatternSrcOnly,
			wanInt:     20,
		},
		{
			name: "双写，源库为主",
			before: func(t *testing.T) {
				err := m.doubleDb.UpdatePattern(doubleWritePoolx.PatternSrcFirst)
				assert.NoError(t, err)
				var uc events.TestUser
				// 使用双写连接池插入数据
				for i := 0; i < 10; i++ {
					uc = events.TestUser{
						Name:      "",
						Email:     "testInsert" + strconv.Itoa(i+1) + "@testInsert.com",
						UpdatedAt: 0,
						Ctime:     0,
						Utime:     0,
					}
					err = m.db.Model(&uc).WithContext(context.Background()).Create(&uc).Error
					assert.NoError(t, err)
				}
			},
			beforeInsert: 10,
			after: func(t *testing.T) {
				var srcUc []events.TestUser
				err := m.srcDb.Model(&events.TestUser{}).WithContext(context.Background()).Where("email like ?", "testInsert%").Find(&srcUc).Error
				assert.NoError(t, err)
				assert.True(t, len(srcUc) == 10)

				var dstUc []events.TestUser
				err = m.dstDb.Model(&events.TestUser{}).WithContext(context.Background()).Where("email like ?", "testInsert%").Find(&dstUc).Error
				assert.NoError(t, err)
				assert.Equal(t, 10, len(dstUc))
				m.clearTableTest() // 清空数据
			},
			wanPattern: doubleWritePoolx.PatternSrcFirst,
			wanInt:     10,
		},
		{
			name: "双写，目标库为主",
			before: func(t *testing.T) {
				err := m.doubleDb.UpdatePattern(doubleWritePoolx.PatternDstFirst)
				assert.NoError(t, err)
				var uc events.TestUser
				// 使用双写连接池插入数据
				for i := 0; i < 10; i++ {
					uc = events.TestUser{
						Name:      "",
						Email:     "testInsert" + strconv.Itoa(i+1) + "@testInsert.com",
						UpdatedAt: 0,
						Ctime:     0,
						Utime:     0,
					}
					err = m.db.Model(&uc).WithContext(context.Background()).Create(&uc).Error
					assert.NoError(t, err)
				}
			},
			beforeInsert: 10,
			after: func(t *testing.T) {
				var srcUc []events.TestUser
				err := m.srcDb.Model(&events.TestUser{}).WithContext(context.Background()).Where("email like ?", "testInsert%").Find(&srcUc).Error
				assert.NoError(t, err)
				assert.True(t, len(srcUc) == 10)

				var dstUc []events.TestUser
				err = m.dstDb.Model(&events.TestUser{}).WithContext(context.Background()).Where("email like ?", "testInsert%").Find(&dstUc).Error
				assert.NoError(t, err)
				assert.Equal(t, 10, len(dstUc))
				m.clearTableTest() // 清空数据
			},
			wanPattern: doubleWritePoolx.PatternDstFirst,
			wanInt:     10,
		},
		{
			name: "只目标库读写",
			before: func(t *testing.T) {
				err := m.doubleDb.UpdatePattern(doubleWritePoolx.PatternDstOnly)
				assert.NoError(t, err)
				var uc events.TestUser
				// 使用双写连接池插入数据
				for i := 0; i < 10; i++ {
					uc = events.TestUser{
						Name:      "",
						Email:     "testInsert" + strconv.Itoa(i+1) + "@testInsert.com",
						UpdatedAt: 0,
						Ctime:     0,
						Utime:     0,
					}
					err = m.db.Model(&uc).WithContext(context.Background()).Create(&uc).Error
					assert.NoError(t, err)
				}
			},
			beforeInsert: 10,
			after: func(t *testing.T) {
				var srcUc []events.TestUser
				err := m.srcDb.Model(&events.TestUser{}).WithContext(context.Background()).Where("email like ?", "testInsert%").Find(&srcUc).Error
				assert.NoError(t, err)
				assert.True(t, len(srcUc) == 0)

				var dstUc []events.TestUser
				err = m.dstDb.Model(&events.TestUser{}).WithContext(context.Background()).Where("email like ?", "testInsert%").Find(&dstUc).Error
				assert.NoError(t, err)
				assert.Equal(t, 10, len(dstUc))
				m.clearTableTest() // 清空表数据，用于每次测试完清空，测试之间互不影响
			},
			wanPattern: doubleWritePoolx.PatternDstOnly,
			wanInt:     0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			assert.Equal(t, tc.wanPattern, m.doubleDb.Pattern.Load())
		})
	}

}

// TestScheduler 测试调度器、校验逻辑(全量校验、增量校验)
func (m *MoveTest) TestScheduler() {
	t := m.T()
	sd := scheduler.NewScheduler[events.TestUser, sarama.SyncProducer](initLog(), m.srcDb, m.dstDb, m.doubleDb, m.producer)
	sd.RegisterRoutes(m.server.Group("/DbMove"))
	ginx.InitCounter(prometheus.CounterOpts{ // 启动Prometheus-CounterVec统计接口调用次数
		Namespace: "hgg",
		Subsystem: "hgg_XiaoWeiShu",
		Name:      "biz_code",
		Help:      "统计业务错误码",
	})
	ginx.NewLogMdlHandlerFunc(initLog())
	go m.server.Run("127.0.0.1:8080") // 启动服务

	testCases := []struct {
		name string

		reqBuild func(t *testing.T) *http.Request

		before func(t *testing.T) // 测试时插入数据数量
		after  func(t *testing.T) // 测试后的验证

		// 迁移状态
		state scheduler.MigrationState
		// 双写模式
		pattern string
		// 迁移统计信息
		stats scheduler.MigrationStats

		wantCode int
		wantBody string
	}{
		{
			name: "http调用切换--》只源库读写，成功",
			reqBuild: func(t *testing.T) *http.Request {
				req := httptest.NewRequest("POST", "http://127.0.0.1:8080/DbMove/src_only", nil)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			before: func(t *testing.T) {
				var uc events.TestUser
				// 使用双写连接池插入数据
				for i := 0; i < 10; i++ {
					uc = events.TestUser{
						Name:      "",
						Email:     "testInsert" + strconv.Itoa(i+1) + "@testInsert.com",
						UpdatedAt: 0,
						Ctime:     0,
						Utime:     0,
					}
					err := m.db.Model(&uc).WithContext(context.Background()).Create(&uc).Error
					assert.NoError(t, err)
				}
			},
			after: func(t *testing.T) {
				var srcUc []events.TestUser
				err := m.srcDb.Model(&events.TestUser{}).WithContext(context.Background()).Where("email like ?", "testInsert%").Find(&srcUc).Error
				assert.NoError(t, err)
				assert.True(t, len(srcUc) == 10)

				var dstUc []events.TestUser
				err = m.dstDb.Model(&events.TestUser{}).WithContext(context.Background()).Where("email like ?", "testInsert%").Find(&dstUc).Error
				assert.NoError(t, err)
				assert.Equal(t, len(dstUc), 0)
				m.clearTableTest() // 清空数据
			},
			state:    doubleWritePoolx.PatternSrcOnly,
			stats:    scheduler.MigrationStats{CurrentState: "src_only"},
			pattern:  doubleWritePoolx.PatternSrcOnly,
			wantCode: 200,
			wantBody: `{"code":0,"msg":"已切换到源库只写模式","data":null}`,
		},
		{
			name: "http调用切换--》双写src源库为主，成功",
			reqBuild: func(t *testing.T) *http.Request {
				req := httptest.NewRequest("POST", "http://127.0.0.1:8080/DbMove/src_first", nil)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			before: func(t *testing.T) {
				var uc events.TestUser
				// 使用双写连接池插入数据
				for i := 0; i < 10; i++ {
					uc = events.TestUser{
						Name:      "",
						Email:     "testInsert" + strconv.Itoa(i+1) + "@testInsert.com",
						UpdatedAt: 0,
						Ctime:     0,
						Utime:     0,
					}
					err := m.db.Model(&uc).WithContext(context.Background()).Create(&uc).Error
					assert.NoError(t, err)
				}
			},
			after: func(t *testing.T) {
				var srcUc []events.TestUser
				err := m.srcDb.Model(&events.TestUser{}).WithContext(context.Background()).Where("email like ?", "testInsert%").Find(&srcUc).Error
				assert.NoError(t, err)
				assert.True(t, len(srcUc) == 10)

				var dstUc []events.TestUser
				err = m.dstDb.Model(&events.TestUser{}).WithContext(context.Background()).Where("email like ?", "testInsert%").Find(&dstUc).Error
				assert.NoError(t, err)
				assert.Equal(t, 10, len(dstUc))
				m.clearTableTest() // 清空数据
			},
			state:    doubleWritePoolx.PatternSrcFirst,
			stats:    scheduler.MigrationStats{CurrentState: "src_first"},
			pattern:  doubleWritePoolx.PatternSrcFirst,
			wantCode: 200,
			wantBody: `{"code":0,"msg":"已切换到双写，源库优先模式","data":null}`,
		},
		{
			name: "http调用切换--》双写dst目标库为主，成功",
			reqBuild: func(t *testing.T) *http.Request {
				req := httptest.NewRequest("POST", "http://127.0.0.1:8080/DbMove/dst_first", nil)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			before: func(t *testing.T) {
				var uc events.TestUser
				// 使用双写连接池插入数据
				for i := 0; i < 10; i++ {
					uc = events.TestUser{
						Name:      "",
						Email:     "testInsert" + strconv.Itoa(i+1) + "@testInsert.com",
						UpdatedAt: 0,
						Ctime:     0,
						Utime:     0,
					}
					err := m.db.Model(&uc).WithContext(context.Background()).Create(&uc).Error
					assert.NoError(t, err)
				}
			},
			after: func(t *testing.T) {
				var srcUc []events.TestUser
				err := m.srcDb.Model(&events.TestUser{}).WithContext(context.Background()).Where("email like ?", "testInsert%").Find(&srcUc).Error
				assert.NoError(t, err)
				assert.True(t, len(srcUc) == 10)

				var dstUc []events.TestUser
				err = m.dstDb.Model(&events.TestUser{}).WithContext(context.Background()).Where("email like ?", "testInsert%").Find(&dstUc).Error
				assert.NoError(t, err)
				assert.Equal(t, 10, len(dstUc))
				m.clearTableTest() // 清空数据
			},
			state:    doubleWritePoolx.PatternDstFirst,
			stats:    scheduler.MigrationStats{CurrentState: "dst_first"},
			pattern:  doubleWritePoolx.PatternDstFirst,
			wantCode: 200,
			wantBody: `{"code":0,"msg":"已切换到双写，目标库优先模式","data":null}`,
		},
		{
			name: "http调用切换--》只目标库读写，成功",
			reqBuild: func(t *testing.T) *http.Request {
				req := httptest.NewRequest("POST", "http://127.0.0.1:8080/DbMove/dst_only", nil)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			before: func(t *testing.T) {
				var uc events.TestUser
				// 使用双写连接池插入数据
				for i := 0; i < 10; i++ {
					uc = events.TestUser{
						Name:      "",
						Email:     "testInsert" + strconv.Itoa(i+1) + "@testInsert.com",
						UpdatedAt: 0,
						Ctime:     0,
						Utime:     0,
					}
					err := m.db.Model(&uc).WithContext(context.Background()).Create(&uc).Error
					assert.NoError(t, err)
				}
			},
			after: func(t *testing.T) {
				var srcUc []events.TestUser
				err := m.srcDb.Model(&events.TestUser{}).WithContext(context.Background()).Where("email like ?", "testInsert%").Find(&srcUc).Error
				assert.NoError(t, err)
				assert.True(t, len(srcUc) == 0)

				var dstUc []events.TestUser
				err = m.dstDb.Model(&events.TestUser{}).WithContext(context.Background()).Where("email like ?", "testInsert%").Find(&dstUc).Error
				assert.NoError(t, err)
				assert.Equal(t, 10, len(dstUc))
				m.clearTableTest() // 清空数据
			},
			state:    doubleWritePoolx.PatternDstOnly,
			stats:    scheduler.MigrationStats{CurrentState: "dst_only"},
			pattern:  doubleWritePoolx.PatternDstOnly,
			wantCode: 200,
			wantBody: `{"code":0,"msg":"已切换到目标库只写模式","data":null}`,
		},
		{
			name: "http调用切换--》开启全量校验",
			reqBuild: func(t *testing.T) *http.Request {
				req := httptest.NewRequest("POST", "http://127.0.0.1:8080/DbMove/full/start", nil)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			before: func(t *testing.T) {
				var uc events.TestUser
				// 使用双写连接池插入数据
				for i := 0; i < 10; i++ {
					uc = events.TestUser{
						Name:      "",
						Email:     "testInsert" + strconv.Itoa(i+1) + "@testInsert.com",
						UpdatedAt: 0,
						Ctime:     0,
						Utime:     0,
					}
					err := m.db.Model(&uc).WithContext(context.Background()).Create(&uc).Error
					assert.NoError(t, err)
				}
			},
			after: func(t *testing.T) {
				m.clearTableTest() // 清空数据
			},
			state:    doubleWritePoolx.PatternDstOnly,
			stats:    scheduler.MigrationStats{CurrentState: "dst_only", FullValidationRuns: 1},
			pattern:  doubleWritePoolx.PatternDstOnly,
			wantCode: 200,
			wantBody: `{"code":0,"msg":"OK, 启动全量校验成功","data":null}`,
		},
		{
			name: "http调用切换--》开启增量校验",
			reqBuild: func(t *testing.T) *http.Request {
				body := scheduler.StartIncrRequest{
					Utime:    time.Now().UnixMilli(),
					Interval: 2,
				}
				bytBody, er := json.Marshal(body)
				assert.NoError(t, er)
				req := httptest.NewRequest("POST", "http://127.0.0.1:8080/DbMove/incr/start", bytes.NewReader(bytBody))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			before: func(t *testing.T) {
				var uc events.TestUser
				// 使用双写连接池插入数据
				for i := 0; i < 10; i++ {
					uc = events.TestUser{
						Name:      "",
						Email:     "testInsert" + strconv.Itoa(i+1) + "@testInsert.com",
						UpdatedAt: 0,
						Ctime:     0,
						Utime:     0,
					}
					err := m.db.Model(&uc).WithContext(context.Background()).Create(&uc).Error
					assert.NoError(t, err)
				}
			},
			after: func(t *testing.T) {
				m.clearTableTest() // 清空数据
			},
			state:    doubleWritePoolx.PatternDstOnly,
			stats:    scheduler.MigrationStats{CurrentState: "dst_only", FullValidationRuns: 1, IncrValidationRuns: 1},
			pattern:  doubleWritePoolx.PatternDstOnly,
			wantCode: 200,
			wantBody: `{"code":0,"msg":"OK, 启动增量校验成功","data":null}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			time.Sleep(time.Second)
			// 发送请求
			req := tc.reqBuild(t)
			resp := httptest.NewRecorder()
			m.server.ServeHTTP(resp, req)

			time.Sleep(time.Second)
			tc.before(t)
			defer tc.after(t)

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())
			assert.Equal(t, tc.pattern, m.doubleDb.Pattern.Load())
			assert.Equal(t, tc.stats, sd.Stats)
		})
	}

}

// 测试校验不一致数据，消费者消费不一致数据处理
func (m *MoveTest) TestConsumer() {
	var addr []string = []string{"localhost:9094"}
	cfg := sarama.NewConfig()
	cm := consumerx.NewConsumer(consumerx.ConsumerConf{Addr: addr, GroupId: "test_group", SaramaConf: cfg}, consumerx.DbConf{SrcDb: m.srcDb, DstDb: m.dstDb}, initLog())
	err := cm.InitConsumer(context.Background(), "dbMove")
	assert.NoError(m.T(), err)
}

// 测试套件的入口
func TestMoveTest(t *testing.T) {
	suite.Run(t, &MoveTest{})
}

// ======================================
// ======================================
// 执行测试结束
// ======================================
// ======================================
// ==============================
// ==============================
// ==============================
// 以下是初始化函数函数

func initDb(key string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(key), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}

func (m *MoveTest) initDbDouble() *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: m.doubleDb}))
	assert.NoError(m.T(), err)
	return db
}

func initLog() logx.Loggerx {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// Level日志级别【可以考虑作为参数传】，测试传zerolog.InfoLevel/NoLevel不打印
	// 模块化: Str("module", "userService模块")
	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel).With().CallerWithSkipFrameCount(4).Timestamp().Logger()
	return zerologx.NewZeroLogger(&logger)
}

// 初始化双写连接池
func (m *MoveTest) initDouble() *doubleWritePoolx.DoubleWritePool {
	//return NewDoubleWritePool(m.srcDb, m.dstDb, initLog(), DoubleWriteConfig{
	//	StrictMode:    false,
	//	EnableMetrics: false,
	//	RetryAttempts: 3,
	//})
	return doubleWritePoolx.NewDoubleWritePool(m.srcDb, m.dstDb, initLog())
}

// 初始化消息队列生产者
func newProducer() mqX.Producer {
	var addr []string = []string{"localhost:9094"}
	//cfg := saramaX.NewConfig()
	////========同步发送==========
	//cfg.Producer.Return.Successes = true
	//
	//syncPro, err := saramaX.NewSyncProducer(addr, cfg)
	//if err != nil {
	//	panic(err)
	//}
	//pro := saramaProducerx.NewSaramaProducerStr[saramaX.SyncProducer](syncPro, cfg)
	pro, err := producerX.NewKafkaProducer(addr, &producerX.ProducerConfig{Async: false})
	// CloseProducer 关闭生产者Producer，请在main函数最顶层defer住生产者的Producer.Close()，优雅关闭防止goroutine泄露
	if err == nil {
		return pro
	}
	return nil
}

func initGinServer() *gin.Engine {
	s := gin.Default()
	return s
}
