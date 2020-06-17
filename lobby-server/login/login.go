package login

import (
	"hash/crc32"
	"math/rand"
	"servers/base-library/base_net"
	"servers/common-library/connData"
	"servers/common-library/digitalId"
	"servers/common-library/function"
	"servers/common-library/gameKeyPrefix"
	"servers/common-library/log"
	"servers/common-library/proto/errorCodeProto"
	"servers/common-library/proto/lobbyProto"
	"servers/common-library/proto/msgIdProto"
	"servers/common-library/redisKeyPrefix"
	"servers/common-library/redisOpt"
	"servers/iface"
	"servers/lobby-server/player"
	"servers/model"
	"strconv"
	"time"

	"github.com/hh8456/go-common/redisObj"
	"github.com/hh8456/redisSession"
	"github.com/jinzhu/gorm"
	"github.com/lifei6671/gorand"
)

type Login struct {
	chanConnData chan *connData.ConnData
	db           *gorm.DB
	storePlayer  func(wxid string, player iface.IPlayer, newConnId, oldConnId int64)
	findPlayer   func(wxid string) iface.IPlayer
}

func New(db *gorm.DB,
	storePlayer func(wxid string, player iface.IPlayer, newConnId, oldConnId int64),
	findPlayer func(wxid string) iface.IPlayer) *Login {
	return &Login{
		chanConnData: make(chan *connData.ConnData, 20000), // 两万并发登录
		db:           db,
		storePlayer:  storePlayer,
		findPlayer:   findPlayer,
	}
}

func (l *Login) Run() {
	// 10 个协程并发处理登录账号
	for i := 0; i < 10; i++ {
		go func() {
			for {
				connData := <-l.chanConnData
				l.login(connData)
			}
		}()
	}
}

func (l *Login) login(connData *connData.ConnData) {
	gateId := uint32(0)
	v := connData.GetProperty(gameKeyPrefix.GateId)
	if v != nil {
		if value, ok := v.(uint32); ok {
			gateId = value
		}
	}

	dp := &base_net.DataPack{}
	c2sPb := &lobbyProto.C2SWxLogin{}
	connId := dp.UnpackClientConnId(connData.BinData)
	if function.ProtoUnmarshal(connData.BinData[dp.GetHeadLen():], c2sPb, "lobbyProto.C2SWxLogin") {
		if len(c2sPb.Wxid) == 0 {
			log.Debugf("收到客户端 connId: %d 发来的登录消息, lobbyProto.C2SWxLogin.Wxid 字段为空, 登录失败", connId)
			return
		}

		log.Debugf("收到客户端 connId: %d, wxid: %s 发来的登录消息", connId, c2sPb.Wxid)
		replyMsg := &lobbyProto.S2CWxLogin{}
		oldPlayer := l.findPlayer(c2sPb.Wxid)
		if oldPlayer != nil && oldPlayer.GetConnId() != connId {
			oldConnId := oldPlayer.GetConnId()
			connData.SendPbBuf(oldConnId, msgIdProto.MsgId_s2cKick, nil)
			oldPlayer.SetConnId(connId)
			oldPlayer.SetGateId(gateId)
			l.storePlayer(c2sPb.Wxid, oldPlayer, connId, oldConnId)
			replyMsg.PlayerBaseInfo = oldPlayer.(*player.Player).GetPlayerBaseInfo()
			log.Debugf("在内存中找到了 player 缓存对象 uid: %d, 新的 player 对象 connid: %d",
				replyMsg.PlayerBaseInfo.Uid, oldPlayer.GetConnId())
			connData.SendPbMsg(connId, msgIdProto.MsgId_s2cWxLogin, replyMsg)
			log.Debugf("客户端 connId: %d, wxid: %s, uid: %d 登录成功",
				connId, c2sPb.Wxid, oldPlayer.(*player.Player).GetUid())
			return
		}

		// 用分布式锁来保护登录流程
		rdsLogin := redisObj.NewSessionWithPrefix(redisKeyPrefix.Login)
		reply, e := rdsLogin.SetExNx(c2sPb.Wxid, 0, 10*time.Second)
		defer rdsLogin.Del(c2sPb.Wxid)
		if e != nil {
			log.Errorf("创建账号时 redis 锁发生错误: %v", e)
			connData.SendErrorCode(connId, errorCodeProto.ErrorCode_redis_setnx_has_error_when_login)
			return
		}

		if reply == "NX" {
			log.Errorf("登录时 redis 锁检测到正在有其他人登录")
			connData.SendErrorCode(connId, errorCodeProto.ErrorCode_other_already_login)
			return
		}

		strUid, e := redisOpt.GetUidByWxid(c2sPb.Wxid)

		if e != nil {
			if e == redisSession.ErrNil {
				playerBaseInfo, b := l.initNewAccount(c2sPb.Wxid)
				if b {
					// 新建账号
					l.createAccount(playerBaseInfo, connData)
				} else {
					connData.SendErrorCode(connId, errorCodeProto.ErrorCode_lobby_init_new_account_fail)
				}
			} else {
				log.Errorf("登录时从 redis 中通过 wxid 查询 uid 出错: %v", e)
				connData.SendErrorCode(connId, errorCodeProto.ErrorCode_redis_get_uid_by_wxid_error_when_login)
			}

			return
		}

		playerBaseInfo, b := redisOpt.LoadPlayerBaseInfo(strUid)
		if b {
			log.Debugf("客户端 connId: %d, wxid: %s, uid: %d 登录成功",
				connId, c2sPb.Wxid, playerBaseInfo.UId)
			player := player.NewPlayer(connId, gateId, uint32(playerBaseInfo.UId),
				c2sPb.Wxid, l.db)
			l.storePlayer(c2sPb.Wxid, player, connId, 0)
			player.Run()

			replyMsg.PlayerBaseInfo = player.GetPlayerBaseInfo()
			connData.SendPbMsg(connId, msgIdProto.MsgId_s2cWxLogin, replyMsg)
		} else {
			connData.SendErrorCode(connId, errorCodeProto.ErrorCode_redis_get_player_base_info_by_uid_error_when_login)
		}

	}
}

