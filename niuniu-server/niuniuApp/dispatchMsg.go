package niuniuApp

import (
	"servers/base-library/base_net"
	"servers/common-library/connData"
	"servers/common-library/function"
	"servers/common-library/gameKeyPrefix"
	"servers/common-library/log"
	"servers/common-library/proto/errorCodeProto"
	"servers/common-library/proto/msgIdProto"
	"servers/common-library/proto/niuniuProto"
	"servers/niuniu-server/playerPkg"
	"time"
)

// 把 gate 发来的数据分发给 player 对象
// 先在 player 中路由消息( player.Handle ),
// 如果没有路由到,就路由给 player 所在的 room 对象( 详见 player.hand  )
func (n *niuniuApp) dispathMsgToPlayer(connData *connData.ConnData) {
	dp := &base_net.DataPack{}
	connId := dp.UnpackClientConnId(connData.BinData)
	msgId := dp.UnpackMsgId(connData.BinData)
	switch msgIdProto.MsgId(msgId) {

	// 进入牛牛房间时创建 player 对象
	case msgIdProto.MsgId_c2sEnterNiuniuRoom:
		c2sPb := &niuniuProto.C2SEnterNiuniuRoom{}
		if function.ProtoUnmarshal(connData.BinData[dp.GetHeadLen():],
			c2sPb, "niuniuProto.C2SEnterNiuniuRoom") {
			gateId := uint32(0)
			v := connData.GetProperty(gameKeyPrefix.GateId)
			if v != nil {
				if value, ok := v.(uint32); ok {
					gateId = value
				}
			}

			oldPlayer := n.FindPlayer(c2sPb.Wxid)
			if oldPlayer != nil {
				oldConnId := oldPlayer.GetConnId()
				if oldConnId != connId {
					log.Debugf("发现顶号, oldConnId: %d, newConnId: %d, roomId: %d",
						oldConnId, connId, c2sPb.RoomId)
					connData.SendPbMsg(oldConnId, msgIdProto.MsgId_s2cKick, nil)
					oldPlayer.SetConnId(connId)
					oldPlayer.SetGateId(gateId)
					n.StorePlayer(c2sPb.Wxid, oldPlayer, connId, oldConnId)
				} else {
					log.Debugf("执行到这里,表示用户进入房间没坐下,然后退出,再进. "+
						"uid: %d, connId: %d, roomId: %d",
						oldPlayer.(*playerPkg.Player).GetUid(), connId, c2sPb.RoomId)
				}

				oldPlayer.(*playerPkg.Player).Online()
				oldPlayer.(*playerPkg.Player).Handle(connData)

			} else {
				log.Debugf("玩家进入房间, newPlayer, uid: %d,connId: %d, roomId: %d", c2sPb.Uid, connId, c2sPb.RoomId)
				player := playerPkg.NewPlayer(connId, gateId, c2sPb.Uid, c2sPb.Wxid, n.GetGate)
				n.StorePlayer(c2sPb.Wxid, player, connId, 0)
				player.Run()
				player.Handle(connData)
			}
		}

	// 离开牛牛房间时删除对象
	case msgIdProto.MsgId_c2sLeaveNiuniuRoom:
		log.Debugf("niuniuApp.dispathMsgToPlayer 收到玩家离开房间的请求,"+
			" msgIdProto.MsgId_c2sLeaveNiuniuRoom, connId: %d", connId)
		iPlayer := n.FindPlayerByConnId(connId)
		if iPlayer != nil {
			player := iPlayer.(*playerPkg.Player)
			player.Handle(connData)
		} else {
			log.Errorf("niuniuApp.dispathMsgToPlayer 收到玩家离开房间的请求,"+
				" msgIdProto.MsgId_c2sLeaveNiuniuRoom, 但没找到 player 对象,"+
				" connId: %d", connId)
		}

		// gate 感知到客户端断开连接
	case msgIdProto.MsgId_gate2NiuniuClientDisconnect:
		iPlayer := n.FindPlayerByConnId(connId)
		if iPlayer != nil {
			player := iPlayer.(*playerPkg.Player)
			player.Handle(connData)
		}

		// 心跳保活
	case msgIdProto.MsgId_c2sNiuniuPing:
		iPlayer := n.FindPlayerByConnId(connId)
		if iPlayer != nil {
			c2sPb := &niuniuProto.C2SNiuniuPing{}
			if function.ProtoUnmarshal(connData.BinData[dp.GetHeadLen():],
				c2sPb, "niuniuProto.C2SNiuniuPing") {
				replyMsg := &niuniuProto.S2CNiuniuPong{Timestamp: c2sPb.Timestamp}
				connData.SendPbMsg(connId, msgIdProto.MsgId_s2cNiuniuPong, replyMsg)
				iPlayer.SetLastAliveTimestamp(time.Now().Unix())
			}
		}

	case msgIdProto.MsgId_c2sCreateNiuniuRoom:
		log.Debugf("开始自建房操作")
		c2sPb := &niuniuProto.C2SSelfBuildNiuNiuRoom{}
		if function.ProtoUnmarshal(connData.BinData[dp.GetHeadLen():],
			c2sPb, "niuniuProto.C2SSelfBuildNiuNiuRoom") {
			gateId := uint32(0)
			v := connData.GetProperty(gameKeyPrefix.GateId)
			if v != nil {
				if value, ok := v.(uint32); ok {
					gateId = value
				}
			}

			player := playerPkg.NewPlayer(connId, gateId, c2sPb.Uid, c2sPb.Wxid, n.GetGate)
			n.StorePlayer(c2sPb.Wxid, player, connId, 0)
			player.Run()
			player.Handle(connData)
		}

		// 其他消息,直接投递给 room 进行处理
	default:
		iPlayer := n.FindPlayerByConnId(connId)
		if iPlayer != nil {
			player := iPlayer.(*playerPkg.Player)
			room := player.GetRoom()
			if room != nil {
				room.Handle(player, connData)
			} else {
				player.SendErrorCode(errorCodeProto.ErrorCode_has_not_in_niuniu_room_when_play)

			}
		}

	}
}
