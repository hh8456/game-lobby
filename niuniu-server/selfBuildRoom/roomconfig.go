package selfBuildRoom

import (
	"servers/common-library/log"
	"servers/common-library/proto/niuniuProto"
)

const (
	// 由于1金币在后台是用 100 分表示,所以这里是 100 和 200, 在前端展示为 1 和 2
	Bet1  = 100     // 一倍下注
	Bet2  = 2 * 100 // 两倍下注
	Bet3  = 3 * 100 // 三倍下注
	Bet4  = 4 * 100
	Bet5  = 5 * 100
	Bet6  = 6 * 100
	Bet8  = 8 * 100
	Bet10 = 10 * 100
	Bet12 = 12 * 100
	Bet15 = 15 * 100
	Bet16 = 16 * 100
	Bet20 = 20 * 100
	Bet24 = 24 * 100
	Bet32 = 32 * 100
)

func GetSelfBuildRoomConfig(req *niuniuProto.C2SSelfBuildNiuNiuRoom) *niuniuProto.RoomConfig {
	log.Debugf("收到自建房消息：", req)

	cfg := &niuniuProto.RoomConfig{}
	cfg.BaseScore = 1
	//抢庄倍数
	cfg.MapRobZhuangRate = map[uint32]uint32{}
	for i := int32(1); i <= req.TheBiggestRobZhuang; i++ {
		cfg.MapRobZhuangRate[uint32(i)] = 0
	}
	//暗抢庄家
	if req.DarkGrabBanker == 0 {
		cfg.KnownRobZhuang = false
	} else {
		cfg.KnownRobZhuang = true
	}

	// 下注倍数
	cfg.MapBetRate = map[uint32]uint32{}
	if req.LowGradeType == 1 {
		if req.LowGrade == 0 {
			cfg.MapBetRate[Bet1] = 1
			cfg.MapBetRate[Bet2] = 1
		} else if req.LowGrade == 1 {
			cfg.MapBetRate[Bet2] = 1
			cfg.MapBetRate[Bet4] = 1
		} else if req.LowGrade == 2 {
			cfg.MapBetRate[Bet3] = 1
			cfg.MapBetRate[Bet6] = 1
		} else if req.LowGrade == 3 {
			cfg.MapBetRate[Bet4] = 1
			cfg.MapBetRate[Bet8] = 1
		}
	} else if req.LowGradeType == 2 {
		if req.LowGrade == 0 {
			cfg.MapBetRate[Bet1] = 1
			cfg.MapBetRate[Bet2] = 1
			cfg.MapBetRate[Bet3] = 1
			cfg.MapBetRate[Bet4] = 1
		} else if req.LowGrade == 1 {
			cfg.MapBetRate[Bet2] = 1
			cfg.MapBetRate[Bet3] = 1
			cfg.MapBetRate[Bet4] = 1
			cfg.MapBetRate[Bet5] = 1
		} else if req.LowGrade == 2 {
			cfg.MapBetRate[Bet2] = 1
			cfg.MapBetRate[Bet4] = 1
			cfg.MapBetRate[Bet6] = 1
			cfg.MapBetRate[Bet8] = 1
		} else if req.LowGrade == 3 {
			cfg.MapBetRate[Bet3] = 1
			cfg.MapBetRate[Bet6] = 1
			cfg.MapBetRate[Bet12] = 1
			cfg.MapBetRate[Bet24] = 1
		} else if req.LowGrade == 4 {
			cfg.MapBetRate[Bet4] = 1
			cfg.MapBetRate[Bet8] = 1
			cfg.MapBetRate[Bet16] = 1
			cfg.MapBetRate[Bet32] = 1
		} else if req.LowGrade == 5 {
			cfg.MapBetRate[Bet5] = 1
			cfg.MapBetRate[Bet10] = 1
			cfg.MapBetRate[Bet15] = 1
			cfg.MapBetRate[Bet20] = 1
		}
	}

	// 牌型 - 倍数
	var doubleRules []uint32
	cfg.MapCardTypeRate = map[uint32]uint32{}
	if req.DoubleRules == 1 {
		doubleRules = []uint32{1, 1, 1, 1, 1, 1, 1, 2, 2, 3, 4}
	} else if req.DoubleRules == 2 {
		doubleRules = []uint32{1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 3}
	} else if req.DoubleRules == 3 {
		doubleRules = []uint32{1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 3}
	} else {
		doubleRules = []uint32{1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	}

	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_0)] = doubleRules[0]
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_1)] = doubleRules[1]
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_2)] = doubleRules[2]
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_3)] = doubleRules[3]
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_4)] = doubleRules[4]
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_5)] = doubleRules[5]
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_6)] = doubleRules[6]
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_7)] = doubleRules[7]
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_8)] = doubleRules[8]
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_9)] = doubleRules[9]
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_niu)] = doubleRules[10]
	//特殊牌型
	//特殊牌型翻倍规则
	var specialCardTypeRate []uint32
	if req.SpecialDoubleRules == 1 {
		specialCardTypeRate = []uint32{15, 15, 20, 25, 30, 35, 40}
	} else {
		specialCardTypeRate = []uint32{5, 5, 10, 15, 50, 25, 30}
	}
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_seq)] = specialCardTypeRate[0]
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_fiveFlower)] = specialCardTypeRate[1]
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_sameFlower)] = specialCardTypeRate[2]
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_hulu)] = specialCardTypeRate[3]
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_boom)] = specialCardTypeRate[4]
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_five)] = specialCardTypeRate[5]
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_straightFlush)] = specialCardTypeRate[6]

	// 是否有蕃, 有蕃表示 牌型倍数需要代入公式计算
	cfg.CardTypeHasKind = true

	//加速模式
	if req.TheShoot == 1 {
		cfg.Faster = true
	} else {
		cfg.Faster = false
	}
	// 各个阶段的等待时长, 这个字段是服务器使用,客户端忽略
	// 阶段 - 等待时长
	cfg.MapWaitTime = map[uint32]int64{}
	// 房间准备状态的等待时长
	cfg.MapWaitTime[uint32(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusReady)] = 3
	// 房间倒计时状态的等待时长
	cfg.MapWaitTime[uint32(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusCountDown)] = 1
	// 房间抢庄状态的等待时长
	cfg.MapWaitTime[uint32(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusRobZhuang)] = 5
	// 房间下注状态的等待时长, 5 秒; 但客户端需要额外用1秒来展示定庄效果,所以是 6 秒
	cfg.MapWaitTime[uint32(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusBet)] = 6
	// 房间亮牌状态的等待时长
	cfg.MapWaitTime[uint32(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusPlay)] = 7
	// 房间结算状态的等待时长
	cfg.MapWaitTime[uint32(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusSettle)] = 1

	// 1:明牌抢庄 2：暗牌抢庄
	if req.PlayingMethod == 1 {
		cfg.KnownCard = true
	} else {
		cfg.KnownCard = false
	}

	//坎斗
	if req.HeFights == 1 {
		cfg.Kan = true
	} else {
		cfg.Kan = false
	}

	//顺斗
	if req.AlongTheBucket == 1 {
		cfg.Shun = true
	} else {
		cfg.Shun = false
	}

	//总局数
	//cfg.TotalNumberOfGame = uint32(req.NumberOfGame)
	//总局数
	cfg.TotalNumberOfGame = 2
	return cfg
}
