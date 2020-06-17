package redisOpt

import (
	"servers/common-library/log"
	"servers/common-library/proto/commonProto"
	"servers/common-library/redisKeyPrefix"
	"servers/model"
	"strconv"
	"time"

	"github.com/hh8456/go-common/redisObj"
)

// 本文件中 设置了 redis.Expire 的地方, 在游戏服中也要定时的设置 redis.Expire

// 获得房间中坐下的玩家 id 和座位号
func GetNiuniuPlayerRoomIdAndSeatIdx(roomId uint32) []string {
	strRoomId := strconv.Itoa(int(roomId))
	//zset: roomId - uid - seatIdx
	prefix := redisKeyPrefix.NiuniuPlayerInRoom + ":" + strRoomId + ":" + redisKeyPrefix.NiuniuSeatIndex
	rdsPlayerRoomIdAndSeatIdx := redisObj.NewSessionWithPrefix(prefix)
	// 座位 [0, 9]
	uidsInSeat, _ := rdsPlayerRoomIdAndSeatIdx.MGet([]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"})

	return uidsInSeat
}

// 设置玩家的房间和座位 ,
// redis 中, niuniu_player_room_id:uid - uid, niuniu_seat_index:uid - uid, niuniu_player_room_id:roomid:niuniu_seat_index:seatIdx - uid
// 这3个键值对的生命周期是一致的
func SetNiuniuPlayerRoomIdAndSeatIdx(strUid string, roomId, seatIdx uint32) {
	strRoomId := strconv.Itoa(int(roomId))
	strSeatIdx := strconv.Itoa(int(seatIdx))
	// niuniu_room:房间号:niuniu_seat_index:座位号
	prefix := redisKeyPrefix.NiuniuPlayerInRoom + ":" + strRoomId + ":" + redisKeyPrefix.NiuniuSeatIndex
	rdsPlayerRoomIdAndSeatIdx := redisObj.NewSessionWithPrefix(prefix)
	// baseServer.time 定时器是每 30 秒一次, 所以这里的生命周期设置为 1 分钟
	rdsPlayerRoomIdAndSeatIdx.Setex(strSeatIdx, time.Minute, strUid)

	log.Debugf("在 redis 中设置 ( roomid:seatIdx - uid )键值对 %s - %s, 存活期 1 分钟 ",
		prefix+":"+strSeatIdx, strUid)

	rdsPlayerRoomId := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuPlayerRoomId)
	rdsPlayerRoomId.Setex(strUid, time.Minute, strRoomId)
	log.Debugf("在 redis 中设置( uid - roomid )键值对 %s - %s, 存活期 1 分钟 ",
		redisKeyPrefix.NiuniuPlayerRoomId+":"+strUid, strRoomId)

	rdsPlayerSeatIdx := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuSeatIndex)
	rdsPlayerSeatIdx.Setex(strUid, time.Minute, strSeatIdx)
	log.Debugf("在 redis 中设置( uid - seatIdx )键值对 %s - %s, 存活期 1 分钟 ",
		redisKeyPrefix.NiuniuSeatIndex+":"+strUid, strSeatIdx)
}

// 玩家离开座位
// redis 中, niuniu_player_room_id:uid - uid, niuniu_seat_index:uid - uid, niuniu_player_room_id:roomid:niuniu_seat_index:seatIdx - uid
// 这3个键值对的生命周期是一致的
func DelNiuniuPlayerRoomIdAndSeatIdx(strUid string, roomId, seatIdx uint32) {
	strRoomId := strconv.Itoa(int(roomId))
	strSeatIdx := strconv.Itoa(int(seatIdx))
	rdsPlayerRoomId := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuPlayerRoomId)
	rdsPlayerRoomId.Del(strUid)

	rdsPlayerSeatIdx := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuSeatIndex)
	rdsPlayerSeatIdx.Del(strUid)

	// niuniu_room:房间号:niuniu_seat_index:座位号 - uid
	prefix := redisKeyPrefix.NiuniuPlayerInRoom + ":" + strRoomId + ":" + redisKeyPrefix.NiuniuSeatIndex
	rdsPlayerRoomIdAndSeatIdx := redisObj.NewSessionWithPrefix(prefix)
	rdsPlayerRoomIdAndSeatIdx.Del(strSeatIdx)

}

func GetNiuniuPlayerRoomId(uid uint32) uint32 {
	rdsPlayerRoomId := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuPlayerRoomId)
	strRoomId, e := rdsPlayerRoomId.Get(strconv.Itoa(int(uid)))
	if e == nil {
		roomId, _ := strconv.ParseInt(strRoomId, 10, 64)
		return uint32(roomId)
	}

	return 0
}

func GetSomePlayersGold(uids []uint32) (map[uint32]int32, error) {
	m := map[uint32]int32{}

	for _, uid := range uids {
		strCoin, e := GetOnePlayerGold(uid)
		if e != nil {
			return nil, e
		}

		coin, e2 := strconv.Atoi(strCoin)
		if e2 != nil {
			return nil, e2
		}

		m[uint32(uid)] = int32(coin)
	}

	return m, nil
}

func SetOnePlayerGold(uid uint32, gold int32) (int, error) {
	rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.PlayerBaseInfo)
	strUid := strconv.Itoa(int(uid))
	return rds.HashSet(strUid, redisKeyPrefix.PlayerBaseInfo_Gold, gold)
}

func GetOnePlayerGold(uid uint32) (string, error) {
	rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.PlayerBaseInfo)
	strUid := strconv.Itoa(int(uid))
	return rds.GetHashSetField(strUid, redisKeyPrefix.PlayerBaseInfo_Gold)
}

