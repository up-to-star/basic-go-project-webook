package main

import (
	"basic-project/webook/internal/web"
	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()
	u := web.NewUserHandle()
	u.RegisterRoutes(server)
	_ = server.Run(":8080")
}
