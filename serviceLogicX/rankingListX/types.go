package rankingListX

type RankingTopN[T any] interface {
	GetTopN() []T
}
