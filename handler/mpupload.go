package handler

import (
	"fmt"
	"github.com/cloud/config"
	"github.com/cloud/db"
	"github.com/cloud/util"
	rPool "github.com/cloud/cache/redis"
	"github.com/gin-gonic/gin"
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

// MultipartUploadInfo : 分块上传时文件元信息
type MultipartUploadInfo struct {
	FileHash string
	FilSize int
	UploadID string
	ChunkSize int
	ChunkCount int
	// 已经上传完成的文件块index
	ChunkExists []int
}

const (
	// ChunkDir : 上传文件块存储路径
	ChunkDir = config.ChunkLocalRootDir
	// MergeDir : 合并文件块存储路径
	MergeDir = config.MergeLocalRootDir
	// ChunkKeyPrefix : 文件块元信息在redis中存储时key的前缀
	ChunkKeyPrefix = "MP_"
	// 文件hash映射uploadID对应的redis中key的前缀
	UploadIDKeyPrefix = "UPLOAD_ID_KEY_PREFIX"
)

func init() {
	if err := os.MkdirAll(ChunkDir, 0744); err != nil {
		log.Println("Failed to create chunk file directory")
		os.Exit(1)
	}
	if err := os.MkdirAll(MergeDir, 0744); err != nil {
		log.Println("Failed to create merge file directory")
		os.Exit(1)
	}
}

// InitialMultipartUploadHandler : 初始化分块上传
func InitialMultipartUploadHandler(c *gin.Context) {
	// 1 解析用户请求
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filesize, err := strconv.Atoi(c.Request.FormValue("filesize"))
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg": "Invalid parameters",
			"code": -1,
		})
		return
	}
	// 判断文件是否存在
	if db.IsUserFileUploaded(username, filehash) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "File has been uploaded before",
			"code": -1,
		})
		return
	}

	// 2 获取redis连接
	rConn := rPool.Pool().Get()
	defer rConn.Close()

	// 3 通过filehash获取uploadID, 判断是否进行断点续传
	uploadID := ""
	keyExists, _ := redis.Bool(rConn.Do("EXISTS", UploadIDKeyPrefix + filehash))
	if keyExists {
		uploadID, err = redis.String(rConn.Do("GET", UploadIDKeyPrefix + filehash))
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusOK, gin.H{
				"msg": "Internal server error",
				"code": -1,
			})
			return
		}
	}

	// 4.1 首次上传则新建uploadID
	// 4.2 断点续传根据uploadID获取已经上传的文件块索引列表
	var chunkExists []int
	if len(uploadID) == 0 {
		uploadID = username + fmt.Sprintf("%x", time.Now().UnixNano())
	} else {
		chunks, err := redis.Values(rConn.Do("HGETALL", ChunkKeyPrefix + uploadID))
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusOK, gin.H{
				"msg": "Internal server error",
				"code": -1,
			})
			return
		}

		for i := 0; i < len(chunks); i += 2 {
			key := string(chunks[i].([]byte))
			value := string(chunks[i + 1].([]byte))
			if strings.HasPrefix(key, "chkidx_") && value == "1" {
				// chkidx_6 -> 6
				chunkIndex, _ := strconv.Atoi(key[7:])
				chunkExists = append(chunkExists, chunkIndex)
			}
		}
	}

	// 5 生成分块上传初始化信息
	uploadInfo := MultipartUploadInfo {
		FileHash:   filehash,
		FilSize:    filesize,
		UploadID:   uploadID, // username + timestamp
		ChunkSize:  5 * 1024 * 1024, // 5MB
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
		ChunkExists: chunkExists,
	}
	
	// 6 首次上传文件，将初始化信息写入redis
	if len(uploadInfo.ChunkExists) <= 0 {
		hkey := ChunkKeyPrefix + uploadInfo.UploadID
		rConn.Do("HSET", hkey, "chunkcount", uploadInfo.ChunkCount)
		rConn.Do("HSET", hkey, "filehash", uploadInfo.FileHash)
		rConn.Do("HSET", hkey, "filesize", uploadInfo.FilSize)
		rConn.Do("EXPIRE", hkey, 43200)
		rConn.Do("Set", UploadIDKeyPrefix + filehash, uploadInfo.UploadID, "EX", 43200) // 半天过期时间
	}

	// 7 将初始化信息返回客户端
	c.Data(http.StatusOK, "application/json", util.NewRespMsg(http.StatusOK, "OK", uploadInfo).JSONBytes())
}

