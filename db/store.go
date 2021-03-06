package db

import (
	mydb "cloudstorage/v1/db/mysql"
	"database/sql"
	"log"
)

// TabFileDataInsert  向表tabfile插入数据
func TabFileDataInsert(fileSha1 string, fileName string, fileSize int64, fileAddr string) bool {
	// 准备sql
	// 这里也有防止sql注入的情况
	stmt, err := mydb.DBConnect().Prepare("INSERT ignore into tabfile (`file_sha1`, `file_name`, `file_size`," +
		" `file_addr`, `status`) values (?, ?, ?, ?, 1)")
	if err != nil {
		log.Printf("Failed to prepare statement: %s\n", err)
		return false
	}

	// 关闭连接
	defer stmt.Close()

	// 执行sql
	ret, err := stmt.Exec(fileSha1, fileName, fileSize, fileAddr)
	if err != nil { // sql 执行失败
		log.Printf("err: %s\n", err)
		return false
	}

	if rf, err := ret.RowsAffected(); err == nil {
		if rf < 0 { // rf < 0 sql 插入数据失败
			log.Printf("File with hash[%s] already exists\n", fileSha1)
			return false
		}
	}
	return true
}

// TabFile file表结构体
type TabFile struct {
	FileSha1 string
	FileName sql.NullString
	FileAddr sql.NullString
	FileSize sql.NullInt64
}

// TabFileDataQuery 从数据库中获取文件信息
func TabFileDataQuery(fileSha1 string) (*TabFile, error) {
	stmt, err := mydb.DBConnect().Prepare("SELECT file_sha1, file_name, file_addr, file_size from tabfile " +
		"WHERE file_sha1=? AND status=1 LIMIT 1")
	if err != nil {
		log.Printf("Failed query sql: %s\n", err)
		return nil, err
	}
	defer stmt.Close()
	file := TabFile{}
	err = stmt.QueryRow(fileSha1).Scan(&file.FileSha1, &file.FileName, &file.FileAddr, &file.FileSize)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Not Found source: %s\n", err)
			return nil, err
		}
		log.Printf("Query Failed: %s\n", err)
		return nil, err
	}
	return &file, nil
}

// TabFileDataDelete 删除表中的数据，逻辑删除
func TabFileDataDelete(fileSha1 string, act int8) bool {
	if act == 0 { // 逻辑删除
		// 准备sql
		stmt, err := mydb.DBConnect().Prepare("UPDATE tabfile SET status=0 WHERE file_sha1=?")
		if err != nil {
			log.Printf("Failed to prepare statement: %s\n", err)
			return false
		}
		// 关闭连接
		defer stmt.Close()

		// 执行sql
		ret, err := stmt.Exec(fileSha1)
		if err != nil { // sql 执行失败
			log.Printf("err: %s\n", err)
			return false
		}

		if rf, err := ret.RowsAffected(); err == nil {
			if rf < 0 { // rf < 0 sql 更新数据失败
				log.Printf("File with hash[%s] update failed\n", fileSha1)
				return false
			}
			log.Printf("File update succeed %d\n", rf)
		}
		return true

	} else if act == 1 {
		// todo  物理删除

		return true
	}
	log.Println("Not allow delete action")
	return false
}
