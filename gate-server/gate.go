package main

import (
	"fmt"
	"servers/common-library/log"
	"servers/gate-server/gateApp"

	"github.com/hh8456/go-common/snowflake"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	logger := &lumberjack.Logger{
		// 日志输出文件路径
		//Filename: "gate_" + time.Now().Format("2006-01-02.15:04:05") + ".log",
		Filename: "../log/gate.log",
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

	// XXX 是否加载这个 hook, 由配置决定
	log.AddHook(&log.WriteConsoleHook{})

	snowflake.SetMachineId(201)
	gateObj := gateApp.CreateGateApp()
	gateObj.Run("0.0.0.0:3201")

	select {}

	fmt.Println("vim-go")
}
