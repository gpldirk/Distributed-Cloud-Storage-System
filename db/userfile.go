package db

import (
	"github.com/cloud/db/mysql"
	"log"
	"time"
)

// UserFile ： 用户文件表的model, 记录用户和文件的关联关系
type UserFile struct {
	UserName string
	FileHash string
	FileName string
	FileSize int64
	UploadAt string
	LastUpdated string
}

// OnUserFileUploadFinished : 更新用户文件表
func OnUserFileUploadFinished(username, filehash, filename string, filesize int64) bool {
	stmt, err := mysql.DBConn().Prepare("insert ignore into tbl_user_file (`user_name`, `file_sha1`, `file_name`, " +
		"`file_size`, `upload_at`, `status`) values (?, ?, ?, ?, ?, 1)")
	if err != nil {
		log.Println(err.Error())
		return false
	}
	defer stmt.Close()

	rf, err := stmt.Exec(username, filehash, filename, filesize, time.Now())
	if err != nil {
		log.Println(err.Error())
		return false
	}
	if rows, err := rf.RowsAffected(); err == nil {
		if rows <= 0 {
			log.Printf("File with hash:%s has been uploaded before\n", filehash)
		}
		return true
	}
	return false
}

// QueryUserFileMetas : 批量获取指定用户拥有的文件元信息
func QueryUserFileMetas(username string, limit int) ([]UserFile, error) {
	stmt, err := mysql.DBConn().Prepare("select file_sha1, file_name, file_size, upload_at, last_update " +
		"from tbl_user_file where user_name=? and status!=2 limit ?")
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(username, limit)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	var userFiles []UserFile
	for rows.Next() {
		userFile := UserFile{}
		err := rows.Scan(&userFile.FileHash, &userFile.FileName, &userFile.FileSize, &userFile.UploadAt, &userFile.LastUpdated)
		if err != nil {
			log.Println(err.Error())
			break
		}
		userFiles = append(userFiles, userFile)
	}

	return userFiles, nil
}

// RenameFileName : 文件重命名
func RenameFileName(username, filehash, filename string) bool {
	stmt, err := mysql.DBConn().Prepare("update tbl_user_file set file_name=? where user_name=? and file_sha1=? limit 1")
	if err != nil {
		log.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(filename, username, filehash)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	return true
}

// DeleteUserFile : 删除文件(标记删除)
func DeleteUserFile(username, filehash string) bool {
	stmt, err := mysql.DBConn().Prepare("update tbl_user_file set status=2 where user_name=? and file_sha1=? limit 1")
	if err != nil {
		log.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, filehash)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	return true
}

// QueryUserFileMeta : 获取当前用户单个文件信息
func QueryUserFileMeta(username string, filehash string) (*UserFile, error) {
	stmt, err := mysql.DBConn().Prepare("select file_sha1, file_name, file_size, upload_at, last_update " +
		"from tbl_user_file where user_name=? and file_sha1=? limit 1")
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	userFile := UserFile{}
	err = stmt.QueryRow(username, filehash).Scan(&userFile.FileHash, userFile.FileName, userFile.FileSize, userFile.UploadAt, userFile.LastUpdated)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return &userFile, nil
}

// IsUserFileUploaded : 判断当前user是否上传过当前文件
func IsUserFileUploaded(username, filehash string) bool {
	stmt, err := mysql.DBConn().Prepare("select 1 from tbl_user_file where user_name=? and file_sha1=? and status=1 limit 1")
	if err != nil {
		log.Println(err.Error())
		return false
	}
	rows, err := stmt.Query(username, filehash)
	if err != nil {
		log.Println(err.Error())
		return false
	} else if rows == nil || !rows.Next() {
		return false
	}

	return true
}


