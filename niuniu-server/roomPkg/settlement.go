package roomPkg

import (
	"servers/common-library/function"
	"servers/common-library/log"
	"servers/common-library/niuniuAlgorithm"
	"servers/common-library/proto/commonProto"
	"servers/common-library/proto/errorCodeProto"
	"servers/common-library/proto/msgIdProto"
	"servers/common-library/proto/niuniuProto"
	"servers/common-library/redisKeyPrefix"
	"servers/common-library/redisOpt"
	"strconv"

	"github.com/hh8456/go-common/redisObj"
	"github.com/jinzhu/gorm"
)

func (room *Room) settle(timestamp int64) {

	// 1.  庄闲家赢的钱,不能超过身上的钱
	// 2.1 输的钱,不能超过身上的钱
	// 2.2 庄家输的钱超过了身上的钱,就把身上的钱按照比例,分给各个闲家
	// 定输赢
	if room.status.settled == false {
		room.status.settled = true
		banker := room.mapPlayers[room.bankerId]
		if banker == nil {
			return
		}

		room.calcWinOrLose()

		// 写 redis
		mapWinOrLoseFinal, mapPlayerGold, b := room.updatePlayerGold(room.mapWinOrLose)
		if b {
			replyPbMsg := &niuniuProto.S2CNiuniuSettle{}
			replyPbMsg.MapWinOrLose = room.mapWinOrLose
			replyPbMsg.MapPlayerGold = mapPlayerGold
			replyPbMsg.MapWinOrLoseFinal = mapWinOrLoseFinal
			room.broadcast(msgIdProto.MsgId_s2cNiuniuSettle, replyPbMsg)
		} else {
			// 提示系统出错, 输赢无效
			room.broadcastErrorCode(errorCodeProto.ErrorCode_write_mysql_error_when_niuniu_settlement)
		}
	} else {
		if timestamp-room.roomStatusTimestamp >= room.cfg.MapWaitTime[uint32(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusSettle)] { // 结算阶段结束
			//用于自建房，局数是否完毕
			if room.roomType != commonProto.RoomType_roomTypePublic && room.haveNext() == false {
				//广播结算信息
				replyMsg := &niuniuProto.S2CNiuNiuRoomWinLose{}
				replyMsg.MapWinLose = room.mapTotalWinLose
				log.Debugf("房间输赢结果:", replyMsg.MapWinLose)
				room.broadcast(msgIdProto.MsgId_s2cSelfBuildRoomWinLose, replyMsg)
				room.mapPlayers= map[uint32]IClient{}
				room.roomStatus = niuniuProto.NiuniuRoomStatus_niuniuRoomStatusGameOver
				return
			}

			// TODO 从 redis 读取配置
			log.Debugf("游戏结束, 从 redis 中重新获取配置")
			rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuRoomConfig)
			binCfg, e := rds.Get(strconv.Itoa(int(room.roomId)))
			if e == nil {
				cfg := &niuniuProto.RoomConfig{}
				if function.ProtoUnmarshal([]byte(binCfg), cfg, "niuniuProto.RoomConfig") {
					room.cfg = cfg
				}
			}
			// 重新准备开始游戏
			log.Debugf("游戏结束, 重新准备开始游戏, room.reReady")
			room.reReady(timestamp)
			// 清除离线玩家
			room.clearOfflineClient()
			log.Debugf("游戏结束, 清除完毕离线玩家, room.clearOfflineClient")
		}
	}
}

