package iface

import (
	"servers/common-library/proto/commonProto"
)

type IRoom interface {
	Run()
	Close()
	GetRoomId() uint32
	GetPlayerNum() uint32    // 获取参与游戏的玩家数量
	GetBystanderNum() uint32 // 获取旁观者数量
	GetRoomType() commonProto.RoomType
	//HandleMsg(IClient, []byte)
}
