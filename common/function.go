package common

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v7"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"io"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
	//导入gorm
	"github.com/jinzhu/gorm"
)

func ConnRedis() {
	//如果没有连接redis则链接,已经连接了就不再连接
	if Redis_Connected == false {
		RedisClient = redis.NewClient(&redis.Options{
			Addr:     REDIS_HOST,
			Password: REDIS_PWD,
			DB:       REDIS_DB,
		})
		Redis_Connected = true //设置redis已经连接,不要关闭连接,它已经维护连接池了
	}
}
func ConnDb() {
	//配置MySQL连接参数
	//通过前面的数据库参数，拼接MYSQL DSN， 其实就是数据库连接串（数据源名称）
	//MYSQL dsn格式： {username}:{password}@tcp({host}:{port})/{Dbname}?charset=utf8&parseTime=True&loc=Local
	//类似{username}使用花括号包着的名字都是需要替换的参数
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", DB_USER, DB_PWD, DB_HOST, DB_PORT, DB_NAME)
	//连接MYSQL
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		panic("连接数据库失败, error=" + err.Error())
	}
	db.LogMode(true)
	db.DB().SetMaxOpenConns(100) //设置数据库连接池最大连接数
	db.DB().SetMaxIdleConns(20)  //连接池最大允许的空闲连接数,如果没有sql任务需要执行的连接数大于20,超过的连接会被连接池关闭。
	DB_CLIENT = db
}

// 发送GET请求
// url：         请求地址
// response：    请求返回的内容
func Get(url string, timeOut time.Duration) (info string, status bool) {
	defer LastError()
	// 超时时间：30秒
	client := &http.Client{Timeout: timeOut * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return fmt.Sprintf("%v", err), false
		}
	}
	return result.String(), true
}

// 发送POST请求
// url：         请求地址
// data：        POST请求提交的数据
// contentType： 请求体格式，如：application/json
func Post(url string, data interface{}, contentType string, timeOut time.Duration) (info string, status bool) {
	defer LastError()
	// 超时时间
	client := &http.Client{Timeout: timeOut * time.Second}
	jsonStr, _ := json.Marshal(data)
	resp, err := client.Post(url, contentType, bytes.NewBuffer(jsonStr))
	if err != nil {
		return fmt.Sprintf("%v", err), false
	}
	defer resp.Body.Close()

	result, _ := ioutil.ReadAll(resp.Body)
	return string(result), true
}

//获取当前时间字符串
func NowString() string {
	return time.Now().In(TIME_ZONE).Format("2006-01-02 15:04:05")
}

//时间字符串转换成int64时间戳
func TimeStringToInt(timeString string) int64 {
	strTimeToIntTime, err := time.ParseInLocation("2006-01-02 15:04:05", timeString, TIME_ZONE)
	if err != nil {
		return 0
	}
	return strTimeToIntTime.Unix()
}

//获取当前时间字符串
func NowTimeInt() int64 {
	return time.Now().In(TIME_ZONE).Unix()
}

//获取相差时间
func TimeDiffer(start_time, end_time string) (int64, error) {
	t1, err1 := time.Parse("2006-01-02 15:04:05", start_time)
	t2, err2 := time.Parse("2006-01-02 15:04:05", end_time)
	if err1 != nil {
		return 0, err1
	}

	if err2 != nil {
		return 0, err1
	}
	return t2.In(TIME_ZONE).Unix() - t1.In(TIME_ZONE).Unix(), nil
}

//检查map里面是否存在某个key
func MapExist(m map[string]interface{}, key string) string {
	if _, ok := m[key]; ok {
		return fmt.Sprintf("%v", m[key])
	} else {
		return ""
	}
}

//解析json字符串成 map
func JsonStringToMap(jsonStr string) (m map[string]interface{}, err error) {
	a := map[string]interface{}{}
	unmarsha1Err := json.Unmarshal([]byte(jsonStr), &a)
	if unmarsha1Err != nil {
		return nil, unmarsha1Err
	}
	return a, nil
}

//记录最终出错,处理致命错误
func LastError() {
	if r := recover(); r != nil {
		shutError := fmt.Sprintf("发生致命错误%v", r)
		FileLog("server", shutError)
		fmt.Println(NowString(), shutError)
	}
}
func Md5(str string) string {
	h := md5.New()
	io.WriteString(h, str) //e10adc3949ba59abbe56e057f20f883e
	return fmt.Sprintf("%x", h.Sum(nil))
}

//返回系统字符串：linux	windows
func osType() (sysType string) {
	sysType = runtime.GOOS
	return
}

//中文字符串截取函数
func Substring(s string, start int, subLen int) string {
	strArr := []rune(s) //这里不要byte,byte的范围是0-255,中文是多少万的
	var endLen int = start + subLen
	var returnString string
	strlen := len(strArr)
	for i := 0; i < strlen; i++ {
		if i >= start && i < endLen {
			returnString = returnString + string(strArr[i])
		} else if i > subLen {
			break //长度超过就跳出循环
		}
	}
	return returnString
}

