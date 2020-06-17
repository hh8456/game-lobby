package redisKeyPrefix

const (
	DigitalId                 = "digital_id"                    // 数字 id
	Login                     = "login"                         // 登录
	NiuniuRoom                = "niuniu_room"                   //创建房间, string: roomId - binData
	SelfBuildNiuNiuRoomMsg    = "self_build_niuniu_room_msg"    //牛牛自建房信息
	SelfBuildNiuNiuRoom       = "self_build_niuniu_room"        //单个牛牛自建房id
	IdSet                     = "id_set"                        //id集合
	NiuniuRoomConfig          = "niuniu_room_config"            //房间配置, string: roomId - binData
	NiuNiuRoomEnterAfterBegin = "niuniu_room_enter_after_begin" //是否允许中途进入房间
	PlayerBaseInfo            = "player_base_info"              // 玩家个人信息 string: uid - binData
	WxidAndUId                = "wxid"                          // string: wxid - uid
	NiuniuPlayerRoomId        = "niuniu_player_room_id"         // string: uid - roomid
	SuperiorUid               = "superior_uid"                  // string: uid - superiorUid 上级 id
	NiuniuSeatIndex           = "niuniu_seat_index"
	NiuniuPlayerInRoom        = "niuniu_player_in_room" //创建房间, string: roomId - binData
)

// 对应 model.PlayerBaseInfo 的字段
const (
	PlayerBaseInfo_Wxid       = "Wxid"
	PlayerBaseInfo_WxidCrc32  = "wxidCrc32"
	PlayerBaseInfo_UId        = "UId"
	PlayerBaseInfo_HeadPic    = "HeadPic"
	PlayerBaseInfo_InviteCode = "InviteCode"
	PlayerBaseInfo_Diamond    = "Diamond"
	PlayerBaseInfo_Gold       = "Gold"
	PlayerBaseInfo_Sex        = "Sex"
	PlayerBaseInfo_Name       = "Name"
	PlayerBaseInfo_NameCrc32  = "NameCrc32"
	PlayerBaseInfo_RegDate    = "RegDate"
)
