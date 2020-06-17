package main

import (
	"fmt"
	"servers/common-library/config"
	"servers/common-library/digitalId"
	"servers/common-library/log"
	"servers/lobby-server/lobbyApp"
	"servers/lobby-server/logic"

	"github.com/hh8456/go-common/redisObj"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	err := config.InitSConfig("../config/server_config.yml")
	if err != nil {
		log.Error(err)
	}

	logger := &lumberjack.Logger{
		// 日志输出文件路径
		//Filename: "lobby_" + time.Now().Format("2006-01-02.15:04:05") + ".log",
		Filename: "../log/lobby.log",
		// 日志文件最大 size, 单位是 MB
		MaxSize: 4096, // megabytes
		// 最大过期日志保留的个数
		MaxBackups: 100,
		// 保留过期文件的最大时间间隔,单位是天
		MaxAge: 28, //days
		// 是否需要压缩滚动日志, 使用的 gzip 压缩
		//Compress: true, // disabled by default
	}
	log.SetOutput(logger) //调用 logrus 的 SetOutput()函数
	log.SetLevel(log.TraceLevel)

	log.AddHook(&log.WriteConsoleHook{})

	logic.Init()
	redisObj.Init(config.Cfg.RdsAddr, "")
	digitalId.Gen()
	lobbyObj := lobbyApp.CreateLobbyApp()
	lobbyObj.Run("0.0.0.0:3301")

	select {}

	fmt.Println("vim-go")
}
