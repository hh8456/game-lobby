package roomPkg

import (
	"servers/common-library/log"
	"servers/common-library/niuniuAlgorithm"
	"servers/common-library/proto/msgIdProto"
	"servers/common-library/proto/niuniuProto"
	"sync/atomic"
)

func (room *Room) showCard(timestamp int64) {
	if timestamp-room.roomStatusTimestamp >= room.cfg.MapWaitTime[uint32(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusPlay)] { // 游戏结束,即亮牌阶段结束
		// 把剩余没亮牌的闲家进行亮牌
		if false == room.isAllPlayerShowCard() {
			for uid, c := range room.mapPlayers {
				if _, find := room.mapShowCard[uid]; !find {
					if uid != room.bankerId {
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
						room.mapShowCard[uid] = replyPbMsg
						room.broadcast(msgIdProto.MsgId_s2cNiuniuShowCard, replyPbMsg)
						log.Debugf("广播闲家亮牌, uid: %d, roomId: %d, gameTimes: %d",
							uid, room.roomId, room.gameTimes)
					}
				}
			}
		}

		if room.status.broadcastBankerCards == false {
			room.bankerShowCard()
		}

		// 通知客户端切换到结算阶段
		// 阶段阶段开始时,就要庄家亮牌
		room.changeRoomStatus(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusSettle, timestamp)
	} else {
		// 所有人都亮牌了, 就让庄家亮牌, 然后提前进入结算阶段
		if room.isAllPlayerShowCard() {
			if room.status.broadcastBankerCards == false {
				room.bankerShowCard()
			} else {
				room.changeRoomStatus(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusSettle, timestamp)
			}

		}
	}
}

func (room *Room) bankerShowCard() {
	log.Debugf("广播庄家亮牌, uid: %d, roomId: %d, gameTimes: %d, 即将广播",
		room.bankerId, room.roomId, atomic.LoadUint32(&room.gameTimes))

	room.status.broadcastBankerCards = true

	if banker, ok := room.mapPlayers[room.bankerId]; ok {
		replyPbMsg := &niuniuProto.S2CNiuniuShowCard{}
		replyPbMsg.Uid = room.bankerId
		replyPbMsg.Cards = banker.GetHandCard()
		cardType, cardComb := niuniuAlgorithm.CalCardType(banker.GetHandCard())
		//是否坎斗
		if room.cfg.Kan {
			kanCardType, kanCardComb, _ := niuniuAlgorithm.Kan(banker.GetHandCard())
			if kanCardType > cardType {
				cardType, cardComb = kanCardType, kanCardComb
			}
		}
		//是否顺斗
		if room.cfg.Shun {
			shunCardType, shunCardComb, _ := niuniuAlgorithm.Shun(banker.GetHandCard())
			if shunCardType > cardType {
				cardType, cardComb = shunCardType, shunCardComb
			}
		}

		replyPbMsg.CardType, replyPbMsg.CardComb = uint32(cardType), cardComb
		room.mapShowCard[room.bankerId] = replyPbMsg
		room.broadcast(msgIdProto.MsgId_s2cNiuniuShowCard, replyPbMsg)
		log.Debugf("广播庄家亮牌, uid: %d, roomId: %d, gameTimes: %d",
			room.bankerId, room.roomId, atomic.LoadUint32(&room.gameTimes))
	}
}
