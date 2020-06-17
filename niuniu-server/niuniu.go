package main

import (
	"fmt"
	"math/rand"
	"servers/common-library/config"
	"servers/common-library/function"
	"servers/common-library/log"
	"servers/niuniu-server/niuniuApp"
	"time"

	"github.com/hh8456/go-common/redisObj"
	"github.com/hh8456/go-common/snowflake"
	"gopkg.in/natefinch/lumberjack.v2"

	"net/http"
	_ "net/http/pprof"
)

func main() {
	err := config.InitSConfig("../config/server_config.yml")
	if err != nil {
		log.Error(err)
	}

	logger := &lumberjack.Logger{
		// 日志输出文件路径
		//Filename: "../log/niuniu" + time.Now().Format("2006-01-02.15:04:05") + ".log",
		Filename: "../log/niuniu.log",
		// 日志文件最大 size, 单位是 MB
		MaxSize: 4096, // megabytes
		// 最大过期日志保留的个数
		MaxBackups: 100,
		// 保留过期文件的最大时间间隔,单位是天
		MaxAge: 3, //days
		// 是否需要压缩滚动日志, 使用的 gzip 压缩
		//Compress: true, // disabled by default
	}
	log.SetOutput(logger) //调用 logrus 的 SetOutput()函数
	log.SetLevel(log.DebugLevel)

	defer function.Catch()
	rand.Seed(time.Now().UnixNano())

	go func() {
		// http://192.168.0.155:3411/debug/pprof/
		pprofAddr := "0.0.0.0:3411"
		http.ListenAndServe(pprofAddr, nil)
	}()

	snowflake.SetMachineId(401)

	redisObj.Init(config.Cfg.RdsAddr, "")
	//redisObj.Init("192.168.0.155:6379", "")
	niuniuObj := niuniuApp.CreateNiuniuApp()
	niuniuObj.Run("0.0.0.0:3401")

	select {}

	fmt.Println("vim-go")
}
