package dbproxy

import (
	"github.com/cloud/config"
	dbProxy "github.com/cloud/service/dbproxy/proto"
	dbRpc "github.com/cloud/service/dbproxy/rpc"
	"github.com/micro/go-micro"
	"log"
	"time"
)

func startRPCService() {
	service := micro.NewService(
		micro.Name("go.micro.service.dbproxy"),
		micro.RegisterTTL(time.Second * 10),
		micro.RegisterInterval(time.Second * 5),
		micro.Registry(config.RegistryConsul()))

	service.Init()
	dbProxy.RegisterDBProxyServiceHandler(service.Server(), new(dbRpc.DBProxy))
	if err := service.Run(); err != nil {
		log.Println(err)
	}
}

func main() {
	startRPCService()
}