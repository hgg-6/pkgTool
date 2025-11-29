package consumerX

import "time"

type ConsumerConfig struct {
	// BatchSize 批量消费的最大消息数（仅 BatchConsumerHandler 生效）
	// 默认：100
	BatchSize int

	// BatchTimeout 批量消费的超时时间（从第一条消息进入缓冲区开始计时）
	// 默认：5秒
	// 若为 0，则禁用超时，仅按数量触发
	BatchTimeout time.Duration
}

func DefaultConsumerConfig() *ConsumerConfig {
	return &ConsumerConfig{
		BatchSize:    100,
		BatchTimeout: 5 * time.Second,
	}
}

func (c *ConsumerConfig) Validate() {
	if c.BatchSize <= 0 {
		c.BatchSize = 100
	}
	if c.BatchTimeout < 0 {
		c.BatchTimeout = 5 * time.Second
	}
}
