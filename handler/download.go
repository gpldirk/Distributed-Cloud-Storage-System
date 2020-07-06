package handler

import (
	"github.com/cloud/db"
	"github.com/cloud/store/oss"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// DownloadHandler : 通过指定filehash下载某个文件
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	fileMeta := meta.GetFileMeta(filehash)
	file, err := os.Open(fileMeta.Location)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("content-disposition", "attachment; filename=\"" + fileMeta.FileName + "\"")
	w.Write(data)
}

// DownloadURLHandler : 获取下载文件的URL
func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	fileMeta, _ := db.GetFileMeta(filehash)

	// 判断文件存在oss还是ceph
	signedURL := oss.DownloadURL(fileMeta.FileAddr.String)
	w.Write([]byte(signedURL))
}