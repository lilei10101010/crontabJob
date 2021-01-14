package main

import (
	"cronJob/common"
	"cronJob/controllers"
	"cronJob/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/robfig/cron"

	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"time"
)

var (
	CronIsRun bool
)

// 入口函数
func main() {
	//设置时区为东8区
	common.TIME_ZONE = time.FixedZone("CST", 8*3600)
	// 禁用控制台颜色
	gin.DisableConsoleColor()
	// 创建记录日志的文件
	f, _ := os.Create("./log/gin.log")
	//gin.DefaultWriter = io.MultiWriter(f)//日志写入到文件

	// 如果需要将日志同时写入文件和控制台，请使用以下代码
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	// 初始化一个http服务对象
	r := gin.Default()

	r.Use(gin.Recovery()) //设置Recovery中间件，主要用于拦截paic错误，不至于导致进程崩掉

	// 创建基于cookie的存储引擎，59OrSq7YIER3mqlQ 参数是用于加密的密钥
	store := cookie.NewStore([]byte("59OrSq7YIER3mqlQ"))
	r.Use(sessions.Sessions("mysession", store))
	gin.DisableConsoleColor()
	common.ROOTPATH, _ = os.Getwd()
	r.LoadHTMLGlob("views/*")
	r.Static("/static", common.ROOTPATH+"/static")

	// 设置一个get请求的路由，url为/ping, 处理函数（或者叫控制器函数）是一个闭包函数。
	r.GET("/", controllers.LoginPage)

	// 设置一个get请求的路由，url为/ping, 处理函数（或者叫控制器函数）是一个闭包函数。
	r.GET("/LoginPage", controllers.LoginPage)
	r.POST("/LoginApi", controllers.LoginApi) //处理登录逻辑
	r.GET("/LoginOut", controllers.LoginOut)  //退出登录逻辑

	r.GET("/SystemIndex", controllers.SystemIndex)    //默认页面
	r.GET("/AddCronJob", controllers.AddCronJob)      //添加任务
	r.POST("/AddJobApi", controllers.AddJobApi)       //添加修改任务API
	r.POST("/EditJobApi", controllers.EditJobApi)     //添加修改任务API
	r.POST("/DeleteJobApi", controllers.DeleteJobApi) //删除任务API
	r.GET("/JobLog", controllers.JobLog)              //任务日志
	r.POST("/ADetailJob", controllers.ADetailJob)     //任务详情

	r.GET("/EmailServer", controllers.EmailServer) //邮件服务器列表
	r.GET("/AddSmtp", controllers.AddSmtp)         //添加发件邮箱
	r.POST("/AddSmtpApi", controllers.AddSmtpApi)  //添加发件邮箱API
	r.POST("/DeleteSmtp", controllers.DeleteSmtp)  //删除发件邮箱

	r.GET("/ReceiveUsers", controllers.ReceiveUsers)         //收件邮箱
	r.GET("/UpdateEmailUsers", controllers.UpdateEmailUsers) //更新发送邮箱
	r.GET("/AddEmailUsers", controllers.AddEmailUsers)       //添加收件邮箱
	r.POST("/UploadFile", controllers.UploadFile)            //上传文件,默认32MB
	r.POST("/AddEmailApi", controllers.AddEmailApi)          //添加收件邮箱API
	r.POST("/DeleteEmailApi", controllers.DeleteEmailApi)    //删除收件邮箱API

	r.GET("/AdminAccount", controllers.AdminAccount)    //后台账号列表
	r.GET("/AddAccount", controllers.AddAccount)        //添加后台账号
	r.POST("/AddAccountApi", controllers.AddAccountApi) //添加后台账号Api
	r.POST("/UpateAccount", controllers.UpateAccount)   //更新后台账号Api
	r.POST("/DeleteAccount", controllers.DeleteAccount) //删除后台账号Api

	r.GET("/Queue", controllers.QueueList)          //队列列表
	r.GET("/QueueAdd", controllers.QueueAdd)        //添加队列
	r.POST("/QueueAddApi", controllers.QueueAddApi) //添加队列API
	r.POST("/QueueDelApi", controllers.QueueDelApi) //删除队列API

	r.GET("/Logs", controllers.Logs)       //日志列表
	r.GET("/DownLog", controllers.DownLog) //下载日志
	r.GET("/ViewLog", controllers.ViewLog) //查看日志

	authorized := r.Group("/admin")
	authorized.Use(LoginMiddleWare())
	{
		//authorized.GET("/index", controllers.UserIndex)
	}
	//common.ConnDb()    //初始化数据库连接
	common.ConnRedis() //Redis连接

	//如果定时任务没有启动,则启动定时任务
	if CronIsRun == false {
		cronClient := cron.New()
		spec1 := "*/1 * * * * ?"                                     //每秒运行一次
		cronErr := cronClient.AddFunc(spec1, models.RunRepeatSecond) //秒级间隙的重复运行任务
		if cronErr != nil {
			fmt.Println(" GET启动失败!", cronErr)
			common.FileLog("server", "GET启动失败!", cronErr)
		} else {
			fmt.Println("GET启动成功!")
			common.FileLog("server", "GET启动成功!")
		}

		spec2 := "* */1 * * * ?"                                      //每分钟运行一次
		cronErr2 := cronClient.AddFunc(spec2, models.RunRepeatMinute) //运行函数Test1
		if cronErr2 != nil {
			fmt.Println("RunRepeatMinute启动失败!", cronErr)
			common.FileLog("server", "RunRepeatMinute启动失败!", cronErr)
		} else {
			fmt.Println("RunRepeatMinute启动成功!")
			common.FileLog("server", "RunRepeatMinute启动成功!")
		}

		spec4 := "*/1 * * * * ?"                               //每秒运行一次
		cronErr4 := cronClient.AddFunc(spec4, models.WatchKey) //执行函数WatchKey
		if cronErr4 != nil {
			fmt.Println(" WatchKey启动失败!", cronErr4)
			common.FileLog("server", "WatchKey启动失败!", cronErr4)
		} else {
			fmt.Println("WatchKey启动成功!")
			common.FileLog("server", "WatchKey启动成功!")
		}

		spec5 := "*/1 * * * * ?"                              //每秒运行一次
		cronErr5 := cronClient.AddFunc(spec5, models.RunOnce) //执行函数RunOnce
		if cronErr4 != nil {
			fmt.Println(" RunOnce启动失败!", cronErr5)
			common.FileLog("server", "RunOnce启动失败!", cronErr5)
		} else {
			fmt.Println("RunOnce启动成功!")
			common.FileLog("server", "RunOnce启动成功!")
		}
		CronIsRun = true
		cronClient.Start()
	}

	serverErr := r.Run("0.0.0.0:9527") // 监听并在 0.0.0.0:8080 上启动服务
	if serverErr != nil {
		fmt.Println("main监听端口失败", serverErr)
	}
}

//登录中间件
func LoginMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		//if user, _ := c.Get(controllers.CONTEXT_USER_KEY); user != nil {
		//	if _, ok := user.(*models.User); ok {
		//		c.Next()
		//		return
		//	}
		//}
		//seelog.Warnf("User not authorized to visit %s", c.Request.RequestURI)
		//c.HTML(http.StatusForbidden, "errors/error.html", gin.H{
		//	"message": "Forbidden!",
		//})
		//c.Abort()
	}
}
