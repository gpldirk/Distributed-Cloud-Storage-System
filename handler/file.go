package handler

import (
	"database/sql"
	"encoding/json"
	"github.com/cloud/db"
	"github.com/cloud/meta"
	"github.com/cloud/util"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// 返回上传文件html页面
		data, err := ioutil.ReadFile("/static/view/index.html")
		if err != nil {
			io.WriteString(w, "Internal server error")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		// 客户端以表单形式提交文件
		// 接收文件流，存储到本地目录
		file, header, err := r.FormFile("file")
		if err != nil {
			log.Println(err.Error())
			io.WriteString(w, "Internal server error")
			return
		}
		defer file.Close()

		// 创建文件元信息对象
		fileMeta := meta.FileMeta{
			FileName: header.Filename,
			Location: "/tmp/" + header.Filename,
			UploadAt: time.Now().Format("2016-01-02 15:04:05"),
		}
		
		// 创建本地文件获取句柄
		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			log.Println(err.Error())
			return
		}
		defer newFile.Close()

		// 将接收文件流copy到新的文件中
		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			log.Println(err.Error())
			return
		}

		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)
		// 写入唯一文件表
		if success := meta.UpdateFileMetaDB(fileMeta); !success {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// 写入用户文件表
		r.ParseForm()
		username := r.Form.Get("username")
		if success := db.OnUserFileUploadFinished(username, fileMeta.FileSha1,
			fileMeta.FileName, fileMeta.FileSize); !success {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/x/view/home.html", http.StatusFound)
	}
}

// UploadSucHandler : 返回上传文件成功页面
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload Finished!")
}

// GetFileMetaHandler : 通过指定filehash获取对应文件的元信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form["filehash"][0]
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(fileMeta)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// FileQueryHandler : 通过指定limit获取最近上传文件生成的文件元信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	limitCnt, err := strconv.Atoi(r.Form.Get("limit"))
	if err != nil {
		log.Println(err.Error())
		io.WriteString(w, "Internal server error")
		return
	}

	// 获取 last limit file Metas from DB or map
	// fileMetas := meta.GetLastFileMetas(limitCnt)
	fileMetas, err := db.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// json 序列化
	data, err := json.Marshal(fileMetas)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

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

// UpdateFileMetaHandler : 更新文件元信息，通过op指定更新类型(op = 1 -> 修改文件名)
func UpdateFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	opType := r.Form.Get("op")
	filehash := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")
	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	curFileMeta := meta.GetFileMeta(filehash)
	curFileMeta.FileName = newFileName
	meta.UpdateFileMeta(curFileMeta)

	data, err := json.Marshal(curFileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// DeleteFileHandler :  删除文件及其元信息
func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form.Get("filehash")

	fileMeta := meta.GetFileMeta(filehash)
	os.Remove(fileMeta.Location)

	meta.RemoveFileMeta(filehash)

	w.WriteHeader(http.StatusOK)
}

// TryFastUploadHandler : 尝试秒传接口(判断当前文件是否上传过)
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))

	// 在唯一文件表中查询对应文件是否存在
	fileMeta, err := meta.GetFileMetaDB(filehash)

	// 查询不到则返回秒传失败
	if err != nil && err != sql.ErrNoRows {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err == sql.ErrNoRows || fileMeta.FileSha1 == "" {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请使用普通上传接口",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 上传过当前文件，则将文件信息和用户信息写入用户文件表(fast upload)
	if success := db.OnUserFileUploadFinished(username, filehash, filename, int64(filesize)); success {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
	} else {
		resp := util.RespMsg{
			Code: -2,
			Msg:  "秒传失败，请稍后重试",
		}
		w.Write(resp.JSONBytes())
	}

	return
}
