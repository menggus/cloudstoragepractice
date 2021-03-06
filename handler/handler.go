/*
	web接口
*/
package handler

import (
	"cloudstorage/v1/db"
	"cloudstorage/v1/meta"
	"cloudstorage/v1/utils"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

//FileHandler 文件接口
func FileHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		// GET 返回文件上传 page
		// ioutil 包读取文件
		page, err := ioutil.ReadFile("static/view/index.html")
		if err != nil {
			log.Fatal(err)
			io.WriteString(w, "访问的页面不触存在")
			return
		}
		io.WriteString(w, string(page))

	} else if r.Method == "POST" {
		// POST 接收文件存放道本地
		ff, head, err := r.FormFile("file")
		if err != nil {
			log.Printf("Faild recive file data: %s\n", err)
			return
		}
		defer ff.Close()
		// 上传文件元信息
		fileMeta := meta.FileMeta{
			FileName:   head.Filename,
			FilePath:   "tmp/" + head.Filename,
			UploadTime: time.Now().Format("2006-01-02 15:04:05"),
		}

		nf, err := os.Create(fileMeta.FilePath)
		if err != nil {
			log.Printf("Failed create local new file: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer nf.Close()
		// 上传文件大小
		fileMeta.FileSize, err = io.Copy(nf, ff)
		if err != nil {
			log.Printf("Failed save upload file: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// 在计算大文件的hash值时，是非常花费时间的，可以抽取为独立的微服务进行异步处理
		// 上传文件的sha1值
		nf.Seek(0, 0) // 游标重新回到文件头部
		fileMeta.FileSha1 = utils.FileSha1(nf)

		// 上传文件到用户表
		username := r.FormValue("username")
		ok := db.TabUserFileInsert(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
		if !ok {
			log.Printf("user file insert failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// 上传存储元信息
		//meta.UpdateFileMetas(fileMeta)
		ok = meta.UpdateFileMetasDB(fileMeta)
		if !ok {
			log.Printf("File upload failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/user/home", http.StatusFound)
	}
}

// SuccedHandler 上传成功跳转提示接口
func SuccedHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "文件上传成功")
}

// QueryFileInfoHandler 查询 file 信息接口
// url： /file/meta?sha1=3bc5f45eb1cf75eff7f3e56c514748a11e84cdba
func QueryFileInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		sha1 := r.FormValue("sha1")

		//filemeta := meta.GetFileMeta(sha1)
		filemeta := meta.GetFileMetaDB(sha1)

		fj, err := json.Marshal(filemeta)
		if err != nil {
			log.Printf("Failed change to json: %s\n", err)
			return
		}
		io.WriteString(w, string(fj))
	}
}

// DownloadFileHandler 下载文件接口
func DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// r.ParseForm()
		sha1 := r.FormValue("sha1")
		//filemeta := meta.GetFileMeta(sha1)
		filemeta := meta.GetFileMetaDB(sha1)
		// 这里只是针对于小文件的读取，大文件需要以流的方式来读
		data, err := ioutil.ReadFile(filemeta.FilePath)
		if err != nil {
			log.Printf("File Not Found: %s\n", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// 返回文件数据设置header
		w.Header().Set("Content-Type", "application/octect-stream")
		w.Header().Set("content-disposition", fmt.Sprintf("attachment;filename=\"%s\"", filemeta.FileName))
		w.Write(data)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// RenameHandler 重命名文件名
func RenameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PUT" {
		sha1 := r.FormValue("sha1")
		op := r.FormValue("op")
		rname := r.FormValue("name")
		if op != "0" {
			log.Println("操作无效........")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		file := meta.GetFileMeta(sha1)
		file.FileName = rname

		// 更新 filemeta
		meta.UpdateFileMetas(file)

		data, err := json.Marshal(file)
		if err != nil {
			log.Printf("Rename Failed：%s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		io.WriteString(w, string(data))

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// DeleteFileHandler 删除文件接口
func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		sha1 := r.FormValue("sha1")
		ok := meta.DeletaFileMeta(sha1)
		if !ok {
			log.Printf("delete file failed\n")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// 云存储中文件的删除基本上都是 不删除实物
		io.WriteString(w, "删除成功....")
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
