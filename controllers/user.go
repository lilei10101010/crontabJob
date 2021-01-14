package controllers

import (
	"cronJob/common"
	"cronJob/models"
	"cronJob/usual"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	path2 "path"
	"strings"
	"time"
)

func EmailServer(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		c.Redirect(302, "/LoginPage")
		return
	}

	smtpServer, _ := usual.SmtpConfig("all")

	c.HTML(http.StatusOK, "emailServer.html", gin.H{
		"seoTitle":   "邮件设置",
		"smtpServer": smtpServer,
		"totalCount": len(smtpServer),
		"path":       fmt.Sprintf("%v", c.Request.URL.Path),
	})
}
func DeleteSmtp(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		c.Redirect(302, "/LoginPage")
		return
	}

	var code int
	var tips string
	account := c.DefaultPostForm("account", "")

	_, server := usual.SmtpConfig(account)
	jsonByte, jErr := json.MarshalIndent(server, "", "") //把struct转成json
	if jErr != nil {
		common.FileLog("server", "DeleteSmtp MarshalIndent出错", jErr)
		code = 3010
		tips = "删除SMTP配置失败"
	}

	keyString := string(jsonByte)
	err := common.RedisClient.LRem("GoJob_SmtpJsonQueue", 1, keyString).Err()
	if err != nil {
		common.FileLog("server", "删除SMTP配置出错", err)
		code = 3020
		tips = "删除SMTP配置失败"
	}

	if err == nil && jErr == nil {
		code = 2000
		tips = "删除成功"
	}
	c.JSON(200, gin.H{
		"msg":  tips,
		"code": code,
	})
}

func AddSmtp(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		c.Redirect(302, "/LoginPage")
		return
	}
	editKey := c.DefaultQuery("key", "")
	var seoTitle string
	var smtpDetail usual.EmailServer
	if editKey != "" {
		_, smtpDetail = usual.SmtpConfig(editKey)
		seoTitle = "修改发件邮箱"
	} else {
		seoTitle = "添加发件邮箱"
	}

	c.HTML(http.StatusOK, "addServer.html", gin.H{
		"name":     "Denny Yang",
		"seoTitle": seoTitle,
		"editKey":  editKey,
		"smtp":     smtpDetail,
		"path":     fmt.Sprintf("%v", c.Request.URL.Path),
	})
}

//添加smtp api
func AddSmtpApi(c *gin.Context) {
	code, tips, redirectUrl := models.AddSmtpApi(c)
	c.JSON(200, gin.H{
		"msg":         tips,
		"code":        code,
		"redirectUrl": redirectUrl,
	})
}

//收件人
func ReceiveUsers(c *gin.Context) {
	var seoTitle string = "收件人列表"
	emalList, code, tips := models.ReceiveEmailUsers(c)
	c.HTML(http.StatusOK, "receiveUsers.html", gin.H{
		"msg":        tips,
		"code":       code,
		"emalList":   emalList,
		"totalCount": len(emalList),
		"seoTitle":   seoTitle,
		"path":       fmt.Sprintf("%v", c.Request.URL.Path),
	})
}

//收件人详情
func UpdateEmailUsers(c *gin.Context) {
	var seoTitle string = "收件人详情"
	emalList, code, tips := models.ReceiveEmailUsers(c)
	keyName := c.DefaultQuery("key", "")
	c.HTML(http.StatusOK, "receiveDetail.html", gin.H{
		"msg":        tips,
		"code":       code,
		"emalList":   emalList,
		"totalCount": len(emalList),
		"seoTitle":   seoTitle,
		"keyName":    keyName,
		"path":       fmt.Sprintf("%v", c.Request.URL.Path),
	})
}

//添加收件人
func AddEmailUsers(c *gin.Context) {
	var seoTitle string = "添加收件人"
	//emalList, code, tips := models.ReceiveEmailUsers(c)

	keyName := c.DefaultQuery("key", "")
	c.HTML(http.StatusOK, "receiveAdd.html", gin.H{
		"seoTitle": seoTitle,
		"keyName":  keyName,
		"path":     fmt.Sprintf("%v", c.Request.URL.Path),
	})
}

