package usual

import (
	"cronJob/common"
	"encoding/json"
)

type EmailServer struct {
	Host     string
	Port     string
	Account  string
	Password string
}

//获取发送邮箱配置,如果参数不是all则会根据参数匹配账号返回
func SmtpConfig(filterByAccount string) ([]EmailServer, EmailServer) {
	common.LastError()
	emailConfig, _ := common.RedisClient.LRange("GoJob_SmtpJsonQueue", 0, -1).Result()
	ecl := len(emailConfig)
	var m = make([]EmailServer, ecl)
	var es = EmailServer{}

	if ecl > 0 {
		for k, v := range emailConfig {
			eConfig, deError := common.Json_decode(v)
			if deError != nil {
				common.FileLog("emailLog", "邮箱服务器配置出错", deError)
			} else {
				var host, port, account, pwd string
				host = common.MapExist(eConfig, "Host")
				port = common.MapExist(eConfig, "Port")
				account = common.MapExist(eConfig, "Account")
				pwd = common.MapExist(eConfig, "Password")
				if filterByAccount == "all" {
					m[k].Host = host
					m[k].Port = port
					m[k].Account = account
					m[k].Password = pwd
				} else if filterByAccount == account {
					es.Host = host
					es.Port = port
					es.Account = account
					es.Password = pwd
				}
			}
		}
	}
	return m, es
}

//发送邮件的用户
func SendUsers(key string) ([]string, bool) {
	common.LastError()
	emails, _ := common.RedisClient.LRange("GoJob_multiEmails", 0, -1).Result()
	ulen := len(emails)
	var m = make([]string, ulen)
	if ulen > 0 {
		if key != "" {
			data, err := common.RedisClient.Get(key).Result()
			if err != nil {
				return m, false
			}
			deErr := json.Unmarshal([]byte(data), &m)
			if deErr != nil {
				return m, false
			}
		} else {
			m = emails
		}
	}
	return m, true
}
