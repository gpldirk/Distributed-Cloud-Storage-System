package download

import (
	"fmt"
	"github.com/cloud/service/download/config"
	dlProto "github.com/cloud/service/download/proto"
	"github.com/cloud/service/download/route"
	"github.com/micro/go-micro"
	"time"
)

func startRPCService() {
	service := micro.NewService(
		micro.Name("go.micro.service.download"),
		micro.RegisterTTL(time.Second * 10),
		micro.Registry(time.Second * 5),
		micro.Registry(config.RegistryConsul()))
	service.Init()

	dlProto.RegisterDownloadServiceHandler(service.Server(), new(dlRpc.Download))
	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}

func startAPIService() {
	router := route.Router()
	router.Run(config.DownloadServiceHost)
}

func main() {
	// 启动API服务
	startAPIService()

	// 启动RPC服务
	startRPCService()
}
