package fixer

import (
	"context"
	"errors"
	"github.com/basic-go-project-webook/webook/pkg/migrator"
	"github.com/basic-go-project-webook/webook/pkg/migrator/events"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OverrideFixer[T migrator.Entity] struct {
	base   *gorm.DB
	target *gorm.DB

	columns []string
}

func NewOverrideFixerWithColumns[T migrator.Entity](base *gorm.DB, target *gorm.DB, columns []string) *OverrideFixer[T] {
	return &OverrideFixer[T]{
		base:    base,
		target:  target,
		columns: columns,
	}
}

func NewOverrideFixer[T migrator.Entity](base *gorm.DB, target *gorm.DB) (*OverrideFixer[T], error) {
	rows, err := base.Model(new(T)).Order("id").Rows()
	if err != nil {
		return nil, err
	}
	column, err := rows.Columns()
	return &OverrideFixer[T]{
		base:    base,
		target:  target,
		columns: column,
	}, err
}

func (f *OverrideFixer[T]) Fix(ctx context.Context, id int64) error {
	var t T
	err := f.base.WithContext(ctx).Where("id = ?", id).First(&t).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return f.target.WithContext(ctx).Model(&t).Delete("id = ?", id).Error
	case err == nil:
		// upsert 语义
		return f.target.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns(f.columns),
		}).Create(&t).Error
	default:
		return err
	}

}

func (f *OverrideFixer[T]) FixV1(evt events.InconsistentEvent) error {
	switch evt.Type {
	case events.InconsistentEventTypeNEQ, events.InconsistentEventTypeTargetMissing:
		var t T
		err := f.base.Where("id =?", evt.ID).First(&t).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return f.target.Model(&t).Delete("id = ?", evt.ID).Error
		case err == nil:
			return f.target.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns(f.columns),
			}).Create(&t).Error
		default:
			return err
		}
	case events.InconsistentEventTypeBaseMissing:
		return f.target.Model(new(T)).Delete("id = ?", evt.ID).Error
	default:
		return nil
	}
}
