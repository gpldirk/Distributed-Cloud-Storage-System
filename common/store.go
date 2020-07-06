package common

type StoreType int

const (
	_ StoreType = iota
	// StoreLocal : 本地存储
	StoreLocal
	// StoreCeph : Ceph存储
	StoreCeph
	// StoreOSS : OSS存储
	StoreOSS
	// StoreMix : 混合存储(Ceph + OSS)
	StoreMix
	// StoreAll : 启用所有类型的存储
	StoreAll
)
