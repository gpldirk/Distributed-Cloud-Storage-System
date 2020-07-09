package handler

import (
	"fmt"
	"github.com/cloud/db"
	"github.com/cloud/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const (
	password_salt = "*#890"
	token_salt = "token_salt"
)

// SignUpHandler : 处理用户注册get请求
func SignUpHandler(c *gin.Context) {
	// 返回signup页面
	c.Redirect(http.StatusFound, "/static/view/signup.html")
	return
}

// DoSignUpHandler : 处理用户注册post请求
func DoSignUpHandler(c *gin.Context) {
	// 用户注册信息以表单形式提交
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")
	if len(username) < 3 || len(password) < 5 {
		c.JSON(http.StatusOK, gin.H{
			"msg": "Invalid parameters",
			"code": -1,
		})
		return
	}

	encodedPWD := util.Sha1([]byte(password + password_salt))
	success := db.UserSignUp(username, encodedPWD)
	if success {
		c.JSON(http.StatusOK, gin.H{
			"msg": "Signup succeeded",
			"code": 0,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg": "Signup failed",
			"code": -2,
		})
	}
}

// SignInHandler ： 处理用户登陆get请求
func SignInHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signin.html")
	return
}

// DoSignInHandler : 处理用户登陆post请求
func DoSignInHandler(c *gin.Context) {
	// 1 解析请求参数
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")

	// 2 校验用户名和密码
	encodedPWD := util.Sha1([]byte(password + password_salt))
	PWDchecked := db.UserSignIn(username, encodedPWD)
	if !PWDchecked {
		c.JSON(http.StatusOK, gin.H{
			"msg": "Signin failed",
			"code": -1,
		})
		return
	}

	// 3 生成返回访问凭证40位token
	token := GenToken(username)
	success := db.UpdateToken(username, token)
	if !success {
		c.JSON(http.StatusOK, gin.H{
			"msg": "Signin failed",
			"code": -2,
		})
		return
	}

	// 4 重定向到首页: 发送重定向的url
	resp := util.RespMsg {
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}

	c.Data(http.StatusOK, "application/json", resp.JSONBytes())
}

// UserInfoHandler : 获取指定用户信息
func UserInfoHandler(c *gin.Context) {
	// 1 解析请求参数
	username := c.Request.FormValue("username")
	// token := r.Form.Get("token")

	// 2 验证token是否有效
	//valid := IsTokenValid(token)
	//if !valid {
	//	w.WriteHeader(http.StatusForbidden)
	//	return
	//}

	// 3 查询用户信息
	user, err := db.GetUserInfo(username)
	if err != nil {
		c.JSON(http.StatusOK, "Internal server error")
		return
	}

	// 4 响应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	c.Data(http.StatusOK, "application/json", resp.JSONBytes())
}

// GenToken : 为指定user生成40位token　
func GenToken(username string) string {
	// 40位token = MD5(username + timestamp + token_salt) + timestamp[:8]
	timestamp := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + timestamp + token_salt))
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
	if db.GetUserToken(username) != token {
		return false
	} else {
		return true
	}
}