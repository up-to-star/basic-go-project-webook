package domain

import "time"

type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
	Ctime   time.Time
	Utime   time.Time
	Status  ArticleStatus
}

func (a Article) Abstract() string {
	content := []rune(a.Content)
	if len(content) < 100 {
		return string(content)
	}
	return string(content[:100])
}

type ArticleStatus uint8

const (
	ArticleStatusUnknown ArticleStatus = iota
	ArticleStatusUnpublished
	ArticleStatusPublished
	ArticleStatusPrivate
)

func (s ArticleStatus) ToUint8() uint8 {
	return uint8(s)
}

func (s ArticleStatus) NonPublished() bool {
	return s != ArticleStatusPublished
}

func (s ArticleStatus) Valid() bool {
	return s.ToUint8() > 0
}

func (s ArticleStatus) String() string {
	switch s {
	case ArticleStatusUnpublished:
		return "Unpublished"
	case ArticleStatusPublished:
		return "Published"
	case ArticleStatusPrivate:
		return "Private"
	default:
		return "Unknown"
	}
}

type Author struct {
	Id   int64
	Name string
}
