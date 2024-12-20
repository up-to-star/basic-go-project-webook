package web

import (
	"basic-project/webook/internal/domain"
	"strconv"
)

type Page struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type ArticleVO struct {
	Id         string `json:"id"`
	Title      string `json:"title"`
	Abstract   string `json:"abstract"`
	Content    string `json:"content"`
	AuthorId   int64  `json:"author_id"`
	AuthorName string `json:"author_name"`
	Status     uint8  `json:"status"`
	ReadCnt    int64  `json:"read_cnt"`
	LikeCnt    int64  `json:"like_cnt"`
	CollectCnt int64  `json:"collect_cnt"`

	Liked     bool `json:"liked"`
	Collected bool `json:"collected"`

	Ctime string `json:"ctime"`
	Utime string `json:"utime"`
}

func toArticleVOs(arts []domain.Article) []ArticleVO {
	result := make([]ArticleVO, 0)
	for _, art := range arts {
		result = append(result, ArticleVO{
			Id:       strconv.FormatInt(art.Id, 10),
			Title:    art.Title,
			Abstract: art.Abstract(),
			Status:   art.Status.ToUint8(),
			Ctime:    art.Ctime.Format("2006-01-02 15:04:05"),
			Utime:    art.Utime.Format("2006-01-02 15:04:05"),
		})
	}
	return result
}
