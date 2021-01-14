package models

import (
	"cronJob/common"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"time"
)

type CronJob struct {
	JobName     string //任务名称
	IsRepeat    string //是否重复0不重复,1重复
	RunBetween  string //运行间隙
	BetweenNum  string //运行间隙数值
	RunDateTime string //开始时间
	Exec        string //执行操作
	Content     string //执行内容
	TimeOut     int64  //超时时间,支持GET/POST
	RetryNum    string //失败重试次数,支持GET/POST,失败后每隔一分钟连接一次
	RetryNumInt int
	SetKey      string //设置成的KEY
	SmtpServer  string //邮件SMTP SERVER
	EmailTitle  string //邮件标题
	SendName    string //发送者名称
	EmailUsers  string //发送的用户,如果是发送多个用户则是用户组的KEY
	SuccessFlag string //成功标志,支持GET/POST
	RunStatus   int    //0失败,1运行成功,2运行失败,3运行失败等待重试中
	FailRetry   int    //当前失败重试次数
	AddUser     int    //添加用户
}

//检查任务字段是否完整
func checkJobFields(c *gin.Context) (code int, msg string, fieldStruct CronJob) {
	jobName := c.DefaultPostForm("job_name", "")
	fieldStruct = CronJob{}
	if jobName == "" {
		return 2010, "任务名称不能为空", fieldStruct
	}
	isRepeat := c.DefaultPostForm("isRepeat", "")
	if isRepeat == "" {
		return 2020, "请选择任务是否重复运行", fieldStruct
	}

	runBetween := c.DefaultPostForm("run_between", "")
	if isRepeat == "1" && runBetween == "" {
		return 2050, "任务运行间隙不能为空", fieldStruct
	}

	betweenNum := c.DefaultPostForm("between_num", "")
	if isRepeat == "1" && betweenNum == "" {
		return 2060, "运行间隙数值不能为空", fieldStruct
	}

	runDateTime := c.DefaultPostForm("RunDateTime", "")
	if runDateTime == "" {
		return 2070, "运行时间不能为空", fieldStruct
	}
	nowTime := time.Now().In(common.TIME_ZONE).Format("2006-01-02 15:04:05")
	diffSecond, err := common.TimeDiffer(nowTime, runDateTime)

	if err != nil {
		return 2073, "时间格式出错", fieldStruct
	}
	//if diffSecond < 0 {
	if diffSecond < -600 {
		return 2073, "时间已过期,请选择未来时间", fieldStruct
	}

	exec := c.DefaultPostForm("exec", "")
	if exec == "" {
		return 2090, "执行类型不能为空", fieldStruct
	}

	//get post请求才有超时的概念
	timeOut := c.DefaultPostForm("time_out", "")
	var timeOutInt int
	if timeOut != "" {
		timeOutInt, _ = strconv.Atoi(timeOut)
	}

	content := strings.TrimSpace(c.DefaultPostForm("content", ""))
	if content == "" {
		return 2100, "运行内容不能为空", fieldStruct
	}

	successFlag := strings.TrimSpace(c.DefaultPostForm("success_flag", ""))

	var isUrl bool //判断是否是网址
	if exec == "get" || exec == "post" {
		isUrl = strings.Contains(content, "http")
		if isUrl == false {
			return 2095, "URL校验失败", fieldStruct
		}
		fieldStruct.TimeOut = int64(timeOutInt)
		fieldStruct.SuccessFlag = successFlag
	} else if exec == "multiGet" || exec == "multiPost" {
		urls := common.Explode("|", content)
		var errUrl string
		for _, v := range urls {
			isUrl = strings.Contains(content, "http")
			if isUrl == false {
				errUrl = errUrl + v + "\n"
			}
		}
		if errUrl != "" {
			errUrl += "不是网址"
			return 2095, errUrl, fieldStruct
		}
		fieldStruct.TimeOut = int64(timeOutInt)
	}

	retryNum := c.DefaultPostForm("retry_num", "")
	RetryNumInt, rnErr := strconv.Atoi(retryNum)
	if rnErr != nil {
		return 2097, "失败重试次数不正确", fieldStruct
	}

	smtpServer := c.DefaultPostForm("smtp_server", "")
	emailTitle := c.DefaultPostForm("email_title", "")
	sendName := c.DefaultPostForm("send_name", "")

	if exec == "email" || exec == "multiEmail" {
		if smtpServer == "" {
			return 2110, "请选择SMTP服务器", fieldStruct
		}
		if emailTitle == "" {
			return 2120, "邮件标题不能为空", fieldStruct
		}
		if sendName == "" {
			return 2130, "发件人名称不能为空", fieldStruct
		}

		//单个用户发送邮件填写email,批量用户填写key
		if exec == "email" {
			emailUsers := c.DefaultPostForm("email_users", "")
			if emailUsers == "" {
				return 2140, "收件人不能为空", fieldStruct
			}
			fieldStruct.EmailUsers = emailUsers
		} else {
			emailUsers := c.DefaultPostForm("email_users_key", "")
			if emailUsers == "" {
				return 2140, "收件人不能为空,请先添加收件人列表", fieldStruct
			}
			fieldStruct.EmailUsers = emailUsers
		}

		fieldStruct.SmtpServer = smtpServer
		fieldStruct.EmailTitle = emailTitle
		fieldStruct.SendName = sendName
	}

	fieldStruct.JobName = jobName
	fieldStruct.IsRepeat = isRepeat
	fieldStruct.RunBetween = runBetween
	fieldStruct.BetweenNum = betweenNum
	fieldStruct.RunDateTime = runDateTime
	fieldStruct.Exec = exec
	fieldStruct.Content = content
	fieldStruct.RetryNum = retryNum
	fieldStruct.RetryNumInt = RetryNumInt

	return 2000, "字段校验通过", fieldStruct
}

