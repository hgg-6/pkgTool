package buildGormX

import (
	"context"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/serviceLogicX/rankingListX/rankingServiceX/types"
	"gorm.io/gorm"
)

// BuildDataSource 通用数据源构建器
//   - baseQuery: 查询时有可能会有status等字段需要过滤，自定义构建数据源的status等，
//     所以暴漏外部来控制数据源查询条件【构造Select/Where/Order，注意不要加分页！】，
//   - baseQuery: 一、是为了灵活构造查询。二、如果where无法命中索引，就不要加where了，性能影响较大
//   - mapper: 映射数据源结构体到分数结构体，返回分数结构体
func BuildDataSource[T any](
	ctx context.Context,
	db *gorm.DB,                       // 数据库连接
	baseQuery func(*gorm.DB) *gorm.DB, // 构造Select/Where/Order，注意不要加分页！
	mapper func(T) types.HotScore,     // 映射数据源结构体到业务结构体
	logger logx.Loggerx,               // 日志
) func(offset, limit int) ([]types.HotScore, error) {

	return func(offset, limit int) ([]types.HotScore, error) {
		query := db.Model(new(T)).WithContext(ctx)
		if baseQuery != nil {
			query = baseQuery(query)
		}
		// 内部统一加 Offset/Limit，确保分页正确
		query = query.Offset(offset).Limit(limit)

		var items []T
		if err := query.Find(&items).Error; err != nil {
			logger.Error("ranking: DB query failed, 计算榜单时查询数据库数据失败",
				logx.Int("offset", offset),
				logx.Int("limit", limit),
				logx.Error(err))
			return nil, err
		}

		result := make([]types.HotScore, len(items))
		for i, item := range items {
			result[i] = mapper(item)
		}
		return result, nil
	}
}
