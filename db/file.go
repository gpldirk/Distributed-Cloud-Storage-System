package db

import (
	"database/sql"
	"github.com/cloud/db/mysql"
	"log"
)

// File : 文件元信息的model
type File struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// OnFileUploadFinished : 上传新文件时在唯一文件表插入文件元信息
func OnFileUploadFinished(filehash string, filename string, filesize int64, fileaddr string) bool {
	stmt, err := mysql.DBConn().Prepare("insert ignore into tbl_file " +
		"(`file_sha1`, `file_name`, `file_size`, `file_addr`, `status`) values (?, ?, ?, ?, 1)")
	if err != nil {
		log.Println(err.Error())
		return false
	}
	defer stmt.Close()

	rf, err := stmt.Exec(filehash, filename, filesize, fileaddr)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	if rows, err := rf.RowsAffected(); err == nil {
		if rows <= 0 {
			log.Printf("File with hash:%d has been uploaded before\n", filehash)
		}
		return true
	}

	return false
}

// GetFileMeta : 指定filehash，从DB中读取对应文件的元信息
func GetFileMeta(filehash string) (*File, error) {
	stmt, err := mysql.DBConn().Prepare("select file_sha1, file_name, file_size, file_addr " +
		"from tbl_file where file_sha1=? and status=1 limit 1")
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	tfile := File{}
	err = stmt.QueryRow(filehash).Scan(&tfile.FileHash, &tfile.FileName, &tfile.FileSize, &tfile.FileAddr)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No file with hash: %s\n", filehash)
			return nil, nil
		} else {
			log.Println(err.Error())
			return nil, err
		}
	}

	return &tfile, nil
}

// IsFileUploaded : 文件是否已经上传过
func IsFileUploaded(filehash string) bool {
	stmt, err := mysql.DBConn().Prepare("select 1 from tbl_file where file_sha1=? and status=1 limit 1")
	if err != nil {
		log.Println(err.Error())
		return false
	}
	defer stmt.Close()

	rows, err := stmt.Query(filehash)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	if rows == nil || !rows.Next() {
		return false
	} else {
		return true
	}
}

// GetFileMetaList : 从mysql批量获取文件元信息
func GetFileMetaList(limit int) ([]File, error) {
	stmt, err := mysql.DBConn().Prepare("select file_sha1, file_name, file_size, file_addr from tbl_file " +
		"where status=1 limit ?")
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(limit)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	columns, _ := rows.Columns()
	values := make([]sql.RawBytes, len(columns))
	var files []File
	for i := 0; i < len(values) && rows.Next(); i++ {
		file := File{}
		err := rows.Scan(&file.FileHash, &file.FileName, &file.FileSize, &file.FileAddr)
		if err != nil {
			log.Println(err.Error())
			break
		}

		files = append(files, file)
	}

	return files, nil
}

// OnFileRemoved : 文件删除(这里只做标记删除，即改为status=2)
func OnFileRemoved(filehash string) bool {
	stmt, err := mysql.DBConn().Prepare("update tbl_file set status=2 where file_sha1=? and status=1 limit 1")
	if err != nil {
		log.Println(err.Error())
		return false
	}
	defer stmt.Close()

	rf, err := stmt.Exec(filehash)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	if rows, err := rf.RowsAffected(); err == nil {
		if rows <= 0 {
			log.Printf("File with hash:%d has not been uploaded\n", filehash)
		}
		return true
	}

	return false
}
