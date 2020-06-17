package roomPkg

import (
	"servers/common-library/connData"
	"servers/common-library/log"
	"servers/common-library/proto/commonProto"
	"servers/common-library/proto/msgIdProto"
	"servers/common-library/proto/niuniuProto"
	"servers/common-library/redisOpt"
	"time"
)

// 进入牛牛房间
func c2sEnterNiuniuRoom(room *Room, c IClient, connData *connData.ConnData) {
	log.Debugf("uid %d 进入牛牛房间 roomId: %d, 这里是房间内逻辑", c.GetUid(), room.roomId)
	s2cPbMsg := &niuniuProto.S2CEnterNiuniuRoom{}
	s2cPbMsg.RoomId = room.roomId
	if false == room.playerInRoom(c) {
		room.addBystanders(c)
	}

	s2cPbMsg.RcSnap = room.makeReconnectSnap(c)

	m := room.getPlayerInSeat()
	// 要把座位上的玩家信息发给后来进入的人
	s2cPbMsg.MapPlayerBaseInfo = map[uint32]*commonProto.PlayerBaseInfo{}
	for seatIdx, client := range m {
		s2cPbMsg.MapPlayerBaseInfo[seatIdx] = client.GetPlayerBaseInfo()
	}

	s2cPbMsg.PlayerBaseInfo = c.GetPlayerBaseInfo()
	s2cPbMsg.RoomStatus = room.roomStatus
	s2cPbMsg.BystanderNum = uint32(len(room.mapBystanders))
	nowTs := time.Now().Unix()
	//log.Debugf("当前时刻: %d - 上次房间时刻: %d = %d", nowTs, room.roomStatusTimestamp, nowTs-room.roomStatusTimestamp)
	s2cPbMsg.CountDown = uint32(room.cfg.MapWaitTime[uint32(room.roomStatus)] - (nowTs - room.roomStatusTimestamp))
	//房主
	s2cPbMsg.Owner = room.owner
	s2cPbMsg.NumberOfGame = room.numberOfGame
	//加入房间配置信息
	s2cPbMsg.Config = room.cfg

	c.SendPbMsg(msgIdProto.MsgId_s2cEnterNiuniuRoom, s2cPbMsg)

	p := &niuniuProto.S2COtherPlayerEnterNiuniuRoom{}
	p.RoomId = room.roomId
	p.BystanderNum = uint32(len(room.mapBystanders))
	room.sendToOthers(c.GetUid(),
		msgIdProto.MsgId_s2cOtherPlayerEnterNiuniuRoom, p)

	// 执行到这里表示断线重连,已经在座位上
	if c.GetSeatIndex() > -1 {
		log.Debugf("执行到这里表示玩家在座位上时断线重连, "+
			"重新回到房间座位上时, 设置信息到 redis 中; "+
			"uid: %d, roomId: %d, seatIdx: %d",
			c.GetUid(), room.roomId, c.GetSeatIndex())

		redisOpt.SetNiuniuPlayerRoomIdAndSeatIdx(c.GetUidString(), room.roomId, uint32(c.GetSeatIndex()))
	}
}
