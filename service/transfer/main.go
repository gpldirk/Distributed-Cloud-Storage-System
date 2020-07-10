package main

import (
	"bufio"
	"encoding/json"
	"github.com/cloud/config"
	"github.com/cloud/mq"
	"github.com/cloud/store/oss"
	"github.com/micro/go-micro"
	dbCli "github.com/cloud/service/dbproxy/client"
	"log"
	"os"
	"time"
)

func ProcessTransferData(msg []byte) bool {
	// 1 解析msg
	pubData := mq.TransferData{}
	err := json.Unmarshal(msg, pubData)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// 2 获取当前文件临时存储路径
	file, err := os.Open(pubData.Location)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	defer file.Close()

	// 3 将文件写入OSS
	err = oss.Bucket().PutObject(pubData.DestLocation, bufio.NewReader(file))
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// 4 更新唯一文件表中文件的存储路径为OSS
	if resp, success := dbCli.UpdateFileLocation(pubData.FileHash, pubData.DestLocation); !success {
		log.Println(err.Error())
		return false
	} else if !resp.Suc {
		log.Println("更新数据库异常，请检查:" + pubData.FileHash)
		return false
	} else {
		return true
	}
}

func startRPCService() {
	service := micro.NewService(
		micro.Name("go.micro.service.transfer"),
		micro.RegisterTTL(time.Second * 10),
		micro.RegisterInterval(time.Second * 5),
		micro.Registry(config.RegistryConsul()))
	service.Init()

	if err := service.Run(); err != nil {
		log.Println(err.Error())
	}
}

func startTransferService() {
	if !config.AsyncTransferEnable {
		log.Println("异步转移文件功能目前被禁用，请检查相关配置")
		return
	}
	log.Println("文件转移服务启动中，开始监听转移队列...")
	mq.StartConsume(config.TransOSSQueueName, "transfer_oss", ProcessTransferData)
}

func main() {
	// 异步启动文件转移服务
	go startTransferService()

	// rpc 服务
	startRPCService()
}

