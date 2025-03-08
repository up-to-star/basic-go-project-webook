package domain

import "time"

type Comment struct {
	Id            int64     `json:"id"`
	Commentator   User      `json:"user"`
	Content       string    `json:"content"`
	Biz           string    `json:"biz"`
	BizId         int64     `json:"bizId"`
	RootComment   *Comment  `json:"rootComment"`
	ParentComment *Comment  `json:"parentComment"`
	Children      []Comment `json:"children"`
	Ctime         time.Time `json:"ctime"`
	Utime         time.Time `json:"utime"`
}

type User struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}
