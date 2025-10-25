package rankingServiceX

import (
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/slicex/queueX"
	"sort"
)

// ScoreProvider 定义如何从任意类型 T 中提取得分
type ScoreProvider[T any] interface {
	Score(item T) float64
}

// RankingServiceBatch 泛型榜单服务
type RankingServiceBatch[T any] struct {
	batchSize int                             // 没批的数据批量大小
	topN      int                             // 列表长度，前多少排名
	source    func(batchSize int) ([]T, bool) // 批量数据源逻辑
	scoreProv ScoreProvider[T]                // 得分提取器
	logx      logx.Loggerx
}

// NewRankingServiceBatch 创建泛型榜单服务
//   - topN 榜单长度，即统计榜单前xx的排名
//   - prov 获取得分的逻辑
func NewRankingServiceBatch[T any](topN int, prov ScoreProvider[T], logx logx.Loggerx) *RankingServiceBatch[T] {
	if topN <= 0 {
		topN = 100
	}
	return &RankingServiceBatch[T]{
		batchSize: 100,
		topN:      topN,
		scoreProv: prov,
	}
}

func (r *RankingServiceBatch[T]) SetBatchSize(size int) {
	if size > 0 {
		r.batchSize = size
	}
}

func (r *RankingServiceBatch[T]) SetSource(source func(batchSize int) ([]T, bool)) {
	r.source = source
}

// GetTopN 返回按得分从高到低排序的 Top-N 列表
func (r *RankingServiceBatch[T]) GetTopN() []T {
	if r.source == nil {
		var empty []T
		r.logx.Error("source is nil，批量数据源未设置【server.SetSource】")
		return empty
	}

	// 注意：queueX.NewPriorityQueue 应该也支持泛型比较函数
	pq := queueX.NewPriorityQueue(func(a, b T) bool {
		return r.scoreProv.Score(a) < r.scoreProv.Score(b) // 小顶堆
	}, r.topN)

	for {
		batch, hasMore := r.source(r.batchSize)
		if len(batch) == 0 && !hasMore {
			break
		}

		for _, item := range batch {
			if pq.Size() < r.topN {
				pq.Enqueue(item)
			} else {
				// 查看堆顶（最小值）
				if topItem, ok := pq.Peek(); ok {
					if r.scoreProv.Score(item) > r.scoreProv.Score(topItem) {
						pq.Dequeue() // 弹出堆顶
						pq.Enqueue(item)
					}
				}
			}
		}

		if !hasMore {
			break
		}
	}

	// 提取结果
	result := make([]T, 0, pq.Size())
	for pq.Size() > 0 {
		if item, ok := pq.Dequeue(); ok {
			result = append(result, item)
		} else {
			break
		}
	}

	// 降序排序：得分从高到低
	sort.Slice(result, func(i, j int) bool {
		return r.scoreProv.Score(result[i]) > r.scoreProv.Score(result[j])
	})

	return result
}
