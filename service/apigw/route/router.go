package route

import (
	"github.com/cloud/middleware"
	"github.com/cloud/service/apigw/handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Router : 网关api路由
func Router() *gin.Engine {
	router := gin.Default()

	router.Static("/static/", "./static")

	// 注册
	router.GET("/user/signup", handler.SignUpHandler)
	router.POST("/user/signup", handler.DoSignUpHandler)
	// 登录
	router.GET("/user/signin", handler.SignInHandler)
	router.POST("/user/signin", handler.DoSignInHandler)

	// 使用gin插件支持跨域请求
	router.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"}, // []string{"http://localhost:8080"},
		AllowMethods:  []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Range", "x-requested-with", "content-Type"},
		ExposeHeaders: []string{"Content-Length", "Accept-Ranges", "Content-Range", "Content-Disposition"},
		// AllowCredentials: true,
	}))

	router.Use(middleware.HTTPInterceptor())

	// 用户查询
	router.POST("/user/info", handler.UserInfoHandler)

	// 用户文件查询
	router.POST("/file/query", handler.FileQueryHandler)
	// 用户文件修改(重命名)
	router.POST("/file/update", handler.FileMetaUpdateHandler)

	return router
}