//处理添加逻辑
func handelAdd(info CronJob) (code int, msg string) {
	uniqid := common.UniqueId()
	var redisKey string
	jsonByte, jErr := json.MarshalIndent(info, "", "") //把struct转成json
	if jErr != nil {
		common.FileLog("debug", "MarshalIndent出错", jErr)
		return 3010, "添加任务出错"
	}
	jString := string(jsonByte)

	if info.IsRepeat == "0" { //不重复
		redisKey = "GoJob_detail_once_" + uniqid
	} else if info.IsRepeat == "1" { //重复运行
		redisKey = "GoJob_detail_repeat_" + uniqid
	}

	setErr := common.RedisClient.Set(redisKey, jString, 0).Err() //设置key永不过期
	if setErr != nil {
		common.FileLog("debug", "KEY SET出错", setErr)
		return 3020, "操作失败,KEY存储出错"
	}
	pushERR := common.RedisClient.LPush("GoJob_totalJobQueue", redisKey).Err()
	if pushERR != nil {
		common.FileLog("debug", "KEY 推送出错", setErr)
		return 3030, "操作失败,KEY存储出错"
	}

	if info.IsRepeat == "0" { //不重复运行的任务
		err := common.RedisClient.LPush("GoJob_runOnce", redisKey).Err()
		if err != nil {
			common.FileLog("debug", "KEY 推送出错", err)
			return 3035, "添加到消费key失败"
		}
	} else if info.IsRepeat == "1" { //重复运行的任务
		var errSec error
		if info.RunBetween == "second" { //一直运行,且间隙是秒级的
			errSec = common.RedisClient.LPush("GoJob_runAlways_second", redisKey).Err()
		} else { //一直运行,且间隙是分钟级的
			errSec = common.RedisClient.LPush("GoJob_runAlways_minute", redisKey).Err()
		}
		if errSec != nil {
			return 3035, "添加到任务失败"
		}
	}
	return 2000, "添加任务成功"
}

