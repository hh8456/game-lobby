package playerPkg

import (
	"servers/base-library/base_net"
	"servers/common-library/baseServer"
	"servers/common-library/config"
	"servers/common-library/connData"
	"servers/common-library/function"
	"servers/common-library/log"
	"servers/common-library/proto/commonProto"
	"servers/common-library/proto/errorCodeProto"
	"servers/common-library/proto/msgIdProto"
	"servers/common-library/proto/niuniuProto"
	"servers/common-library/redisKeyPrefix"
	"servers/niuniu-server/roomPkg"
	"strconv"
	"strings"
	"time"

	"github.com/hh8456/go-common/redisObj"
)

var (
	gRoomId            uint32
	mapPlayerLogicFunc map[msgIdProto.MsgId]func(*Player, *connData.ConnData)
)

func init() {
	mapPlayerLogicFunc = map[msgIdProto.MsgId]func(*Player, *connData.ConnData){}
	// 进入牛牛房间
	mapPlayerLogicFunc[msgIdProto.MsgId_c2sEnterNiuniuRoom] = c2sEnterNiuniuRoom
	// 离开牛牛房间
	mapPlayerLogicFunc[msgIdProto.MsgId_c2sLeaveNiuniuRoom] = c2sLeaveNiuniuRoom
	// gate 通知 niuniu 客户端断线
	mapPlayerLogicFunc[msgIdProto.MsgId_gate2NiuniuClientDisconnect] = gate2NiuniuClientDisconnect
	//自建牛牛房间
	mapPlayerLogicFunc[msgIdProto.MsgId_c2sCreateNiuniuRoom] = C2SSelfBuildRoom
}

// 进入牛牛房间
func c2sEnterNiuniuRoom(c *Player, connData *connData.ConnData) {
	dp := &base_net.DataPack{}
	c2sPbMsg := &niuniuProto.C2SEnterNiuniuRoom{}
	uid := c.GetUid()

	if function.ProtoUnmarshal(connData.BinData[dp.GetHeadLen():],
		c2sPbMsg, "niuniuProto.C2SEnterNiuniuRoom") {
		roomId := c2sPbMsg.GetRoomId()
		if c.isInRoom() == false {
			room := roomPkg.GetRoom(roomId)
			if room != nil {
				if AllowEnterSelfBuildRoomAfterBegin(room) == false {
					c.SendErrorCode(errorCodeProto.ErrorCode_can_not_enter_self_build_room)
					log.Debugf("进入不被允许中途加入的房间")
					return
				}

				c.SetRoomId(c2sPbMsg.GetRoomId(), room)
				log.Debugf("uid %d 进入牛牛房间 %d", uid, roomId)
				room.Handle(c, connData)
			} else {
				log.Errorf("%d 请求进入的牛牛房间 %d 不存在时",
					uid, roomId)

				// 请求进入大厅的某个牛牛房间时,发现房间不存在
				c.SendErrorCode(errorCodeProto.ErrorCode_niuniu_public_room_is_not_exist_when_request_enter_room)
			}
		} else {
			if c2sPbMsg.GetRoomId() == c.GetRoomId() {
				log.Debugf("uid %d 进入牛牛房间 %d 时发现已经在房间中了", uid, roomId)
				room := c.GetRoom()
				if room != nil {
					room.Handle(c, connData)
				}
			} else {
				// 牛牛断线重连时,上发的 room id 有误
				log.Errorf("牛牛断线重连时, 客户端 uid: %d, wxid: %s, 上发得 roomid: %d 和服务器保存的玩家 roomid: %d 不同, "+
					"两者应该相等 ", c.GetUid(), c.GetWxid(), c2sPbMsg.GetRoomId(), c.GetRoomId())
				c.SendErrorCode(errorCodeProto.ErrorCode_niuniu_room_id_error_when_reconnect)
			}
		}
	}

}

// 主动离开牛牛房间; room.mapPlayers 和 room.mapBystanders 中的 client 会删除掉
func c2sLeaveNiuniuRoom(c *Player, connData *connData.ConnData) {
	uid := c.GetUid()
	roomId := c.GetRoomId()
	if c.IsOnline() {
		log.Debugf("niuniu server 上收到 msgIdProto.MsgId_c2sLeaveNiuniuRoom 消息, "+
			"玩家 uid: %d, wxid: %s, 离开房间: roomid: %d, connid: %d, 1",
			uid, c.GetWxid(), roomId, c.GetConnId())

		room := c.GetRoom()
		if room != nil {
			if room.IsPlaying(uid) == false {
				c.Offline()

				room.Handle(c, connData)
			} else {
				log.Debugf("niuniu server 上收到 msgIdProto.MsgId_c2sLeaveNiuniuRoom 消息, "+
					"由于正在游戏中所以无法退出connid: %d , uid: %d, ", c.GetConnId(), uid)
				c.SendErrorCode(errorCodeProto.ErrorCode_can_not_leave_niuniu_room_when_playing)
			}
		} else {
			c.SendErrorCode(errorCodeProto.ErrorCode_has_not_in_niuniu_room_when_req_leave_niuniu_room)
		}

	} else {
		log.Errorf("发现 uid %d 重复发送离开牛牛房间 %d", uid, roomId)
	}
}

