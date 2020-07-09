package handler

import (
	"fmt"
	"github.com/cloud/config"
	"github.com/cloud/db"
	"github.com/cloud/meta"
	"github.com/cloud/store/ceph"
	"github.com/cloud/store/oss"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// DownloadHandler : 某个用户下载某个文件
func DownloadHandler(c *gin.Context) {
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
	userFile, err := db.QueryUserFileMeta(username, filehash)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg": "Query user file meta from DB failed",
			"code": -1,
		})
		return
	}

	var Data []byte
	// 文件存储在本地
	if strings.HasPrefix(fileMeta.Location, config.MergeLocalRootDir) {
		fmt.Println("To download file from local directory")
		f, err := os.Open(fileMeta.Location)
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusOK, gin.H{
				"msg": "Failed to open local file",
				"code": -1,
			})
			return
		}
		defer f.Close()

		Data, err = ioutil.ReadAll(f)
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusOK, gin.H{
				"msg": "Failed to read local file",
				"code": -1,
			})
			return
		}
	} else if strings.HasPrefix(fileMeta.Location, "/ceph") {
		fmt.Println("To download file from ceph")
		bucket := ceph.GetCephBucket("userfile")
		Data, err = bucket.Get(fileMeta.Location)
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusOK, gin.H{
				"msg": "Failed to get object from ceph bucket",
				"code": -1,
			})
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
			c.JSON(http.StatusOK, gin.H{
				"msg": "Failed to read file from oss",
				"code": -1,
			})
			return
		}
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg": "File not found",
			"code": -1,
		})
		return
	}

	c.Header("content-disposition", "attachment; filename=\"" + userFile.FileName + "\"")
	c.Data(http.StatusOK, "application/octect-stream", Data)
}

// DownloadURLHandler : 获取下载文件的URL
func DownloadURLHandler(c *gin.Context) {
	filehash := c.Request.FormValue("filehash")
	fileMeta, _ := db.GetFileMeta(filehash)

	// 判断文件存在于本地/ceph，还是OSS
	if strings.HasPrefix(fileMeta.FileAddr.String, config.MergeLocalRootDir) ||
		strings.HasPrefix(fileMeta.FileAddr.String, "/ceph") {
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")
		tmpURL := fmt.Sprintf("http://%s/file/download?username=%s&filehash=%s&token=%s", c.Request.Host, username, token)
		c.String(http.StatusOK, tmpURL)
	} else if strings.HasPrefix(fileMeta.FileAddr.String, "oss/") {
		signedURL := oss.DownloadURL(fileMeta.FileAddr.String)
		c.String(http.StatusOK, signedURL)
	} else {
		c.String(http.StatusOK, "Error: 暂时无法生成下载链接")
	}
}

// RangeDownloadHandler : 支持断点的文件下载接口
func RangeDownloadHandler(c *gin.Context) {
	filehash := c.Request.FormValue("filehash")
	username := c.Request.FormValue("username")
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg": "Get file meta from DB failed",
			"code": -1,
		})
		return
	}
	userFile, err := db.QueryUserFileMeta(username, filehash)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg": "Query user file from DB failed",
			"code": -1,
		})
		return
	}

	// 使用本地文件目录
	filePath := config.MergeLocalRootDir + fileMeta.FileSha1
	fmt.Println("range-download-file-path: " + filePath)

	f, err := os.Open(filePath)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg": "Failed to open local file",
			"code": -1,
		})
		return
	}
	defer f.Close()

	c.Header("Content-Type", "application/octect-stream")
	// attachment表示文件将会提示下载到本地，而不是直接在浏览器中打开
	c.Header("content-disposition", "attachment; filename=\"" + userFile.FileName + "\"")
	http.ServeFile(c.Writer, c.Request, filePath)
}
