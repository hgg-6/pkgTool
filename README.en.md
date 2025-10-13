# pkg_tool

This is feature-rich toolkit project containing various tool modules, suitable for different development needs.

## Features

- **Cache Count Service**: Provides counting services based on Redis and local cache.
- **Database Migration Tool**: Supports dual-write pools, migration schedulers, and data validation.
- **Message Queue**: Implements Kafka producers and consumers.
- **Configuration Management**: Provides configuration management interfaces based on Viper.
- **Type Conversion**: Offers multiple utility functions for type conversion.
- **Rate Limiting & Locking**: Implements sliding window rate limiting and Redis distributed locks.
- **Logging & Monitoring**: Supports multiple logging frameworks and Prometheus monitoring.
- **Web Middleware**: Provides middleware support for the Gin framework, including JWT authentication, rate limiting, and logging.

## Installation

Ensure you have a Go environment installed, then use the following command to get the project:

```bash
go get https://gitee.com/hgg_test/pkg_tool
```

## Usage Examples

### Cache Count Service

```go
redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

localCache := cacheLocalx.NewCacheLocalRistrettoStr[string, string](ristretto.NewCache[string, string]())

countService := cacheCountServicex.NewCount[string, string](redisClient, localCache)
```

### Database Migration

```go
srcDB, _ := gorm.Open(mysql.Open("user:pass@tcp(localhost:3306)/src_db"), &gorm.Config{})
dstDB, _ := gorm.Open(mysql.Open("user:pass@tcp(localhost:3306)/dst_db"), &gorm.Config{})

doubleWritePool := dbMovex.NewDoubleWritePool(srcDB, dstDB, logger, config...)
```

### Message Queue Producer

```go
config := sarama.NewConfig()
config.Producer.Return.Successes = true

producer, _ := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
messageProducer := saramaProducerx.NewSaramaProducerStr[sarama.SyncProducer](producer, config)
```

### Message Queue Consumer

```go
consumerGroup, _ := sarama.NewConsumerGroup([]string{"localhost:9092"}, "group_id", sarama.NewConfig())
consumer := saramaConsumerx.NewConsumerIn(consumerGroup, handler)
```

### Configuration Management

```go
viperConfig := viper.New()
viperConfig.SetDefault("mysql.dsn", "user:pass@tcp(localhost:3306)/dbname")

configService := viperx.NewViperConfigStr()
configService.InitViperLocal("config.yaml", DefaultConfig{})
```

### Type Conversion

```go
intValue, ok := toanyx.ToAny[int](someValue)
stringValue, ok := toanyx.ToAny[string](someValue)
```

### Rate Limiting

```go
redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

limiter := redis_slide_window.NewRedisSlideWindowKLimiter(redisClient, time.Minute, 100)
```

### Distributed Lock

```go
redisClients := []*redis.Client{
    redis.NewClient(&redis.Options{Addr: "localhost:6379"}),
}

lock := redsyncx.NewLockRedsync(redisClients, logger, redsyncx.Config{})
```

### Logging

```go
zapLogger, _ := zap.NewProduction()
logger := zaplogx.NewZapLogger(zapLogger)
```

### Web Middleware

```go
r := gin.Default()

jwtMiddleware := jwtx.NewJwtxMiddlewareGinx(redisClient, &jwtx.JwtxMiddlewareGinxConfig{})
r.Use(jwtMiddleware.VerifyToken)
```

## Contribution

Contributions and issue reports are welcome. Please read the [Contribution Guide](CONTRIBUTING.md) first.

## License

This project uses the MIT License. For details, please see the [LICENSE](LICENSE) file.