package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var (
	token_maker RandStringMaker
	Users_info  map[string]Uinfo
)

func handleHtml(c *gin.Context, name string) {
	c.HTML(http.StatusOK, name, nil)
}

func handleLogin(c *gin.Context) {
	var user, pw, token string
	var ok bool
	var uinfo Uinfo

	token, err := c.Cookie("token")
	if err == nil {
		delete(Users_info, token)
		c.SetCookie("token", "", -1, "/", config.App.Domain, config.App.Https, true)
		log.Println("Delete token", token)
	}

	user, ok = c.GetPostForm("user")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"res": -1,
			"msg": "账号不能为空",
		})
		return
	}
	pw, ok = c.GetPostForm("password")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"res": -1,
			"msg": "密码不能为空",
		})
		return
	}
	uinfo, ok = dblogin(user, pw)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"res": -1,
			"msg": "账号或密码错误",
		})
		return
	}
	uinfo.Enable = user != pw
	for {
		token = token_maker.Get()
		if _, v := Users_info[token]; v {
			continue
		}
		Users_info[token] = uinfo
		break
	}
	c.SetCookie("token", token, 0, "/", config.App.Domain, config.App.Https, true)
	log.Println("Add token", token)
	if user == pw {
		c.JSON(http.StatusOK, gin.H{
			"res": 1,
			"msg": "请修改密码",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"res": 0,
		"msg": "成功",
	})
}
func handleLogout(c *gin.Context) {
	var token string
	token, err := c.Cookie("token")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"res": -1,
			"msg": "请先登录",
		})
		return
	}
	_, ok := Users_info[token]
	if !ok {
		c.SetCookie("token", "", -1, "/", config.App.Domain, config.App.Https, true)
		c.JSON(http.StatusOK, gin.H{
			"res": -1,
			"msg": "无效cookie，请重新登录",
		})
		return
	}
	delete(Users_info, token)
	c.SetCookie("token", "", -1, "/", config.App.Domain, config.App.Https, true)
	log.Println("Delete token", token)
	c.JSON(http.StatusOK, gin.H{
		"res": 0,
		"msg": "成功",
	})
}
func handleUinfo(c *gin.Context) {
	token, err := c.Cookie("token")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"res":  -1,
			"msg":  "请先登录",
			"data": "",
		})
		return
	}
	uinfo, ok := Users_info[token]
	if !ok {
		c.SetCookie("token", "", -1, "/", config.App.Domain, config.App.Https, true)
		c.JSON(http.StatusOK, gin.H{
			"res":  -1,
			"msg":  "无效cookie，请重新登录",
			"data": "",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"res": 0,
		"msg": "成功",
		"data": gin.H{
			"no":     uinfo.No,
			"name":   uinfo.Name,
			"level":  uinfo.Level,
			"enable": uinfo.Enable,
		},
	})
}
func handleChpw(c *gin.Context) {
	token, err := c.Cookie("token")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"res": -1,
			"msg": "请先登录",
		})
		return
	}
	uinfo, ok := Users_info[token]
	if !ok {
		c.SetCookie("token", "", -1, "/", config.App.Domain, config.App.Https, true)
		c.JSON(http.StatusOK, gin.H{
			"res": -1,
			"msg": "无效cookie，请重新登录",
		})
		return
	}

	pw, ok := c.GetPostForm("password")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"res": -1,
			"msg": "密码不能为空",
		})
		return
	}
	res, msg := checkPw(pw)
	if !res {
		c.JSON(http.StatusOK, gin.H{
			"res": -1,
			"msg": msg,
		})
		return
	}
	ok = dbchpw(uinfo.No, pw)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"res": -1,
			"msg": "未知错误",
		})
	}
	delete(Users_info, token)
	c.SetCookie("token", "", -1, "/", config.App.Domain, config.App.Https, true)
	log.Println("Delete token", token)

	c.JSON(http.StatusOK, gin.H{
		"res": 0,
		"msg": "成功，请重新登录",
	})
}

func handleResetpw(c *gin.Context) {
	token, err := c.Cookie("token")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"res": -1,
			"msg": "请先登录",
		})
		return
	}
	uinfo, ok := Users_info[token]
	if !ok {
		c.SetCookie("token", "", -1, "/", config.App.Domain, config.App.Https, true)
		c.JSON(http.StatusOK, gin.H{
			"res": -1,
			"msg": "无效cookie，请重新登录",
		})
		return
	}
	if !uinfo.Enable {
		c.JSON(http.StatusOK, gin.H{
			"res": -1,
			"msg": "请先改密",
		})
		return
	}
	if uinfo.Level != "1" {
		c.JSON(http.StatusOK, gin.H{
			"res": -1,
			"msg": "您没有权限",
		})
		return
	}

	user, ok := c.GetPostForm("user")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"res": -1,
			"msg": "账号不能为空",
		})
		return
	}
	pw, ok := c.GetPostForm("password")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"res": -1,
			"msg": "密码不能为空",
		})
		return
	}
	// 不检查密码
	// res, msg := checkPw(pw)
	// if !res {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"res": -1,
	// 		"msg": msg,
	// 	})
	// }
	ok = dbchpw(user, pw)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"res": -1,
			"msg": "账号不存在",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"res": 0,
		"msg": "成功",
	})
}

func checkPw(pw string) (res bool, msg string) {
	res = true
	msg = ""
	if !config.Password.Strict_mode {
		return
	}
	pw_byte := []byte(pw)

	if len(pw) < config.Password.Length {
		res = false
		msg += "密码长度不小于" + strconv.Itoa(config.Password.Length) + ";"
	}
	var u, l, d, o = 0, 0, 0, 0
	for _, c := range pw_byte {
		switch {
		case 'A' <= c && c <= 'Z':
			u++
		case 'a' <= c && c <= 'z':
			l++
		case '0' <= c && c <= '9':
			d++
		default:
			o++
		}
	}
	if u < config.Password.Upper {
		res = false
		msg += "大写字符数不应少于" + strconv.Itoa(config.Password.Upper) + ";"
	}
	if l < config.Password.Lower {
		res = false
		msg += "小写字符数不应少于" + strconv.Itoa(config.Password.Lower) + ";"
	}
	if d < config.Password.Digital {
		res = false
		msg += "数字字符数不应少于" + strconv.Itoa(config.Password.Digital) + ";"
	}
	if o < config.Password.Other {
		res = false
		msg += "特殊字符数不应少于" + strconv.Itoa(config.Password.Other) + ";"
	}
	return
}
