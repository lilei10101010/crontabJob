![Image text](https://github.com/lilei10101010/crontabJob/blob/master/static/image/login.png?raw=true)
![Image text](https://github.com/lilei10101010/crontabJob/blob/master/static/image/systemIndex.png?raw=true)
![Image text](https://github.com/lilei10101010/crontabJob/blob/master/static/image/addCrontab.png?raw=true)
![Image text](https://github.com/lilei10101010/crontabJob/blob/master/static/image/addlisten.png?raw=true)
![Image text](https://github.com/lilei10101010/crontabJob/blob/master/static/image/listen.png?raw=true)

##crontab job可以干嘛？  
1.他可以帮你在指定的时间请求某个url,指定的时间发送邮件  
2.它可以帮你监听redis的队列，请求队列的url或者发送队列里面的邮件  
3.你也可以把它看做是一个秒级的定时器  
 

##服务器设置  
1.默认监听9527端口,在main.go文件更改  
2.时区默认设置为北京时间，在main.go文件TIME_ZONE更改  
3.默认使用redis 127.0.0.1:6379，如需要更改在common/config.go更改,然后重新编译  

##初始化设置  
1.在redis里面设置 set  GoJob_user_admin "e10adc3949ba59abbe56e057f20f883e" （即账号:admin密码123456）,这是必须的不然没有账号登录后台  
2.所有的Redis Key都以GoJob_开头，除了监听的redis key,监听的KEY是你写什么就监听什么KEY  

##如果您不需要更改配置文件，我们还提供编译好的文件  
windows用户解压main.zip文件得到main.exe双击即可运行  
linux用户解压cronJob.tar.gz文件得到linux的可执行文件  

如果您不需要二次开发的话上面的设置完成"服务器设置"和"初始化设置"即可使用,输入：http://localhost:9527/ 即可到登录页面

##下面是开发文档
这是项目所有使用的第三方包,其余的都是官方提供的包  
github.com/gin-contrib/sessions v0.0.3  
github.com/gin-gonic/gin v1.6.3  
github.com/go-redis/redis/v7 v7.4.0  
github.com/jinzhu/gorm v1.9.16  
github.com/robfig/cron v1.2.0  


##监控队列：  
例如监听一个叫aaa的key执行的操作是GET/POST则(用任何语言按照下面格式写进redis即可),  
在Redis直接 lpush aaa "{\"Exec\":\"get\",\"Url\":\"https:\/\/www.buruyouni.com/spider/request?rid=watchKeyGetMethod\"}"  
在Redis直接 lpush bbb "{\"Exec\":\"post\",\"Url\":\"https:\/\/www.buruyouni.com/spider/request?rid=watchKeyPostMethod\"}"  
程序读取到aaa有内容,则会去执行GET或者POST请求,如果填写了"成功字符串",例如success，那么请求的返回结果是success则这个请求是成功的,否则请求就记录是失败的,如果不填写成功字符串,请求了就算成功(注意字段名Exec,Url一定要完全一致,大小写敏感)

###例如监听一个叫ccc的key执行的操作是email发送邮件,
用任何语言写入Redis的list ccc如下字段的数据即可：  
lpush ccc "{\"Exec\":\"email\",\"SmtpUser\":\"admin@buruyouni.com\",\"SmtpPwd\":\"yourPassword\",\"SmtpHost\":\"smtpdm.aliyun.com\",\"SmtpPort\":\"80\",\"Receive\":\"xxx@xx.com\",\"Sender\":\"MMOGA LTD\",\"Title\":\"Please send testing email for me\",\"Content\":\"When you think it's too late, it's the earliest time\"}"

###Exec是执行操作email必填,SmtpUser是SMTP的用户，SmtpPwd是用户的密码
SmtpHost是SMTP的服务器地址,SmtpPort是端口号
Receive是收件人邮箱,Sender是发件人名称
Title是邮件标题,Content是邮件内容
程序读取到ccc有内容,则会去检查字段是否完整.(注意字段名一定要完全一致,大小写敏感)

##如果需要nginx代理可以配置
 server {  
        listen 80;  
        server_name task.buruyouni.com;  
        location / {  
            rewrite ^ https://$http_host$request_uri? permanent;  
        }  
   }  
 server {  
        listen 443 ssl http2 default_server;  
        server_name task.buruyouni.com;  
        ssl_certificate  /etc/cert/task.buruyouni.com.pem;  
        ssl_certificate_key /etc/cert/task.buruyouni.com.key;  
        ssl_session_timeout 5m;  
        ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE:ECDH:AES:HIGH:!NULL:!aNULL:!MD5:!ADH:!RC4;  
        ssl_protocols TLSv1 TLSv1.1 TLSv1.2;  
        ssl_prefer_server_ciphers on;  
        access_log logs/task.log;  
        error_log logs/task_error.log;  
        location / {  
                proxy_pass http://localhost:9527;  
        }  
   }  

