package logic

import (
	"servers/base-library/base_net"
	"servers/common-library/connData"
	"servers/common-library/function"
	"servers/common-library/log"
	"servers/common-library/proto/commonProto"
	"servers/common-library/proto/lobbyProto"
	"servers/common-library/proto/msgIdProto"
	"servers/common-library/proto/niuniuProto"
	"servers/common-library/redisKeyPrefix"
	"servers/common-library/redisOpt"
	"servers/lobby-server/player"
	"strconv"

	"github.com/hh8456/go-common/redisObj"
)

func c2sNiuNiuGetAllPublicRoom(player *player.Player, connData *connData.ConnData) {
	dp := base_net.NewDataPack()
	connId := dp.UnpackClientConnId(connData.BinData)
	rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuRoom)
	// TODO niuniu-server - roomPkg.CreateRoom 上现在是建立编号 1-10 的房间
	// 后期需要把系统房间号设置为统一配置,供 niuniu-server 和 lobby-server 读取
	// 并在此处使用
	keys := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	roomInfos, e := rds.MGet(keys)
	if e != nil {
		log.Errorf("玩家在大厅查询 redis 房间信息时错误, redis 发生错误: %v", e)
		return
	}

	replyMsg := &niuniuProto.S2CNiuNiuGetAllRoom{}
	for _, binStr := range roomInfos {
		roomInfo := &niuniuProto.RoomInfo{}
		if function.ProtoUnmarshal([]byte(binStr), roomInfo, "niuniuProto.RoomInfo") {
			replyMsg.RoomInfos = append(replyMsg.RoomInfos, roomInfo)
		}
	}

	connData.SendPbMsg(connId, msgIdProto.MsgId_s2cNiuNiuGetAllPublicRoom, replyMsg)
}

//获取自建房信息
func c2sNiuNiuGetAllPublicSelfBuildRoom(player *player.Player, connData *connData.ConnData) {
	dp := base_net.NewDataPack()
	connId := dp.UnpackClientConnId(connData.BinData)
	rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuRoom)

	idListRds := redisObj.NewSessionWithPrefix(redisKeyPrefix.SelfBuildNiuNiuRoom)
	result, err := idListRds.GetSetMembers(redisKeyPrefix.IdSet)
	if err != nil {
		log.Errorf("玩家在大厅查询 redis 自建房间列表信息时错误, redis 发生错误: %v", err)
		return
	}

	var keys []string
	for _, v := range result {
		keys = append(keys, v.(string))
	}

	roomInfos, e := rds.MGet(keys)
	if e != nil {
		log.Errorf("玩家在大厅查询 redis 自建房信息时错误, redis 发生错误: %v", e)
		return
	}

	replyMsg := &niuniuProto.S2CNiuNiuGetAllRoom{}
	for _, binStr := range roomInfos {
		roomInfo := &niuniuProto.RoomInfo{}
		if function.ProtoUnmarshal([]byte(binStr), roomInfo, "niuniuProto.RoomInfo") &&
			roomInfo.RoomType == commonProto.RoomType_selfBuildRoomTypePublic {
			replyMsg.RoomInfos = append(replyMsg.RoomInfos, roomInfo)
		}
	}

	connData.SendPbMsg(connId, msgIdProto.MsgId_s2cNiuNiuGetAllPublicSelfBuildRoom, replyMsg)
}

// XXX 这里要通知 gate, gate 通知 niuniu服务器, 同时在 gate client 上绑定 niuniu 服 id
// 进入牛牛房间
func c2sEnterNiuniuRoom(player *player.Player, connData *connData.ConnData) {
	dp := base_net.NewDataPack()
	connId := dp.UnpackClientConnId(connData.BinData)
	pb := &niuniuProto.C2SEnterNiuniuRoom{}

	if function.ProtoUnmarshal(connData.BinData[dp.GetHeadLen():],
		pb, "niuniuProto.C2SEnterNiuniuRoom") {
		pb.Uid = player.GetUid()
		pb.Wxid = player.GetWxid()
		log.Debugf("收到客户端 uid: %d 进入牛牛房间 roomId: %d 的请求", pb.Uid, pb.RoomId)

		binData, _ := function.ProtoMarshal(pb, "niuniuProto.C2SEnterNiuniuRoom")
		// XXX NiuniuServerId: 401 应该改为从配置中获取
		buf := make([]byte, int(dp.GetHeadLen())+len(binData))
		dp.SetClientConnId(buf, connId)
		dp.SetMsgId(buf, uint32(msgIdProto.MsgId_c2sEnterNiuniuRoom))
		dp.SetMsgLen(buf, uint32(len(binData)))
		copy(buf[dp.GetHeadLen():], binData)
		toGateMsg := &niuniuProto.Lobby2GateEnterNiuniuRoom{C2SEnterNiuniuRoom: buf,
			NiuniuServerId: 401}
		log.Debugf("收到客户端进入牛牛房间的请求, lobby 通知 gate 告之 niuniu server")
		connData.SendPbMsg(connId, msgIdProto.MsgId_lobby2GateEnterNiuniuRoom, toGateMsg)
	}
}

