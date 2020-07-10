package handler

import (
	"context"
	"github.com/cloud/common"
	"github.com/cloud/config"
	userProto "github.com/cloud/service/account/proto"
	dlproto "github.com/cloud/service/download/proto"
	uploadProto "github.com/cloud/service/upload/proto"
	"github.com/cloud/util"
	"github.com/gin-gonic/gin"
	micro "github.com/micro/go-micro"
	"log"
	"net/http"
)

var (
	userCli userProto.UserService
	upCli uploadProto.UploadService
	dlCli dlproto.DownloadService
)

func init() {
	service := micro.NewService(
		micro.Registry(config.RegistryConsul()),
	)
	// 初始化，解析命令行参数
	service.Init()
	// 初始化一个account服务的rpcCli
	userCli = userProto.NewUserService("go.micro.service.user", service.Client())
	// 初始化一个upload服务的rpcCli
	upCli = uploadProto.NewUploadService("go.micro.service.upload", service.Client())
	// 初始化一个download服务的rpcCli
	dlCli = dlproto.NewDownloadService("go.micro.service.download", service.Client())
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

	rpcResp, err := userCli.Signin(context.TODO(), &userProto.ReqSignin{
		Username:             username,
		Password:             password,
	})
	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	if rpcResp.Code != http.StatusOK {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Failed to login",
			"code": rpcResp.Code,
		})
		return
	}

	uploadEntryRes, err := upCli.UploadEntry(context.TODO(), &uploadProto.ReqEntry{})
	if err != nil {
		log.Println(err.Error())
	} else if uploadEntryRes.Code != http.StatusOK {
		log.Println(uploadEntryRes.Message)
	}


	downloadEntryRes, err := dlCli.DownloadEntry(context.TODO(), &dlproto.ReqEntry{})
	if err != nil {
		log.Println(err.Error())
	} else if downloadEntryRes.Code != http.StatusOK {
		log.Println(downloadEntryRes.Message)
	}

	cliResp := util.RespMsg{
		Code: int(common.StatusOK),
		Msg:  "登陆成功",
		Data: struct {
			Location      string
			Username      string
			Token         string
			UploadEntry   string
			DownloadEntry string
		}{
			Location:      "/static/view/home.html",
			Username:      username,
			Token:         rpcResp.Token,
			UploadEntry:   uploadEntryRes.Entry,
			DownloadEntry: downloadEntryRes.Entry,
		},
	}

	c.Data(http.StatusOK, "application/json", cliResp.JSONBytes())
}

// UserInfoHandler ： 查询用户信息
func UserInfoHandler(c *gin.Context) {
	// 1. 解析请求参数
	username := c.Request.FormValue("username")

	resp, err := userCli.UserInfo(context.TODO(), &userProto.ReqUserInfo{
		Username: username,
	})

	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	// 3. 组装并且响应用户数据
	cliResp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: gin.H{
			"Username": username,
			"SignupAt": resp.SignupAt,
			// TODO: 完善其他字段信息
			"LastActive": resp.LastActiveAt,
		},
	}
	c.Data(http.StatusOK, "application/json", cliResp.JSONBytes())
}





