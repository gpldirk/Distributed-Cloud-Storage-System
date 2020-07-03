package handler

import (
	"fmt"
	"github.com/cloud/db"
	"github.com/cloud/util"
	rPool "github.com/cloud/cache/redis"
	"github.com/gomodule/redigo/redis"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// MultipartUploadInfo : 分块上传时每块文件元信息
type MultipartUploadInfo struct {
	FileHash string
	FilSize int
	UploadID string
	ChunkSize int
	ChunkCount int

}

// 初始化分块上传
func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1 解析用户请求
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		log.Println(err.Error())
		w.Write(util.NewRespMsg(http.StatusBadRequest, "Invalid parameters", nil).JSONBytes())
		return
	}

	// 2 获取redis连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3 生成分块上传初始化信息
	uploadInfo := MultipartUploadInfo {
		FileHash:   filehash,
		FilSize:    filesize,
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()), // username + timestamp
		ChunkSize:  5 * 1024 * 1024, // 5MB
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
	}
	
	// 4 将初始化信息写入redis
	rConn.Do("HSET", "MP_" + uploadInfo.UploadID, "chunkcount", uploadInfo.ChunkCount)
	rConn.Do("HSET", "MP_" + uploadInfo.UploadID, "filehash", uploadInfo.FileHash)
	rConn.Do("HSET", "MP_" + uploadInfo.UploadID, "filesize", uploadInfo.FilSize)

	// 5 将初始化信息返回客户端
	w.Write(util.NewRespMsg(http.StatusOK, "OK", uploadInfo).JSONBytes())
}

// 上传文件分块
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// 1 解析用户请求参数
	r.ParseForm()
	// username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")

	// 2 获取redis的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3 获取文件句柄，存储当前文件块
	filepath := "/data/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(filepath), 0744) // 首先创建文件path指定权限
	f, err := os.Create(filepath) // 然后创建文件获得文件句柄
	if err != nil {
		log.Println(err.Error())
		w.Write(util.NewRespMsg(http.StatusInternalServerError, "Upload part failed", nil).JSONBytes())
		return
	}
	defer f.Close()

	buff := make([]byte, 1024 * 1024) // 每次读取1MB, 分块hash校验 - 和客户端文件块hash值对比判断文件块是否修改或丢失
	for {
		n, err := r.Body.Read(buff)
		f.Write(buff[0:n])
		if err != nil {
			break
		}
	}

	// 4 更新redis缓存
	rConn.Do("HSET", "MP_" + uploadID, "chkidx_" + chunkIndex, 1)

	// 5 返回处理结果给客户端
	w.Write(util.NewRespMsg(http.StatusOK, "OK", nil).JSONBytes())
}

// 进行文件块合并
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1 解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	filehash := r.Form.Get("filehash")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))
	filename := r.Form.Get("filename")

	// 2 获取redis的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3 通过uploadID查询redis，判断是否所有文件块已经完成上传
	data, err := redis.Values(rConn.Do("HGETALL", "MP_" + uploadID))
	if err != nil {
		log.Println(err.Error())
		w.Write(util.NewRespMsg(http.StatusInternalServerError, "complete upload failed", nil).JSONBytes())
		return
	}
	totalCount := 0
	chunkCount := 0
	for i := 0; i < len(data); i += 2 { // key 和 value在同一array
		key := string(data[i].([]byte))
		value := string(data[i + 1].([]byte))
		if key == "chunkcount" {
			totalCount, _ = strconv.Atoi(value)
		} else if strings.HasPrefix(key, "chkidx_") && value == "1" {
			chunkCount++
		}
	}
	if totalCount != chunkCount {
		w.Write(util.NewRespMsg(http.StatusBadRequest, "Invalid request", nil).JSONBytes())
		return
	}

	// 4 合并文件块


	// 5 更新唯一文件表和用户文件表
	db.OnFileUploadFinished(filehash, filename, int64(filesize), "")
	db.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))

	// 6 返回处理结果给客户端
	w.Write(util.NewRespMsg(http.StatusOK, "OK", nil).JSONBytes())
}
