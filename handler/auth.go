package handler

import (
	"github.com/cloud/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

// HTTPInterceptor : HTTP请求拦截器使用闭包实现
func HTTPInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")
		if len(username) < 3 || !IsTokenValid(username, token) {
			c.Abort() // 告知后面的handler不再执行
			resp := util.RespMsg{
				Code: http.StatusOK,
				Msg:  "Invalid Token",
				Data: nil,
			}
			c.Data(http.StatusOK, "application/json", resp.JSONBytes())
			return
		}
		c.Next() // 执行下一个handler
	}
}
