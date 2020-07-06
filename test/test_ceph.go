package test

import (
	"github.com/cloud/store/ceph"
	"gopkg.in/amz.v1/s3"
	"log"
)

func main() {
	bucket := ceph.GetCephBucket("testbucket1")
	// 1 创建一个新的bucket
	err := bucket.PutBucket(s3.PublicRead)
	log.Println(err)

	// 2 查询当前bucket中指定的object keys
	res, err := bucket.List("", "", "", 100)
	log.Println(res)

	// 3 新上传一个对象到bucket
	err = bucket.Put("/testupload/a.txt", []byte("just for test"), "octet-stream", s3.PublicRead)
	log.Println(err)

	// 4 再次查询当前bucket中指定的object keys
	res, err = bucket.List("", "", "", 100)
	log.Println(res)
}


