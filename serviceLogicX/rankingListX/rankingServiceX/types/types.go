package types

// ScoreProvider 定义如何从任意类型 T 中提取得分
type ScoreProvider[T any] interface {
	Score(item T) float64
}

type HotScore struct {
	Biz   string
	BizID string
	Score float64
	Title string
}

type HotScoreProvider struct{}

func (p HotScoreProvider) Score(item HotScore) float64 {
	return item.Score
}
