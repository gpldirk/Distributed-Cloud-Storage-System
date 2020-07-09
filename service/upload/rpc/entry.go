package rpc

import (
	"context"
	"github.com/cloud/service/upload/config"
	upProto "github.com/cloud/service/upload/proto"
)

type Upload struct {}

func (upload *Upload) UploadEntry(ctx context.Context, req *upProto.ReqEntry, res *upProto.RespEntry) error {
	res.Entry = config.UploadEntry
	return nil
}

