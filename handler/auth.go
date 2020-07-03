package handler

import (
	"github.com/cloud/util"
	"net/http"
)

// HTTPInterceptor : HTTP请求拦截器使用闭包实现
func HTTPInterceptor(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		username := r.Form.Get("username")
		token := r.Form.Get("token")
		if len(username) < 3 || len(token) < 5 {
			resp := util.RespMsg{
				Code: http.StatusForbidden,
				Msg:  "Invalid Token",
				Data: nil,
			}
			w.Write(resp.JSONBytes())
			return
		}
		h(w, r)
	}
}
