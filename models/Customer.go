package models

import (
	"cronJob/common"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"log"
	"strconv"
	"time"
)

type CustomerStruct struct {
	Customer_id int    `gorm:"column:customer_id"`
	Roleid      int    `gorm:"column:roleid"`
	Store_id    int    `gorm:"column:store_id"`
	Language_id int    `gorm:"column:language_id"`
	Firstname   string `gorm:"column:firstname"`
	Lastname    string `gorm:"column:lastname"`
	Email       string `gorm:"column:email"`
	Telephone   string `gorm:"column:telephone"`
	Password    string `gorm:"column:password"`
	Salt        string `gorm:"column:salt"`
	Wishlist    string `gorm:"column:wishlist"`
	Newsletter  int    `gorm:"column:newsletter"`
	Address_id  int    `gorm:"column:address_id"`
	Ip          string `gorm:"column:ip"`
	Status      int    `gorm:"column:status"`
	Code        string `gorm:"column:code"`
	Date_added  string `gorm:"column:date_added"`
}

//处理登录逻辑,如果没有设置账号密码则默认为admin 123456
func HandelRedisAccount(c *gin.Context) (bool, string) {
	username := c.DefaultPostForm("username", "")
	password := c.DefaultPostForm("password", "")
	accountKey := "GoJob_user_" + username
	if username == "" || password == "" {
		return false, "用户名或密码不能为空"
	}

	redisPwd, err := common.RedisClient.Get(accountKey).Result()
	if err != nil || redisPwd == "" {
		return false, "账号不存在或者密码未设置"
	}

	//获取当前账号登录次数
	attackKey := "GoJob_attackNum_" + username
	tryNumStr, _ := common.RedisClient.Get(attackKey).Result()
	tryNum, _ := strconv.Atoi(tryNumStr)
	limitNum := 10 //一小时限制登录出错次数10
	if tryNum > limitNum {
		return false, "账号或密码错误,输入密码次数超过10次"
	}

	password = common.Md5(password)
	if redisPwd == password {
		session := sessions.Default(c)
		session.Set("username", username) //设置
		sErr := session.Save()
		if sErr != nil {
			log.Println(sErr)
		}
		return true, "登陆成功"
	} else {
		setErr := common.RedisClient.Set(attackKey, tryNum+1, 600*time.Second).Err()
		if setErr != nil {
			common.FileLog("server", "HandelRedisAccount", setErr)
		}
		num := limitNum - tryNum - 1
		return false, fmt.Sprintf("账号或密码错误,剩余尝试次数%d", num)
	}
}

//处理登录逻辑,如果没有设置账号密码则默认为admin 123456
func HandelUpdateAccount(c *gin.Context) (code int, msg, redirectUrl string) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		return 1900, "请先登录", "/LoginPage"
	}
	key := c.DefaultPostForm("key", "")
	if key == "" {
		return 2010, "Key不能为空", ""
	}
	key = "GoJob_user_" + key

	passwd := c.DefaultPostForm("pwd", "")
	if passwd == "" {
		return 2020, "密码不能为空", ""
	}

	if len(passwd) < 6 {
		return 2020, "请输入六位数或以上的密码", ""
	}

	newPdw := common.Md5(passwd)
	err := common.RedisClient.Set(key, newPdw, 0).Err()
	if err != nil {
		return 2020, "Key出错", ""
	}

	//退出当前登录状态,让用户重新登录
	session.Delete("username")
	delSessionerr := session.Save()
	var tips string
	if delSessionerr != nil {
		tips = "退出登录失败"
	} else {
		tips = "更新成功,请重新登录!"
	}

	return 2000, tips, "/LoginPage?tips=" + tips
}

//处理添加后台账号的逻辑
func HandelAddAccountApi(c *gin.Context) (code int, msg, redirectUrl string) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		return 1900, "请先登录", "/LoginPage"
	}
	account := c.DefaultPostForm("Account", "")
	if account == "" {
		return 2010, "账号不能为空", ""
	}
	key := "GoJob_user_" + account

	passwd := c.DefaultPostForm("Password", "")
	if passwd == "" {
		return 2020, "密码不能为空", ""
	}
	if len(passwd) < 6 {
		return 2020, "请输入六位数或以上的密码", ""
	}

	newPdw := common.Md5(passwd)
	err := common.RedisClient.Set(key, newPdw, 0).Err()
	if err != nil {
		return 2020, "Key出错", ""
	}
	return 2000, "操作成功", "/AdminAccount"
}

//处理删除后台账号的逻辑
func HandelDeleteccountApi(c *gin.Context) (code int, msg, redirectUrl string) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		return 1900, "请先登录", "/LoginPage"
	}
	key := c.DefaultPostForm("key", "")
	if key == "" {
		return 2010, "账号不能为空", ""
	}
	key = "GoJob_user_" + key

	err := common.RedisClient.Del(key).Err()
	if err != nil {
		return 2020, "Key删除出错", ""
	}
	return 2000, "操作成功", "/AdminAccount"
}
