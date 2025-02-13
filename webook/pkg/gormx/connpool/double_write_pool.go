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

func NewDoubleWritePool(src, dst gorm.ConnPool) *DoubleWritePool {
	return &DoubleWritePool{
		src:     src,
		dst:     dst,
		pattern: atomic.NewString(PatternSrcOnly),
	}
}
