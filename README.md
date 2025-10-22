# pkg_tool

这是一个功能丰富的工具包项目，包含了多种工具模块，适用于不同的开发需求。

## 功能特性

- **Web 中间件**：提供 Gin 框架的中间件支持，包括 JWT 认证、限流、日志记录等。
- **Rpc 中间件**：多种限流算法【滑动窗口、计数器、令牌桶等】、负载均衡算法、熔断拦截器、可观测性平台。
- **数据库迁移工具**：支持数据库的双写池、迁移调度器、数据校验等功能。
- **限流与锁**：实现滑动窗口限流和 Redis 分布式锁。
- **消息队列**：支持 Kafka 的生产者和消费者实现。
- **配置管理**：提供基于 Viper 的配置管理接口。
- **类型转换**：提供多种类型转换工具函数。
- **缓存计数服务**：提供基于 Redis 和本地缓存的计数服务。
- **日志与监控**：支持多种日志框架和 Prometheus 监控。


## 安装

确保你已经安装了 Go 环境，然后使用以下命令获取项目：

```bash
go get gitee.com/hgg_test/pkg_tool/v2@latest
```

## 使用示例

### 缓存计数服务

```go
redisClient := redis.NewClient(&redis.Options{
Addr: "localhost:6379",
})

localCache := cacheLocalx.NewCacheLocalRistrettoStr[string, string](ristretto.NewCache[string, string]())

countService := cacheCountServicex.NewCount[string, string](redisClient, localCache)
```

### 数据库迁移

```go
srcDB, _ := gorm.Open(mysql.Open("user:pass@tcp(localhost:3306)/src_db"), &gorm.Config{})
dstDB, _ := gorm.Open(mysql.Open("user:pass@tcp(localhost:3306)/dst_db"), &gorm.Config{})

doubleWritePool := dbMovex.NewDoubleWritePool(srcDB, dstDB, logger, config...)
```

### 消息队列生产者

```go
config := sarama.NewConfig()
config.Producer.Return.Successes = true

producer, _ := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
messageProducer := saramaProducerx.NewSaramaProducerStr[sarama.SyncProducer](producer, config)
```

### 消息队列消费者

```go
consumerGroup, _ := sarama.NewConsumerGroup([]string{"localhost:9092"}, "group_id", sarama.NewConfig())
consumer := saramaConsumerx.NewConsumerIn(consumerGroup, handler)
```

### 配置管理

```go
viperConfig := viper.New()
viperConfig.SetDefault("mysql.dsn", "user:pass@tcp(localhost:3306)/dbname")

configService := viperx.NewViperConfigStr()
configService.InitViperLocal("config.yaml", DefaultConfig{})
```

### 类型转换

```go
intValue, ok := toanyx.ToAny[int](someValue)
stringValue, ok := toanyx.ToAny[string](someValue)
```

### 限流

```go
redisClient := redis.NewClient(&redis.Options{
Addr: "localhost:6379",
})

limiter := redis_slide_window.NewRedisSlideWindowKLimiter(redisClient, time.Minute, 100)
```

### 分布式锁

```go
redisClients := []*redis.Client{
redis.NewClient(&redis.Options{Addr: "localhost:6379"}),
}

lock := redsyncx.NewLockRedsync(redisClients, logger, redsyncx.Config{})
```

### 日志记录

```go
zapLogger, _ := zap.NewProduction()
logger := zaplogx.NewZapLogger(zapLogger)
```

### Web 中间件

```go
r := gin.Default()

jwtMiddleware := jwtx.NewJwtxMiddlewareGinx(redisClient, &jwtx.JwtxMiddlewareGinxConfig{})
r.Use(jwtMiddleware.VerifyToken)
```

## 贡献

欢迎贡献代码和提出问题。请先阅读 [贡献指南](CONTRIBUTING.md)。

## 许可证

该项目使用 MIT 许可证。详情请查看 [LICENSE](LICENSE) 文件。