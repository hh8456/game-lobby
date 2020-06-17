package logic

import (
	"servers/base-library/base_net"
	"servers/common-library/connData"
	"servers/common-library/function"
	"servers/common-library/log"
	"servers/common-library/proto/commonProto"
	"servers/common-library/proto/lobbyProto"
	"servers/common-library/proto/msgIdProto"
	"servers/common-library/redisOpt"
	"servers/iface"
	"servers/lobby-server/player"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
)

type Logic struct {
	chanConnData       chan *connData.ConnData
	db                 *gorm.DB
	findPlayerByConnId func(connId int64) iface.IPlayer
}

func New(db *gorm.DB,
	findPlayerByConnId func(connId int64) iface.IPlayer) *Logic {
	return &Logic{
		chanConnData:       make(chan *connData.ConnData, 20000), // 两万并发登录
		db:                 db,
		findPlayerByConnId: findPlayerByConnId,
	}
}

func (l *Logic) DispatchMsgToPlayer(connData *connData.ConnData) {
	dp := base_net.NewDataPack()
	connId := dp.UnpackClientConnId(connData.BinData)
	iPlayer := l.findPlayerByConnId(connId)
	if iPlayer != nil {
		iPlayer.(*player.Player).Handle(connData)
	} else {
		log.Debugf("没有通过 connid: %d 找到客户端", connId)
	}
}

func c2sPing(player *player.Player, connData *connData.ConnData) {
	dp := base_net.NewDataPack()
	connId := dp.UnpackClientConnId(connData.BinData)
	pb := &commonProto.C2SPing{}
	if function.ProtoUnmarshal(connData.BinData[dp.GetHeadLen():],
		pb, "commonProto.C2SPing") {
		player.SetLastAliveTimestamp(time.Now().Unix())
		replyMsg := &commonProto.S2CPing{Timestamp: pb.Timestamp}
		connData.SendPbMsg(connId, msgIdProto.MsgId_s2cPing, replyMsg)
	}
}

func c2sGetPlayerGold(player *player.Player, connData *connData.ConnData) {
	strGold, _ := redisOpt.GetOnePlayerGold(player.GetUid())
	gold, _ := strconv.Atoi(strGold)
	replyMsg := &lobbyProto.S2CGetPlayerGold{Gold: int32(gold)}
	connData.SendPbMsg(player.GetConnId(), msgIdProto.MsgId_s2cGetPlayerGold, replyMsg)
}
