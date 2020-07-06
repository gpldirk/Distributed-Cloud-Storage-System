package mq

import "github.com/cloud/common"

// TransferData : 转移队列中消息题的格式
type TransferData struct {
	FileHash string
	Location string
	DestLocation string
	DestStoreType common.StoreType
}



