package main

import (
	"fmt"
	"github.com/cloud/config"
	"github.com/cloud/route"
)

func main() {
	r := route.Router()
	r.Run(config.UploadServiceHost)
	// 监听端口
	fmt.Printf("上传服务已启动，监听： %s\n", config.UploadServiceHost)
}
