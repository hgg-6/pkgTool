package rankingServiceRdbZsetX

import (
	"context"
	"fmt"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/localCahceX"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/serviceLogicX/rankingListX/rankingServiceRdbZsetX/types"
	"github.com/redis/go-redis/v9"
	"sort"
	"sync"
	"time"
)

// BizRankingService 是具体业务榜单实例（如 article 榜）
type BizRankingService struct {
	parent   *RankingServiceZset
	bizType  string // e.g., "article"
	provider types.ScoreProvider
}

// RankingServiceZset 实时排行榜服务
type RankingServiceZset struct {
	shardCount int                                                // 分片数，如 16，默认10【区间10-256】小型系统（< 10万数据)10 ~ 32, 中型系统（10万~100万)64 ~ 128, 大型系统（> 100万）128 ~ 256
	redisCache redis.Cmdable                                      // Redis 客户端（或 Cluster）
	localCache localCahceX.CacheLocalIn[string, []types.HotScore] // 本地缓存
	//provider   types.ScoreProvider                                // 分数提供器
	logger logx.Loggerx

	// 用于控制后台 goroutine 生命周期
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
}

// NewRankingService 创建全局服务
//   - shardCount: redis的Zset分片数，如 16，默认10【区间10-256】小型系统（< 10万数据)10 ~ 32, 中型系统（10万~100万)64 ~ 128, 大型系统（> 100万）128 ~ 256
func NewRankingService(
	shardCount int,
	redisCache redis.Cmdable,
	localCache localCahceX.CacheLocalIn[string, []types.HotScore],
	logger logx.Loggerx,
) *RankingServiceZset {
	return &RankingServiceZset{
		shardCount: shardCount,
		//keyPrefix:  "hot_",
		redisCache: redisCache,
		localCache: localCache,
		//provider:   provider,
		cancel: func() {},
		logger: logger,
	}
}

// WithBizType 获取具体业务榜单服务
func (s *RankingServiceZset) WithBizType(bizType string, provider types.ScoreProvider) *BizRankingService {
	return &BizRankingService{
		parent:   s,
		bizType:  bizType,
		provider: provider,
	}
}

// Start 启动后台缓存刷新（建议显式调用）
func (s *RankingServiceZset) Start(refreshInterval time.Duration) {
	s.once.Do(func() {
		go s.backgroundRefresh(s.ctx, refreshInterval)
	})
}

// Stop 停止后台任务
func (s *RankingServiceZset) Stop() {
	s.cancel()
}

// backgroundRefresh 后台定时刷新缓存
func (s *RankingServiceZset) backgroundRefresh(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("background refresh stopped")
			return
		case <-ticker.C:
			s.backgroundRefresh(ctx, interval) // 预加载 Top 100
		}
	}
}

// GetTopN 获取 TopN 榜单（优先本地缓存）
func (b *BizRankingService) GetTopN(ctx context.Context, topN int) ([]types.HotScore, error) {
	cacheKey := b.buildCacheKey()

	// 尝试本地缓存
	if items, ok := b.parent.localCache.Get(cacheKey); ok == nil && len(items) >= topN {
		return items[:topN], nil
	}

	// 从 Redis ZSET 拉取 BizID + Score
	bizScores, err := b.fetchTopNBizIDs(ctx, topN)
	if err != nil {
		return nil, err
	}

	// 批量补全元数据（Title, Biz 等）
	fullItems, err := b.enrichHotScores(ctx, bizScores)
	if err != nil {
		b.parent.logger.Warn("enrichHotScores failed, return scores only", logx.Error(err))
		// 即使失败，也返回基础分数（Biz/Title 为空）
		fullItems = bizScores // 降级
	}

	// 异步回写本地缓存
	go b.parent.localCache.Set(cacheKey, fullItems, 15*time.Second, 1)

	if len(fullItems) > topN {
		fullItems = fullItems[:topN]
	}
	return fullItems, nil
}

func (b *BizRankingService) fetchTopNBizIDs(ctx context.Context, topN int) ([]types.HotScore, error) {
	type result struct {
		items []types.HotScore
		err   error
	}
	ch := make(chan result, b.parent.shardCount)

	for i := 0; i < b.parent.shardCount; i++ {
		go func(shard int) {
			key := b.buildZSetKey(shard)
			zs, err := b.parent.redisCache.ZRevRangeWithScores(ctx, key, 0, int64(topN-1)).Result()
			if err != nil {
				ch <- result{err: err}
				return
			}

			var items []types.HotScore
			for _, z := range zs {
				if bizID, ok := z.Member.(string); ok {
					items = append(items, types.HotScore{
						Biz:   b.bizType, // 这里可以填 bizType！
						BizID: bizID,
						Score: z.Score,
						// Title 留空，后续补
					})
				}
			}
			ch <- result{items: items}
		}(i)
	}

	var all []types.HotScore
	for i := 0; i < b.parent.shardCount; i++ {
		r := <-ch
		if r.err != nil {
			return nil, r.err
		}
		all = append(all, r.items...)
	}

	// 全局排序
	sort.Slice(all, func(i, j int) bool {
		return b.provider.Score(all[i]) > b.provider.Score(all[j])
	})

	if len(all) > topN {
		all = all[:topN]
	}
	return all, nil
}

