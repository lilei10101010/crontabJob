package controllers

import (
	"cronJob/models"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func LoginPage(c *gin.Context) {
	tips := c.DefaultQuery("tips", "")
	session := sessions.Default(c)
	username := session.Get("username")

	if username != nil { //已经登录的用户,跳转到任务列表页面
		c.Redirect(302, "/SystemIndex")
		return
	}

	c.HTML(http.StatusOK, "adminLogin.html", gin.H{
		"seoTitle": "登录页面",
		"tips":     tips,
		"path":     fmt.Sprintf("%v", c.Request.URL.Path),
	})
}

func LoginApi(c *gin.Context) {
	loginSuccess, tips := models.HandelRedisAccount(c)
	var apiCode int
	if loginSuccess == true {
		apiCode = 2000
	} else {
		apiCode = 6000
	}
	c.JSON(200, gin.H{
		"code":         apiCode,
		"msg":          tips,
		"redirect_url": "/SystemIndex",
	})
}
func LoginOut(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("username")
	err := session.Save()
	var tips string

	if err != nil {
		tips = "退出登录失败"
	} else {
		tips = "已退出登录"
	}
	c.Redirect(302, "/LoginPage?tips="+tips)
}

func SystemIndex(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		c.Redirect(302, "/LoginPage")
		return
	}

	nowPage, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	if nowPage <= 0 {
		nowPage = 1
	}
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("size", "10"), 10, 64)
	var prevPage int64
	if pageSize <= 1 {
		pageSize = 1
		prevPage = 1
	} else {
		prevPage = nowPage - 1
	}

	jobList, totalCount, nextPage := models.AllTasks("all", nowPage, pageSize)

	c.HTML(http.StatusOK, "systemIndex.html", gin.H{
		"name":       "Denny Yang",
		"seoTitle":   "任务管理",
		"jobList":    jobList,    //任务列表
		"totalCount": totalCount, //总结果数
		"nowPage":    nowPage,    //当前页码
		"pageSize":   pageSize,   //每页条数
		"prevPage":   prevPage,   //上下一页页码,如果大于0则有下一页
		"nextPage":   nextPage,   //下一页页码,如果大于0则有下一页
		"path":       fmt.Sprintf("%v", c.Request.URL.Path),
	})
}
