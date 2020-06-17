package backEndServer

import (
	"servers/base-library/base_net"
	"servers/common-library/function"
	"servers/common-library/log"
	"servers/common-library/proto/lobbyProto"
	"servers/common-library/proto/msgIdProto"
	"servers/common-library/proto/niuniuProto"
	"sync/atomic"
)

var (
	gSerialNum uint32
)

type BackEndServer struct {
	*base_net.Connection
	localGateServerId   uint32
	remoteServerId      uint32 // 远程 BackEndServer 服务器 id
	chanPingReq         chan struct{}
	isRun               uint32
	disconnect          func(uint32)
	sendToClient        func([]byte)
	setClientGameSrvId  func(int64, uint32)
	sendToBackEndServer func(uint32, []byte)
	kickClient          func(int64)
}

func NewBackEndServer(c *base_net.Socket,
	localGateServerId, remoteServerId, sendBufSize, recvBufSize uint32,
	disconnect func(uint32), sendToClient func([]byte),
	setClientGameSrvId func(int64, uint32),
	sendToBackEndServer func(uint32, []byte),
	kickClient func(int64)) *BackEndServer {

	connId := int64(atomic.AddUint32(&gSerialNum, 1))
	conn := base_net.NewConnection(c, connId, sendBufSize, recvBufSize)
	conn.SetProperty("connId", connId)

	return &BackEndServer{
		Connection:          conn,
		localGateServerId:   localGateServerId,
		remoteServerId:      remoteServerId,
		chanPingReq:         make(chan struct{}, 5),
		disconnect:          disconnect,
		sendToClient:        sendToClient,
		setClientGameSrvId:  setClientGameSrvId,
		sendToBackEndServer: sendToBackEndServer,
		kickClient:          kickClient,
	}
}

func (s *BackEndServer) Run() {
	if 0 == atomic.LoadUint32(&s.isRun) {
		atomic.AddUint32(&s.isRun, 1)

		s.Connection.Run()
		s.register()

		// 先给 backEndServer 发送注册消息,然后接收回复
		go s.forwardLoop()
	}
}

// 把 lobby 或者游戏服务器发来的数据转发给客户端
func (s *BackEndServer) forwardLoop() {
	defer func() {
		function.Catch()
		log.Debugf("BackEndServer forwardLoop 协程退出")
	}()

	log.Debugf("BackEndServer forwardLoop 协程启动")
	for {
		select {
		case binData := <-s.GetChanRecvMsg():
			s.handleBinData(binData)

		case <-s.GetExitSignal():
			s.disconnect(s.remoteServerId)
			return
		}
	}
}

