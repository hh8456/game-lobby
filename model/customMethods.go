package model

import (
	"servers/common-library/proto/commonProto"
)

func (pbi *PlayerBaseInfo) ToPbMsg() *commonProto.PlayerBaseInfo {
	p := &commonProto.PlayerBaseInfo{Wxid: pbi.Wxid,
		Uid:     uint32(pbi.UId),
		HeadPic: pbi.HeadPic,
		// 邀请码
		InviteCode: uint32(pbi.InviteCode),
		// 钻石
		Diamond: int32(pbi.Diamond),
		// 金币
		Gold: int32(pbi.Gold),
		// 0 未填写性别, 1 女, 2 男
		Sex:  int32(pbi.Sex),
		Name: pbi.Name,
	}

	return p
}
