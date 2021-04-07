package handler

import (
	"cloudstorage/v1/db"
	"cloudstorage/v1/utils"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const Secrete = "#@$$@dosage@!#$"

// UserRegisterHandler 用户注册
func UserRegisterHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		con, err := ioutil.ReadFile("static/view/signup.html")
		if err != nil {
			log.Printf("未找到html文件")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(con)

		return

	} else if r.Method == "POST" {
		username := r.FormValue("username")
		passwrod := r.FormValue("password")

		// 校验用户的输入
		if len(username) < 3 || len(passwrod) < 6 {
			log.Printf("无效的账户密码")
			w.Write([]byte("非法的账户密码"))
			return
		}
		passwrodSecret := utils.Sha1([]byte(passwrod + Secrete))
		log.Println(passwrodSecret)

		// 调用数据库接口，插入数据
		ok := db.TabUserDataInsert(username, passwrodSecret)
		if !ok {
			log.Printf("插入用户数据失败")
			w.Write([]byte("用户注册失败，请重新尝试...."))
			return
		}
		w.Write([]byte("SUCCESS"))
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)

}

// UserLoginHandler 用户登录
func UserLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		con, err := ioutil.ReadFile("static/view/signin.html")
		if err != nil {
			log.Printf("未找到html文件")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(con)

		return
	} else if r.Method == http.MethodPost {
		// 校验账户密码
		username := r.FormValue("username")
		password := r.FormValue("password")

		oku := db.TabUserDataQuery(username, password)
		if !oku {
			log.Println("查询用户不存在")
			io.WriteString(w, "查询用户不存在")
			return
		}

		log.Println(username, password)
		// 生成token 并写入数据库中
		token := GetToken(username)

		okt := db.TabTokenDataInsert(username, token)
		if !okt {
			log.Println("token 写入失败")
			w.Write([]byte("请稍后重试"))
			return
		}

		// 返回token值写入到cookie
		//r.AddCookie()

		// 重定向首页
		w.Write([]byte("http://" + r.Host + "/static/view/home.html"))

	}
}

const token_salt = "#$#ad@!#"

func GetToken(username string) string {
	// todo 40位字符：md5(username + timestamp + token_salt) + timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	token_prefix := utils.MD5([]byte(username + ts + token_salt))
	return token_prefix + ts[:8]
}