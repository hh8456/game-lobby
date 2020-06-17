package roomPkg

import (
	"servers/base-library/base_net"
	"servers/common-library/connData"
	"servers/common-library/function"
	"servers/common-library/log"
	"servers/common-library/niuniuAlgorithm"
	"servers/common-library/proto/commonProto"
	"servers/common-library/proto/msgIdProto"
	"servers/common-library/proto/niuniuProto"
	"servers/common-library/redisKeyPrefix"
	"servers/common-library/redisOpt"
	"strconv"
	"sync/atomic"

	"github.com/hh8456/go-common/redisObj"
)

var mapRoomLogicFunc map[msgIdProto.MsgId]func(*Room, IClient, *connData.ConnData)

func init() {
	mapRoomLogicFunc = map[msgIdProto.MsgId]func(*Room, IClient, *connData.ConnData){}
	// 进入牛牛房间
	mapRoomLogicFunc[msgIdProto.MsgId_c2sEnterNiuniuRoom] = c2sEnterNiuniuRoom
	// 离开牛牛房间
	mapRoomLogicFunc[msgIdProto.MsgId_c2sLeaveNiuniuRoom] = c2sLeaveNiuniuRoom
	// 入座
	// 玩家入座后,才会写 redis
	mapRoomLogicFunc[msgIdProto.MsgId_c2sNiuniuHaveASeat] = c2sNiuniuHaveASeat
	// 玩家点击"准备"按钮
	mapRoomLogicFunc[msgIdProto.MsgId_c2sNiuniuReady] = c2sNiuniuReady
	// 抢庄
	mapRoomLogicFunc[msgIdProto.MsgId_c2sNiuniuRobZhuang] = c2sRobZhuang
	// 下注
	mapRoomLogicFunc[msgIdProto.MsgId_c2sNiuniuBet] = c2sNiuniuBet
	// 亮牌
	mapRoomLogicFunc[msgIdProto.MsgId_c2sNiuniuShowCard] = c2sShowCard
	// 客户端断线
	mapRoomLogicFunc[msgIdProto.MsgId_gate2NiuniuClientDisconnect] = gate2NiuniuClientDisconnect
}

// 离开牛牛房间
func c2sLeaveNiuniuRoom(room *Room, c IClient, connData *connData.ConnData) {
	uid := c.GetUid()
	strUid := c.GetUidString()
	p := &niuniuProto.S2CBroadcastLeaveNiuniuRoom{}
	p.RoomId = room.roomId
	p.SeatIndex = int32(c.GetSeatIndex())
	room.removePlayer(c)
	// 必须调用完 room.RemovePlayer 以后,才能调用 room.getBystanderNum
	p.BystanderNum = room.getBystanderNum()
	room.sendToOthers(uid, msgIdProto.MsgId_s2cBroadcastLeaveNiuniuRoom, p)
	redisOpt.DelNiuniuPlayerRoomIdAndSeatIdx(strUid, room.roomId, uint32(p.SeatIndex))
	log.Debugf("niuniu server 上收到 msgIdProto.MsgId_c2sLeaveNiuniuRoom 消息, "+
		"connid: %d, uid: %d (这里是房间内逻辑), 从 redis 中把玩家从 房间 roomid: %d, "+
		"桌子座位 seatIdx: %d 上清除", c.GetConnId(), uid, room.roomId, p.SeatIndex)

	//自建房玩家和观战玩家全部离开后删除房间
	deleteSelfBuildRoom(room)

	s2cPbMsg := &niuniuProto.S2CLeaveNiuniuRoom{RoomId: room.roomId}
	c.SendPbMsg(msgIdProto.MsgId_s2cLeaveNiuniuRoom, s2cPbMsg)
}

