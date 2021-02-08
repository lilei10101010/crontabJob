package common

import (
	"github.com/go-redis/redis/v7"
	"github.com/jinzhu/gorm"
	"sync"
	"time"
)

var (
	DB_CLIENT       *gorm.DB
	Redis_Connected bool
	RedisClient     *redis.Client
	TIME_ZONE       *time.Location         //时区设置,在入口文件的地方就设置了
	ROOTPATH        string                 //根目录
	LockQueue       sync.Mutex             //给QueueStruct加锁
	ProcessNum      int            = 10000 //最大并发量
)

//redis连接配置
const REDIS_HOST string = "127.0.0.1:6379"
const REDIS_PWD string = ""
const REDIS_DB int = 0

//mysql连接配置,当前没有使用
const DB_HOST string = "127.0.0.1"
const DB_NAME string = "yangxingyi"
const DB_USER string = "root"
const DB_PWD string = "root"
const DB_PORT int = 3306