//去除左右的空白字符串
func TrimSpace(text string) string {
	return strings.TrimSpace(text)
}

//去除左右的指定字符串
func Trim(text, delimiter string) string {
	return strings.Trim(text, delimiter)
}
func Ltrim(text, delimiter string) string {
	return strings.TrimLeft(text, delimiter)
}
func Rtrim(text, delimiter string) string {
	return strings.TrimRight(text, delimiter)
}
func Explode(delimiter, text string) []string {
	if len(delimiter) > len(text) {
		return strings.Split(delimiter, text)
	} else {
		return strings.Split(text, delimiter)
	}
}
func Implode(glue string, pieces []string) string {
	return strings.Join(pieces, glue)
}

//翻转切片
func ReverseSlice(reverseSlice []string) []string {
	rlen := len(reverseSlice)
	if rlen < 1 { //切片为空直接返回
		return reverseSlice
	}
	var newSlice []string
	for i := rlen - 1; i >= 0; i-- {
		newSlice = append(newSlice, reverseSlice[i])
	}
	return newSlice
}

//用用户输入的密码跟盐进行加密跟保存的密码进行对比
func PasswordEncode(str, salt string) string {
	return Md5(Md5(str) + salt)
}
func Base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}
func Base64Decode(str string) (string, error) {
	b64DeString, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}
	return string(b64DeString), nil
}
func UrlEncode(str string) string {
	return url.QueryEscape(str)
}
func UrlDecode(str string) string {
	deStr, err := url.QueryUnescape(str)
	if err != nil {
		return ""
	}
	return deStr
}
func Json_decode(data string) (map[string]interface{}, error) {
	var dat map[string]interface{}
	err := json.Unmarshal([]byte(data), &dat)
	return dat, err
}
func Json_encode(data interface{}) (string, error) {
	jsons, err := json.Marshal(data)
	return string(jsons), err
}

//获取相差时间
func TimeDifferSecond(start_time, end_time string) (int64, error) {
	t1, err1 := time.Parse("2006-01-02 15:04:05", start_time)
	t2, err2 := time.Parse("2006-01-02 15:04:05", end_time)
	if err1 != nil {
		return 0, err1
	}

	if err2 != nil {
		return 0, err1
	}
	return t2.Unix() - t1.Unix(), nil
}

//判断文件是否存在  存在返回 true 不存在返回false
func FileExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

//记录文本日志,100M分一个文件
func FileLog(fileName string, logInfos ...interface{}) {
	filePath := ROOTPATH + "/log/" + fileName + ".log"
	fileInfo, fiErr := os.Stat(filePath)
	if fiErr == nil {
		if fileInfo.Size() > 104857600 { //100M分一个文件
			nowDayTime := time.Now().Format("20060102150405")
			backupName := ROOTPATH + "/log/" + fileName + "_" + nowDayTime + ".log"
			reErr := os.Rename(filePath, backupName)
			if reErr != nil {
				fmt.Println("重命名失败", reErr)
			}
		}
	}
	var f *os.File
	var err error
	if FileExist(filePath) {
		f, err = os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0666)
	} else {
		f, err = os.Create(filePath)
	}
	if err != nil {
		fmt.Println("文件操作异常", err)
		return
	}

	logValue := time.Now().Format("2006-01-02 15:04:05") + " "
	for _, v := range logInfos {
		logValue += fmt.Sprintln(v)
	}
	_, writeErr := io.WriteString(f, logValue)
	if writeErr != nil {
		fmt.Println("写入文件出错", writeErr)
	}
	defer f.Close() //一定要关闭文件句柄
}

//清空文件内容
func FlushFile(file string) error {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		return err
	}
	_, wErr := f.WriteString("")
	if wErr != nil {
		return wErr
	}
	return nil
}

//读取CSV文件
func ReadCsv(filePath string) (readData [][]string, readError error) {
	file, err := os.Open(filePath)
	if err != nil {
		return readData, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	result, err := reader.ReadAll()
	return result, err
}

//判断是否是文件,返回bool和FileInfo
func IsFile(filePath string) (bool, os.FileInfo) {
	fileInfo, _ := os.Stat(filePath)
	if fileInfo != nil {
		return true, fileInfo
	} else {
		return false, fileInfo
	}
}

//按行读取，7600W行数据2.6G文件5秒钟读取
func ReadFileLine(fileName, keyword string, startLine, limitLine int) (result string, nowReadNum, totalNumber int) {
	var readNum, readNumEnd, alreadyRead int
	if startLine < 0 {
		startLine = 0
	}
	readNumEnd = startLine + limitLine //读取的行数

	var readResult string
	LastError()
	if file, err := os.Open(fileName); err != nil {
		FileLog("server", fmt.Sprintf("文件%s读取出错,%v", fileName, err))
	} else {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if keyword != "" { //关键词过滤
				if strings.Contains(line, keyword) {
					//fmt.Println(readNum, startLine, alreadyRead, limitLine)
					if readNum >= startLine && alreadyRead < limitLine {
						alreadyRead++
						line = strings.TrimSpace(line) //读取到每行的结果
						readResult = readResult + line + "\n"
						//fmt.Println("读取到的结果", readResult)
					} else if startLine > 0 {
						//break//读取够了,且不是直接退出
					}
					readNum++
					totalNumber++ //筛选到的文件的总行数
				}
			} else {
				if readNum >= startLine && alreadyRead < limitLine {
					alreadyRead++
					line = strings.TrimSpace(line) //读取到每行的结果
					readResult = readResult + line + "\n"
				}
				readNum++
				totalNumber++ //文件的总行数
			}
		}
	}
	return readResult, readNumEnd, totalNumber
}

