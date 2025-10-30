package producerX

import "time"

type ProducerConfig struct {
	// BatchSize 最大批量大小（达到即发送）
	BatchSize int

	// BatchTimeout 超时自动 flush（从第一条消息进入缓冲区开始计时）
	// 0 表示禁用超时，仅按数量触发
	BatchTimeout time.Duration

	// Async 是否使用异步生产者（推荐 true）
	Async bool
}

func DefaultProducerConfig() *ProducerConfig {
	return &ProducerConfig{
		BatchSize:    100,
		BatchTimeout: 5 * time.Second,
		Async:        true, // 默认异步
	}
}

func (c *ProducerConfig) Validate() {
	if c.BatchSize <= 0 {
		c.BatchSize = 100
	}
	if c.BatchTimeout < 0 {
		c.BatchTimeout = 5 * time.Second
	}
}
