package handler

import (
	"github.com/cloud/db"
	userProto "github.com/cloud/service/account/proto"
	"github.com/gin-gonic/gin"
	micro "github.com/micro/go-micro"
	"log"
	"net/http"
	"context"
)

var (
	userCli userProto.UserService
)

func init() {
	service := micro.NewService(
		micro.Registry(config.RegistryConsul()),
	)
	// 初始化，解析命令行参数
	service.Init()
	// 初始化一个account服务的rpcCli
	userCli = userProto.NewUserService("go.micro.service.user", service.Client())
}

// SignUpHandler : 处理用户注册get请求
func SignUpHandler(c *gin.Context) {
	// 返回signup页面
	c.Redirect(http.StatusFound, "/static/view/signup.html")
}

// DoSignUpHandler : 处理用户注册post请求
func DoSignUpHandler(c *gin.Context) {
	// 用户注册信息以表单形式提交
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")
	resp, err := userCli.Signup(context.TODO(), &userProto.ReqSignup{
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": resp.Message,
		"code": resp.Code,
	})
}

// SignInHandler ： 处理用户登陆get请求
func SignInHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signin.html")
}

// DoSignInHandler : 处理用户登陆post请求
func DoSignInHandler(c *gin.Context) {
	// 1 解析请求参数
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")

	resp, err := userCli.Signin(context.TODO(), &userProto.ReqSignin{
		Username:             username,
		Password:             password,
	})
	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	if resp.Code != http.StatusOK {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Failed to login",
			"code": resp.Code,
		})
		return
	}


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





