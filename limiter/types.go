package limiter

import "context"

type Limiter interface {
	// Limit 限制、限流，是否触发限流。
	// 返回True就是触发限流。
	Limit(ctx context.Context, key string) (bool, error)
}
