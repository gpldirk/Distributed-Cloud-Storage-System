package handler

import "C"
import (
	"database/sql"
	"encoding/json"
	"github.com/cloud/common"
	"github.com/cloud/config"
	"github.com/cloud/db"
	"github.com/cloud/meta"
	"github.com/cloud/mq"
	"github.com/cloud/store/ceph"
	"github.com/cloud/store/oss"
	"github.com/cloud/util"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func init() {
	if err := os.MkdirAll(config.TempLocalRootDir, 0744); err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
	if err := os.MkdirAll(config.MergeLocalRootDir, 0744); err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
}

// UploadHandler : 响应文件上传get请求
func UploadHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/index.html")
	return
}

// DoUploadHandler : 响应文件上传post请求
func DoUploadHandler(c *gin.Context) {
	// 客户端以表单形式提交文件
	// 接收文件流，存储到本地目录
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusFound, gin.H{
			"msg": "Failed to upload file to ceph",
			"code": -1,
		})
		return
	}
	defer file.Close()

	// 创建文件元信息对象
	tmpPath := config.TempLocalRootDir + header.Filename
	fileMeta := meta.FileMeta{
		FileName: header.Filename,
		Location: tmpPath,
		UploadAt: time.Now().Format("2016-01-02 15:04:05"),
	}

	// 创建本地文件获取句柄
	newFile, err := os.Create(fileMeta.Location)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer newFile.Close()

	// 将接收文件流copy到本地文件中
	fileMeta.FileSize, err = io.Copy(newFile, file)
	if err != nil {
		log.Println(err.Error())
		return
	}

	newFile.Seek(0, 0)
	fileMeta.FileSha1 = util.FileSha1(newFile)

	newFile.Seek(0, 0)
	mergePath := config.MergeLocalRootDir + fileMeta.FileSha1
	// 将文件以同步/异步方式转移到Ceph/OSS
	if config.CurrentStoreType == common.StoreCeph {
		// 文件写入ceph
		data, _ := ioutil.ReadAll(newFile)
		cephPath := "/ceph/" + fileMeta.FileSha1
		err = ceph.PutObject("userfile", cephPath, data)
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusOK, gin.H{
				"msg": "Failed to get ceph bucket",
				"code": -1,
			})
			return
		}
		fileMeta.Location = cephPath
	} else if config.CurrentStoreType == common.StoreOSS {
		// 文件写入OSS
		ossPath := "oss/" + fileMeta.FileSha1

		// 文件的同步转移逻辑
		if !config.AsyncTransferEnable {
			err = oss.Bucket().PutObject(ossPath, newFile)
			if err != nil {
				log.Println(err.Error())
				c.JSON(http.StatusOK, gin.H{
					"msg": "Failed to get oss bucket",
					"code": -1,
				})
				return
			}
			fileMeta.Location = ossPath
		} else {
			// 文件异步转移初始阶段存储在本地
			fileMeta.Location = mergePath

			// 借助转移队列的异步转移逻辑
			data := mq.TransferData{
				FileHash:      fileMeta.FileSha1,
				Location:      fileMeta.Location,
				DestLocation:  ossPath,
				DestStoreType: common.StoreOSS,
			}
			pubData, err := json.Marshal(data)
			if err != nil {
				log.Println(err.Error())
				// TODO: 写入重试队列进行消息的再次consume
			}
			if success := mq.Publish(config.TransExchangeName, config.TransOSSRoutingKey, pubData); !success {
				// TODO: 写入重试队列进行消息的再次consume
			}
		}
	}

	// (普通上传/分块上传) 文件统一存储在mergePath
	err = os.Rename(tmpPath, mergePath)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg": "Merge local file failed",
			"code": -1,
		})
		return
	}

	// 写入唯一文件表
	if success := meta.UpdateFileMetaDB(fileMeta); !success {
		c.JSON(http.StatusOK, gin.H{
			"msg": "Update file meta from DB failed",
			"code": -1,
		})
		return
	}

	// 写入用户文件表
	username := c.Request.FormValue("username")
	if success := db.OnUserFileUploadFinished(username, fileMeta.FileSha1,
		fileMeta.FileName, fileMeta.FileSize); !success {
		c.JSON(http.StatusOK, gin.H{
			"msg": "Update user file table from DB failed",
			"code": -1,
		})
		return
	}

	c.Redirect(http.StatusFound, "/static/view/home.html")
}

// UploadSucHandler : 返回上传文件成功页面
func UploadSucHandler(c * gin.Context) {
	c.String(http.StatusOK, "Upload Finished!")
}

