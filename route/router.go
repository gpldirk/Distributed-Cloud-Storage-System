package route

import (
	"github.com/cloud/handler"
	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	// gin framework, 包括logger/recovery等中间件
	router := gin.Default()

	// 处理静态资源
	router.Static("/static/", "./static")

	// 不需要验证即可访问的接口
	router.GET("/user/signup", handler.SignUpHandler)
	router.POST("/user/signup", handler.DoSignUpHandler)
	router.GET("/user/signin", handler.SignInHandler)
	router.POST("/user/signin", handler.DoSignInHandler)

	// 加入中间件，进行token校验
	router.Use(handler.HTTPInterceptor())

	// use之后的handler都需要经过拦截器
	router.GET("/user/info", handler.UserInfoHandler)

	// 文件路由接口设置
	router.GET("/file/upload", handler.UploadHandler)
	router.POST("/file/upload", handler.DoUploadHandler)
	router.GET("/file/upload/suc", handler.UploadSucHandler)
	router.POST("/file/meta", handler.GetFileMetaHandler)
	router.POST("/file/query", handler.FileQueryHandler)
	router.GET("/file/download", handler.DownloadHandler)
	router.POST("/file/update", handler.UpdateFileMetaHandler)
	router.POST("/file/delete", handler.DeleteFileHandler)
	router.POST("/file/downloadurl", handler.DownloadURLHandler)

	// 秒传接口
	router.POST("/file/fastupload", handler.TryFastUploadHandler)

	// 分块上传接口
	router.POST("/file/mpupload/init", handler.InitialMultipartUploadHandler)
	router.POST("/file/mpupload/uppart", handler.UploadPartHandler)
	router.POST("/file/mpupload/complete", handler.CompleteUploadHandler)
	router.POST("/file/mpupload/cancel", handler.CancelUploadHandler)

	return router
}
