package handler

import (
	"fmt"
	"github.com/cloud/config"
	"github.com/cloud/db"
	"github.com/cloud/meta"
	"github.com/cloud/store/ceph"
	"github.com/cloud/store/oss"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// DownloadHandler : 某个用户下载某个文件
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	userFile, err := db.QueryUserFileMeta(username, filehash)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var Data []byte
	// 文件存储在本地
	if strings.HasPrefix(fileMeta.Location, config.MergeLocalRootDir) {
		fmt.Println("To download file from local directory")
		f, err := os.Open(fileMeta.Location)
		if err != nil {
			log.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer f.Close()

		Data, err = ioutil.ReadAll(f)
		if err != nil {
			log.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else if strings.HasPrefix(fileMeta.Location, "/ceph") {
		fmt.Println("To download file from ceph")
		bucket := ceph.GetCephBucket("userfile")
		Data, err = bucket.Get(fileMeta.Location)
		if err != nil {
			log.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else if strings.HasPrefix(fileMeta.Location, "oss/") {
		fmt.Println("To download file from OSS")
		f, err := oss.Bucket().GetObject(fileMeta.Location)
		defer f.Close()
		if err == nil {
			Data, err = ioutil.ReadAll(f)
		}
		if err != nil {
			log.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.Write([]byte("File not found"))
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("content-disposition", "attachment; filename=\"" + userFile.FileName + "\"")
	w.Write(Data)
}

// DownloadURLHandler : 获取下载文件的URL
func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	fileMeta, _ := db.GetFileMeta(filehash)

	// 判断文件存在于本地/ceph，还是OSS
	if strings.HasPrefix(fileMeta.FileAddr.String, config.MergeLocalRootDir) ||
		strings.HasPrefix(fileMeta.FileAddr.String, "/ceph") {
		username := r.Form.Get("username")
		token := r.Form.Get("token")
		tmpURL := fmt.Sprintf("http://%s/file/download?username=%s&filehash=%s&token=%s", r.Host, username, token)
		w.Write([]byte(tmpURL))
	} else if strings.HasPrefix(fileMeta.FileAddr.String, "oss/") {
		signedURL := oss.DownloadURL(fileMeta.FileAddr.String)
		w.Write([]byte(signedURL))
	} else {
		w.Write([]byte("Error: 暂时无法生成下载链接"))
	}
}

// RangeDownloadHandler : 支持断点的文件下载接口
func RangeDownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	username := r.Form.Get("username")
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	userFile, err := db.QueryUserFileMeta(username, filehash)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 使用本地文件目录
	filePath := config.MergeLocalRootDir + fileMeta.FileSha1
	fmt.Println("range-download-file-path: " + filePath)

	f, err := os.Open(filePath)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "application/octect-stream")
	// attachment表示文件将会提示下载到本地，而不是直接在浏览器中打开
	w.Header().Set("content-disposition", "attachment; filename=\"" + userFile.FileName + "\"")
	http.ServeFile(w, r, filePath)
}