//自建房玩家和观战玩家全部离开后删除房间
func deleteSelfBuildRoom(room *Room) {
	if room.roomType == commonProto.RoomType_roomTypePublic {
		return
	}
	if len(room.mapPlayers) != 0 || len(room.mapBystanders) != 0 {
		return
	}

	if room.roomId < 1000 {
		return
	}
	log.Debugf("自建房所有玩家离开，删除房间：", room.roomId)

	DeleteRoom(room.roomId)
	//redis删除自建房信息
	rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuRoomConfig)
	_, err := rds.Del(strconv.Itoa(int(room.roomId)))
	if err != nil {
		log.Error("redis 删除自建房配置失败，房间id:", room.roomId)
	}
	log.Debugf("删除自建房配置:", room.roomId)

	rdsCrRoom := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuRoom)
	_, err = rdsCrRoom.Del(strconv.Itoa(int(room.GetRoomId())))
	if err != nil {
		log.Error("redis 删除自建房信息失败，房间id:", room.roomId)
	}
	log.Debugf("删除自建房信息:", room.roomId)

}

// 客户端断线
func gate2NiuniuClientDisconnect(room *Room, c IClient, connData *connData.ConnData) {
	// 只要进入这里,就表明玩家在房间中,但不在游戏中
	uid := c.GetUid()
	log.Debugf("收到 gate 发来的断线消息, uid: %d 在房间中但没进行游戏(这里是房间内逻辑)",
		uid)
	c2sLeaveNiuniuRoom(room, c, connData)
}

// 入座,
func c2sNiuniuHaveASeat(room *Room, c IClient, connData *connData.ConnData) {
	uid := c.GetUid()
	strUid := c.GetUidString()
	if _, find := room.mapBystanders[uid]; find {
		if c.GetSeatIndex() > -1 {
			log.Warnf("玩家 %d 点击入座, 发现已经在座位上了, 有可能是多点了几次入座",
				uid)

			return
		}

		s2cPbMsg := &niuniuProto.S2CNiuniuHaveASeat{}
		seatIdx, bHaveASeat := room.haveASeat(c)

		// 入座成功
		if bHaveASeat {
			s2cPbMsg.SeatIdx = uint32(seatIdx)
			s2cPbMsg.IsSitDown = true
			s2cPbMsg.Player = c.GetPlayerBaseInfo()
			s2cPbMsg.BystanderNum = uint32(len(room.mapBystanders))

			redisOpt.SetNiuniuPlayerRoomIdAndSeatIdx(strUid, room.roomId, s2cPbMsg.SeatIdx)

			log.Debugf("uid %d 入座, roomId: %d, seatIdx: %d (这里是房间内逻辑), 在 "+
				"redis 中把玩家设置到桌子上", uid, room.roomId, s2cPbMsg.SeatIdx)

			room.broadcast(msgIdProto.MsgId_s2cNiuniuHaveASeat, s2cPbMsg)
			return
		}

		// 座位已满
		c.SendPbMsg(msgIdProto.MsgId_s2cNiuniuHaveASeat, s2cPbMsg)
	} else {
		log.Warnf("玩家 %d 点击入座, 发现不在旁观者队列, 有可能是已经开始进行游戏了",
			uid)
	}

}

// 玩家点击"准备"按钮
func c2sNiuniuReady(room *Room, c IClient, connData *connData.ConnData) {
	uid := c.GetUid()
	// 只有在准备阶段和倒计时阶段才能点击"准备"
	if room.roomStatus == niuniuProto.NiuniuRoomStatus_niuniuRoomStatusReady ||
		room.roomStatus == niuniuProto.NiuniuRoomStatus_niuniuRoomStatusCountDown {
		if room.playerIsInSeat(c) && c.IsPrepare() == false {
			log.Debugf("uid: %d 点击准备, 服务器给了回复", uid)
			// 防止玩家重复点准备
			c.Prepare()

			s2cPbMsg := &niuniuProto.S2CNiuniuReady{}
			s2cPbMsg.Uid = uid
			s2cPbMsg.BystanderNum = uint32(len(room.mapBystanders))
			room.broadcast(msgIdProto.MsgId_s2cNiuniuReady, s2cPbMsg)
		} else {
			if false == room.playerIsInSeat(c) {
				log.Debugf("uid: %d 点击准备, 但由于不在座位上, 所以服务器忽略请求", uid)
			}

			if c.IsPrepare() {
				log.Debugf("uid: %d 点击准备, 但由于已经点击过准备, 所以服务器忽略请求", uid)
			}
		}
	} else {
		log.Debugf("uid: %d 点击准备, 但由于不是 [ 准备, 倒计时 ] 阶段, 所以服务器忽略请求", uid)
	}
}

