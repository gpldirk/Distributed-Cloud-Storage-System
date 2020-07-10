package config

import "github.com/cloud/common"

const (
	// TempLocalRootDir : 文件块在本地的临时存储路径
	TempLocalRootDir = "./data/fileserver_tmp/"
	// MergeLocalRootDir : 文件在本地的存储路径(普通上传和分块上传)
	MergeLocalRootDir = "./data/fileserver_merge/"
	// ChunkLocalRootDir : 文件块在本地的存储路径
	ChunkLocalRootDir = "./data/fileserver_chunk/"
	// CephRootDir : Ceph存储路径的prefix
	CephRootDir = "/ceph"
	// OSSRootDir : OSS存储路径的prefix
	OSSRootDir = "oss/"
	// CurrentStoreType : 当前的存储类型
	CurrentStoreType = common.StoreOSS
)
