package logic

import (
	"servers/common-library/proto/msgIdProto"
	"servers/lobby-server/player"
)

func Init() {
	player.AppendFunc(msgIdProto.MsgId_c2sPing, c2sPing)
	player.AppendFunc(msgIdProto.MsgId_c2sEnterNiuniuRoom, c2sEnterNiuniuRoom)
	player.AppendFunc(msgIdProto.MsgId_s2cSelfBuildRoom, c2sSelfBuildNiuNiuRoom)
	player.AppendFunc(msgIdProto.MsgId_c2sNiuNiuGetAllPublicRoom, c2sNiuNiuGetAllPublicRoom)
	player.AppendFunc(msgIdProto.MsgId_c2sNiuNiuGetAllPublicSelfBuildRoom, c2sNiuNiuGetAllPublicSelfBuildRoom)
	player.AppendFunc(msgIdProto.MsgId_c2sNiuniuRoomConfig, c2sNiuniuRoomConfig)
	player.AppendFunc(msgIdProto.MsgId_c2sGetPlayerGold, c2sGetPlayerGold)
	player.AppendFunc(msgIdProto.MsgId_c2sGetNiuniuPlayerBriefInfoOnSeat, c2sGetNiuniuPlayerBriefInfoOnSeat)
	player.AppendFunc(msgIdProto.MsgId_c2sInputAnotherInviteCode, c2sGetNiuniuPlayerBriefInfoOnSeat)
}