func (s *BackEndServer) handleBinData(binData []byte) {
	dp := &base_net.DataPack{}
	msgId := dp.UnpackMsgId(binData)
	connId := dp.UnpackClientConnId(binData)
	switch msgIdProto.MsgId(msgId) {
	// XXX 注册成功, 待联调
	case msgIdProto.MsgId_lobby2GateRegisterSucc:
		log.Debugf("gate 到 BackEndServer id: %d 注册成功", s.remoteServerId)
		go s.handlePing()

		// 注册失败,
	case msgIdProto.MsgId_lobby2GateRegisterFail:
		log.Errorf("gate 到 BackEndServer 注册失败, 应该有多个 gate 用了相同的 server id: %d", s.localGateServerId)
		s.Close()
		return

		// 心跳保活,不处理
	case msgIdProto.MsgId_bs2fsPong:
		log.Debugf("gate 收到 BackEndServer id: %d 回复的心跳保活,不处理", s.remoteServerId)

		// lobby 发来的消息, 设置 gate 绑定 gameId
	case msgIdProto.MsgId_lobby2GateEnterNiuniuRoom:
		log.Debugf("gate 收到 lobby 发来进入牛牛房间的消息:msgIdProto.MsgId_lobby2GateEnterNiuniuRoom")
		pb := &niuniuProto.Lobby2GateEnterNiuniuRoom{}
		if function.ProtoUnmarshal(binData[dp.GetHeadLen():], pb,
			"niuniuProto.Lobby2GateEnterNiuniuRoom") {
			log.Debugf("gate  收到 lobby 发来进入牛牛房间的消息( 长度 %d ) "+
				"msgIdProto.MsgId_lobby2GateEnterNiuniuRoom, 为玩家绑定 niuniuServer "+
				"id: %d", len(pb.C2SEnterNiuniuRoom), pb.NiuniuServerId)
			s.setClientGameSrvId(connId, pb.NiuniuServerId)
			// 转发给牛牛服务器
			s.sendToBackEndServer(pb.NiuniuServerId, pb.C2SEnterNiuniuRoom)
		}

	case msgIdProto.MsgId_lobby2GateCreateNiuNiuRoom:
		//log.Debugf("gate 收到 lobby 发来进入牛牛房间的消息:msgIdProto.MsgId_lobby2GateEnterNiuniuRoom")
		pb := &niuniuProto.Lobby2GateSelfBuildNiuNiuRoom{}
		if function.ProtoUnmarshal(binData[dp.GetHeadLen():], pb,
			"niuniuProto.Lobby2GateEnterNiuniuRoom") {
			s.setClientGameSrvId(connId, pb.NiuniuServerId)
			// 转发给牛牛服务器
			s.sendToBackEndServer(pb.NiuniuServerId, pb.C2SSelfBuildNiuNiuRoom)
		}

		// 牛牛服务器发来的离开房间
	case msgIdProto.MsgId_s2cLeaveNiuniuRoom:
		// 解除和牛牛服务器的绑定
		s.setClientGameSrvId(connId, 0)
		// 转发给客户端
		s.sendToClient(binData)

		// 服务器掐断连接
	case msgIdProto.MsgId_s2cKick:
		s.kickClient(connId)

		// 调试
	//case msgIdProto.MsgId_s2cWxLogin:
	//s2cWxLogin := &lobbyProto.S2CWxLogin{}
	//if function.ProtoUnmarshal(binData[dp.GetHeadLen():], s2cWxLogin,
	//"lobbyProto.S2CWxLogin") {
	//log.Debugf("gate 收到 lobby 创建账号成功 msgIdProto.MsgId_s2cWxLogin 的通知, "+
	//"connid: %d, wxid: %s, uid: %d", connId, s2cWxLogin.PlayerBaseInfo.Wxid,
	//s2cWxLogin.PlayerBaseInfo.Uid)
	//s.sendToClient(binData)
	//}

	// 转发给客户端
	default:
		s.sendToClient(binData)
	}
}

func (s *BackEndServer) register() {
	pb := &lobbyProto.Gate2LobbyRegister{GateServerId: s.localGateServerId}
	pbBuf, b := function.ProtoMarshal(pb, "lobbyProto.Gate2LobbyRegister")
	if b {
		msg := base_net.NewMsgPackage(0,
			uint32(msgIdProto.MsgId_gate2LobbyRegister), pbBuf)

		dp := base_net.DataPack{}
		buf, err := dp.Pack(msg)
		if err != nil {
			log.Errorf("BackEndServer.register base_net.DataPack error %v", err)
		}

		s.Send(buf)
	}
}

func (s *BackEndServer) handlePing() {
	defer func() {
		function.Catch()
		s.clearChanPing()
		log.Debug("BackEndServer handlePing 协程退出")
	}()

	msg := base_net.NewMsgPackage(0, uint32(msgIdProto.MsgId_fs2bsPing), nil)
	dp := base_net.DataPack{}
	buf, err := dp.Pack(msg)
	if err != nil {
		log.Errorf("BackEndServer.handlePing base_net.DataPack.Pack error %v", err)
		return
	}

	for {
		select {
		case <-s.chanPingReq:
			s.Send(buf)

		case <-s.GetExitSignal():
			return
		}
	}

}

func (s *BackEndServer) clearChanPing() {
	for {
		select {
		case _ = <-s.chanPingReq:

		default:
			return
		}
	}
}
func (s *BackEndServer) Ping() {
	// 每隔 20 秒 ping 一次
	select {
	case s.chanPingReq <- struct{}{}:

	default:
	}
}
