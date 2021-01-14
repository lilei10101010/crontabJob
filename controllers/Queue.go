package controllers

import (
	"cronJob/common"
	"cronJob/models"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
	"os"
	"strings"
)

//队列
func QueueList(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		c.Redirect(302, "/LoginPage")
		return
	}

	code, msg, _, data := models.WatchKeyList(c)
	c.HTML(http.StatusOK, "queue.html", gin.H{
		"seoTitle":   "监听队列",
		"totalCount": len(data),
		"list":       data,
		"msg":        msg,
		"code":       code,
		"path":       fmt.Sprintf("%v", c.Request.URL.Path),
	})
}
func QueueAdd(c *gin.Context) {
	tips := c.DefaultQuery("tips", "")
	session := sessions.Default(c)
	username := session.Get("username")

	if username == nil { //已经登录的用户,跳转到任务列表页面
		c.Redirect(302, "/LoginPage")
		return
	}

	c.HTML(http.StatusOK, "queueAdd.html", gin.H{
		"seoTitle": "添加KEY",
		"tips":     tips,
		"path":     fmt.Sprintf("%v", c.Request.URL.Path),
	})
}

//添加监听的key
func QueueAddApi(c *gin.Context) {
	code, tips, redirectUrl := models.AddWatchKeyApi(c)
	c.JSON(200, gin.H{
		"code":         code,
		"msg":          tips,
		"redirect_url": redirectUrl,
	})
}

//添加监听的key
func QueueDelApi(c *gin.Context) {
	code, tips, redirectUrl := models.DelWatchKeyApi(c)
	c.JSON(200, gin.H{
		"code":         code,
		"msg":          tips,
		"redirect_url": redirectUrl,
	})
}

//日志
func Logs(c *gin.Context) {
	tips := c.DefaultQuery("tips", "")
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil { //已经登录的用户,跳转到任务列表页面
		c.Redirect(302, "/LoginPage")
		return
	}

	filePath := common.ROOTPATH + "/log/"
	logList, err := common.ReadDirFile(filePath)
	if err != nil {
		common.FileLog("server", "Logs方法出错", err)
	}

	c.HTML(http.StatusOK, "logList.html", gin.H{
		"seoTitle":   "运行日志",
		"tips":       tips,
		"logList":    logList,
		"totalCount": len(logList),
		"path":       fmt.Sprintf("%v", c.Request.URL.Path),
	})
}

//下载日志
func DownLog(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil { //已经登录的用户,跳转到任务列表页面
		c.Redirect(302, "/LoginPage")
		return
	}

	fileName := c.DefaultQuery("file", "")
	var code int
	var tips, redirectUrl string
	if fileName == "" {
		code = 2010
		tips = "文件名称不能为空"
	}

	Path := common.ROOTPATH + "/log/"
	fullPath := Path + fileName

	//清空文件内容的请求
	flushFile := c.DefaultQuery("flushFile", "") //清空文件
	if flushFile != "" {
		fErr := common.FlushFile(fullPath)
		if fErr != nil {
			code = 2005
			tips = "文件清空失败" + fmt.Sprintf("%v", fErr)
		} else {
			code = 2000
			tips = "文件已清空"
		}
		c.JSON(200, gin.H{
			"code":        code,
			"msg":         tips,
			"redirectUrl": "",
		})
		return
	}

	isFile, fileError := common.IsFile(fullPath)
	if isFile == false {
		code = 2020
		tips = "文件不存在" + fmt.Sprintf("%v", fileError)
	}

	newZipName := "/static/download/" + fileName + "_" + common.UniqueId() + ".zip"
	zipPath := common.ROOTPATH + newZipName

	zipErr := common.Zip(fullPath, zipPath)
	if zipErr != nil {
		code = 2030
		tips = "文件压缩失败" + fmt.Sprintf("%v", zipErr)
	}

	if code < 2010 {
		code = 2000 //成功
		redirectUrl = newZipName
	}

	filePath := common.ROOTPATH + "/static/download/"
	logList, downloadDirErr := common.ReadDirFile(filePath)
	if downloadDirErr != nil {
		common.FileLog("server", "DownLog方法出错", downloadDirErr)
	}

	nowInt := common.NowTimeInt()
	//删除一天之前的文件
	if len(logList) > 0 {
		for _, value := range logList {
			thisFileName := value.FileName
			if thisFileName == "zip.html" {
				continue
			}
			timeInt64 := common.TimeStringToInt(value.ModTime)
			if nowInt-timeInt64 > 86400 {
				rErr := os.Remove(filePath + thisFileName)
				if rErr != nil {
					common.FileLog("server", "DownLog删除文件失败", value.FileName, downloadDirErr)
				}
			}
		}
	}

	c.JSON(200, gin.H{
		"code":        code,
		"msg":         tips,
		"redirectUrl": redirectUrl,
	})
}

//查看日志
func ViewLog(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil { //已经登录的用户,跳转到任务列表页面
		c.Redirect(302, "/LoginPage")
		return
	}

	fileName := c.DefaultQuery("fileName", "")

	filePath := common.ROOTPATH + "/log/"
	logString := common.ReadLastFile(filePath+fileName, "\n", 6000)
	if logString != "" {
		logString = strings.ReplaceAll(strings.TrimSpace(logString), "\n", "<br>")
	}

	isApi := c.DefaultQuery("isApi", "")
	if isApi == "" {
		c.HTML(http.StatusOK, "logDetail.html", gin.H{
			"seoTitle":  "运行日志",
			"logName":   fileName,
			"logString": template.HTML(logString),
			"path":      fmt.Sprintf("%v", c.Request.URL.Path),
		})
	} else {
		c.JSON(200, gin.H{
			"logString": logString,
		})
	}
}