func c2sNiuniuRoomConfig(player *player.Player, connData *connData.ConnData) {
	dp := base_net.NewDataPack()
	connId := dp.UnpackClientConnId(connData.BinData)
	pb := &niuniuProto.C2SNiuniuRoomConfig{}
	if function.ProtoUnmarshal(connData.BinData[dp.GetHeadLen():],
		pb, "niuniuProto.C2SNiuniuRoomConfig") {
		rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuRoomConfig)
		strIds := []string{}
		for _, roomId := range pb.RoomIds {
			strIds = append(strIds, strconv.Itoa(int(roomId)))
		}
		replyMsg := &niuniuProto.S2CNiuniuRoomConfig{}
		strCfgList, e := rds.MGet(strIds)
		if e == nil {
			for _, strCfg := range strCfgList {
				cfg := &niuniuProto.RoomConfig{}
				if function.ProtoUnmarshal([]byte(strCfg), cfg, "niuniuProto.RoomConfig") {
					replyMsg.Cfgs = append(replyMsg.Cfgs, cfg)
				}
			}
		}

		connData.SendPbMsg(connId, msgIdProto.MsgId_s2cNiuniuRoomConfig, replyMsg)
	}
}

func c2sGetNiuniuPlayerBriefInfoOnSeat(player *player.Player, connData *connData.ConnData) {
	dp := base_net.NewDataPack()
	connId := dp.UnpackClientConnId(connData.BinData)
	pb := &lobbyProto.C2SGetNiuniuPlayerBriefInfoOnSeat{}
	if function.ProtoUnmarshal(connData.BinData[dp.GetHeadLen():],
		pb, "lobbyProto.C2SGetNiuniuPlayerBriefInfoOnSeat") {
		replyMsg := &lobbyProto.S2CGetNiuniuPlayerBriefInfoOnSeat{}
		replyMsg.MapPlayerRoomId = map[uint32]uint32{}
		replyMsg.MapPlayerSeatIdx = map[uint32]uint32{}

		for _, roomId := range pb.RoomIds {
			uidsInSeat := redisOpt.GetNiuniuPlayerRoomIdAndSeatIdx(roomId)
			log.Debugf("客户端在大厅查询到房间: %d, 座位信息: %v", roomId, uidsInSeat)
			for seatIdx, strUid := range uidsInSeat {
				if len(strUid) > 0 {
					uid, _ := strconv.Atoi(strUid)
					replyMsg.MapPlayerRoomId[uint32(uid)] = roomId
					replyMsg.MapPlayerSeatIdx[uint32(uid)] = uint32(seatIdx)

					playerBriefInfo, _ := redisOpt.LoadPlayerBriefInfo(strUid)
					if playerBriefInfo != nil {
						//log.Debugf("客户端在大厅查询到房间: %d ,玩家: %d 的简略信息",
						//roomId, uid)
						replyMsg.PlayerBriefInfos = append(replyMsg.PlayerBriefInfos, playerBriefInfo)
					} else {
						log.Errorf("客户端在大厅  未  查询到房间: %d ,玩家: %d 的简略信息",
							roomId, uid)
					}
				}
			}
		}

		replyMsg.RoomIds = pb.RoomIds
		connData.SendPbMsg(connId, msgIdProto.MsgId_s2cGetNiuniuPlayerBriefInfoOnSeat, replyMsg)
	}
}

// 自建牛牛房间
func c2sSelfBuildNiuNiuRoom(player *player.Player, connData *connData.ConnData) {
	log.Debugf("收到客户端进入牛牛房间的请求")
	dp := base_net.NewDataPack()
	connId := dp.UnpackClientConnId(connData.BinData)
	pb := &niuniuProto.C2SSelfBuildNiuNiuRoom{}

	if function.ProtoUnmarshal(connData.BinData[dp.GetHeadLen():],
		pb, "niuniuProto.C2SSelfBuildNiuNiuRoom") {
		pb.Uid = player.GetUid()
		pb.Wxid = player.GetWxid()

		binData, _ := function.ProtoMarshal(pb, "niuniuProto.C2SEnterNiuniuRoom")
		log.Println("lobby 玩家id:", player.GetUid())

		// XXX NiuniuServerId: 401 应该改为从配置中获取
		buf := make([]byte, int(dp.GetHeadLen())+len(binData))
		dp.SetClientConnId(buf, connId)
		dp.SetMsgId(buf, uint32(msgIdProto.MsgId_c2sCreateNiuniuRoom))
		dp.SetMsgLen(buf, uint32(len(binData)))
		copy(buf[dp.GetHeadLen():], binData)
		toGateMsg := &niuniuProto.Lobby2GateSelfBuildNiuNiuRoom{C2SSelfBuildNiuNiuRoom: buf,
			NiuniuServerId: 401}
		log.Debugf("收到客户端进入牛牛房间的请求, lobby 通知 gate 告之 niuniu server")
		connData.SendPbMsg(connId, msgIdProto.MsgId_lobby2GateCreateNiuNiuRoom, toGateMsg)
	}
}
