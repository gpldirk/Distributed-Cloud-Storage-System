package mq

import (
	"github.com/cloud/config"
	"github.com/streadway/amqp"
	"log"
)

var conn *amqp.Connection
var channel *amqp.Channel
// 如果mq异常关闭，会接受到通知
var notifyClose chan *amqp.Error

func init() {
	// 判断是否开启异步转移功能，是否连接rabbitmq
	if !config.AsyncTransferEnable {
		return
	}
	if initChannel() {
		channel.NotifyClose(notifyClose)
	}

	// 如果channel异常关闭，则重新开启 - 短线重联
	go func() {
		for {
			select {
			case msg := <- notifyClose:
				log.Printf("Channel has benn closed: %+v\n", msg)
				conn = nil
				channel = nil
				initChannel()
			}
		}
	}()
}

// initChannel : 获取channel
func initChannel() bool {
	// 1 判断channel是否已经被创建
	if channel != nil {
		return true
	}

	// 2 获取rabbitmq的一个连接
	conn, err := amqp.Dial(config.RabbitURL)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// 3 通过连接创建channel，进行消息的发布
	channel, err = conn.Channel()
	if err != nil {
		log.Println(err.Error())
		return false
	}
	return true
}

// Publish : 发布消息到指定exchange with rk
func Publish(exchange, routingKey string, msg []byte) bool {
	// 1 判断channel是否可用
	if !initChannel() {
		return false
	}

	// 2 发布消息给mq
	err := channel.Publish(exchange, routingKey, false, false, amqp.Publishing{
		ContentType:     "text/plain",
		Body:            msg,
	})
	if err != nil {
		log.Println(err.Error())
		return false
	}

	return true
}