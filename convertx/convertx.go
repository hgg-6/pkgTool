package convertx

// ToPtr 将任意值转换为指针
func ToPtr[T any](t T) *T {
	return &t
}

// ToPtrs 将任意值转换为指针数组
func ToPtrs[T any](ts ...T) []*T {
	var ptrs []*T
	for _, t := range ts {
		ptrs = append(ptrs, ToPtr(t))
	}
	return ptrs
}

// Deref 通用解引用
func Deref[T any](t *T) T {
	if t != nil {
		return *t
	}
	return *new(T)
}

// DerefOr 安全解引用指针，如果为nil时返回指定默认值or
func DerefOr[T any](ptr *T, or T) T {
	if ptr != nil {
		return *ptr
	}
	return or
}