//生成32位md5字串
func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

//生成Guid字串
func UniqueId() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return GetMd5String(base64.URLEncoding.EncodeToString(b))
}

//发送邮件
func SendEmail(smtpUser, sendUserName, password, hostAndPort, toUser, subject, body, mailtype, redisKey string) (string, bool) {
	defer LastError() //记录错误
	if subject == "" || body == "" || IsEmail(toUser) == false {
		FileLog("runningLog", redisKey, toUser, "有标题、内容为空或邮箱不正确的邮件")
		return redisKey + "数据不完整", false
	}

	hp := strings.Split(hostAndPort, ":")
	auth := smtp.PlainAuth("", smtpUser, password, hp[0])
	var content_type string
	if mailtype == "html" {
		content_type = "Content-Type: text/" + mailtype + "; charset=UTF-8"
	} else {
		content_type = "Content-Type: text/plain" + "; charset=UTF-8"
	}

	msg := []byte("To: " + toUser + "\r\nFrom: " + sendUserName + "<" + smtpUser + ">" + "\r\nSubject: " + subject + "\r\n" + content_type + "\r\n\r\n" + body)
	send_to := strings.Split(toUser, ";")
	err := smtp.SendMail(hostAndPort, auth, smtpUser, send_to, msg)

	if err != nil {
		sendErr := fmt.Sprintf("邮件发送错误原因：%v", err)
		fmt.Println(NowString(), toUser+"email发送失败", sendErr, "当前协程数量")
		return sendErr, false
	} else {
		fmt.Println(NowString(), toUser+"email发送成功")
		return "", true
	}
}

//校验电子邮箱
func IsEmail(email string) bool {
	result, _ := regexp.MatchString(`^([\w\.\_\-]{1,32})@(\w{1,}).([a-z]{1,12})$`, email)
	if result {
		return true
	} else {
		return false
	}
}

//校验url
func IsUrl(url string) bool {
	result := strings.Contains(url, "http")
	if result {
		return true
	} else {
		return false
	}
}

type FileStruct struct {
	FileName       string //文件名称
	FilePath       string //文件夹
	ModTime        string //修改时间
	FileByteSize   int64  //大小,单位byte
	FileSizeString string //大小带单位
	Mode           string //权限
}

//读取文件夹的文件
func ReadDirFile(filePath string) (fi []FileStruct, err error) {
	data, err := ioutil.ReadDir(filePath)
	if err != nil {
		return fi, err
	}

	for _, v := range data {
		if v.IsDir() == true {
			continue //如果是dir直接跳走
		}
		var fileDetail FileStruct
		fileDetail.FileName = v.Name()
		fileDetail.FilePath = filePath
		fileDetail.ModTime = Substring(fmt.Sprintf("%v", v.ModTime()), 0, 19)
		fileDetail.FileByteSize = v.Size()

		var fsize string
		if v.Size() < 1024 {
			fsize = fmt.Sprintf("%vB", v.Size())
		} else if v.Size() >= 1024 && v.Size() < 1024*1024 {
			fsize = fmt.Sprintf("%vK", v.Size()/1024)
		} else if v.Size() >= 1024*1024 {
			fsize = fmt.Sprintf("%vM", v.Size()/1024/1024)
		}
		fileDetail.FileSizeString = fsize

		fileDetail.Mode = fmt.Sprintf("%v", v.Mode())
		fi = append(fi, fileDetail)
	}
	return fi, nil
}

//压缩文件 srcFile could be a single file or a directory
func Zip(srcFile string, destZip string) error {
	zipfile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	fkErr := filepath.Walk(srcFile, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = path
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
		}
		return err
	})

	if fkErr != nil {
		return fkErr
	}
	return err
}

//读取文件最后几字节的内容,建议可以被三整除的数,不然中文有一个字会乱码
func ReadLastFile(fname, splitStr string, startSize64 int64) string {
	file, err := os.Open(fname)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	defer file.Close()

	stat, err := os.Stat(fname)

	var startSize int64
	fileSize := stat.Size()
	if fileSize > startSize64 {
		startSize = fileSize - startSize64
	}
	if startSize <= 0 {
		startSize64 = fileSize
	}
	buf := make([]byte, startSize64)

	_, err = file.ReadAt(buf, startSize)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	newSlice := Explode(splitStr, string(buf))
	var sLen = len(newSlice)

	var reverseSlice []string
	for i := sLen - 1; i > 0; i-- {
		reverseSlice = append(reverseSlice, newSlice[i])
	}
	return Implode(splitStr, reverseSlice)
}
