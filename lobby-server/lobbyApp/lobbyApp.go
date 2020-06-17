package lobbyApp

import (
	"servers/base-library/base_net"
	"servers/common-library/baseServer"
	"servers/common-library/config"
	"servers/common-library/connData"
	"servers/common-library/proto/msgIdProto"
	"servers/lobby-server/logic"
	"servers/lobby-server/login"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

const (
	playerExpires = 1800 // player 的缓存时间是半小时
)

type lobbyApp struct {
	*baseServer.BaseServer
	signinModule *login.Login // 玩家登录时,由登录模块 *login.Login 创建 player 对象
	logicModule  *logic.Logic // 逻辑模块 *logic.Logic 把消息分发给创建好的 player 对象
}

func CreateLobbyApp() *lobbyApp {
	dbStr := config.Cfg.MysqlAddr
	//dbStr := "dev:dev123@tcp(192.168.0.155)/games?charset=utf8mb4&parseTime=True&loc=Local"

	p := &lobbyApp{BaseServer: baseServer.New(dbStr, playerExpires)}

	p.signinModule = login.New(p.DB, p.StorePlayer, p.FindPlayer)
	p.signinModule.Run()

	p.logicModule = logic.New(p.DB, p.FindPlayerByConnId)

	return p
}

func (la *lobbyApp) Run(addr string) {
	go la.ListenTcpGate(addr, la.dispathMsgToPlayer)
}

func (la *lobbyApp) dispathMsgToPlayer(connData *connData.ConnData) {
	dp := base_net.DataPack{}
	msgId := dp.UnpackMsgId(connData.BinData)
	switch msgIdProto.MsgId(msgId) {
	case msgIdProto.MsgId_c2sWxLogin:
		la.signinModule.Login(connData)

	default:
		la.logicModule.DispatchMsgToPlayer(connData)
	}
}
