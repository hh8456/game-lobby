package logic

import (
	"servers/base-library/base_net"
	"servers/common-library/connData"
	"servers/common-library/function"
	"servers/common-library/log"
	"servers/common-library/proto/errorCodeProto"
	"servers/common-library/proto/lobbyProto"
	"servers/common-library/proto/msgIdProto"
	"servers/common-library/redisKeyPrefix"
	"servers/common-library/redisOpt"
	"servers/lobby-server/player"
	"servers/model"
	"strconv"
	"time"

	"github.com/hh8456/go-common/redisObj"
	"github.com/jinzhu/gorm"
)

/*
流程
1. 玩家先在 redis 中查询 string: superior_uid:uid - 代理 id
2. 如果找到了, 就提示客户端已经绑定过, 退出
3. 如果没找到, 就在 agent 表中,通过  invite_code 查询出代理 id => agentId
4. insert into subordinate(uid, subordinate_uid, establish_contact_date) values(agentId, player.GetUid(), time.Now())
5. 写 redis: superior_uid:player.GetUid() - agentId
*/

func c2sInputAnotherInviteCode(player *player.Player, connData *connData.ConnData) {
	dp := base_net.NewDataPack()
	connId := dp.UnpackClientConnId(connData.BinData)
	pb := &lobbyProto.C2SInputAnotherInviteCode{}

	if function.ProtoUnmarshal(connData.BinData[dp.GetHeadLen():],
		pb, "lobbyProto.C2SInputAnotherInviteCode") {
		uid := player.GetUid()
		strUid := strconv.Itoa(int(uid))

		// 玩家先在 redis 中查询 string: superior_uid:uid - 代理 id
		rdsSuperior := redisObj.NewSessionWithPrefix(redisKeyPrefix.SuperiorUid)
		if false == rdsSuperior.Exists(strUid) {
			var err error
			modAgent := &model.Agent{}
			player.DB.Transaction(func(tx *gorm.DB) error {
				err = tx.Where("invite_code = ?", pb.InviteCode).First(modAgent).Error
				if err != nil {
					if err == gorm.ErrRecordNotFound {
						// 邀请码无效
						connData.SendErrorCode(connId, errorCodeProto.ErrorCode_invalid_invite_code)
						return err
					} else {
						log.Errorf("用户输入邀请码时, 查询 mysql 表 agent 出现错误: %v\n", err)
						connData.SendErrorCode(connId, errorCodeProto.ErrorCode_mysql_query_agent_has_error_when_player_input_invite_code)
						return err
					}
				}

				if modAgent.UId != 0 {
					subordinate := model.Subordinate{}
					subordinate.UId = modAgent.UId
					// player 成为 modAgent.UId 的下级
					subordinate.SubordinateUId = int(uid)
					subordinate.EstablishContactDate = time.Now()
					err = tx.Create(subordinate).Error
					if err != nil {
						connData.SendErrorCode(connId, errorCodeProto.ErrorCode_mysql_insert_subordinate_has_error_when_player_input_invite_code)
						return err
					}

					// 自动成为代理
					subAgent := &model.Agent{}
					subAgent.UId = int(uid)
					subAgent.IsAgent = 1
					strInviteCode, _ := redisOpt.GetOnePlayerInviteCode(uid)
					subAgent.InviteCode, _ = strconv.Atoi(strInviteCode)
					err = tx.Create(subAgent).Error
					if err != nil {
						connData.SendErrorCode(connId, errorCodeProto.ErrorCode_mysql_insert_agent_has_error_when_player_input_invite_code)
						return err
					}

				} else {
					// 查无此人
					connData.SendErrorCode(connId, errorCodeProto.ErrorCode_invalid_invite_code)
				}

				return nil
			})

			if err == nil && modAgent.UId != 0 {
				reply, e := rdsSuperior.Setnx(strUid, modAgent.UId)
				if e != nil {
					log.Errorf("用户 uid: %d 输入邀请码, 执行 redis setnx 时, 出现错误, error: %v", uid, e)
					connData.SendErrorCode(connId, errorCodeProto.ErrorCode_redis_setnx_has_error_when_player_input_invite_code)
				} else {
					if reply == 1 {
						// 设置成功
						replyMsg := &lobbyProto.S2CInputAnotherInviteCode{}
						replyMsg.SuperiorUid = uint32(modAgent.UId)
						connData.SendPbMsg(connId, msgIdProto.MsgId_s2cInputAnotherInviteCode, replyMsg)
					} else {
						log.Errorf("用户 uid: %d 输入邀请码, 成功写入数据库 subordinate, "+
							"但执行 redis setnx 时, 获取锁失败; 一般不会执行到这里来", uid)
						// 触发 setnx
						connData.SendErrorCode(connId, errorCodeProto.ErrorCode_redis_setnx_fail_when_player_input_invite_code)
					}
				}
			}
		} else {
			// 如果找到了, 就提示客户端已经绑定过, 退出
			connData.SendErrorCode(connId, errorCodeProto.ErrorCode_can_not_repeat_input_invite_code)
		}
	}
}