func GetOnePlayerInviteCode(uid uint32) (string, error) {
	rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.PlayerBaseInfo)
	strUid := strconv.Itoa(int(uid))
	return rds.GetHashSetField(strUid, redisKeyPrefix.PlayerBaseInfo_InviteCode)
}

// 把键值对 wxid - uid 保存到 redis
func SaveWxidAndUid(wxid string, uid int) {
	rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.WxidAndUId)
	rds.Set(wxid, strconv.Itoa(uid))
}

// 通过 wxid 查询 uid
func GetUidByWxid(wxid string) (string, error) {
	rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.WxidAndUId)
	return rds.Get(wxid)
}

func SavePlayerBaseInfo(playerBaseInfo *model.PlayerBaseInfo) {
	rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.PlayerBaseInfo)
	// 写入 redis hash
	hashKey := strconv.Itoa(int(playerBaseInfo.UId))
	m := map[string]interface{}{}
	m[redisKeyPrefix.PlayerBaseInfo_Wxid] = playerBaseInfo.Wxid
	m[redisKeyPrefix.PlayerBaseInfo_WxidCrc32] = playerBaseInfo.WxidCrc32
	m[redisKeyPrefix.PlayerBaseInfo_UId] = playerBaseInfo.UId
	m[redisKeyPrefix.PlayerBaseInfo_HeadPic] = playerBaseInfo.HeadPic
	m[redisKeyPrefix.PlayerBaseInfo_InviteCode] = playerBaseInfo.InviteCode
	m[redisKeyPrefix.PlayerBaseInfo_Diamond] = playerBaseInfo.Diamond
	m[redisKeyPrefix.PlayerBaseInfo_Gold] = playerBaseInfo.Gold
	m[redisKeyPrefix.PlayerBaseInfo_Sex] = playerBaseInfo.Sex
	m[redisKeyPrefix.PlayerBaseInfo_Name] = playerBaseInfo.Name
	m[redisKeyPrefix.PlayerBaseInfo_RegDate] = playerBaseInfo.RegDate.Unix()

	rds.HashMultipleSet(hashKey, m)
}

// 获取玩家简略信息
func LoadPlayerBriefInfo(strUid string) (*commonProto.PlayerBriefInfo, bool) {
	rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.PlayerBaseInfo)
	values, e := rds.GetHashMultipleSet(strUid,
		redisKeyPrefix.PlayerBaseInfo_UId,
		redisKeyPrefix.PlayerBaseInfo_HeadPic,
		redisKeyPrefix.PlayerBaseInfo_Name,
	)

	if e != nil {
		return nil, false
	}

	if values != nil && len(values) != 3 {
		return nil, false
	}

	p := &commonProto.PlayerBriefInfo{}
	if values[0] != nil {
		uid, _ := strconv.ParseInt(string(values[0].([]byte)), 10, 64)
		p.Uid = uint32(uid)
	}

	if values[1] != nil {
		p.HeadPic = string(values[1].([]byte))
	}

	if values[2] != nil {
		p.Name = string(values[2].([]byte))
	}

	return p, true
}

func LoadPlayerBaseInfo(strUid string) (*model.PlayerBaseInfo, bool) {
	rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.PlayerBaseInfo)

	values, e := rds.GetHashMultipleSet(strUid,
		redisKeyPrefix.PlayerBaseInfo_Wxid,
		redisKeyPrefix.PlayerBaseInfo_WxidCrc32,
		redisKeyPrefix.PlayerBaseInfo_UId,
		redisKeyPrefix.PlayerBaseInfo_HeadPic,
		redisKeyPrefix.PlayerBaseInfo_InviteCode,
		redisKeyPrefix.PlayerBaseInfo_Diamond,
		redisKeyPrefix.PlayerBaseInfo_Gold,
		redisKeyPrefix.PlayerBaseInfo_Sex,
		redisKeyPrefix.PlayerBaseInfo_Name,
		redisKeyPrefix.PlayerBaseInfo_RegDate,
	)

	if e != nil {
		return nil, false
	}

	if values != nil && len(values) != 10 {
		return nil, false
	}

	p := &model.PlayerBaseInfo{}

	if values[0] != nil {
		p.Wxid = string(values[0].([]byte))
	} else {
		return nil, false
	}

	if values[1] != nil {
		//p.WxidCrc32 = int64(binary.BigEndian.Uint64(values[1].([]byte)))
		p.WxidCrc32, _ = strconv.ParseInt(string(values[1].([]byte)), 10, 64)
	}

	if values[2] != nil {
		uid, _ := strconv.ParseInt(string(values[2].([]byte)), 10, 64)
		p.UId = int(uid)
	}

	if values[3] != nil {
		p.HeadPic = string(values[3].([]byte))
	}

	if values[4] != nil {
		inviteCode, _ := strconv.ParseInt(string(values[4].([]byte)), 10, 64)
		p.InviteCode = int(inviteCode)
	}

	if values[5] != nil {
		diamond, _ := strconv.ParseInt(string(values[5].([]byte)), 10, 64)
		p.Diamond = int(diamond)
	}

	if values[6] != nil {
		gold, _ := strconv.ParseInt(string(values[6].([]byte)), 10, 64)
		p.Gold = int(gold)
	}

	if values[7] != nil {
		sex, _ := strconv.ParseInt(string(values[7].([]byte)), 10, 64)
		p.Sex = int8(sex)
	}

	if values[8] != nil {
		p.Name = string(values[8].([]byte))
	}

	if values[9] != nil {
		//fmt.Printf("时间戳: %s\n", string(values[10].([]byte)))
		// 带时区的时间戳非常不好解析,为了正确解析,必须要对时间字符串做截断
		ts, _ := strconv.ParseInt(string(values[9].([]byte)), 10, 64)
		p.RegDate = time.Unix(ts, 0)
	}

	return p, true
}
