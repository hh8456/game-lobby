package roomPkg

import (
	"math/rand"
	"servers/common-library/log"
	"time"
)

// 确定庄家, 返回值: 庄家 uid, 倍率, 相同倍率的 uid
func (room *Room) confirmBanker() (uint32, uint32, []uint32) {
	// 8 倍率的玩家
	rate8 := []uint32{}
	// 4 倍率的玩家
	rate4 := []uint32{}
	// 2 倍率的玩家
	rate2 := []uint32{}
	// 1 倍率的玩家; 另外,不抢庄的玩家默认为 1 倍率
	rate1 := []uint32{}
	for uid, rate := range room.mapRobZhuang {
		log.Debugf("confirmBanker 函数,  uid: %d 抢庄倍率: %d", uid, rate)
		switch rate {
		case 1:
			rate1 = append(rate1, uid)
		case 2:
			rate2 = append(rate2, uid)
		case 4:
			rate4 = append(rate4, uid)
		case 8:
			rate8 = append(rate8, uid)
		}
	}

	rand.Seed(time.Now().UnixNano())
	if len(rate8) > 0 {
		idx := rand.Intn(len(rate8))
		return rate8[idx], 8, rate8
	}

	if len(rate4) > 0 {
		idx := rand.Intn(len(rate4))
		return rate4[idx], 4, rate4
	}

	if len(rate2) > 0 {
		idx := rand.Intn(len(rate2))
		return rate2[idx], 2, rate2
	}

	if len(rate1) > 0 {
		idx := rand.Intn(len(rate1))
		return rate1[idx], 1, rate1
	}

	// 对于没有操作的玩家,默认为不抢庄
	for uid, _ := range room.mapPlayers {
		log.Debugf("confirmBanker 函数, 检查玩家 uid: %d 是否抢庄倍率: 1", uid)
		if _, find := room.mapRobZhuang[uid]; !find {
			log.Debugf("confirmBanker 函数, 检查玩家 uid: %d 没抢庄, 默认倍率: 1", uid)
			rate1 = append(rate1, uid)
		} else {
			log.Debugf("confirmBanker 函数, 检查玩家 uid: %d 抢庄", uid)
		}
	}

	if len(rate1) > 0 {
		idx := rand.Intn(len(rate1))
		return rate1[idx], 1, rate1
	}

	// 不可能执行到这里
	log.Errorf("牛牛定庄时 Room.confirmBanker 发生错误, 不应该出现这条日志")

	return 0, 0, []uint32{}
}
