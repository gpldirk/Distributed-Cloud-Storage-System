package ceph

import (
	"github.com/cloud/config"
	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

var cephConn *s3.S3

// GetCephConn : 获取ceph集群连接
func GetCephConn() *s3.S3 {
	if cephConn != nil {
		return cephConn
	}

	// 1 初始化ceph信息
	auth := aws.Auth{ // aws验证信息
		AccessKey: config.CephAccessKey,
		SecretKey: config.CephSecretKey,
	}
	region := aws.Region{ // aws region信息
		Name:                 "default",
		EC2Endpoint:          config.CephGWEndpoint, // 9080端口会映射到容器内部的80端口
		S3Endpoint:           config.CephGWEndpoint,
		S3BucketEndpoint:     "",
		S3LocationConstraint: false,
		S3LowercaseBucket:    false,
		Sign:                 aws.SignV2,
	}

	// 2 创建S3类型的连接
	return s3.New(auth, region)
}

// GetCephBucket : 返回ceph指定bucket
func GetCephBucket(bucket string) *s3.Bucket {
	conn := GetCephConn()
	return conn.Bucket(bucket)
}

// PutObject : 向指定bucket的指定path存储data
func PutObject(bucket, path string, data []byte) error {
	return GetCephBucket(bucket).Put(path, data, "octet-stream", s3.PublicRead)
}
