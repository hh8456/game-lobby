package niuniuConfig

import (
	"servers/common-library/function"
	"servers/common-library/proto/niuniuProto"
	"servers/common-library/redisKeyPrefix"
	"strconv"

	"github.com/hh8456/go-common/redisObj"
)

const (
	// 由于1金币在后台是用 100 分表示,所以这里是 100 和 200, 在前端展示为 1 和 2
	Bet1 = 100 // 一倍下注
	Bet2 = 200 // 两倍下注
)

func DefaultRoomConfig() *niuniuProto.RoomConfig {
	cfg := &niuniuProto.RoomConfig{}
	cfg.BaseScore = 1
	cfg.MapRobZhuangRate = map[uint32]uint32{}
	// 不抢庄为 1 倍抢庄
	cfg.MapRobZhuangRate[1] = 0
	// 2 倍抢庄
	cfg.MapRobZhuangRate[2] = 0
	// 4 倍抢庄
	cfg.MapRobZhuangRate[4] = 0
	// 8 倍抢庄
	cfg.MapRobZhuangRate[8] = 0
	cfg.MapBetRate = map[uint32]uint32{}
	// 下注倍数
	cfg.MapBetRate[Bet1] = 1
	cfg.MapBetRate[Bet2] = 1
	// 牌型 - 倍数
	cfg.MapCardTypeRate = map[uint32]uint32{}
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_0)] = 1
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_1)] = 1
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_2)] = 2
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_3)] = 3
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_4)] = 4
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_5)] = 5
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_6)] = 6
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_7)] = 7
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_8)] = 8
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_9)] = 9
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_niu)] = 10
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_seq)] = 15
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_fiveFlower)] = 15
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_sameFlower)] = 20
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_hulu)] = 25
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_boom)] = 30
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_five)] = 35
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_straightFlush)] = 40

	// 是否有蕃, 有蕃表示 牌型倍数需要代入公式计算
	cfg.CardTypeHasKind = true

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

	// 默认为明牌抢庄
	cfg.KnownCard = true

	return cfg
}

// TODO 这里要做成 web 后台功能, 把房间配置写入 redis
func SetRoomConfig(roomIds []uint32) {
	rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuRoomConfig)
	cfg := &niuniuProto.RoomConfig{}
	cfg.BaseScore = 1
	cfg.MapRobZhuangRate = map[uint32]uint32{}
	// 不抢庄为 1 倍抢庄
	cfg.MapRobZhuangRate[1] = 0
	// 2 倍抢庄
	cfg.MapRobZhuangRate[2] = 0
	// 4 倍抢庄
	cfg.MapRobZhuangRate[4] = 0
	// 8 倍抢庄
	cfg.MapRobZhuangRate[8] = 0
	cfg.MapBetRate = map[uint32]uint32{}
	// 下注倍数
	cfg.MapBetRate[Bet1] = 1
	cfg.MapBetRate[Bet2] = 1
	cfg.MapCardTypeRate = map[uint32]uint32{}
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_0)] = 1
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_1)] = 1
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_2)] = 2
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_3)] = 3
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_4)] = 4
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_5)] = 5
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_6)] = 6
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_7)] = 7
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_8)] = 8
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_9)] = 9
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_niu)] = 10
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_seq)] = 15
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_fiveFlower)] = 15
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_sameFlower)] = 20
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_hulu)] = 25
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_boom)] = 30
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_five)] = 35
	cfg.MapCardTypeRate[uint32(niuniuProto.NiuniuCardType_niuniuCardType_straightFlush)] = 40

	cfg.CardTypeHasKind = true

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
	cfg.MapWaitTime[uint32(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusPlay)] = 6
	// 房间结算状态的等待时长
	cfg.MapWaitTime[uint32(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusSettle)] = 1

	// 前面4个明牌抢庄
	for i := 0; i < 4; i++ {
		cfg.RoomId = roomIds[i]
		cfg.KnownCard = true
		buf, b := function.ProtoMarshal(cfg, "niuniuProto.RoomConfig")
		if b {
			rds.Set(strconv.Itoa(int(cfg.RoomId)), string(buf))
		}
	}

	// 后面 6 个暗牌抢庄
	for i := 4; i < 10; i++ {
		cfg.RoomId = roomIds[i]
		cfg.KnownCard = false
		buf, b := function.ProtoMarshal(cfg, "niuniuProto.RoomConfig")
		if b {
			rds.Set(strconv.Itoa(int(cfg.RoomId)), string(buf))
		}
	}
}