// UploadPartHandler : 上传文件块，并保存其元信息到redis
func UploadPartHandler(c *gin.Context) {
	// 1 解析用户请求参数
	// username := c.Request.FormValue("username")
	uploadID := c.Request.FormValue("uploadid")
	chunkhash := c.Request.FormValue("chkhash") // 校验文件块是否完整
	chunkIndex := c.Request.FormValue("index")

	// 2 获取redis的一个连接
	rConn := rPool.Pool().Get()
	defer rConn.Close()

	// 3 获取文件句柄，存储当前文件块
	filepath := ChunkDir + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(filepath), 0744) // 首先创建文件path指定权限
	f, err := os.Create(filepath) // 然后创建文件获得文件句柄
	if err != nil {
		log.Println(err.Error())
		c.Data(http.StatusOK, "application/json", util.NewRespMsg(http.StatusInternalServerError, "Upload part failed", nil).JSONBytes())
		return
	}
	defer f.Close()

	buff := make([]byte, 1024 * 1024) // 每次读取1MB, 分块hash校验 - 和客户端文件块hash值对比判断文件块是否修改或丢失
	for {
		n, err := c.Request.Body.Read(buff)
		f.Write(buff[:n])
		if err != nil {
			break
		}
	}

	// 校验文件块hash值
	cmpHash, err := util.ComputeSha1ByShell(filepath)
	if err != nil || cmpHash != chunkhash {
		log.Printf("Verify chunk failed, computing hash: %s, chunk hash: %s\n", cmpHash, chunkhash)
		c.JSON(http.StatusOK, gin.H{
			"msg": "Verify chunk hash failed",
			"code": -1,
		})
		return
	}

	// 4 将文件块元信息写入redis
	rConn.Do("HSET", ChunkKeyPrefix + uploadID, "chkidx_" + chunkIndex, 1)

	// 5 返回处理结果给客户端
	c.JSON(http.StatusOK, gin.H{
		"msg": "OK",
		"code": 0,
	})
}

// CompleteUploadHandler : 进行文件块合并
func CompleteUploadHandler(c *gin.Context) {
	// 1 解析请求参数
	username := c.Request.FormValue("username")
	uploadID := c.Request.FormValue("uploadid")
	filehash := c.Request.FormValue("filehash")
	filesize, _ := strconv.Atoi(c.Request.FormValue("filesize"))
	filename := c.Request.FormValue("filename")

	// 2 获取redis的一个连接
	rConn := rPool.Pool().Get()
	defer rConn.Close()

	// 3 通过uploadID查询redis，判断是否所有文件块已经完成上传
	data, err := redis.Values(rConn.Do("HGETALL", ChunkKeyPrefix + uploadID))
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg": "Internal server error",
			"code": -1,
		})
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
		c.JSON(http.StatusOK, gin.H{
			"msg": "Invalid request",
			"code": -1,
		})
		return
	}

	// 4 合并文件块
	if success := util.MergeChuncksByShell(ChunkDir + uploadID, MergeDir + filehash, filehash); !success {
		log.Println("Merge chunk files failed")
		c.JSON(http.StatusOK, gin.H{
			"msg": "Merge chunk files failed",
			"code":  -1,
		})
		return
	}

	// 5 更新唯一文件表和用户文件表
	db.OnFileUploadFinished(filehash, filename, int64(filesize), MergeDir + filehash)
	db.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))

	// 删除redis分块信息
	_, delHashErr := rConn.Do("DEL", UploadIDKeyPrefix + filehash)
	delUploadID, delUploadInfoErr := redis.Int64(rConn.Do("DEL", ChunkKeyPrefix + uploadID))
	if delUploadID != 1 || delHashErr != nil || delUploadInfoErr != nil {
		log.Println("Failed to delete meta data from redis")
		c.JSON(http.StatusOK, gin.H{
			"msg": "Failed to delete meta data from redis",
			"code": -1,
		})
		return
	}

	// 删除已经上传的文件块
	if delRes := util.RemovePathByShell(ChunkDir + uploadID); ! delRes {
		log.Printf("Failed to delete chunk files with uploadID: %s\n", uploadID)
		c.JSON(http.StatusOK, gin.H{
			"msg": "Failed to delete local file",
			"code": -1,
		})
	}

	// 6 返回处理结果给客户端
	c.JSON(http.StatusOK, gin.H{
		"msg": "ok",
		"code": 0,
	})
}

// CancelUploadHandler : 取消文件分块上传
func CancelUploadHandler(c *gin.Context) {
	// 1 解析请求参数
	filehash := c.Request.FormValue("filehash")

	// 2 获取redis的一个连接
	rConn := rPool.Pool().Get()
	defer rConn.Close()

	// 3 检测uploadID是否存在，如果存在则删除
	uploadID, err := redis.String(rConn.Do("GET", UploadIDKeyPrefix + filehash))
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg": "Internal server error",
			"code": -1,
		})
		return
	}
	_, delHashErr := rConn.Do("DEL", UploadIDKeyPrefix + filehash)
	_, delUploadInfoErr := rConn.Do("DEL", ChunkKeyPrefix + uploadID)
	if delHashErr != nil || delUploadInfoErr != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg": "Internal server error",
			"code": -1,
		})
		return
	}

	// 4 删除已经上传的文件块
	if delChunkSuc := util.RemovePathByShell(ChunkDir + uploadID); !delChunkSuc {
		// 如果删除失败，可以后期定期清理，无须返回错误信息给用户
		log.Printf("Failed to delete chunks with uploadID: %s\n", uploadID)
	}

	// 5 返回处理结果给用户
	c.JSON(http.StatusOK, gin.H{
		"msg": "ok",
		"code": 0,
	})
}