// 抢庄
func c2sRobZhuang(room *Room, c IClient, connData *connData.ConnData) {
	dp := &base_net.DataPack{}
	c2sPbMsg := &niuniuProto.C2SNiuniuRobZhuang{}
	if function.ProtoUnmarshal(connData.BinData[dp.GetHeadLen():],
		c2sPbMsg, "niuniuProto.C2SRobZhuang") {
		uid := c.GetUid()
		if _, find := room.mapRobZhuang[uid]; !find {
			if _, find := room.cfg.MapRobZhuangRate[c2sPbMsg.Rate]; find {
				room.mapRobZhuang[uid] = c2sPbMsg.Rate
				//自建房暗抢庄家
				if room.roomType != commonProto.RoomType_roomTypePublic && room.cfg.KnownRobZhuang {
					return
				}
				replyPbMsg := &niuniuProto.S2CNiuniuRobZhuang{}
				replyPbMsg.Rate = c2sPbMsg.Rate
				replyPbMsg.Uid = c.GetUid()
				room.broadcast(msgIdProto.MsgId_s2cNiuniuRobZhuang, replyPbMsg)
				log.Debugf("收到 uid: %d 抢庄, 倍率: %d", uid, c2sPbMsg.Rate)
			}
		}
	}
}

// 闲家下注
func c2sNiuniuBet(room *Room, c IClient, connData *connData.ConnData) {
	c2sPbMsg := &niuniuProto.C2SNiuniuBet{}
	// 旁观者不能下注
	if c.GetSeatIndex() > -1 {
		dp := &base_net.DataPack{}
		if function.ProtoUnmarshal(connData.BinData[dp.GetHeadLen():],
			c2sPbMsg, "niuniuProto.C2SNiuniuBet") {
			// 只有闲家才能下注
			uid := c.GetUid()
			if _, find := room.mapBet[uid]; !find {
				if uid != room.bankerId {
					if _, ok := room.cfg.MapBetRate[c2sPbMsg.Bet]; ok {
						room.mapBet[uid] = c2sPbMsg.Bet
						replyPbMsg := &niuniuProto.S2CNiuniuBet{}
						replyPbMsg.Bet = c2sPbMsg.Bet
						replyPbMsg.Uid = c.GetUid()
						room.broadcast(msgIdProto.MsgId_s2cNiuniuBet, replyPbMsg)
					}
				}
			}
		}
	}
}

// 亮牌
func c2sShowCard(room *Room, c IClient, connData *connData.ConnData) {
	uid := c.GetUid()
	replyPbMsg := &niuniuProto.S2CNiuniuShowCard{}
	replyPbMsg.Uid = uid
	replyPbMsg.Cards = c.GetHandCard()

	cardType, cardComb := niuniuAlgorithm.CalCardType(c.GetHandCard())
	//是否坎斗
	if room.cfg.Kan {
		kanCardType, kanCardComb, _ := niuniuAlgorithm.Kan(c.GetHandCard())
		if kanCardType > cardType {
			cardType, cardComb = kanCardType, kanCardComb
		}
	}
	//是否顺斗
	if room.cfg.Shun {
		shunCardType, shunCardComb, _ := niuniuAlgorithm.Shun(c.GetHandCard())
		if shunCardType > cardType {
			cardType, cardComb = shunCardType, shunCardComb
		}
	}

	replyPbMsg.CardType, replyPbMsg.CardComb = uint32(cardType), cardComb
	// 只转发闲家亮牌
	if room.isNormalPlayer(uid) {
		if uid != room.bankerId {
			room.mapShowCard[uid] = replyPbMsg
			room.broadcast(msgIdProto.MsgId_s2cNiuniuShowCard, replyPbMsg)
			log.Debugf("广播闲家亮牌, uid: %d, roomId: %d, gameTimes: %d",
				uid, room.roomId, atomic.LoadUint32(&room.gameTimes))
		}
	}

	// 庄家亮牌不转发
	if room.isBanker(uid) {
		room.mapShowCard[uid] = replyPbMsg
		log.Debugf("庄家亮牌, 不广播, uid: %d, roomId: %d, gameTimes: %d",
			uid, room.roomId, atomic.LoadUint32(&room.gameTimes))
	}
}
