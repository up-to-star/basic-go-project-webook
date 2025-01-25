package errs

// 用户模块
const (
	// UserInputValid 用户模块输入错误
	UserInputValid = 4010001
	// UserInvalidOrPassword 用户不存在或密码错误
	UserInvalidOrPassword   = 4010002
	UserInternalServerError = 501001
)

// 文章模块
const (
	ArticleInternalServerError = 502001
	ArticleInvalidInput        = 402001
)
