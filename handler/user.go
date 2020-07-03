package handler

import (
	"fmt"
	"github.com/cloud/db"
	"github.com/cloud/util"
	"net/http"
	"time"
)

const (
	password_salt = "*#890"
	token_salt = "token_salt"
)

// SignUpHandler : 处理用户注册请求
func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	// 返回signup页面
	if r.Method == "GET" {
		http.Redirect(w, r, "/static/view/signup.html", http.StatusFound)
		return
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		if len(username) < 3 || len(password) < 5 {
			w.Write([]byte("Invalid parameters"))
			return
		}

		encodedPWD := util.Sha1([]byte(password + password_salt))
		success := db.UserSignUp(username, encodedPWD)
		if success {
			w.Write([]byte("SUCCESS"))
		} else {
			w.Write([]byte("FAILED"))
		}
	}
}

// SignInHandler ： 处理用户登陆请求
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.Redirect(w, r, "/static/view/signin.html", http.StatusFound)
		return
	} else if r.Method == "POST" {
		// 1 校验用户名和密码
		r.ParseForm()
		username := r.Form.Get("username")
		password := r.Form.Get("password")

		// 2 校验用户名和密码
		encodedPWD := util.Sha1([]byte(password + password_salt))
		PWDchecked := db.UserSignIn(username, encodedPWD)
		if !PWDchecked {
			w.Write([]byte("FAILED"))
			return
		}

		// 3 生成返回访问凭证40位token
		token := GenToken(username)
		success := db.UpdateToken(username, token)
		if !success {
			w.Write([]byte("FAILED"))
			return
		}

		// 4 重定向到首页: 发送重定向的url
		// w.Write([]byte("http://" + r.Host + "/x/view/home.html"))
		resp := util.RespMsg {
			Code: 0,
			Msg:  "OK",
			Data: struct {
				Location string
				Username string
				Token    string
			}{
				Location: "http://" + r.Host + "/static/view/home.html",
				Username: username,
				Token:    token,
			},
		}

		w.Write(resp.JSONBytes())
	}
}

// UserInfoHandler : 获取指定user信息
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// 1 解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 4 响应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	w.Write(resp.JSONBytes())
}

// GenToken : 位指定user生成40位token　
func GenToken(username string) string {
	// 40位token = MD5(username + timestamp + token_salt) + timestamp[:8]
	timestamp := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + timestamp + token_salt))
	return tokenPrefix + timestamp[:8]
}

// IsTokenValid : 判断token是否有效
func IsTokenValid(token string) bool {
	// 1 判断token的时效性
	if len(token) != 40 {
		return false
	}
	tokenTS := token[32:40]
	if util.Hex2Dec(tokenTS) < time.Now().Unix() - 86400 {
		return false
	}

	// 2 从DB查询username对应的token进行对比是否一致


	return true
}