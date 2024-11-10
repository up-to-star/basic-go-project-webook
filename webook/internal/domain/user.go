package domain

// User 领域对象，DDD中的聚合根
type User struct {
	Id       int64
	Email    string
	Password string
}
