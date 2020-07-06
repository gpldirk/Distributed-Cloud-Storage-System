package mq

import "log"

var done chan bool

// StartConsume : 开始监听队列，处理消息
func StartConsume(qName, cName string, callback func(msg []byte) bool)  {
	// 1 通过channel.Consume获取channel
	msgs, err := channel.Consume(
		qName,
		cName,
		true,  //自动应答
		false, // 非唯一的消费者
		false, // rabbitMQ只能设置为false
		false, // noWait, false表示会阻塞直到有消息过来
		nil)

	if err != nil {
		log.Println(err.Error())
		return
	}

	done := make(chan bool)
	// 2 循环从msgs信道获取msg，没有则阻塞等待
	go func() {
		for msg := range msgs {
			// 3 每次获取新的消息则调用callback进行处理
			if success := callback(msg.Body); !success {
				// TODO: 写入重试队列进行消息的再次consume
			}
		}
	}()

	// ...

	// 队列没有处理完成，就会一直处于阻塞装态
	<- done

	// 完成所有信息的处理之后关闭channel
	channel.Close()
}

// StopConsume : 停止监听队列
func StopConsume() {
	done <- true
}
