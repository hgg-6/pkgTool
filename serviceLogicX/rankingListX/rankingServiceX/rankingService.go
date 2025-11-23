package rankingServiceX

import (
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/serviceLogicX/rankingListX/rankingServiceX/types"
	"gitee.com/hgg_test/pkg_tool/v2/sliceX/queueX"
)

// RankingServiceBatch 泛型榜单服务
type RankingServiceBatch[T any] struct {
	batchSize int                                  // 没批的数据批量大小
	topN      int                                  // 列表长度，前多少排名
	source    func(offset, limit int) ([]T, error) // 批量数据源逻辑
	scoreProv types.ScoreProvider[T]               // 得分提取器
	l         logx.Loggerx
}

// NewRankingServiceBatch 创建泛型榜单服务
//   - topN 榜单长度，即统计榜单前xx的排名
//   - prov 获取得分的逻辑
func NewRankingServiceBatch[T any](topN int, prov types.ScoreProvider[T], log logx.Loggerx) *RankingServiceBatch[T] {
	const maxTopN = 10000 // 列表长度，前多少排名，不可大于10000，一万
	if topN <= 0 {
		topN = 100
	}
	if topN > maxTopN {
		log.Error("topN is too large, 榜单长度不可大于10000，请设置合理值", logx.Int("topN", topN))
		topN = maxTopN
	}
	return &RankingServiceBatch[T]{
		batchSize: 100,
		topN:      topN,
		scoreProv: prov,
		l:         log,
	}
}

// SetBatchSize 设置批量数据源的批量大小
func (r *RankingServiceBatch[T]) SetBatchSize(size int) {
	if size > 0 {
		r.batchSize = size
	}
}

// SetSource 设置批量数据源逻辑
func (r *RankingServiceBatch[T]) SetSource(source func(offset, limit int) ([]T, error)) {
	r.source = source
}

// GetTopN 返回按得分从高到低排序的 Top-N 列表
func (r *RankingServiceBatch[T]) GetTopN() []T {
	if r.source == nil {
		r.l.Error("source is nil，批量数据源未设置【server.SetSource】")
		return nil
	}

	pq := queueX.NewPriorityQueue(func(a, b T) bool {
		return r.scoreProv.Score(a) < r.scoreProv.Score(b)
	}, r.topN)

	offset := 0
	for {
		batch, err := r.source(offset, r.batchSize)
		if err != nil {
			r.l.Error("fetch batch error at offset, 在偏移处提取批次错误", logx.Int("offset", offset), logx.Error(err))
			break
		}
		if len(batch) == 0 {
			break
		}

		for _, item := range batch {
			if pq.Size() < r.topN {
				pq.Enqueue(item)
			} else if topItem, ok := pq.Peek(); ok {
				if r.scoreProv.Score(item) > r.scoreProv.Score(topItem) {
					pq.Dequeue()
					pq.Enqueue(item)
				}
			}
		}

		offset += len(batch)
		if len(batch) < r.batchSize {
			break
		}
	}

	result := make([]T, 0, pq.Size())
	for pq.Size() > 0 {
		if item, ok := pq.Dequeue(); ok {
			result = append(result, item)
		}
	}

	// 降序
	//sort.Slice(result, func(i, j int) bool {
	//	return r.scoreProv.Score(result[i]) > r.scoreProv.Score(result[j])
	//})
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result
}