func (b *BizRankingService) enrichHotScores(ctx context.Context, items []types.HotScore) ([]types.HotScore, error) {
	if len(items) == 0 {
		return items, nil
	}

	// 1. 收集所有 bizID
	keys := make([]string, len(items))
	for i, item := range items {
		keys[i] = b.buildMetaKey(item.BizID)
	}

	// 2. Pipeline 批量 HGETALL
	pipe := b.parent.redisCache.Pipeline()
	var cmds []*redis.MapStringStringCmd
	for _, key := range keys {
		cmds = append(cmds, pipe.HGetAll(ctx, key))
	}
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	// 3. 填充 Title（假设 meta 中有 "title" 字段）
	for i, cmd := range cmds {
		meta, _ := cmd.Result() // 即使 key 不存在，meta 也是空 map
		if title, ok := meta["title"]; ok {
			items[i].Title = title
		}
		// 可扩展：cover, author, etc.
	}

	return items, nil
}

func (b *BizRankingService) buildZSetKey(shard int) string {
	return fmt.Sprintf("hot_%s_%d", b.bizType, shard)
}

func (b *BizRankingService) buildMetaKey(bizID string) string {
	return fmt.Sprintf("hot_meta:%s:%s", b.bizType, bizID)
}

func (b *BizRankingService) buildCacheKey() string {
	return fmt.Sprintf("hot_%s_topttt", b.bizType)
}

func (b *BizRankingService) getShard(bizID string) int {
	hash := fnv1a32(bizID)
	return int(hash % uint32(b.parent.shardCount))
}

func (b *BizRankingService) StartRefresh(interval time.Duration) {
	go b.backgroundRefresh(context.Background(), interval)
}

func (b *BizRankingService) backgroundRefresh(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			items, err := b.fetchTopNBizIDs(ctx, 100)
			if err != nil {
				continue
			}
			full, _ := b.enrichHotScores(ctx, items)
			_ = b.parent.localCache.Set(b.buildCacheKey(), full, 15*time.Second, 1)
		}
	}
}

// IncrScore 更新某个 BizID 的分数，并可选更新元数据
//   - bizID: 业务 ID
//   - delta: 分数增量 【正数：增加分数，负数：减少分数，零值无变化】eg:分享	+2.0, 评论+0.5, 点踩/举报-1.0
//   - meta: 元数据，可选，如：title, cover, author, etc.
func (b *BizRankingService) IncrScore(ctx context.Context, bizID string, delta float64, meta map[string]string) error {
	shard := b.getShard(bizID)
	zsetKey := b.buildZSetKey(shard)

	// 1. 更新 ZSET 分数
	if _, err := b.parent.redisCache.ZIncrBy(ctx, zsetKey, delta, bizID).Result(); err != nil {
		b.parent.logger.Error("ZIncrBy failed", logx.String("biz_type", b.bizType), logx.String("biz_id", bizID), logx.Error(err))
		return err
	}

	// 2. 如果提供了 meta，更新 Hash（非覆盖，用 HMSET 或 HSET 多字段）
	if len(meta) > 0 {
		metaKey := b.buildMetaKey(bizID)
		// 转换 map[string]string -> []interface{}
		args := make([]interface{}, 0, len(meta)*2)
		for k, v := range meta {
			args = append(args, k, v)
		}
		if err := b.parent.redisCache.HMSet(ctx, metaKey, args).Err(); err != nil {
			b.parent.logger.Warn("HMSet meta failed (non-fatal)", logx.String("meta_key", metaKey), logx.Error(err))
			// 元数据失败不影响核心分数更新
		}
	}
	return nil
}

// buildZSetKey 构建 Redis ZSET key
//func (s *RankingServiceZset) buildZSetKey(shard int) string {
//	return fmt.Sprintf("%s%d", s.keyPrefix, shard)
//}

// getShard 根据 BizID 哈希分片
func (s *RankingServiceZset) getShard(bizID string) int {
	hash := fnv1a32(bizID)
	return int(hash % uint32(s.shardCount))
}

// FNV-1a 32位哈希
func fnv1a32(data string) uint32 {
	const (
		offset32 = 2166136261
		prime32  = 16777619
	)
	hash := uint32(offset32)
	for _, c := range data {
		hash ^= uint32(c)
		hash *= prime32
	}
	return hash
}
