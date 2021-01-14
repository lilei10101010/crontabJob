package models

import (
	"cronJob/common"
	"cronJob/usual"
	"encoding/json"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"os"
	"strings"
)

//处理添加SMTP服务器逻辑
func AddSmtpApi(c *gin.Context) (code int, tips, redirect string) {
	Host := c.DefaultPostForm("Host", "")
	Port := c.DefaultPostForm("Port", "")
	Account := c.DefaultPostForm("Account", "")
	Password := c.DefaultPostForm("Password", "")
	isEdit := c.DefaultPostForm("isEdit", "")

	if Host == "" {
		return 3010, "Host不能为空", "/"
	}
	if Port == "" {
		return 3020, "Port不能为空", "/"
	}
	if Account == "" {
		return 3030, "Account不能为空", "/"
	}
	if Password == "" {
		return 3040, "Password不能为空", "/"
	}

	//新增
	if isEdit == "" {
		_, existInfo := usual.SmtpConfig(Account)
		if existInfo.Account != "" {
			return 3050, Account + "账号已存在", ""
		}
	}

	var smtpData usual.EmailServer
	smtpData.Host = Host
	smtpData.Port = Port
	smtpData.Account = Account
	smtpData.Password = Password
	jsonByte, jErr := json.MarshalIndent(smtpData, "", "") //把struct转成json
	if jErr != nil {
		common.FileLog("server", "AddSmtpApi MarshalIndent出错", jErr)
		return 3060, "Json转换出错", ""
	}

	keyString := string(jsonByte)

	err := common.RedisClient.LPush("GoJob_SmtpJsonQueue", keyString).Err()
	if err != nil {
		return 3100, "操作失败", ""
	}

	//如果是编辑操作,添加完成,删掉旧数据
	if isEdit != "" {
		_, oldData := usual.SmtpConfig(isEdit)
		oldByte, oldErr := json.MarshalIndent(oldData, "", "") //把struct转成json
		if oldErr != nil {
			common.FileLog("server", "AddSmtpApi edit MarshalIndent出错", oldErr)
			return 3070, "Json转换出错", ""
		}

		oldString := string(oldByte)
		err := common.RedisClient.LRem("GoJob_SmtpJsonQueue", 1, oldString).Err()
		if err != nil {
			return 3080, "编辑失败", ""
		}
	}
	return 2000, "操作成功", "/EmailServer"
}

//收件人列表,收件人详情
func ReceiveEmailUsers(c *gin.Context) (emailList []string, code int, tips string) {
	getDetailBykey := c.DefaultQuery("key", "")
	userList, result := usual.SendUsers(getDetailBykey)
	if result != false {
		return userList, 2500, "数据获取失败"
	}
	return userList, 2000, "数据获取成功"
}

//添加批量收件人
func AddEmailUsers(c *gin.Context) (code int, tips, redirectUrl string) {
	var emailQueueKey = "GoJob_multiEmails"
	keyName := strings.TrimSpace(c.DefaultPostForm("name", ""))
	if keyName == "" {
		return 2010, "收件人名称不能为空", ""
	}
	keyName = strings.ReplaceAll(keyName, " ", "")
	keyName = "GoJob_multiUser_" + keyName //加上前缀

	checkExist, _ := common.RedisClient.LRange(emailQueueKey, 0, -1).Result()
	for _, v := range checkExist {
		if v == keyName {
			return 2020, "收件人名称已存在", ""
		}
	}

	fileName := c.DefaultPostForm("fileValue", "")
	if fileName == "" {
		return 2030, "请上传收件人文件", ""
	}

	dir, _ := os.Getwd()
	filepath := dir + "/" + fileName
	isFile, _ := common.IsFile(filepath)
	if isFile == false {
		return 2040, "请上传收件人文件", ""
	}

	readData, err := common.ReadCsv(filepath)
	if err != nil {
		return 2050, "文件读取出错", ""
	}

	if len(readData) < 1 {
		return 2050, "文件数据不能为空", ""
	}

	var emails []string
	for k, value := range readData {
		if k != 0 { //去掉第一个标题
			emails = append(emails, strings.TrimSpace(value[0])) //第一行是email
		}
	}

	jsonByte, err := json.MarshalIndent(emails, "", "")
	if err != nil {
		return 2060, "json encode出错", ""
	}
	jsonStr := string(jsonByte)

	err1 := common.RedisClient.Set(keyName, jsonStr, 0).Err()
	if err1 != nil {
		return 2070, "set key出错", ""
	}

	err2 := common.RedisClient.LPush(emailQueueKey, keyName).Err()
	if err2 != nil {
		common.RedisClient.Del(keyName)
		return 2070, "push key出错", ""
	}

	return 2000, "数据添加成功", "/ReceiveUsers"
}

//删除批量邮件用户的数据
func HandelDeleteEmailUsers(c *gin.Context) (code int, tips, redirectUrl string) {

	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		return 1900, "请先登录", "/LoginPage"
	}

	keyName := c.DefaultPostForm("keyName", "")
	if keyName == "" {
		return 2010, "KEY NAME不能为空", "/"
	}

	delErr := common.RedisClient.Del(keyName).Err()
	if delErr != nil {
		return 2020, "数据删除失败", "/"
	}

	lrErr := common.RedisClient.LRem("GoJob_multiEmails", 1, keyName).Err()
	if lrErr != nil {
		return 2030, "数据删除失败", "/"
	}

	return 2000, "数据删除成功", "/"
}
