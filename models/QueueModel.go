package models

import (
	"cronJob/common"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"strconv"
)

type QueueStruct struct {
	QueueName   string //任务名称
	KeyName     string //KEY名称
	Exec        string //执行操作触发发送email,get/post请求
	Url         string //请求的URL
	SmtpHost    string //发送的SMTP服务器
	SmtpPort    string //发送的SMTP服务器
	SmtpUser    string //发送的SMTP服务器
	SmtpPwd     string //发送的SMTP服务器
	Title       string //邮件标题
	Content     string //邮件内容
	Receive     string //收件人
	Sender      string //发送者名称
	TimeOut     int64  //请求超时
	SuccessFlag string //请求成功标记

	CreateUser string //创建人
	CreateTime string //创建时间
}

func WatchKeyList(c *gin.Context) (code int, msg, redirectUrl string, data []QueueStruct) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		return 1900, "请先登录", "/LoginPage", data
	}

	result, err := common.RedisClient.LRange("GoJob_watchKeyList", 0, -1).Result()
	if err != nil {
		return 2010, "获取数据出错", "", data
	}

	for _, value := range result {
		keyName := "GoJob_wkl_" + value
		detail, _ := common.RedisClient.Get(keyName).Result()
		var qc QueueStruct
		decodeErr := json.Unmarshal([]byte(detail), &qc)
		if decodeErr != nil {
			common.FileLog("server", "WatchKeyList decode error "+keyName, decodeErr)
			continue
		}

		data = append(data, qc)
	}
	return 2000, "获取数据成功", "", data
}

//添加api
func AddWatchKeyApi(c *gin.Context) (code int, msg, redirectUrl string) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		return 1900, "请先登录", "/LoginPage"
	}

	var qc QueueStruct
	QueueName := c.DefaultPostForm("QueueName", "")
	if QueueName == "" {
		return 2010, "请输入监听的队列名称", ""
	}

	KeyName := c.DefaultPostForm("KeyName", "")
	if KeyName == "" {
		return 2020, "请输入监听的KEY", ""
	}

	ExecType := c.DefaultPostForm("ExecType", "")
	if ExecType == "" {
		return 2030, "请选择执行的操作", ""
	}

	TimeOutStr := c.DefaultPostForm("TimeOut", "30")
	TimeOut, _ := strconv.ParseInt(TimeOutStr, 10, 64)
	if TimeOut < 1 {
		TimeOut = 1
	}

	SuccessFlag := c.DefaultPostForm("SuccessFlag", "")

	exist, err := common.RedisClient.LRange("GoJob_watchKeyList", 0, -1).Result()
	if err != nil {
		common.FileLog("server", "AddWatchKeyApi", err)
		return 2050, "连接出错", ""
	}

	if len(exist) > 0 {
		for _, v := range exist {
			if v == KeyName {
				return 2060, KeyName + "已在监听列表", ""
			}
		}
	}

	if ExecType == "get" || ExecType == "post" {
		qc.TimeOut = TimeOut
		qc.SuccessFlag = SuccessFlag
	}
	qc.QueueName = QueueName
	qc.KeyName = KeyName
	qc.Exec = ExecType
	qc.CreateTime = common.NowString()
	qc.CreateUser = fmt.Sprintf("%v", username)

	byteStr, encodeErr := json.MarshalIndent(qc, "", "")
	if encodeErr != nil {
		common.FileLog("server", "AddWatchKeyApi json error", encodeErr)
		return 2060, KeyName + "保存数据出错", ""
	}

	resultStr := string(byteStr)
	setErr := common.RedisClient.Set("GoJob_wkl_"+KeyName, resultStr, 0).Err()
	if setErr != nil {
		common.FileLog("server", "AddWatchKeyApi set key error", setErr)
		return 2070, KeyName + "保存数据出错", ""
	}

	lpushErr := common.RedisClient.LPush("GoJob_watchKeyList", KeyName).Err()
	if lpushErr != nil {
		common.RedisClient.Del("GoJob_wkl_" + KeyName)
		common.FileLog("server", "AddWatchKeyApi lpush key error", lpushErr)
		return 2080, KeyName + "保存数据出错", ""
	}

	return 2000, "添加成功", "/Queue"
}

//删除监听的队列api
func DelWatchKeyApi(c *gin.Context) (code int, msg, redirectUrl string) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		return 1900, "请先登录", "/LoginPage"
	}

	KeyName := c.DefaultPostForm("keyName", "")
	if KeyName == "" {
		return 2010, "KeyName不能为空", ""
	}

	err := common.RedisClient.Del("GoJob_wkl_" + KeyName).Err()
	if err != nil {
		return 2020, "删除出错", ""
	}

	remErr := common.RedisClient.LRem("GoJob_watchKeyList", 0, KeyName).Err()
	if remErr != nil {
		return 2030, "删除出错", ""
	}

	return 2000, "删除成功", "/Queue"
}
