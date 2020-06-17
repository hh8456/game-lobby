package roomPkg

import (
	"servers/common-library/function"
	"servers/common-library/log"
	"servers/common-library/proto/commonProto"
	"servers/common-library/proto/msgIdProto"
	"servers/common-library/proto/niuniuProto"
	"servers/common-library/redisKeyPrefix"
	"servers/common-library/redisOpt"
	"strconv"
	"time"

	"github.com/hh8456/go-common/redisObj"
)

func (room *Room) Run() {
	rdsRoom := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuRoom)
	key := strconv.Itoa(int(room.roomId))
	go func() {
		for {
			//删除自建房
			if room.roomStatus == niuniuProto.NiuniuRoomStatus_niuniuRoomStatusGameOver {
				DeleteRoom(room.roomId)
				//redis删除自建房信息
				rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuRoomConfig)
				_, err := rds.Del(strconv.Itoa(int(room.roomId)))
				if err != nil {
					log.Error("redis 删除自建房配置失败，房间id:", room.roomId)
				}

				rdsCrRoom := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuRoom)
				_, err = rdsCrRoom.Del(strconv.Itoa(int(room.GetRoomId())))
				if err != nil {
					log.Error("redis 删除自建房信息失败，房间id:", room.roomId)
				}

				log.Debugf("自建房游戏结束：", room.roomId)
				return
			}

			select {
			case p := <-room.chanClientAndMsg:
				room.handle(p.Client, p.connData)

				// 房间定时器是每秒执行一次
			case timestamp := <-room.chanTimer:
				room.running(timestamp)
				// 每秒钟刷新房间玩家数量, 用来支持后台查询
				room.setRoomPlayerNum(room.roomId, uint32(len(room.mapPlayers)))
				roomInfo := &niuniuProto.RoomInfo{RoomId: room.roomId,
					PlayerNum: uint32(len(room.mapPlayers)),
					RoomType:  commonProto.RoomType_roomTypePublic}

				binBuf, b := function.ProtoMarshal(roomInfo, "niuniuProto.RoomInfo")
				if b {
					rdsRoom.Setex(key, 10*time.Second, binBuf)
				}
			}
		}
	}()
}

