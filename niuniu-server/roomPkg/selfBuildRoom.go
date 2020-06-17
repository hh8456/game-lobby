package roomPkg

import (
	"crypto/rand"
	"math/big"
	"servers/common-library/function"
	"servers/common-library/log"
	"servers/common-library/proto/commonProto"
	"servers/common-library/proto/niuniuProto"
	"servers/common-library/redisKeyPrefix"
	"servers/common-library/utility"
	"servers/niuniu-server/selfBuildRoom"
	"strconv"
	"time"

	"github.com/hh8456/go-common/redisObj"
	"github.com/jinzhu/gorm"
)

func SelfBuildNewRoom(c IClient, db *gorm.DB, req *niuniuProto.C2SSelfBuildNiuNiuRoom) (*Room, error) {
	var roomType commonProto.RoomType

	if req.RoomPublic == 0 {
		roomType = commonProto.RoomType_selfBuildRoomTypePublic
	} else {
		roomType = commonProto.RoomType_roomTypePrivate
	}

	cfg := selfBuildRoom.GetSelfBuildRoomConfig(req)
	err, roomId := GetRoomId()
	if err != nil {
		log.Error("自建房获取房间id发生错误")
		return nil, err
	}

	rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuRoomConfig)
	buf, b := function.ProtoMarshal(cfg, "niuniuProto.RoomConfig")
	if b {
		rds.Set(strconv.Itoa(int(roomId)), string(buf))
	}

	//是否允许中途加入
	roomRedis := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuNiuRoomEnterAfterBegin)
	err = roomRedis.Setex(strconv.Itoa(int(roomId)), 24*60*60*time.Second, req.Add)
	if err != nil {
		log.Error("自建房存房间中途加入字段错误")
		return nil, err
	}

	players := make([]IClient, req.PlayerNumber)
	players[0] = c

	//ctx, cancel := context.WithCancel(context.Background())
	p := &Room{
		serverId:         c.GetUid(),
		roomId:           roomId,
		roomType:         roomType,
		roomStatus:       niuniuProto.NiuniuRoomStatus_niuniuRoomStatusReady,
		chanClientAndMsg: make(chan *ClientAndMsg, 100),
		//createTimestamp:  time.Now().UnixNano() / 1e6,
		chanTimer: make(chan int64, 10),
		//ctx:              ctx,
		//cancel:           cancel,
		seat:             players,
		mapBystanders:    map[uint32]IClient{},
		mapPlayers:       map[uint32]IClient{},
		cardHeap:         utility.GetPokerHeap(),
		mapRobZhuang:     map[uint32]uint32{},
		mapWinOrLose:     map[uint32]int32{}, // uid - 输赢金币
		mapBet:           map[uint32]uint32{},
		mapShowCard:      map[uint32]*niuniuProto.S2CNiuniuShowCard{},
		cfg:              cfg,
		setRoomPlayerNum: SetRoomPlayerNum,
		db:               db,
		owner:            c.GetUid(),
		mapTotalWinLose:  map[uint32]int32{},
	}

	return p, err
}

func SelfBuildRoomEnterRoom(room *Room, c IClient) {
	c2sEnterNiuniuRoom(room, c, nil)
	//房主
	c2sNiuniuHaveASeat(room, c, nil)
}

//crypto/rand 的rand.Int不需要随机数种子，math/rand需要
// FIXME 这里需要返回唯一错误码(自行定义在 error_code.proto 中); 如果错误码为 0 表示成功,非 0 就表示失败, 同时要把失败的错误码下发给客户端,以方便定位 bug
func GetRoomId() (err error, roomIdResult uint32) {
	for {
		result, _ := rand.Int(rand.Reader, big.NewInt(900000))
		roomId := result.Int64() + 100000
		roomIdString := strconv.FormatInt(roomId, 10)
		rdsLogin := redisObj.NewSessionWithPrefix(redisKeyPrefix.SelfBuildNiuNiuRoom)
		reply, e := rdsLogin.SetExNx(roomIdString, 0, 24*60*60*time.Second)
		if e != nil {
			log.Error("自建房获取房间id时 redis 锁发生错误: %v", e)
			return e, 0
		}

		if reply != "NX" {
			rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.SelfBuildNiuNiuRoom)
			result, e := rds.AddSetMembers(redisKeyPrefix.IdSet, roomIdString)
			if result == 0 {
				log.Error("自建房获取房间id,id加入集合是 redis 锁发生错误: %v", e)
				return e, 0
			}

			return e, uint32(roomId)
		}
	}
}

func GetSelfBuildRoomMessage(uid, roomId uint32) *niuniuProto.S2CSelfBuildNiuNiuRoom {
	msg := &niuniuProto.S2CSelfBuildNiuNiuRoom{}
	room := GetRoom(roomId)

	msg.RoomId = roomId
	msg.BaseScore = room.cfg.BaseScore
	msg.MapRobZhuangRate = room.mapRobZhuang
	msg.MapBetRate = room.cfg.MapBetRate
	msg.MapCardTypeRate = room.cfg.MapCardTypeRate
	msg.CardTypeHasKind = room.cfg.CardTypeHasKind
	msg.MapWaitTime = room.cfg.MapWaitTime
	msg.KnownCard = room.cfg.KnownCard

	m := room.getPlayerInSeat()
	msg.MapPlayerBaseInfo = map[uint32]*commonProto.PlayerBaseInfo{}
	for seatIdx, client := range m {
		msg.MapPlayerBaseInfo[seatIdx] = client.GetPlayerBaseInfo()
	}

	return msg
}
