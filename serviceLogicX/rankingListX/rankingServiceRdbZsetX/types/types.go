package types

// ScoreProvider 分数提供器（用于排序）
type ScoreProvider interface {
	Score(item HotScore) float64
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
