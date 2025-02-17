package connpool

import (
	"context"
	"database/sql"
	"errors"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	PatternSrcOnly  = "src_only"
	PatternDstOnly  = "dst_only"
	PatternSrcFirst = "src_first"
	PatternDstFirst = "dst_first"
)

var errUnknownPattern = errors.New("未知的双写模式")

type DoubleWritePool struct {
	src     gorm.ConnPool
	dst     gorm.ConnPool
	pattern *atomic.String
}

// BeginTx 实现TxBeginner interface
func (d *DoubleWritePool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWriteTx{src: src, pattern: atomic.NewString(pattern)}, err
	case PatternDstOnly:
		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWriteTx{dst: dst, pattern: atomic.NewString(pattern)}, err
	case PatternSrcFirst:
		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			zap.L().Error("双写目标表事务启动失败", zap.Error(err))
		}
		return &DoubleWriteTx{src: src, dst: dst, pattern: atomic.NewString(pattern)}, err
	case PatternDstFirst:
		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			zap.L().Error("双写源表事务启动失败", zap.Error(err))
		}
		return &DoubleWriteTx{src: src, dst: dst, pattern: atomic.NewString(pattern)}, err
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWritePool) UpdatePattern(pattern string) error {
	switch pattern {
	case PatternSrcOnly, PatternDstOnly, PatternSrcFirst, PatternDstFirst:
		d.pattern.Store(pattern)
		return nil
	default:
		return errUnknownPattern
	}
}

// PrepareContext 预编译的，不支持
func (d *DoubleWritePool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	panic("双写模式不支持PrepareContext")
}

func (d *DoubleWritePool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err == nil {
			_, err1 := d.dst.ExecContext(ctx, query, args...)
			if err1 != nil {
				zap.L().Error("双写 write 写入dst失败", zap.Error(err1), zap.String("query", query), zap.Any("args", args))
			}
		}
		return res, err
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err == nil {
			_, err1 := d.src.ExecContext(ctx, query, args...)
			if err1 != nil {
				zap.L().Error("双写 write 写入src失败", zap.Error(err1), zap.String("query", query), zap.Any("args", args))
			}
		}
		return res, err
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWritePool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWritePool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		// 这写没有带上错误信息
		// return &sql.Row{}
		panic(errUnknownPattern)
	}
}

func NewDoubleWritePool(src *gorm.DB, dst *gorm.DB) *DoubleWritePool {
	return &DoubleWritePool{
		src:     src.ConnPool,
		dst:     dst.ConnPool,
		pattern: atomic.NewString(PatternSrcOnly),
	}
}

type DoubleWriteTx struct {
	src     *sql.Tx
	dst     *sql.Tx
	pattern *atomic.String
}

// Commit 实现 TxCommitter interface
func (d *DoubleWriteTx) Commit() error {
	switch d.pattern.Load() {
	case PatternSrcOnly:
		return d.src.Commit()
	case PatternDstOnly:
		return d.dst.Commit()
	case PatternSrcFirst:
		err := d.src.Commit()
		if err != nil {
			return err
		}
		if d.dst != nil {
			err = d.dst.Commit()
			if err != nil {
				zap.L().Error("目标表提交事物失败")
			}
		}
		return nil
	case PatternDstFirst:
		err := d.dst.Commit()
		if err != nil {
			return err
		}
		if d.src != nil {
			err = d.src.Commit()
			if err != nil {
				zap.L().Error("源表提交事务失败")
			}
		}
		return nil
	default:
		return errUnknownPattern
	}
}

// Rollback 实现 TxCommitter interface
func (d *DoubleWriteTx) Rollback() error {
	switch d.pattern.Load() {
	case PatternSrcOnly:
		return d.src.Rollback()
	case PatternDstOnly:
		return d.dst.Rollback()
	case PatternSrcFirst:
		err := d.src.Rollback()
		if err != nil {
			return err
		}
		if d.dst != nil {
			err = d.dst.Rollback()
			if err != nil {
				zap.L().Error("目标表提交事务失败")
			}
		}
		return nil
	case PatternDstFirst:
		err := d.dst.Rollback()
		if err != nil {
			return err
		}
		if d.src != nil {
			err = d.src.Rollback()
			if err != nil {
				zap.L().Error("源表提交事务失败")
			}
		}
		return nil
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWriteTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	panic("双写模式不支持")
}

func (d *DoubleWriteTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err == nil {
			_, err1 := d.dst.ExecContext(ctx, query, args...)
			if err1 != nil {
				zap.L().Error("双写写入 dst 失败", zap.Error(err1), zap.String("query", query), zap.Any("args", args))
			}
		}
		return res, err
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err == nil {
			_, err1 := d.src.ExecContext(ctx, query, args...)
			if err1 != nil {
				zap.L().Error("双写写入 src 失败", zap.Error(err1), zap.String("query", query), zap.Any("args", args))
			}
		}
		return res, err
	default:
		return nil, errUnknownPattern
	}

}

func (d *DoubleWriteTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWriteTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		panic(errUnknownPattern)
	}
}
