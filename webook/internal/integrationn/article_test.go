package integrationn

import (
	"basic-project/webook/internal/domain"
	"basic-project/webook/internal/integrationn/startup"
	gorm2 "basic-project/webook/internal/repository/dao/article"
	ijwt "basic-project/webook/internal/web/jwt"
	"basic-project/webook/ioc"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 测试套件
type ArticleTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func (s *ArticleTestSuite) SetupTest() {
	s.server = startup.InitWebServer()
	s.db = ioc.InitDBDefault()
}

//// 每一个测试都会执行
//func (s *ArticleTestSuite) TearDownTest() {
//	// 清空所有数据, 并且自增组件恢复到1
//	s.db.Exec("TRUNCATE TABLE articles")
//}

func (s *ArticleTestSuite) TestABC() {
	s.T().Log("hello world")
}

func (s *ArticleTestSuite) TestEdit() {
	testCases := []struct {
		name  string
		token string
		// 集成测试准备数据
		before func(t *testing.T)
		// 集成测试验证数据
		after func(t *testing.T)
		// 预期中的输入
		art      Article
		wantCode int
		wantRes  Result[int64]
	}{
		{
			name:  "新建帖子-保存成功",
			token: generateToken(123),
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				// 验证数据库
				var art gorm2.Article
				err := s.db.Where("id = ?", 1).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, gorm2.Article{
					Id:       1,
					Title:    "test",
					Content:  "hello world",
					AuthorId: 123,
					Ctime:    0,
					Utime:    0,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}, art)
				s.db.Exec("TRUNCATE TABLE articles")
			},
			art: Article{
				Title:   "test",
				Content: "hello world",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 0,
				Msg:  "OK",
				Data: 1,
			},
		},
		{
			name:  "修改已有的帖子，并保存",
			token: generateToken(123),
			before: func(t *testing.T) {
				err := s.db.Create(gorm2.Article{
					Id:       2,
					Title:    "test",
					Content:  "hello world",
					AuthorId: 123,
					Ctime:    123,
					Utime:    234,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证数据库
				var art gorm2.Article
				err := s.db.Where("id = ?", 2).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Utime > 234)
				art.Utime = 0
				assert.Equal(t, gorm2.Article{
					Id:       2,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorId: 123,
					Ctime:    123,
					Utime:    0,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}, art)
				s.db.Exec("TRUNCATE TABLE articles")
			},
			art: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 0,
				Msg:  "OK",
				Data: 2,
			},
		},
		{
			name:  "修改别人的帖子",
			token: generateToken(123),
			before: func(t *testing.T) {
				err := s.db.Create(gorm2.Article{
					Id:       3,
					Title:    "test",
					Content:  "hello world",
					AuthorId: 789, // 在修改别人的数据
					Ctime:    123,
					Utime:    234,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证数据库
				var art gorm2.Article
				err := s.db.Where("id = ?", 3).First(&art).Error
				assert.NoError(t, err)
				assert.Equal(t, gorm2.Article{
					Id:       3,
					Title:    "test",
					Content:  "hello world",
					AuthorId: 789,
					Ctime:    123,
					Utime:    234,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, art)
				s.db.Exec("TRUNCATE TABLE articles")
			},
			art: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
				Data: 0,
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/edit",
				bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tc.token)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()
			s.server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}
			var result Result[int64]
			err = json.NewDecoder(recorder.Body).Decode(&result)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, result)
		})
	}
}

func TestArticle(t *testing.T) {
	suite.Run(t, &ArticleTestSuite{})
}

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

func generateToken(uid int64) string {
	claims := ijwt.UserClaims{
		Uid: uid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString(ijwt.AtKey)
	return tokenStr
}
