package account

import (
	micro "github.com/micro/go-micro"
	proto "github.com/cloud/service/account/proto"
	"github.com/cloud/service/account/handler"
	"log"
	"time"
)

func main() {
	// 创建一个service
	service := micro.NewService(
		micro.Name("go.micro.service.user"), // 微服务模块名称
		micro.RegisterTTL(time.Second * 10), // 微服务模块响应超时时间，超时之后会从consul中删除其注册信息
		micro.RegisterInterval(time.Second * 5), // 微服务模块的心跳报时时间
		)
	service.Init()

	proto.RegisterUserServiceHandler(service.Server(), new(handler.User))
	if err := service.Run(); err != nil {
		log.Println(err)
	}
}