// GetFileMetaHandler : 通过指定filehash获取对应文件的元信息
func GetFileMetaHandler(c *gin.Context) {
	filehash := c.Request.FormValue("filehash")
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg": "Get file meta from DB failed",
			"code": -1,
		})
		return
	}

	data, err := json.Marshal(fileMeta)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg": "JSON parse failed",
			"code": -1,
		})
		return
	}

	c.Data(http.StatusOK, "application/json", data)
}

// FileQueryHandler : 通过指定limit获取最近上传文件生成的文件元信息
func FileQueryHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	limitCnt, err := strconv.Atoi(c.Request.FormValue("limit"))
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg": "Request parse failed",
			"code": -1,
		})
		return
	}

	// 获取 last limit file Metas from DB or map
	// fileMetas := meta.GetLastFileMetas(limitCnt)
	fileMetas, err := db.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg": "Query db failed",
			"code": -1,
		})
		return
	}

	// json 序列化
	data, err := json.Marshal(fileMetas)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg": "JSON parse failed",
			"code": -1,
		})
		return
	}

	c.Data(http.StatusOK, "application/json", data)
}

// UpdateFileMetaHandler : 更新文件元信息，通过op指定更新类型(op = 1 -> 修改文件名)
func UpdateFileMetaHandler(c *gin.Context) {
	opType := c.Request.FormValue("op")
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	newFileName := c.Request.FormValue("filename")
	if opType != "0" || len(newFileName) < 1 {
		c.JSON(http.StatusOK, gin.H{
			"msg": "Invalid operation",
			"code": -1,
		})
		return
	}
	if c.Request.Method != "POST" {
		c.JSON(http.StatusOK, gin.H{
			"msg": "Invalid method",
			"code": -1,
		})
		return
	}

	// 更新用户文件表中的文件名，不用更新唯一文件表
	if success := db.RenameFileName(username, filehash, newFileName); !success {
		c.JSON(http.StatusOK, gin.H{
			"msg": "DB update failed",
			"code": -1,
		})
		return
	}

	// 获取最新的文件元信息
	userFile, err := db.QueryUserFileMeta(username, filehash)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg": "DB query failed",
			"code": -1,
		})
		return
	}

	// 将fileMeta序列化返回给用户
	data, err := json.Marshal(userFile)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg": "JSON parse failed",
			"code": -1,
		})
		return
	}

	c.Data(http.StatusOK, "application/json", data)
}

// DeleteFileHandler :  删除文件及其元信息
func DeleteFileHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")

	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg": "Get file meta from DB failed",
			"code": -1,
		})
		return
	}

	// 删除本地文件
	os.Remove(fileMeta.Location)
	// TODO: 可考虑删除Ceph/OSS上的文件
	// 可以不立即删除，加个超时机制，
	// 比如该文件10天后也没有用户再次上传，那么就可以真正的删除了


	/// 删除用户文件表中的一条记录
	if success := db.DeleteUserFile(username, filehash); !success {
		c.JSON(http.StatusOK, gin.H{
			"msg": "Delete user file from DB failed",
			"code": -1,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "Delete file successfully",
		"code": 0,
	})
}

// TryFastUploadHandler : 尝试秒传接口(判断当前文件是否上传过)
func TryFastUploadHandler(c *gin.Context) {
	// 解析请求参数
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filename := c.Request.FormValue("filename")
	filesize, _ := strconv.Atoi(c.Request.FormValue("filesize"))

	// 在唯一文件表中查询对应文件是否存在
	fileMeta, err := meta.GetFileMetaDB(filehash)

	// 查询不到则返回秒传失败
	if err != nil && err != sql.ErrNoRows {
		log.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg": "Get file meta from DB failed",
			"code": -1,
		})
		return
	}

	if err == sql.ErrNoRows || fileMeta.FileSha1 == "" {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请使用普通上传接口",
		}
		c.Data(http.StatusOK, "application/json", resp.JSONBytes())
		return
	}

	// 上传过当前文件，则将文件信息和用户信息写入用户文件表(fast upload)
	if success := db.OnUserFileUploadFinished(username, filehash, filename, int64(filesize)); success {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		c.Data(http.StatusOK, "application/json", resp.JSONBytes())
	} else {
		resp := util.RespMsg{
			Code: -2,
			Msg:  "秒传失败，请稍后重试",
		}
		c.Data(http.StatusOK, "application/json", resp.JSONBytes())
	}
	return
}
