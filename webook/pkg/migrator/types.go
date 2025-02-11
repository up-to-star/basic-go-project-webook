package migrator

type Entity interface {
	// ID 返回ID
	ID() int64
	CompareTo(dst Entity) bool
}
