package db

import (
	"github.com/cloud/db/mysql"
	"log"
)

// User :  用户表的model
type User struct {
	Username string
	Email string
	Phone string
	SignUpAt string
	LastActiveAt string
	Status int
}

// UserSignUp : 注册新用户时更新user table
func UserSignUp(username, password string) bool {
	stmt, err := mysql.DBConn().Prepare("insert ignore into tbl_user (`user_name`, `user_pwd`) values (?, ?)")
	if err != nil {
		log.Println(err.Error())
		return false
	}
	defer stmt.Close()

	rf, err := stmt.Exec(username, password)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	if rows, err := rf.RowsAffected(); err == nil && rows > 0{
		return true
	} else {
		return false
	}
}

// UserSignIn : 读取数据库得到username对应的PWD进行校验
func UserSignIn(username, encodedPWD string) bool {
	stmt, err := mysql.DBConn().Prepare("select * from tbl_user where user_name=? limit 1")
	if err != nil {
		log.Println(err.Error())
		return false
	}
	defer stmt.Close()

	rows, err := stmt.Query(username)
	if err != nil {
		log.Println(err.Error())
		return false
	} else if rows == nil {
		log.Println("Username not found")
		return false
	}

	pRows := mysql.ParseRows(rows)
	if len(pRows) > 0 && string(pRows[0]["user_pwd"].([]byte)) == encodedPWD {
		return true
	}

	return false
}

// GetUserInfo : 从DB中读取对应username的信息
func GetUserInfo(username string) (User, error) {
	user := User{}
	stmt, err := mysql.DBConn().Prepare("select user_name, signup_at from tbl_user where user_name=? limit 1")
	if err != nil {
		log.Println(err.Error())
		return user, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignUpAt)
	if err != nil {
		log.Println(err.Error())
		return user, err
	}

	return user, nil
}

// UpdateToken : 更新数据库中username所有的token信息
func UpdateToken(username, token string) bool {
	stmt, err := mysql.DBConn().Prepare("replace into tbl_user_token (`user_name`, `user_token`) values (?, ?)")
	if err != nil {
		log.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, token)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	return true
}

// GetUserToken : 获取用户登录token
func GetUserToken(username string) string {
	stmt, err := mysql.DBConn().Prepare("select user_token from tbl_user_token where user_name=? limit 1")
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	defer stmt.Close()

	var token string
	err = stmt.QueryRow(username).Scan(&token)
	if err != nil {
		log.Println(err.Error())
		return ""
	}

	return token
}
