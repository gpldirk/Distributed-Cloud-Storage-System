package config

const (
	// AsyncTransferEnable : 是否启动文件的异步转移
	AsyncTransferEnable = true
	// RabbitURL ： rabbitmq服务入口URL
	RabbitURL = "amqp://guest:guest@127.0.0.1:5672/"
	// TransExchangeName ：交换机名称
	TransExchangeName = "uploadserver.trans"
	// TransOSSQueueName ：文件转移到OSS使用的queue
	TransOSSQueueName = "uploadserver.trans.oss"
	// TransOSSErrQueueName ：文件转移失败使用的queue
	TransOSSErrQueueName = "uploadserver.trans.oss.err"
	// TransOSSRoutingKey : routing key
	TransOSSRoutingKey = "oss"
)