//处理添加任务
func AddJob(c *gin.Context) (code int, msg, redirectUrl string) {
	session := sessions.Default(c)
	username := session.Get("username")

	if username == nil {
		return 3000, "请先登录", "/LoginPage"
	}
	code, msg, fields := checkJobFields(c) //处理表单提交数据
	common.FileLog("debug", msg, code, fields)

	if code == 2000 {
		code, msg = handelAdd(fields) //添加逻辑
		redirectUrl = "/SystemIndex"
	} else {
		redirectUrl = "/LoginPage"
	}
	return code, msg, redirectUrl
}

//获取所有任务或正在运行的任务
func AllTasks(runFlag string, pageSize, limitCount int64) (list []CronJob, totalCount, nextPage int64) {
	var cj CronJob
	var startSize int64
	var subLen = 90
	if pageSize <= 0 {
		startSize = 0
		pageSize = 1
	} else {
		startSize = (pageSize - 1) * limitCount
	}
	limitCountEnd := startSize + limitCount - 1

	//获取全部任务
	if runFlag == "all" {
		totalCount, _ = common.RedisClient.LLen("GoJob_totalJobQueue").Result()
		jobKeys, _ := common.RedisClient.LRange("GoJob_totalJobQueue", startSize, limitCountEnd).Result()

		resCount := int64(len(jobKeys))
		if resCount > 0 {
			for _, keyName := range jobKeys {
				detail, _ := common.RedisClient.Get(keyName).Result()
				_ = json.Unmarshal([]byte(detail), &cj)
				cj.SetKey = keyName
				if len(cj.Content) > subLen {
					cj.Content = common.Substring(cj.Content, 0, subLen)
				}
				list = append(list, cj)
			}
		}
	} else { //获取将要运行或运行中的任务,

	}

	if limitCount*pageSize < totalCount {
		nextPage = pageSize + 1
	}
	return list, totalCount, nextPage
}

//获取删除任务
func TaskDelete(c *gin.Context) (int, string) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		return 3000, "请先登录"
	}

	var logString string

	keyName := c.DefaultPostForm("keyName", "")
	if keyName == "" {
		return 2110, "keyName不能为空"
	}

	resErr := common.RedisClient.LRem("GoJob_totalJobQueue", 1, keyName).Err()
	if resErr != nil {
		logString = fmt.Sprintf("删除Key Total List:%s失败%v", keyName, resErr)
		return 2260, "删除失败"
	}

	isRepeat := c.DefaultPostForm("isRepeat", "")
	runBetween := c.DefaultPostForm("runBetween", "")

	//执行一次的任务
	if isRepeat == "0" || isRepeat == "" {
		err := common.RedisClient.LRem("GoJob_runOnce", 1, keyName).Err()

		if err != nil {
			logString = fmt.Sprintf("删除Key:%s失败%v", keyName, err)
			return 2200, "删除失败"
		}
	} else if isRepeat == "1" && runBetween == "second" {
		resErr := common.RedisClient.LRem("GoJob_runAlways_second", 1, keyName).Err()
		if resErr != nil {
			logString = fmt.Sprintf("删除Key:%s失败%v", keyName, resErr)
			return 2230, "删除失败"
		}
	} else {
		resErr := common.RedisClient.LRem("GoJob_runAlways_minute", 1, keyName).Err()
		if resErr != nil {
			logString = fmt.Sprintf("删除Key:%s失败%v", keyName, resErr)
			return 2240, "删除失败"
		}
	}

	waitingErr := common.RedisClient.LRem("GoJob_WaitingToRunTask", 1, keyName).Err()
	if waitingErr != nil {
		logString = fmt.Sprintf("删除Key:%s失败%v", keyName, resErr)
		return 2250, "删除失败"
	}

	keyDelErr := common.RedisClient.Del(keyName).Err()
	if keyDelErr != nil {
		logString = fmt.Sprintf("删除Key Set:%s失败%v", keyName, keyDelErr)
		return 2260, "删除失败"
	}

	if logString != "" {
		common.FileLog("server", logString)
	}
	return 2000, "操作成功"
}
