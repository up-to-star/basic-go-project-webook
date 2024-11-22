package domain

import "time"

// User 领域对象，DDD中的聚合根
type User struct {
	Id         int64
	Email      string
	Password   string
	Nickname   string
	Phone      string
	Birthday   time.Time
	AboutMe    string
	Ctime      time.Time
	Utime      time.Time
	WechatInfo WechatInfo
}
