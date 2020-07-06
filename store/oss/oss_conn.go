package oss

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/cloud/config"
	"log"
)

var ossCli *oss.Client

// Client : 创建OSS连接
func Client() *oss.Client {
	if ossCli != nil {
		return ossCli
	}

	ossCli, err := oss.New(config.OSSEndPoint, config.OSSAccessKey, config.OSSSecretkey)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	return ossCli
}

// Bucket : 基于oss cli创建/获取bucket对象
func Bucket() *oss.Bucket {
	cli := Client()
	if cli != nil {
		bucket, err := cli.Bucket(config.OSSBucket)
		if err != nil {
			log.Println(err.Error())
			return nil
		}
		return bucket
	}
	return nil
}

// DownloadURL : 获取临时授权的URL
func DownloadURL(objName string) (signedURL string) {
	signedURL, err := Bucket().SignURL(objName, oss.HTTPGet, 3600) // 过期时间为一小时
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	return signedURL
}

// BuildLifeCycleRules : 为指定bucket设置生命周期管理规则
func BuildLifeCycleRules(bucket string) {
	// 表示前缀为test的对象(文件)的过期时间为30天
	ruleTest := oss.BuildLifecycleRuleByDays("rule1", "test/", true, 30) // 过期时间为30天
	rules := []oss.LifecycleRule{ruleTest}
	Client().SetBucketLifecycle(bucket, rules)
}