// gate 通知 niuniu 客户端断线
func gate2NiuniuClientDisconnect(c *Player, connData *connData.ConnData) {
	uid := c.GetUid()
	roomId := c.GetRoomId()
	if c.IsOnline() {
		// 设置为离线状态
		c.Offline()
		log.Debugf("niuniu server 上收到客户端断线消息 "+
			" msgIdProto.MsgId_gate2NiuniuClientDisconnect 消息, "+
			"connid: %d, uid: %d, wxid: %s", c.GetConnId(), uid, c.GetWxid())
		room := c.GetRoom()
		if room != nil {
			// 如果不在游戏中,就从房间中删除,
			if false == room.IsPlaying(uid) {
				log.Debugf("niuniu server 上收到客户端 uid: %d, wxid: %s, connid: %d, 断线消息 "+
					" msgIdProto.MsgId_gate2NiuniuClientDisconnect , 由于不在游戏中, 所以"+
					"按照离开房间 roomid: %d 进行处理, 并 设置了, player.offline = 1, 表示离线 ",
					uid, c.GetWxid(), c.GetConnId(), roomId)
				room.Handle(c, connData)
			} else {
				log.Debugf("niuniu server 上收到客户端 uid: %d, wxid: %s, connid: %d, 断线消息 "+
					" msgIdProto.MsgId_gate2NiuniuClientDisconnect , 由于正在游戏中, 所以"+
					"所以仅仅设置了, player.offline = 1, 表示离线,  roomid: %d ",
					uid, c.GetWxid(), c.GetConnId(), roomId)
			}
		}
	} else {
		log.Warnf("重复发送离开牛牛房间的请求, 出现这条日志,应该是玩家 uid %d 先发送了离开房间 roomid: %d 的消息, "+
			"属于正常情况", uid, roomId)
	}
}

//自建牛牛房
func C2SSelfBuildRoom(c *Player, connData *connData.ConnData) {
	log.Debugf("收到自建牛牛房消息%d", c.uid)
	if c.isInRoom() != false {
		c.SendErrorCode(errorCodeProto.ErrorCode_can_not_build_room_when_in_room)
		return
	}

	dbStr := config.Cfg.MysqlAddr
	db := baseServer.New(dbStr, 30*60)
	dp := &base_net.DataPack{}
	c2sPbMsg := &niuniuProto.C2SSelfBuildNiuNiuRoom{}
	if function.ProtoUnmarshal(connData.BinData[dp.GetHeadLen():],
		c2sPbMsg, "niuniuProto.C2SSelfBuildNiuNiuRoom") {
		//条件判断
		if c2sPbMsg.PlayerNumber < 4 || c2sPbMsg.PlayerNumber > 10 {
			c.SendErrorCode(errorCodeProto.ErrorCode_can_not_self_build_room_with_bad_request)
			return
		}

		room, err := roomPkg.SelfBuildNewRoom(c, db.DB, c2sPbMsg)
		if err != nil {
			c.SendErrorCode(errorCodeProto.ErrorCode_redis_error_when_self_build_niuniu_room)
			return
		}

		log.Debugf("是否允许中途加入自建房", c2sPbMsg.Add)

		rdsCrRoom := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuRoom)
		roomInfo := &niuniuProto.RoomInfo{RoomId: room.GetRoomId(), RoomType: room.GetRoomType()}
		binBuf, _ := function.ProtoMarshal(roomInfo, "niuniuProto.RoomInfo")

		_, e := rdsCrRoom.SetExNx(strconv.Itoa(int(room.GetRoomId())), string(binBuf), 24*60*60*time.Second)
		if e != nil {
			log.Errorf("自建牛牛房写入房间信息时 redis发生错误: %v", e)
			c.SendErrorCode(errorCodeProto.ErrorCode_redis_error_when_self_build_niuniu_room_save_room_info)
			return
		}

		roomPkg.StoreRoom(room.GetRoomId(), room)
		room.Run()

		c.SetRoomId(room.GetRoomId(), room)

		log.Debugf("uid %d 自建牛牛房间 %d", c.uid, room.GetRoomId())
		room.Handle(c, connData)

		c.SendPbMsg(msgIdProto.MsgId_s2cCreateNiuniuRoom, roomPkg.GetSelfBuildRoomMessage(c.uid, room.GetRoomId()))
		roomPkg.SelfBuildRoomEnterRoom(room, c)
	}
}

func AllowEnterSelfBuildRoomAfterBegin(room *roomPkg.Room) bool {
	if room.GetRoomType() == commonProto.RoomType_roomTypePublic {
		return true
	}

	roomRedis := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuNiuRoomEnterAfterBegin)
	result, _ := roomRedis.Get(strconv.Itoa(int(room.GetRoomId())))
	if strings.Compare(result, "1") != 0 && room.GetRoomBeginStatus() == true {
		// 不允许中途加入的自建房
		return false
	}

	return true
}
