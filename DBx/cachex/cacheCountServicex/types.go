package cacheCountServicex

import (
	"context"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/cachex/cacheLocalx"
)

// CntServiceIn 抽象计数服务接口【只用计数服务的话，此方法即可】
type CntServiceIn[K cacheLocalx.Key, V any] interface {
	// Key(biz string, bizId int64) string
	// RankKey(biz string) string

	SetCnt(ctx context.Context, biz string, bizId int64, num ...int64) *Count[K, V]
	DelCnt(ctx context.Context, biz string, bizId int64) error
	GetCnt(ctx context.Context, biz string, bizId int64, opt ...GetCntType) ([]RankItem, error)
}
