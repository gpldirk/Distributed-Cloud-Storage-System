package rpc

import (
	"context"
	"github.com/cloud/service/download/config"
	dlProto "github.com/cloud/service/download/proto"
)

// Dwonload :download结构体
type Download struct{}

// DownloadEntry : 获取下载入口
func (u *Download) DownloadEntry(
	ctx context.Context,
	req *dlProto.ReqEntry,
	res *dlProto.RespEntry) error {

	res.Entry = config.DownloadEntry
	return nil
}