func (room *Room) running(timestamp int64) {
	switch room.roomStatus {
	case niuniuProto.NiuniuRoomStatus_niuniuRoomStatusReady:
		// 满 2 人准备 && 准备阶段满 x 秒,  就开始倒计时
		if room.getPrepperNum() > 1 &&
			timestamp-room.roomStatusTimestamp >= room.cfg.MapWaitTime[uint32(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusReady)] {
			log.Debugf("满 2 人准备, 马上切换到倒计时状态, roomId: %d, gameTimes: %d", room.roomId, room.gameTimes)
			// 通知客户端切换到倒计时阶段,
			room.roomStatus = niuniuProto.NiuniuRoomStatus_niuniuRoomStatusCountDown
			room.roomStatusTimestamp = timestamp
			room.broadcast(msgIdProto.MsgId_s2cNiuniuBroadcastRoomStatus,
				&niuniuProto.S2CNiuniuBroadcastRoomStatus{RoomStatus: room.roomStatus,
					CountDown: uint32(room.cfg.MapWaitTime[uint32(room.roomStatus)])})

			//自建房有次数限制时，进行局数+1
			if room.cfg.TotalNumberOfGame != 0 {
				room.numberOfGame += 1
				log.Debugf("房间：", room.roomId, "当前局数", room.numberOfGame)
				room.broadcast(msgIdProto.MsgId_s2cSelfBuildRoomNumberOfGame,
					&niuniuProto.S2CNiuNiuGameNumber{RoomId: room.roomId, NumberOfGame: room.numberOfGame})
			}
		}

	case niuniuProto.NiuniuRoomStatus_niuniuRoomStatusCountDown:
		if timestamp-room.roomStatusTimestamp >= room.cfg.MapWaitTime[uint32(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusCountDown)] { // 倒计时阶段结束
			replyMsg := &niuniuProto.S2CNiuniuLeaveTheSeat{}
			replyMsg.MapUidSeat = map[uint32]uint32{}
			// 确定游戏者
			for _, client := range room.seat {
				if client != nil {
					uid := client.GetUid()
					if client.IsPrepare() {
						room.mapPlayers[uid] = client
					}
				}
			}

			// 人数不足,退回准备阶段
			if len(room.mapPlayers) == 1 && len(room.mapPlayers) == 0 {
				log.Debugf("倒计时阶段结束时发现人数不足 2 人,只有 %d 人, 切回准备阶段, roomId: %d",
					len(room.mapPlayers), room.roomId)

				// 重新准备开始游戏
				room.reReady(timestamp)
				// 清除离线玩家
				room.clearOfflineClient()
				return
			}

			// 没有点准备的人,就变成旁观者
			for seatIdx, client := range room.seat {
				if client != nil {
					uid := client.GetUid()
					strUid := client.GetUidString()
					if _, find := room.mapPlayers[uid]; !find {
						// 没有点准备的人,就变成旁观者
						room.mapBystanders[uid] = client
						room.seat[seatIdx] = nil
						client.SetSeatIndex(-1)
						replyMsg.MapUidSeat[uid] = uint32(seatIdx)
						redisOpt.DelNiuniuPlayerRoomIdAndSeatIdx(strUid, room.roomId, uint32(seatIdx))
					}
				}
			}

			// 离开座位的人, uid - 座位号, [0, 10]
			if len(replyMsg.MapUidSeat) > 0 {
				replyMsg.BystanderNum = uint32(len(room.mapBystanders))
				room.broadcast(msgIdProto.MsgId_s2cNiuniuLeaveTheSeat, replyMsg)
			}

			gameStartMsg := &niuniuProto.S2CNiuniuGameStart{}
			for uid, _ := range room.mapPlayers {
				gameStartMsg.Uids = append(gameStartMsg.Uids, uid)
			}
			room.broadcast(msgIdProto.MsgId_s2cNiuniuGameStart, gameStartMsg)

			// 明牌抢庄
			if room.cfg.KnownCard {
				// 发 4 张明牌
				room.sendKnownCard(4)
			}

			log.Debugf("马上切换到抢庄状态, roomId: %d, gameTimes: %d", room.roomId, room.gameTimes)
			room.gameTimes++

			// 通知客户端切换到抢庄阶段
			room.changeRoomStatus(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusRobZhuang, timestamp)
		}

		// 抢庄
	case niuniuProto.NiuniuRoomStatus_niuniuRoomStatusRobZhuang:
		b := false
		if timestamp-room.roomStatusTimestamp >= room.cfg.MapWaitTime[uint32(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusRobZhuang)] { // 抢庄结束

			// 抢庄结束
			b = true
		} else {
			// 所有人都抢庄完毕
			if len(room.mapRobZhuang) == len(room.mapPlayers) {
				b = true
			}
		}

		if b {
			bankerId, rate, uids := room.confirmBanker()
			room.bankerId = bankerId
			room.mapRobZhuang[bankerId] = rate
			p := &niuniuProto.S2CConfirmBanker{Uid: bankerId, Rate: rate, Uids: uids}
			room.broadcast(msgIdProto.MsgId_s2cConfirmBanker, p)

			// 通知客户端切换到下注阶段
			room.changeRoomStatus(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusBet, timestamp)
		}

		// 下注
	case niuniuProto.NiuniuRoomStatus_niuniuRoomStatusBet:
		b := false
		if timestamp-room.roomStatusTimestamp >= room.cfg.MapWaitTime[uint32(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusBet)] { // 下注结束
			b = true
		} else {
			// 所有闲家都下注完毕
			if len(room.mapBet)+1 == len(room.mapPlayers) {
				b = true
			}
		}

		if b {
			// 检测下注情况,代替没下注的闲家下注
			room.checkBet()

			// 明牌抢庄
			if room.cfg.KnownCard {
				// 发暗牌
				room.sendOneUnknownCard()
			} else {
				// 直接发 5 张明牌
				// 发 4 张明牌
				room.sendKnownCard(5)
			}

			// 通知客户端切换到游戏阶段
			room.changeRoomStatus(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusPlay, timestamp)

		}

		// 游戏中, 亮牌
	case niuniuProto.NiuniuRoomStatus_niuniuRoomStatusPlay:
		room.showCard(timestamp)

		// 结算
	case niuniuProto.NiuniuRoomStatus_niuniuRoomStatusSettle:
		room.settle(timestamp)

		//游戏结束
	case niuniuProto.NiuniuRoomStatus_niuniuRoomStatusGameOver:
		return
	}
}
