package middleware

import (
	"github.com/cloud/util"
	"github.com/gin-gonic/gin"
	DBcli "github.com/cloud/service/dbproxy/client"
	"net/http"
	"time"
	"fmt"
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

// GenToken : 为指定user生成40位token　
func GenToken(username string) string {
	// 40位token = MD5(username + timestamp + token_salt) + timestamp[:8]
	timestamp := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + timestamp + "token_salt"))
	return tokenPrefix + timestamp[:8]
}

// IsTokenValid : 判断token是否有效
func IsTokenValid(username, token string) bool {
	// 1 判断token的时效性
	if len(token) != 40 {
		return false
	}
	tokenTS := token[32:40]
	if util.Hex2Dec(tokenTS) < time.Now().Unix() - 86400 { // 假设时效性为一天
		return false
	}

	// 2 从DB查询username对应的token进行对比是否一致
	if DBtoken, err := DBcli.GetUserToken(username); err != nil || DBtoken != token {
		return false
	} else {
		return true
	}
}