func (l *Login) Login(connData *connData.ConnData) {
	select {
	case l.chanConnData <- connData:

	default:
	}
}

func (l *Login) createAccount(playerBaseInfo *model.PlayerBaseInfo, connData *connData.ConnData) {
	b := false

	var err error
	l.db.Transaction(func(tx *gorm.DB) error {
		err = tx.Exec("insert into uid_in_use(uid) values (?)", playerBaseInfo.UId).Error
		if err != nil {
			return err
		}

		err = tx.Exec("insert into invite_code_in_use(invite_code) values (?)", playerBaseInfo.InviteCode).Error
		if err != nil {
			return err
		}

		err = tx.Create(playerBaseInfo).Error
		if err != nil {
			return err
		}

		b = true
		return nil
	})

	dp := &base_net.DataPack{}
	connId := dp.UnpackClientConnId(connData.BinData)
	if b {
		// 注册成功, 写 redis 并生成 player
		redisOpt.SavePlayerBaseInfo(playerBaseInfo)

		redisOpt.SaveWxidAndUid(playerBaseInfo.Wxid, playerBaseInfo.UId)

		gateId := uint32(0)
		v := connData.GetProperty(gameKeyPrefix.GateId)
		if v != nil {
			if value, ok := v.(uint32); ok {
				gateId = value
			}
		}

		player := player.NewPlayer(connId, gateId, uint32(playerBaseInfo.UId),
			playerBaseInfo.Wxid, l.db)
		l.storePlayer(playerBaseInfo.Wxid, player, connId, 0)
		player.Run()

		replyMsg := &lobbyProto.S2CWxLogin{}
		replyMsg.PlayerBaseInfo = player.GetPlayerBaseInfo()
		connData.SendPbMsg(connId, msgIdProto.MsgId_s2cWxLogin, replyMsg)

		log.Debugf("客户端 connId: %d, wxid: %s, uid: %d 登录成功并新建了账号",
			connId, playerBaseInfo.Wxid, playerBaseInfo.UId)
	} else {
		log.Errorf("创建账号时发生错误: %v", err)
		connData.SendErrorCode(connId, errorCodeProto.ErrorCode_mysql_has_error_when_create_account)
	}
}

func (l *Login) initNewAccount(wxid string) (*model.PlayerBaseInfo, bool) {
	playerBaseInfo := &model.PlayerBaseInfo{}
	playerBaseInfo.Wxid = wxid
	strUid := digitalId.Get()

	if strUid == "" {
		return nil, false
	}

	strInviteCode := digitalId.Get()
	if strInviteCode == "" {
		return nil, false
	}

	uid, e1 := strconv.ParseInt(strUid, 10, 64)
	if e1 != nil {
		return nil, false
	}

	inviteCode, e2 := strconv.ParseInt(strInviteCode, 10, 64)
	if e2 != nil {
		return nil, false
	}

	playerBaseInfo.UId = int(uid)
	rand.Seed(time.Now().UnixNano())
	// 头像范围: [1,12]
	HeadPicId := 1 + rand.Intn(12)
	playerBaseInfo.HeadPic = strconv.Itoa(HeadPicId)
	playerBaseInfo.Diamond = 100000
	playerBaseInfo.Gold = 100000
	playerBaseInfo.WxidCrc32 = int64(crc32.ChecksumIEEE([]byte(wxid)))
	playerBaseInfo.InviteCode = int(inviteCode)
	name := gorand.KRand(6, gorand.KC_RAND_KIND_LOWER)
	playerBaseInfo.Name = string(name)
	playerBaseInfo.RegDate = time.Now()
	return playerBaseInfo, true
}
