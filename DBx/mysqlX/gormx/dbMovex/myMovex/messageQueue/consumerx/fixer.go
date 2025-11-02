package consumerx

import (
	"context"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/mysqlX/gormx/dbMovex/myMovex"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OverrideFixer[T myMovex.Entity] struct {
	base   *gorm.DB
	target *gorm.DB

	columns []string
}

func NewOverrideFixer[T myMovex.Entity](base *gorm.DB, target *gorm.DB) (*OverrideFixer[T], error) {
	rows, err := base.Model(new(T)).Order("id").Rows()
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	return &OverrideFixer[T]{base: base, target: target, columns: columns}, err
}
func (f *OverrideFixer[T]) Fix(ctx context.Context, id int64) error {
	// 最最粗暴的
	var t T
	err := f.base.WithContext(ctx).Where("id=?", id).First(&t).Error
	switch err {
	case gorm.ErrRecordNotFound:
		return f.target.WithContext(ctx).Model(&t).Delete("id = ?", id).Error
	case nil:
		// upsert
		return f.target.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns(f.columns),
		}).Create(&t).Error
	default:
		return err
	}
}
