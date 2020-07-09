package main

import (
	"fmt"
	"github.com/cloud/service/upload/config"
	"github.com/cloud/route"
	"github.com/micro/go-micro"
	upProto "github.com/cloud//service/upload/proto"
	upRpc "github.com/cloud/service/upload/rpc"
	"log"
	"time"
)

func startRPCService() {
	service := micro.NewService(
		micro.Name("go.micro.service.upload"),
		micro.RegisterTTL(time.Second * 10),
		micro.RegisterInterval(time.Second * 5),
		micro.Registry(config.RegistryConsul()))
	service.Init()

	upProto.RegisterUploadServiceHandler(service.Server(), new(upRpc.Upload))
	if err := service.Run(); err != nil {
		log.Println(err)
	}
}

func startAPIService() {
	router := route.Router()
	router.Run(config.UploadServiceHost)
}

func main() {
	// 启动API服务
	go startAPIService()

	// 启动RPC服务
	startRPCService()　
}
