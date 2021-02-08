package models

import (
	"cronJob/common"
	"cronJob/usual"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func HandelWatchKey(key string) {
	lenNum, err := common.RedisClient.LLen(key).Result()
	if err != nil {
		common.FileLog("server", fmt.Sprintf("WatchKey获取keys出错%v", err))
		return
	}

	if lenNum > 0 {
		lenNum32 := int(lenNum)

		for i := 1; i <= lenNum32; i++ {
			go execWatchKey(key) //有多少数据就开启多少协程去pop数据
			if i%10000 == 0 {
				time.Sleep(1 * time.Second) //开启1W个协程等待0.5秒
			}
		}
	} else {
		if common.NowTimeInt()%60 == 0 {
			common.FileLog("heartBeat", "WatchKey idel "+key)
		}
	}
}

//执行具体操作
func execWatchKey(key string) {
	result, err := common.RedisClient.BRPop(1*time.Second, key).Result()
	if err != nil {
		common.FileLog("server", fmt.Sprintf("execWatchKey pop数据出错%s %v", key, err))
		return
	}

	resultString := result[1]
	var qs QueueStruct
	decodeErr := json.Unmarshal([]byte(resultString), &qs)
	fmt.Println(decodeErr)
	if decodeErr != nil {
		common.FileLog("server", fmt.Sprintf("execWatchKey decode出错%s %v", key, decodeErr))
		return
	}
	//执行具体操作
	var logValue string
	if strings.ToLower(qs.Exec) == "get" {
		result, status := common.Get(qs.Url, 60) //60秒超时
		if ((qs.SuccessFlag == "") || (qs.SuccessFlag != "" && result == qs.SuccessFlag)) && status == true {
			logValue = " exec success "
		} else {
			logValue = " exec fail "
		}
		common.FileLog("watchLog", key+logValue)
	} else if strings.ToLower(qs.Exec) == "post" {
		result, status := common.Post(qs.Url, "", "", 60) //60秒超时
		if ((qs.SuccessFlag == "") || (qs.SuccessFlag != "" && result == qs.SuccessFlag)) && status == true {
			logValue = " exec success "
		} else {
			logValue = " exec fail "
		}
		common.FileLog("watchLog", key+logValue)
	} else if strings.ToLower(qs.Exec) == "email" {
		err, flag := common.SendEmail(qs.SmtpUser, qs.Sender, qs.SmtpPwd, qs.SmtpHost+":"+qs.SmtpPort, qs.Receive, qs.Title, qs.Content, "html", key)
		if flag == true { //邮件发送成功
			common.FileLog("watchLog", key+" send success")
		} else {
			common.FileLog("watchLog", key+" send fail "+err)
		}
	}
}

//监听redis list,当key发生变化,执行一个GET请求,key存放的内容就是需要请求的URL
func WatchKey() {
	keys, err := common.RedisClient.LRange("GoJob_watchKeyList", 0, -1).Result()
	if err != nil {
		common.FileLog("server", fmt.Sprintf("WatchKey获取keys出错%v", err))
		return
	}

	quelen := len(keys)
	if quelen > 0 {
		for _, key := range keys {
			go HandelWatchKey(key)
		}
	} else {
		if common.NowTimeInt()%600 == 0 {
			common.FileLog("heartBeat", "global WatchKey idel")
		}
	}
}

//永久执行的秒级任务
func RunRepeatSecond() {
	quelen, lErr := common.RedisClient.LLen("GoJob_runAlways_second").Result()
	if lErr != nil {
		common.FileLog("server", "获取数据出错RunRepeatSecond", lErr)
		return
	}
	needOpenNum := int(quelen)
	if needOpenNum > 0 {
		for i := 0; i < needOpenNum; i++ {
			go HandelMultiJob("GoJob_runAlways_second", needOpenNum) //执行任务
		}
	} else {
		if common.NowTimeInt()%60 == 0 {
			common.FileLog("heartBeat", "RunRepeatSecond idel")
		}
	}
}

//永久执行的非秒级任务
func RunRepeatMinute() {
	quelen, lErr := common.RedisClient.LLen("GoJob_runAlways_minute").Result()
	if lErr != nil {
		common.FileLog("server", "获取数据出错RunRepeatMinute", lErr)
		return
	}
	needOpenNum := int(quelen)
	if needOpenNum > 0 {
		for i := 0; i < needOpenNum; i++ {
			go HandelMultiJob("GoJob_runAlways_minute", needOpenNum) //执行任务
		}
	} else {
		if common.NowTimeInt()%60 == 0 {
			common.FileLog("heartBeat", "RunRepeatMinute idel")
		}
	}
}

//运行一次的任务
func RunOnce() {
	//如果是第一次启动或者能整除60的检查一次等待运行的任务
	if common.NowTimeInt()%60 == 0 {
		waitResult, waitErr := common.RedisClient.LRange("GoJob_WaitingToRunTask", 0, -1).Result()
		if waitErr != nil {
			common.FileLog("server", "RunOnce waiting Job lrange error", waitErr)
			return
		}

		waitlen := len(waitResult)
		if waitlen > 0 {
			for _, jobKey := range waitResult {
				waitInfo, wiErr := common.RedisClient.Get(jobKey).Result()
				if wiErr != nil {
					common.FileLog("server", "RunOnce waiting Job GetKey error", jobKey, wiErr)
				} else {
					var data CronJob
					unRrr := json.Unmarshal([]byte(waitInfo), &data)
					if unRrr != nil {
						common.FileLog("server", "RunOnce waiting Job GetKey jsondecode error", unRrr)
					} else {
						//判断时间还有多少秒执行
						diffSecond, _ := common.TimeDiffer(common.NowString(), data.RunDateTime)
						if diffSecond <= 60 && diffSecond > -600 {
							common.RedisClient.LRem("GoJob_WaitingToRunTask", 1, jobKey)
							go readyToRun(jobKey, data)
						} else if diffSecond < -600 { //过期太久也不运行了,直接删掉
							common.RedisClient.LRem("GoJob_WaitingToRunTask", 1, jobKey)
						}
					}
				}
			}
		}
	}

	//读取任务
	quelen, lErr := common.RedisClient.LLen("GoJob_runOnce").Result()
	if lErr != nil {
		common.FileLog("server", "获取数据出错GoJob_runOnce", lErr)
	}

	//确定启动协程数量,如果队列数小于协程数,则启动对应的队列数量即可,如果队列数量大于协程数量,则启动定义的协程数量
	needOpenNum := int(quelen)
	if needOpenNum > 0 {
		for i := 0; i < needOpenNum; i++ {
			//最高启动协程数量,以免队列里面的数据太多会造成系统崩溃
			go HandelOneceJob("GoJob_runOnce") //执行任务
		}
	} else {
		if common.NowTimeInt()%60 == 0 {
			common.FileLog("heartBeat", fmt.Sprintf("RunOnce idel"))
		}
	}
}

//处理重复运行的任务
func AlwaysSecondJobDetail(KeyName string) {
	if KeyName != "" {
		result, _ := common.RedisClient.Get(KeyName).Result()

		var data CronJob
		unRrr := json.Unmarshal([]byte(result), &data)

		if unRrr != nil {
			fmt.Println(KeyName, "AlwaysSecondJobDetail,json解析出错", unRrr)
			common.FileLog("server", KeyName+" HandelOnceJob json解析出错,", unRrr)
			return
		}

		//检查是否到时间运行了
		betWeenNum, _ := strconv.ParseInt(data.BetweenNum, 10, 64)
		if betWeenNum < 0 {
			betWeenNum = 1 //运行间隙数值最小是1秒
		}
		if data.RunBetween == "minute" {
			betWeenNum = betWeenNum * 60
		} else if data.RunBetween == "hour" {
			betWeenNum = betWeenNum * 3600
		} else if data.RunBetween == "day" {
			betWeenNum = betWeenNum * 86400
		}

		var runNow bool                  //标记是否现在运行
		nowTime64 := common.NowTimeInt() //当前时间戳
		//秒级间隙任务,且间隙小于2秒的,直接就运行
		if betWeenNum <= 60 && data.RunBetween == "second" {
			if nowTime64%betWeenNum == 0 {
				runNow = true
			}
		} else {
			//获取上次运行的时间
			lastTimeStr, _ := common.RedisClient.HGet("GoJob_AlwaysBetwween", KeyName).Result()
			//如果没有上次运行时间的,则需要判断是否到了运行时间
			if lastTimeStr == "" {
				isRunTime, _ := common.TimeDiffer(common.NowString(), data.RunDateTime)
				if isRunTime <= 0 && isRunTime > -1200 { //到运行时间了
					//设置这次运行的时间
					_, hsErr := common.RedisClient.HSet("GoJob_AlwaysBetwween", KeyName, nowTime64).Result()
					if hsErr != nil {
						logVal := fmt.Sprintf("AlwaysSecondJobDetail %s 设置间隙出错%v", KeyName, hsErr)
						fmt.Println(logVal)
						common.FileLog("server", logVal)
						return
					}
					runNow = true //可运行
				} else {
					return //没有运行过,且还没到运行时间不运行
				}
			} else { //上次有运行时间的，判断运行间隙是否到了
				lastExecTime, _ := strconv.ParseInt(lastTimeStr, 10, 64)
				runBetweenNum := nowTime64 - lastExecTime
				if runBetweenNum > betWeenNum { //上次运行间隙时间大于设定的时间,则运行
					go readyToRun(KeyName, data)
					//设置这次运行的时间
					_, hsErr := common.RedisClient.HSet("GoJob_AlwaysBetwween", KeyName, nowTime64).Result()
					if hsErr != nil {
						logVal := fmt.Sprintf("AlwaysSecondJobDetail %s 设置间隙出错%v", KeyName, hsErr)
						fmt.Println(logVal)
						common.FileLog("server", logVal)
						return
					}
					runNow = true //设置可执行
				}
			}
		}

		//可马上运行的任务
		if runNow == true {
			go readyToRun(KeyName, data)
		}
	}
}

//处理重复运行任务
func HandelMultiJob(listName string, jobNum int) {
	if jobNum < 1 {
		return
	}
	jobList, _ := common.RedisClient.LRange(listName, 0, -1).Result()
	for _, jobKey := range jobList {
		go AlwaysSecondJobDetail(jobKey) //处理非秒级且重复的定时任务
	}
}

//处理单个任务
func HandelOneceJob(listName string) {
	//jobKey, _ := common.RedisClient.RPop(listName).Result()
	popResult, poperr := common.RedisClient.BRPop(1*time.Second, listName).Result()
	if poperr != nil {
		pError := fmt.Sprintf("HandelOneceJob,rpop err,%v,%v", popResult, poperr)
		fmt.Println(pError)
		common.FileLog("server", pError)
		return
	}

	jobKey := popResult[1]
	if jobKey != "" {
		result, _ := common.RedisClient.Get(jobKey).Result()
		var data CronJob
		unRrr := json.Unmarshal([]byte(result), &data)
		if unRrr != nil {
			deErr := fmt.Sprintf("HandelOnceJob,json解析出错,%v,%v", jobKey, result)
			fmt.Println(deErr)
			common.FileLog("server", deErr)
			return
		}

		diffSecond, _ := common.TimeDiffer(common.NowString(), data.RunDateTime)
		if diffSecond > 60 { //运行时间大于60秒的都会push到等待运行的队列
			common.RedisClient.LPush("GoJob_WaitingToRunTask", jobKey)
			return
		}

		//可以执行的任务
		go readyToRun(jobKey, data)
	}
}

//执行具体操作
func readyToRun(key string, s CronJob) {
	var retStr string
	//判断时间还有多少秒执行
	diffSecond, _ := common.TimeDiffer(common.NowString(), s.RunDateTime)
	if s.Exec == "get" || s.Exec == "post" || s.Exec == "multiGet" || s.Exec == "multiPost" {
		//重复运行的,或者只运行一次的,只运行一次的需要判断运行时间在十分钟内的才执行
		if s.IsRepeat == "1" || (diffSecond <= 0 && diffSecond > -600) {
			//单个get/post允许失败重连,批量的只执行一次,失败不重连
			if s.Exec == "get" {
				retStr, _ = common.Get(s.Content, time.Duration(s.TimeOut))
			} else if s.Exec == "post" {
				retStr, _ = common.Post(s.Content, "", "", time.Duration(s.TimeOut))
			} else if s.Exec == "multiGet" {
				urls := common.Explode("|", s.Content)
				if len(urls) > 0 {
					for k, httpUrl := range urls {
						go execGet(key, httpUrl, s)
						if k > 0 && k%10000 == 0 {
							time.Sleep(1 * time.Second)
						}
					}
				}
			} else if s.Exec == "multiPost" {
				urls := common.Explode("|", s.Content)
				if len(urls) > 0 {
					for k, httpUrl := range urls {
						go execPost(key, httpUrl, s)
						if k > 0 && k%10000 == 0 {
							time.Sleep(1 * time.Second)
						}
					}
				}
			}
		} else if diffSecond > 0 && diffSecond <= 60 {
			time.Sleep(time.Duration(diffSecond) * time.Second)
			//单个get/post允许失败重连,批量的只执行一次,失败不重连
			if s.Exec == "get" {
				retStr, _ = common.Get(s.Content, time.Duration(s.TimeOut))
			} else if s.Exec == "post" {
				retStr, _ = common.Post(s.Content, "", "", time.Duration(s.TimeOut))
			} else if s.Exec == "multiGet" {
				urls := common.Explode("|", s.Content)
				if len(urls) > 0 {
					for _, httpUrl := range urls {
						go execGet(key, httpUrl, s)
					}
				}
			} else if s.Exec == "multiPost" {
				urls := common.Explode("|", s.Content)
				if len(urls) > 0 {
					for k, httpUrl := range urls {
						go execPost(key, httpUrl, s)
						if k > 0 && k%10000 == 0 {
							time.Sleep(1 * time.Second)
						}
					}
				}
			}
		} else if diffSecond <= -600 {
			//删除运行中的任务，过期时间太长了,不再执行了
			common.FileLog("runningLog", key+"任务过期时间超过10分钟,不执行了"+s.Content)
			common.RedisClient.LRem("GoJob_WaitingToRunTask", 1, key)
			return
		}

		var logValue string
		//RunStatus   int //0失败,1运行成功,2运行失败,3运行失败等待重试中
		if s.SuccessFlag == "" && (s.Exec == "get" || s.Exec == "post") && s.RetryNumInt == 0 {
			logValue = " exec success "
			s.RunStatus = 1 //执行成功
			common.FileLog("runningLog", key+logValue+s.Content)
		} else if (s.Exec == "get" || s.Exec == "post") && retStr != s.SuccessFlag && s.RetryNumInt > 0 {
			s.RunStatus = 3
			s.RetryNumInt = s.RetryNumInt - 1
			logValue = " exec retry "
			common.FileLog("runningLog", key+logValue+s.Content)
			time.Sleep(60 * time.Second)
			readyToRun(key, s) //执行失败,重试执行
		} else if (s.Exec == "get" || s.Exec == "post") && retStr != s.SuccessFlag && s.RetryNumInt == 0 {
			s.RunStatus = 2 //执行失败,且重试次数用完
			logValue = " exec fail "
			common.FileLog("runningLog", key+logValue+s.Content)
		} else if (s.Exec == "get" || s.Exec == "post") && retStr == s.SuccessFlag {
			logValue = " exec success "
			s.RunStatus = 1 //执行成功
			common.FileLog("runningLog", key+logValue+s.Content)
		} else if s.Exec == "multiGet" || s.Exec == "multiPost" {
			//multiSet,multiGet日志
			common.FileLog("runningLog", key+" "+s.Exec+" 已执行")
		}
	}

	//邮件和批量邮件
	if s.Exec == "email" || s.Exec == "multiEmail" {
		//获取邮件smtp配置
		var smtpConfig usual.EmailServer
		_, smtpConfig = usual.SmtpConfig(s.SmtpServer)
		if smtpConfig.Account == "" || smtpConfig.Host == "" || smtpConfig.Password == "" || smtpConfig.Port == "" {
			//删除运行中的任务
			common.RedisClient.LRem("GoJob_WaitingToRunTask", 1, key)
			common.FileLog("runningLog", key+" "+s.Exec+" 邮件配置不完整")
			return
		}

		if s.IsRepeat == "1" || (diffSecond <= 0 && diffSecond > -600) { //十分钟内的才执行
			//smtpUser, sendUserName, password, hostAndPort, toUser, subject, body, mailtype, redisKey string
			if s.Exec == "email" {
				fmt.Println("准备发送邮件中1...")
				go runSend(smtpConfig, s, key, "")
			} else if s.Exec == "multiEmail" {
				multiUsers, status := usual.SendUsers(s.EmailUsers)
				if status == false {
					common.FileLog("runningLog", key+" 收件人列表获取失败 "+s.EmailUsers)
				}
				if len(multiUsers) > 0 {
					for k, emailAccount := range multiUsers {
						go runSend(smtpConfig, s, key, emailAccount)
						if k > 0 && k%1000 == 0 {
							time.Sleep(1 * time.Second) //批量邮件每启动1000个协程休息一秒
						}
					}
				}
			}
		} else if diffSecond > 0 && diffSecond <= 60 {
			time.Sleep(time.Duration(diffSecond) * time.Second)
			if s.Exec == "email" {
				go runSend(smtpConfig, s, key, "")
			} else {
				multiUsers, status := usual.SendUsers(s.EmailUsers)
				if status == false {
					fmt.Println(key + " 收件人列表获取失败... " + s.EmailUsers)
					common.FileLog("runningLog", key+" 收件人列表获取失败 "+s.EmailUsers)
				}
				if len(multiUsers) > 0 {
					for k, emailAccount := range multiUsers {
						go runSend(smtpConfig, s, key, emailAccount)
						if k > 0 && k%1000 == 0 {
							time.Sleep(1 * time.Second) //批量邮件每启动1000个协程休息一秒
						}
					}
				}
			}
		} else if diffSecond <= -600 {
			common.FileLog("runningLog", key+" "+s.Exec+" 超过时间太久,不再执行")
			return
		}
	}

	//删除运行中的任务
	common.RedisClient.LRem("GoJob_WaitingToRunTask", 1, key)
}

//multiGet 不做失败重试
func execGet(key, url string, s CronJob) {
	result, status := common.Get(url, time.Duration(s.TimeOut))
	var logValue string
	if result == s.SuccessFlag && status == true {
		logValue = " exec success "
	} else {
		logValue = " exec fail "
	}
	common.FileLog("runningLog", key+logValue+s.Content)
}

//multiPost 也不做失败重试go common.Post(httpUrl, "", "", time.Duration(s.TimeOut))
func execPost(key, url string, s CronJob) {
	result, status := common.Post(url, "", "", time.Duration(s.TimeOut))
	var logValue string
	if result == s.SuccessFlag && status == true {
		logValue = " exec success "
	} else {
		logValue = " exec fail "
	}
	common.FileLog("runningLog", key+logValue+s.Content)
}

//发送邮件
func runSend(sc usual.EmailServer, s CronJob, key, multiUser string) {
	var receiveEmail string
	if multiUser != "" {
		receiveEmail = multiUser
	} else {
		receiveEmail = s.EmailUsers
	}
	err, flag := common.SendEmail(sc.Account, s.SendName, sc.Password, sc.Host+":"+sc.Port, receiveEmail, s.EmailTitle, s.Content, "html", key)
	if flag == true { //邮件发送成功
		common.FileLog("runningLog", key+" send success")
	} else {
		common.FileLog("runningLog", key+" send fail "+err)
	}
}