//上传文件
func UploadFile(c *gin.Context) {
	file, _ := c.FormFile("myfile")
	//log.Println(file.Filename) // 打印上传的文件名
	fileExt := strings.ToLower(path2.Ext(file.Filename))
	var allowExt = []string{".txt", ".mp3", ".mp4", ".wav", ".csv", ".doc", ".xlsx", "xls", ".jpg", ".png", ".gif", ".jpeg", ".gz", ".php", ".go", ".log", ".pdf"}
	var allowFlag = false
	for _, v := range allowExt {
		if fileExt == v {
			allowFlag = true
			continue
		}
	}

	var msg, fileUrl string
	var code int
	if allowFlag == true {
		rooPath, _ := os.Getwd()
		dir := time.Now().Format("200601")
		path := rooPath + "/static/upload/" + dir + "/"
		_, err := os.Stat(path)
		if os.IsExist(err) == false { //如果不是文件夹,则创建文件夹,权限777
			mkErr := os.MkdirAll(path, os.ModePerm)
			if mkErr != nil {
				common.FileLog("server", "文件夹创建失败", path, mkErr)
				msg = "文件夹创建失败"
				code = 2500
			}
		}

		fileName := common.UniqueId() + fileExt
		saveFileName := path + fileName
		// 将上传的文件，保存到文件中
		saveErr := c.SaveUploadedFile(file, saveFileName)
		if saveErr != nil {
			fmt.Println("save error", saveErr)
			msg = "保存文件失败"
			code = 2600
		}

		if msg == "" {
			code = 2000
			msg = fmt.Sprintf("'%s' uploaded!", file.Filename)
			fileUrl = "/static/upload/" + dir + "/" + fileName
		}
	} else {
		msg = "文件后缀不允许"
		code = 2100
	}

	c.JSON(200, gin.H{
		"msg":  msg,
		"code": code,
		"url":  fileUrl, //文件相对路径的url
	})
}

//post添加批量邮件用户Api
func AddEmailApi(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	var code int
	var tips, redirectUrl string
	if username == nil {
		c.JSON(200, gin.H{
			"msg":         tips,
			"code":        code,
			"redirectUrl": redirectUrl,
		})
		return
	}

	code, tips, redirectUrl = models.AddEmailUsers(c)
	c.JSON(200, gin.H{
		"msg":         tips,
		"code":        code,
		"redirectUrl": redirectUrl,
	})
}

//post删除用户列表
func DeleteEmailApi(c *gin.Context) {
	code, tips, redirectUrl := models.HandelDeleteEmailUsers(c)
	c.JSON(200, gin.H{
		"msg":         tips,
		"code":        code,
		"redirectUrl": redirectUrl,
	})
}

//后台账号管理
func AdminAccount(c *gin.Context) {
	var seoTitle string = "账号管理"
	keyName := c.DefaultQuery("key", "")

	userList, err := common.RedisClient.Keys("GoJob_user_*").Result()
	if err != nil {
		common.FileLog("server", "AdminAccount", err)
	}

	totalCount := len(userList)

	if totalCount > 0 {
		for k, detailKey := range userList {
			userList[k] = strings.ReplaceAll(detailKey, "GoJob_user_", "")
		}
	}

	c.HTML(http.StatusOK, "adminAccount.html", gin.H{
		"seoTitle":   seoTitle,
		"keyName":    keyName,
		"userList":   userList,
		"totalCount": totalCount,
		"path":       fmt.Sprintf("%v", c.Request.URL.Path),
	})
}

//添加后台账号
func AddAccount(c *gin.Context) {
	var seoTitle string = "账号管理"
	//emalList, code, tips := models.ReceiveEmailUsers(c)
	keyName := c.DefaultQuery("key", "")
	c.HTML(http.StatusOK, "accountAdd.html", gin.H{
		"seoTitle": seoTitle,
		"keyName":  keyName,
		"path":     fmt.Sprintf("%v", c.Request.URL.Path),
	})
}

//添加后台账号
func AddAccountApi(c *gin.Context) {
	code, tips, redirectUrl := models.HandelAddAccountApi(c)
	c.JSON(200, gin.H{
		"msg":         tips,
		"code":        code,
		"redirectUrl": redirectUrl,
	})

}

//更新后台账号
func UpateAccount(c *gin.Context) {
	code, tips, redirectUrl := models.HandelUpdateAccount(c)
	c.JSON(200, gin.H{
		"msg":         tips,
		"code":        code,
		"redirectUrl": redirectUrl,
	})
}

//删除账号
func DeleteAccount(c *gin.Context) {
	code, tips, redirectUrl := models.HandelDeleteccountApi(c)
	c.JSON(200, gin.H{
		"msg":         tips,
		"code":        code,
		"redirectUrl": redirectUrl,
	})
}
