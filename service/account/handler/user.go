package handler

import (
	"context"
	_ "github.com/aliyun/alibaba-cloud-sdk-go/services/dbs"
	"github.com/cloud/db"
	"github.com/cloud/config"
	"github.com/cloud/util"
	proto "github.com/cloud/service/account/proto"
	_ "github.com/gin-gonic/gin"
	"net/http"
	"time"
	"fmt"
)

// 用于实现UserServiceHandler接口的对象
type User struct {}

// SignUp : 处理用户注册请求
func (user *User) Signup(ctx context.Context, req *proto.ReqSignup, res *proto.RespSignup) error {
	// 用户注册信息以表单形式提交
	username := req.Username
	password := req.Password
	if len(username) < 3 || len(password) < 5 {
		res.Code = http.StatusBadRequest
		res.Message = "Invalid parameters"
		return nil
	}

	encodedPWD := util.Sha1([]byte(password + config.Password_salt))
	success := db.UserSignUp(username, encodedPWD)
	if success {
		res.Code = http.StatusOK
		res.Message = "Signup succeeded"
	} else {
		res.Code = http.StatusInternalServerError
		res.Message = "Signup failed"
	}

	return nil
}

func (user *User) Signin(ctx context.Context, req *proto.ReqSignin, res *proto.RespSignin) error {
	username := req.Username
	password := req.Password

	// 1 校验用户名和密码
	encodedPWD := util.Sha1([]byte(password + config.Password_salt))
	PWDchecked := db.UserSignIn(username, encodedPWD)
	if !PWDchecked {
		res.Message = "Signin failed"
		res.Code = http.StatusUnauthorized
		return nil
	}

	// 3 生成返回访问凭证40位token
	token := GenToken(username)
	success := db.UpdateToken(username, token)
	if success {
		res.Message = "Signin failed"
		res.Code = http.StatusUnauthorized
		return nil
	} else {
		res.Message = "Signin succeeded"
		res.Code = http.StatusOK
		return nil
	}
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


// UserInfoHandler : 获取指定用户信息
func UserInfoHandler(ctx context.Context, req *proto.ReqUserInfo, res *proto.RespUserInfo) error {
	// 1 解析请求参数
	username := req.Username

	// 2 查询用户信息
	user, err := db.GetUserInfo(username)
	if err != nil {
		res.Code = http.StatusOK
		res.Message = "Internal server error"
		return nil
	}

	res.Code = http.StatusOK
	res.Username = user.Username
	res.SignupAt = user.SignUpAt
	res.LastActiveAt = user.LastActiveAt
	res.Status = int32(user.Status)
	res.Email = user.Email
	res.Phone = user.Phone
	return nil
}

