package main

import (
	"github.com/cloud/store/ceph"
	"gopkg.in/amz.v1/s3"
	"log"
)

func main() {
	bucket := ceph.GetCephBucket("userfile")
	// 1 创建一个新的bucket
	err := bucket.PutBucket(s3.PublicRead)
	if err != nil {
		log.Println(err.Error())
	}

	// 2 查询当前bucket中指定的object keys
	res, err := bucket.List("", "", "", 100)
	if err != nil {
		log.Println(err.Error())
	} else {
		log.Println(res)
	}

	// 3 新上传一个对象到bucket
	objAPath := "/testupload/a.txt"
	err = bucket.Put(objAPath, []byte("just for test"), "octet-stream", s3.PublicRead)
	if err != nil {
		log.Println(err.Error())
	}

	// 4 再次查询当前bucket中指定的object keys
	res, err = bucket.List("", "", "", 100)
	if err != nil {
		log.Println(err.Error())
	} else {
		log.Println(res)
	}
}


