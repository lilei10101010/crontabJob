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
	"strconv"
)

func AddCronJob(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		c.Redirect(302, "/LoginPage")
	}

	smtpService, _ := usual.SmtpConfig("all") //获取smtp配置
	multiUsers, _ := usual.SendUsers("")      //获取批量发送邮件的用户列表

	c.HTML(http.StatusOK, "addCronjob.html", gin.H{
		"seoTitle":    "添加任务",
		"smtpService": smtpService,
		"multiUsers":  multiUsers,
		"path":        fmt.Sprintf("%v", c.Request.URL.Path), //当前请求的url，不含参数,c.Request.URL带参数
	})
}

//添加任务POST
func AddJobApi(c *gin.Context) {
	code, msg, redirectUrl := models.AddJob(c)
	c.JSON(200, gin.H{
		"code":        code,
		"msg":         msg,
		"redirectUrl": redirectUrl,
	})
}

//修改任务POST
func EditJobApi(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")

	n := session.Get("username")
	common.FileLog("debug", n)
	if username == nil {
		c.Redirect(302, "/LoginPage")
		return
	}

	c.HTML(http.StatusOK, "systemIndex.html", gin.H{
		"name":  "Denny Yang",
		"title": c.ClientIP(),
		"path":  fmt.Sprintf("%v", c.Request.URL.Path),
	})
}

//删除任务
func DeleteJobApi(c *gin.Context) {
	apiCode, tips := models.TaskDelete(c)

	c.JSON(200, gin.H{
		"code":        apiCode,
		"msg":         tips,
		"redirectUrl": "/SystemIndex",
	})
}

//任务运行日志
func JobLog(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		c.Redirect(302, "/LoginPage")
		return
	}

	jobName := c.DefaultQuery("title", "")
	num := c.DefaultQuery("num", "0")
	limit := c.DefaultQuery("limit", "18")
	startLine, _ := strconv.Atoi(num)
	limiNum, _ := strconv.Atoi(limit)

	key := c.DefaultQuery("key", "")

	//日志文件
	var logName string = common.ROOTPATH + "/log/runningLog.log"

	resultString, nowReadNum, totalNum := common.ReadFileLine(logName, key, startLine, limiNum)

	logList := common.Explode("\n", resultString)

	var nextPage string //判断是否有下一页
	if nowReadNum < totalNum {
		nextPage = "/JobLog?key=" + key + "&title=" + jobName + "&num=" + strconv.Itoa(nowReadNum) + "&limit=" + limit
	}

	isApi := c.DefaultQuery("isApi", "")
	if isApi == "" {
		c.HTML(http.StatusOK, "jobLog.html", gin.H{
			"seoTitle":   "运行日志",
			"jobName":    jobName,
			"key":        key,
			"limitSize":  limiNum,
			"logList":    logList,
			"nextPage":   nextPage,
			"totalCount": totalNum,
			"path":       fmt.Sprintf("%v", c.Request.URL.Path),
		})
	} else {
		var logString string
		reverselogList := common.ReverseSlice(logList)
		for k, v := range reverselogList {
			if k == 0 && v == "" {
				continue //如果第一个值是空,则不输出空值
			}
			logString = logString + "<tr class='listItem'><td>" + v + "</td></tr>"
		}
		c.JSON(200, gin.H{
			"logString":  logString,
			"totalCount": totalNum,
		})
	}
}

//查看任务详情
func ADetailJob(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		c.Redirect(302, "/LoginPage")
		return
	}
	key := c.DefaultPostForm("key", "")
	var code int = 2000
	var err error
	var msg, result string
	if key == "" {
		code = 2010
		msg = "KEY值不能为空"
	} else {
		result, err = common.RedisClient.Get(key).Result()
		if err != nil {
			code = 2020
			msg = fmt.Sprintf("%v", msg)
		}

		var cj models.CronJob
		if result != "" {
			deErr := json.Unmarshal([]byte(result), &cj)
			if deErr != nil {
				code = 2030
				msg = fmt.Sprintf("%v", msg)
			} else {
				result = cj.Content
			}
		}
	}

	c.JSON(200, gin.H{
		"code":        code,
		"msg":         msg,
		"result":      result,
		"redirectUrl": "",
	})
}