// mapWinOrLose 键值: uid - 输赢金币
// TODO 需要做容错性处理, 具体咨询策划
// 返回值: 扣除房费的金币, 玩家身上剩下的金币, 是否成功
func (room *Room) updatePlayerGold(mapWinOrLose map[uint32]int32) (map[uint32]int32, map[uint32]int32, bool) {
	// 更新若干个玩家的金币,需要事务性,要么全部成功,要么全部失败
	uids := []uint32{}
	for uid, _ := range mapWinOrLose {
		uids = append(uids, uid)
	}

	// 从 redis 中查询出玩家剩余金币
	mapPlayerGold, err1 := redisOpt.GetSomePlayersGold(uids)
	if err1 != nil {
		log.Errorf("从 redis 中查询用户金币出错, 返回的 mapPlayerGold 为空, "+
			"error: %v, uids: %v", err1, uids)
	}

	// 扣除房费的金币
	mapWinOrLoseFinal := map[uint32]int32{}
	for uid, winOrLose := range mapWinOrLose {
		if winOrLose > 0 {
			// 扣除 5% 房费
			dec := int32(float64(winOrLose) * 0.95)
			mapWinOrLoseFinal[uid] = dec
		} else {
			mapWinOrLoseFinal[uid] = winOrLose
		}
	}

	for uid, winOrLose := range mapWinOrLoseFinal {
		if coin, find := mapPlayerGold[uid]; find {
			coin += winOrLose
			mapPlayerGold[uid] = coin
		}
	}

	log.Debugf("马上开始写数据库事务, update player_base_info set gold = ? where uid = ? ")

	b := false
	room.db.Transaction(func(tx *gorm.DB) error {
		for uid, coin := range mapPlayerGold {
			err := tx.Exec("update player_base_info set gold = ? where uid = ?", coin, uid).Error
			if err != nil {
				log.Errorf("结算时写数据库事务出错, update player_base_info set gold = ? where uid = ? ==> error: %v", err)
				return err
			}
		}

		b = true
		return nil
	})

	if b {
		for uid, gold := range mapPlayerGold {
			redisOpt.SetOnePlayerGold(uid, gold)
		}

		log.Debugf("结算时写数据库事务成功, update player_base_info set gold = ? where uid = ? ")
		return mapWinOrLoseFinal, mapPlayerGold, true
	}

	log.Errorf("结算时写数据库事务出错, update player_base_info set gold = ? where uid = ? ==> 无效的写入")
	return nil, nil, false
}

