package main

import (
	"encoding/binary"
	"fmt"
	"servers/base-library/base_net"
	"servers/common-library/function"
	"servers/common-library/packet"
	"servers/common-library/proto/fishProto"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	logger := &lumberjack.Logger{
		// 日志输出文件路径
		Filename: "robot_" + time.Now().Format("2006-01-02.15:04:05") + ".log",
		// 日志文件最大 size, 单位是 MB
		MaxSize: 1, // megabytes
		// 最大过期日志保留的个数
		MaxBackups: 10,
		// 保留过期文件的最大时间间隔,单位是天
		MaxAge: 28, //days
		// 是否需要压缩滚动日志, 使用的 gzip 压缩
		//Compress: true, // disabled by default
	}

	log.SetOutput(logger) //调用 logrus 的 SetOutput()函数

	addr := "192.168.0.184:3333"
	socket, err := base_net.ConnectSocket("192.168.0.155:3333", uint32(packet.PacketMaxSize_4MB))
	if err != nil {
		log.Errorf("robot 连接后端服务器 addr: %s 错误: %v", addr, err)
		return
	}

	socket.SendPbMsg(0, int32(fishProto.MsgId_c2sPing), &fishProto.C2SPing{Timestamp: 100})

	ackBuf, e := socket.ReadOne()
	if e != nil {
		log.Errorf("robot 读取 socket 出错: %v", e)
		return
	}

	msgId := binary.BigEndian.Uint32(ackBuf[4:8])
	fmt.Printf("收到服务端的 ping, msg id: %d \n", msgId)
	msgLen := binary.BigEndian.Uint32(ackBuf[8:])
	fmt.Printf("收到服务端的 ping, msg length: %d \n", msgLen)

	pb := &fishProto.S2CPing{}
	if function.ProtoUnmarshal(ackBuf[packet.PacketHeaderSize:], pb, "fishProto.S2CPing") {
		fmt.Printf("收到服务端的 ping, timestamp: %d \n", pb.Timestamp)
	}

	socket.Close()

	fmt.Println("vim-go")
}
