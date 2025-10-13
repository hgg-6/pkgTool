package events

type Entity interface {
	// ID 要求返回 ID
	ID() int64
	// CompareTo dst 必然也是 Entity，正常来说类型是一样的，怎么比较两张表数据
	CompareTo(dst Entity) bool

	Types() string
}
