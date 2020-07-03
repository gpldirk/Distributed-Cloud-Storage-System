package meta

import (
	"github.com/cloud/db"
	"log"
	"sort"
)

// FileMeta : 文件元信息结构
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
}

var fileMetas map[string]FileMeta

func init() {
	fileMetas = make(map[string]FileMeta)
}

// UpdateFileMeta : 在Map中更新文件元信息
func UpdateFileMeta(fileMeta FileMeta) {
	fileMetas[fileMeta.FileSha1] = fileMeta
}

// GetFileMeta : 通过sha1获取文件元信息
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}

// GetLastFileMetas : 获取最近生成的limit个filemeta
func GetLastFileMetas(limit int) []FileMeta {
	fileMetaArr := make([]FileMeta, len(fileMetas))
	for _, v := range fileMetas {
		fileMetaArr = append(fileMetaArr, v)
	}

	// 强制类型转换后进行排序
	sort.Sort(ByUploadTime(fileMetaArr))
	if len(fileMetaArr) < limit {
		return fileMetaArr
	} else {
		return fileMetaArr[0:limit]
	}
}

// RemoveFileMeta : 删除文件元信息
func RemoveFileMeta(filehash string) {
	delete(fileMetas, filehash)
}


// GetFileMetaDB : 从DB获取文件元信息
func GetFileMetaDB(fileSha1 string) (FileMeta, error) {
	tfile, err := db.GetFileMeta(fileSha1)
	if err != nil {
		return FileMeta{}, err
	}
	fmeta := FileMeta{
		FileSha1: tfile.FileHash,
		FileName: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
	}

	return fmeta, nil
}

// GetLastFileMetasDB : 批量从mysql获取文件元信息
func GetLastFileMetasDB(limit int) ([]FileMeta, error) {
	files, err := db.GetFileMetaList(limit)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	var fileMetas []FileMeta
	for _, file := range files {
		fileMeta := FileMeta{
			FileSha1: file.FileHash,
			FileName: file.FileName.String,
			FileSize: file.FileSize.Int64,
			Location: file.FileAddr.String,
		}
		fileMetas = append(fileMetas, fileMeta)
	}

	return fileMetas, nil
}

// UpdateFileMetaFromDB : 在DB中加入/更新文件元信息
func UpdateFileMetaDB(fileMeta FileMeta) bool {
	return db.OnFileUploadFinished(fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize, fileMeta.Location)
}

// OnFileRemovedDB : 删除文件
func OnFileRemovedDB(filehash string) bool {
	return db.OnFileRemoved(filehash)
}



