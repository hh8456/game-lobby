package gateServer

import (
	"servers/base-library/base_net"
	"servers/common-library/connData"
	"servers/common-library/function"
	"servers/common-library/gameKeyPrefix"
	"servers/common-library/log"
	"servers/common-library/proto/lobbyProto"
	"servers/common-library/proto/msgIdProto"
	"sync/atomic"
	"time"
)

var (
	gSerialNum uint32
)

type GateServer struct {
	*base_net.Connection
	remoteGateServerId uint32 // 远程 gate 服务器 id
	isRun              uint32
	// storeGate, removeGate 属于机制层函数,必须醒目的声明
	storeGate      func(uint32, *GateServer) bool
	removeGate     func(uint32)
	outFunc        func(*connData.ConnData)
	startTimestamp int64
}

func NewGateServer(c *base_net.Socket,
	sendBufSize, recvBufSize uint32,
	storeGate func(uint32, *GateServer) bool,
	removeGate func(uint32),
	outFunc func(*connData.ConnData)) *GateServer {

	connId := int64(atomic.AddUint32(&gSerialNum, 1))
	conn := base_net.NewConnection(c, connId, sendBufSize, recvBufSize)

	conn.SetProperty("connId", connId)

	return &GateServer{Connection: conn, storeGate: storeGate, removeGate: removeGate,
		outFunc: outFunc}
}

func (s *GateServer) Run() {
	if 0 == atomic.LoadUint32(&s.isRun) {
		atomic.AddUint32(&s.isRun, 1)

		s.Connection.Run()
		go s.handleLoop()
	}
}

func (s *GateServer) handleLoop() {
	s.startTimestamp = time.Now().Unix()

	defer func() {
		function.Catch()
		log.Debugf("GateServer handleLoop 协程退出, 时间戳: %d", s.startTimestamp)
	}()

	log.Debugf("GateServer handleLoop 协程启动, 时间戳: %d", s.startTimestamp)

	for {
		select {
		case binData := <-s.GetChanRecvMsg():
			connData := &connData.ConnData{
				IConnection: s.Connection,
				BinData:     binData}
			s.handleMsg(connData)

		case <-s.GetExitSignal():
			log.Debugf("GateServer handleLoop 协程感知到底层关闭通知, 删除了底层 map 中的 gate 对象, 即将退出, 时间戳: %d", s.startTimestamp)
			s.removeGate(s.remoteGateServerId)
			return
		}
	}
}

// XXX 这里处理 gate 发来的消息,分发到各个 player
func (s *GateServer) handleMsg(connData *connData.ConnData) {
	defer function.Catch()
	dp := base_net.NewDataPack()
	msgId := dp.UnpackMsgId(connData.BinData)
	log.Debugf("底层 GateServer 模块, 收到 gate 发来的消息: %s", msgIdProto.MsgId(msgId).String())
	switch msgIdProto.MsgId(msgId) {
	// gate server 发来的注册请求
	case msgIdProto.MsgId_gate2LobbyRegister:
		pb := &lobbyProto.Gate2LobbyRegister{}
		if function.ProtoUnmarshal(connData.BinData[dp.GetHeadLen():],
			pb, "lobbyProto.Gate2LobbyRegister") {
			if s.storeGate(pb.GateServerId, s) {
				s.remoteGateServerId = pb.GateServerId
				s.SetProperty(gameKeyPrefix.GateId, pb.GateServerId)
				// 通知 gate 注册成功
				log.Debugf("收到 gate 发来的注册请求, 成功")
				s.SendPbMsg(0, msgIdProto.MsgId_lobby2GateRegisterSucc, nil)
			} else {
				// 通知 gate 注册失败
				log.Debugf("收到 gate 发来的注册请求, 失败, 应该有多个 gate 使用了相同的 server id")
				s.SendPbMsg(0, msgIdProto.MsgId_lobby2GateRegisterFail, nil)
				s.Close()
			}
		}

		// gate server 发来的心跳
	case msgIdProto.MsgId_fs2bsPing:
		log.Debugf("收到 gate 发来的心跳: %s", msgIdProto.MsgId(msgId).String())
		s.SendPbMsg(0, msgIdProto.MsgId_bs2fsPong, nil)

	default:
		s.outFunc(connData)
	}
}