// 计算输赢
// 1.  庄闲家赢的钱,不能超过身上的钱
// 2.1 输的钱,不能超过身上的钱
// 2.2 庄家输的钱超过了身上的钱,就把身上的钱按照比例,分给各个闲家
func (room *Room) calcWinOrLose() {
	banker := room.mapPlayers[room.bankerId]
	if banker == nil {
		return
	}

	// 输赢金额 = 底分 * 庄家抢庄倍数 * 下注倍数 * 牌型倍数
	baseScore := room.cfg.BaseScore // 底分
	robZhuangRate := room.mapRobZhuang[room.bankerId]
	bankerWinOrLose := int32(0)

	uids := []uint32{}
	for uid, _ := range room.mapPlayers {
		uids = append(uids, uid)
	}

	// 结算时, 从 redis 中获取所有玩家金币时, redis 出现错误
	mapUidGold, e := redisOpt.GetSomePlayersGold(uids)
	if e != nil {
		// XXX 这里要补充错误码
		return
	}

	// 结算时, redis 中没有查询出足够玩家信息
	if len(mapUidGold) != len(uids) {
		// XXX 这里要补充错误码
		return
	}

	for uid, c := range room.mapPlayers {
		if uid != room.bankerId {
			betRate := room.mapBet[uid]

			var cardTypeBanker, cardTypeNormalPlayer niuniuProto.NiuniuCardType
			var winOrLose bool
			//判断顺序：只有坎斗、只有顺斗、有坎斗有顺斗、无坎斗无顺斗
			if room.cfg.Kan && !room.cfg.Shun {
				cardTypeBanker, _, cardTypeNormalPlayer, _, winOrLose =
					niuniuAlgorithm.KanCompare(banker.GetHandCard(), c.GetHandCard())
			} else if !room.cfg.Kan && room.cfg.Shun {
				cardTypeBanker, _, cardTypeNormalPlayer, _, winOrLose =
					niuniuAlgorithm.ShunCompare(banker.GetHandCard(), c.GetHandCard())
			} else if room.cfg.Kan && room.cfg.Shun {
				cardTypeBanker, _, cardTypeNormalPlayer, _, winOrLose =
					niuniuAlgorithm.KanShunCompare(banker.GetHandCard(), c.GetHandCard())
			} else {
				cardTypeBanker, _, cardTypeNormalPlayer, _, winOrLose =
					niuniuAlgorithm.Compare(banker.GetHandCard(), c.GetHandCard())
			}

			// 庄家赢
			if winOrLose {
				// cardTypeRate 牌型倍数
				cardTypeRate := room.cfg.MapCardTypeRate[uint32(cardTypeBanker)]

				// 闲家输的钱
				gold := int32(baseScore * robZhuangRate * betRate * cardTypeRate)

				// 闲家输的钱不能超过自己身上的钱, 也就是不能为负数
				goldNow := mapUidGold[uid]
				if gold > goldNow {
					gold = goldNow
				}

				room.mapWinOrLose[uid] = -gold
				// 庄闲家赢的钱
				bankerWinOrLose += gold
				log.Debugf("预先计算: 房间号: %d, 庄家 %d 赢了, 闲家 %d 输了, 底分: %d, "+
					"庄家抢庄倍数: %d, 下注倍数: %d, 牌型倍数: %d, 庄家牌型: %s, "+
					"闲家牌型: %s",
					room.roomId, room.bankerId, uid, baseScore, robZhuangRate, betRate,
					cardTypeRate, niuniuAlgorithm.CardTypeString(cardTypeBanker),
					niuniuAlgorithm.CardTypeString(cardTypeNormalPlayer))
			} else {
				// 闲家赢
				// cardTypeRate 牌型倍数
				cardTypeRate := room.cfg.MapCardTypeRate[uint32(cardTypeNormalPlayer)]
				gold := int32(baseScore * robZhuangRate * betRate * cardTypeRate)
				room.mapWinOrLose[uid] = gold
				// 庄家输给闲家的钱
				bankerWinOrLose -= gold
				log.Debugf("预先计算: 房间号: %d, 庄家 %d 输了, 闲家 %d 赢了, 底分: %d, "+
					"庄家抢庄倍数: %d, 下注倍数: %d, 牌型倍数: %d, 庄家牌型: %s, "+
					"闲家牌型: %s",
					room.roomId, room.bankerId, uid, baseScore, robZhuangRate, betRate,
					cardTypeRate, niuniuAlgorithm.CardTypeString(cardTypeBanker),
					niuniuAlgorithm.CardTypeString(cardTypeNormalPlayer))
			}
		}
	}

	// 本局庄家赢了若干金币
	if bankerWinOrLose >= 0 {
		room.mapWinOrLose[room.bankerId] = bankerWinOrLose
	} else {
		// 本局庄家输了若干金币
		// 未结算前, 庄家身上的金币
		bankerGold := mapUidGold[room.bankerId]
		if bankerGold+bankerWinOrLose >= 0 { // 庄家身上的金币足够支付输掉的
			room.mapWinOrLose[room.bankerId] = bankerWinOrLose
		} else {
			// 如果庄家输的钱超过了身上的钱,就按比例赔付
			// total 是闲家们从庄家处赢的钱, 这里是个正数
			total := int32(0)
			// 此时 room.mapWinOrLose 是闲家的输赢情况
			for _, gold := range room.mapWinOrLose {
				if gold > 0 {
					total += gold
				}
			}

			log.Debugf("房间号: %d, 庄家 uid: %d 输的钱( %d  )超过了身上的钱( %d  ), 按比例赔付, "+
				"共计 %d 金币", room.roomId, room.bankerId, bankerWinOrLose, -bankerGold, bankerGold)

			for uid, gold := range room.mapWinOrLose {
				if gold > 0 {
					rate := float32(gold) / float32(total)
					winGold := int32(rate * float32(bankerGold))
					room.mapWinOrLose[uid] = winGold
					log.Debugf("房间号: %d, 闲家 uid: %d 按比例 %f 赢了 %d 金币 ",
						room.roomId, uid, rate, winGold)
				}
			}

			// 庄家输的钱不能超过身上的钱, 所以把钱扣完
			room.mapWinOrLose[room.bankerId] = -bankerGold
		}
	}

	for uid, gold := range room.mapWinOrLose {
		if uid == room.bankerId {
			if gold >= 0 {
				log.Debugf("房间号: %d, 庄家 uid: %d 赢了 %d 金币", room.roomId, uid, gold)
			} else {
				log.Debugf("房间号: %d, 庄家 uid: %d 输了 %d 金币", room.roomId, uid, gold)
			}
		} else {
			if gold >= 0 {
				log.Debugf("房间号: %d, 闲家 uid: %d 赢了 %d 金币", room.roomId, uid, gold)
			} else {
				log.Debugf("房间号: %d, 闲家 uid: %d 输了 %d 金币", room.roomId, uid, gold)
			}
		}

		//自建房玩家输赢统计
		if room.roomType != commonProto.RoomType_roomTypePublic {
			if _, ok := room.mapTotalWinLose[uid]; ok {
				room.mapTotalWinLose[uid] += gold
			} else {
				room.mapTotalWinLose[uid] = gold
			}
		}
	}
}
