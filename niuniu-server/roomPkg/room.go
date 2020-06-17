package roomPkg

import (
	"github.com/gogo/protobuf/proto"
	"github.com/hh8456/go-common/redisObj"
	"github.com/jinzhu/gorm"
	"servers/base-library/base_net"
	"servers/common-library/connData"
	"servers/common-library/function"
	"servers/common-library/log"
	"servers/common-library/proto/commonProto"
	"servers/common-library/proto/errorCodeProto"
	"servers/common-library/proto/msgIdProto"
	"servers/common-library/proto/niuniuProto"
	"servers/common-library/redisKeyPrefix"
	"servers/common-library/redisOpt"
	"servers/common-library/utility"
	"servers/niuniu-server/niuniuConfig"
	"strconv"
	"sync"
)

// 需要把下面这些方法暴露出来,以便让 room 调用 client
type IClient interface {
	Send([]byte)
	SetSeatIndex(int)
	GetSeatIndex() int32
	GetUid() uint32
	GetUidString() string
	GetPlayerBaseInfo() *commonProto.PlayerBaseInfo
	SendPbMsg(msgId msgIdProto.MsgId, pbMsg proto.Message)
	SendErrorCode(errCode errorCodeProto.ErrorCode)
	AppendHandCard(...*commonProto.PokerCard)
	GetHandCard() []*commonProto.PokerCard
	ReReady()
	UnPrepare()
	Prepare()
	IsPrepare() bool
	GetConnId() int64
	IsOffline() bool
	SetRoomId(roomId uint32, room *Room)
}

type ClientAndMsg struct {
	Client   IClient
	connData *connData.ConnData
}

type roomConfig struct {
	baseScore       uint32            // 底分
	robZhuangRate   uint32            // 庄家抢庄倍数; X 倍抢庄,当庄后输赢都按 X 倍率; 抢庄倍数是 [0, robZhuangRate] 中之一
	betRate         uint32            // 下注倍数, 表示底分的倍数,只有闲家才能下注
	mapCardTypeRate map[uint32]uint32 // 每种牌型对应一种倍数
	addPreBetRate   uint32            // 推注,闲家上盘赢了后, 这次开局时选择了推注,那么就把推注倍数和底分相加,并代入公式计算
}

type roomInsideStatus struct {
	settled              bool // 是否结算了
	broadcastBankerCards bool // 是否广播了庄家手牌
}

type Room struct {
	serverId            uint32 // 如果把房间信息写入 redis/etcd, 那么需要有 server id 来区分
	roomId              uint32
	roomType            commonProto.RoomType
	roomStatus          niuniuProto.NiuniuRoomStatus
	roomStatusTimestamp int64 // 房间切换状态时的时间戳
	chanClientAndMsg    chan *ClientAndMsg
	lock                sync.RWMutex
	seat                []IClient // 座位占用情况, 0 表示空位, > 0 表示 uid
	// 玩家同一时刻必然只在 mapPlayers, mapBystanders, seat 其中之一
	mapBystanders    map[uint32]IClient // 旁观者, uid - IClient, 只要不在座位上都算旁观者
	mapPlayers       map[uint32]IClient // 已经入座,并且正在进行游戏的玩家, uid - IClient
	cardHeap         []*commonProto.PokerCard
	bankerId         uint32 // 庄家 id
	chanTimer        chan int64
	mapRobZhuang     map[uint32]uint32                         // 已经抢庄的人, uid - rate( 倍率, 1, 2, 4, 8, 不抢庄也是 1 )
	mapBet           map[uint32]uint32                         // 已经下注的人, uid - bet
	mapShowCard      map[uint32]*niuniuProto.S2CNiuniuShowCard // 已经亮牌的人, uid - *niuniuProto.S2CNiuniuShowCard
	mapWinOrLose     map[uint32]int32
	cfg              *niuniuProto.RoomConfig
	setRoomPlayerNum func(uint32, uint32)
	status           roomInsideStatus
	db               *gorm.DB
	gameTimes        uint32
	owner            uint32           //房主
	numberOfGame     uint32           //游戏当前局数，有次数限制时使用
	mapTotalWinLose  map[uint32]int32 //自建房玩家输赢
}

var (
	lock             sync.RWMutex
	mapRoom          map[uint32]*Room  // roomId - *Room
	mapRoomPlayerNum map[uint32]uint32 // roomId - 玩家人数
)

func init() {
	mapRoom = map[uint32]*Room{}
	mapRoomPlayerNum = map[uint32]uint32{}
}

func CreateRoom(serverId uint32, roomIds []uint32, db *gorm.DB) bool {
	rdsCrRoom := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuRoom)

	for _, roomId := range roomIds {
		roomInfo := &niuniuProto.RoomInfo{RoomId: roomId,
			RoomType: commonProto.RoomType_roomTypePublic}

		binBuf, b := function.ProtoMarshal(roomInfo, "niuniuProto.RoomInfo")
		if !b {
			return false
		}

		// TODO 第二个参数 0 应该换成填写房间信息
		e := rdsCrRoom.Set(strconv.Itoa(int(roomId)), string(binBuf))
		if e != nil {
			log.Errorf("创建牛牛房间写 redis 时发生错误: %v", e)
			return false
		}
	}

	for _, roomId := range roomIds {
		room := NewRoom(serverId, roomId, 10, commonProto.RoomType_roomTypePublic, db)
		mapRoom[roomId] = room
		mapRoomPlayerNum[roomId] = 0
	}

	for _, room := range mapRoom {
		room.Run()
	}

	return true
}

func StoreRoom(roomId uint32, room *Room) {
	lock.Lock()
	mapRoom[roomId] = room
	lock.Unlock()
}

func DeleteRoom(roomId uint32) {
	lock.Lock()
	delete(mapRoom, roomId)
	lock.Unlock()
}

func SetRoomPlayerNum(roomId uint32, playerNum uint32) {
	lock.Lock()
	mapRoomPlayerNum[roomId] = playerNum
	lock.Unlock()
}

func Timer(timestamp int64) {
	lock.RLock()
	defer lock.RUnlock()
	for _, room := range mapRoom {
		room.timer(timestamp)
	}
}

func NewRoom(serverId, roomId, seatNum uint32,
	roomType commonProto.RoomType, db *gorm.DB) *Room {
	cfg := niuniuConfig.DefaultRoomConfig()
	rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuRoomConfig)
	strCfg, e := rds.Get(strconv.Itoa(int(roomId)))
	if e != nil {
		log.Errorf("牛牛服务器启动并新建系统房时读取 redis 中的房间配置出错: %v, 将使用默认配置继续游戏", e)
	} else {
		tmpCfg := &niuniuProto.RoomConfig{}
		if function.ProtoUnmarshal([]byte(strCfg), tmpCfg, "niuniuProto.RoomConfig") {
			cfg = tmpCfg
		} else {
			log.Errorf("牛牛服务器启动并新建系统房时, 反序列化 redis 中的房间配置出错, 将使用默认配置继续游戏")
		}
	}
	p := &Room{
		serverId:         serverId,
		roomId:           roomId,
		roomType:         roomType,
		roomStatus:       niuniuProto.NiuniuRoomStatus_niuniuRoomStatusReady,
		chanClientAndMsg: make(chan *ClientAndMsg, 100),
		chanTimer:        make(chan int64, 10),
		seat:             make([]IClient, seatNum),
		mapBystanders:    map[uint32]IClient{},
		mapPlayers:       map[uint32]IClient{},
		cardHeap:         utility.GetPokerHeap(),
		mapRobZhuang:     map[uint32]uint32{},
		mapWinOrLose:     map[uint32]int32{}, // uid - 输赢金币
		mapBet:           map[uint32]uint32{},
		mapShowCard:      map[uint32]*niuniuProto.S2CNiuniuShowCard{},
		cfg:              cfg,
		setRoomPlayerNum: SetRoomPlayerNum,
		db:               db,
	}

	return p
}

//用于自建房
func (room *Room) GetRoomBeginStatus() bool {
	if room.numberOfGame == 0 {
		return false
	} else {
		return true
	}
}

func GetRoom(roomId uint32) *Room {
	lock.RLock()
	defer lock.RUnlock()
	if room, find := mapRoom[roomId]; find {
		return room
	}

	return nil
}

func (room *Room) Handle(c IClient, connData *connData.ConnData) {
	select {
	case room.chanClientAndMsg <- &ClientAndMsg{c, connData}:

	default:
		log.Errorf("Room.chanClientAndMsg 的容量满了,接收到的客户端消息溢出了 ")
	}
}

func (room *Room) handle(c IClient, connData *connData.ConnData) {
	dp := &base_net.DataPack{}
	msgId := dp.UnpackMsgId(connData.BinData)

	if f, find := mapRoomLogicFunc[msgIdProto.MsgId(msgId)]; find {
		f(room, c, connData)
	}
}

func (room *Room) GetRoomId() uint32 {
	return room.roomId
}

func (room *Room) timer(timestamp int64) {
	select {
	case room.chanTimer <- timestamp:

	default:
	}
}

// 加入旁观者
func (room *Room) addBystanders(c IClient) {
	room.mapBystanders[c.GetUid()] = c
}

func (room *Room) removePlayer(c IClient) {
	// 玩家同一时刻必然只在 room.mapPlayers/room.seat, room.mapBystanders 其中之一
	uid := c.GetUid()
	seatIdx := c.GetSeatIndex()
	c.SetSeatIndex(-1)
	c.UnPrepare()
	// 从座位中移除
	if seatIdx > -1 && seatIdx < int32(len(room.seat)) {
		room.seat[seatIdx] = nil
	}

	c.SetRoomId(0, nil)

	// 从游戏者集合中移除
	delete(room.mapPlayers, uid)
	// 从旁观者集合中移除
	delete(room.mapBystanders, uid)
}

func (room *Room) isNormalPlayer(uid uint32) bool {
	_, find := room.mapPlayers[uid]
	return find && uid != room.bankerId
}

func (room *Room) isBanker(uid uint32) bool {
	return uid == room.bankerId && uid > 0
}

func (room *Room) GetRoomType() commonProto.RoomType {
	return room.roomType
}

func (room *Room) getBystanderNum() uint32 {
	return uint32(len(room.mapBystanders))
}

func (room *Room) getPrepperId() []uint32 {
	ids := []uint32{}
	for _, c := range room.seat {
		if c != nil && c.IsPrepare() {
			ids = append(ids, c.GetUid())
		}
	}

	return ids
}

func (room *Room) getPrepperNum() uint32 {
	n := 0
	for _, c := range room.seat {
		if c != nil && c.IsPrepare() {
			n++
		}
	}

	return uint32(n)
}

func (room *Room) broadcastErrorCode(errCode errorCodeProto.ErrorCode) {
	s2cErrorCode := &errorCodeProto.S2CErrorCode{}
	s2cErrorCode.Code = int32(errCode)
	room.broadcast(msgIdProto.MsgId_s2cErrorCode, s2cErrorCode)
}

func (room *Room) broadcast(pid msgIdProto.MsgId, pbMsg proto.Message) {
	var (
		pbBuf []byte
		err   error
	)

	if pbMsg == nil {
		pbBuf = []byte{}
	} else {
		pbBuf, err = proto.Marshal(pbMsg)
		if err != nil {
			log.Errorf("proto.Marshal error: %v", err)
			return
		}
	}

	dp := base_net.DataPack{}
	msg := base_net.NewMsgPackage(0, uint32(pid), pbBuf)
	// 这里只对座位上的人和旁观者进行广播
	// 因为座位上的人就包括了游戏者
	for _, c := range room.seat {
		if c != nil {
			msg.SetClientConnId(c.GetConnId())
			buf, err := dp.Pack(msg)
			if err != nil {
				log.Errorf("NiuniuRoom.broadcast base_net.DataPack.Pack error: %v", err)
				return
			}

			c.Send(buf)
		}
	}

	for _, c := range room.mapBystanders {
		msg.SetClientConnId(c.GetConnId())
		buf, err := dp.Pack(msg)
		if err != nil {
			log.Errorf("NiuniuRoom.broadcast base_net.DataPack.Pack error: %v", err)
			return
		}

		c.Send(buf)
	}
}

func (room *Room) sendToOthers(uid uint32, pid msgIdProto.MsgId, pbMsg proto.Message) {

	var (
		pbBuf []byte
		err   error
	)

	if pbMsg == nil {
		pbBuf = []byte{}
	} else {
		pbBuf, err = proto.Marshal(pbMsg)
		if err != nil {
			log.Errorf("proto.Marshal error: %v", err)
			return
		}
	}

	// 第一个参数应该输入连接号 id
	dp := base_net.DataPack{}
	msg := base_net.NewMsgPackage(0, uint32(pid), pbBuf)

	// 这里只对座位上的人和旁观者进行广播
	// 因为座位上的人就包括了游戏者
	for _, c := range room.seat {
		if c != nil && c.GetUid() != uid {
			msg.SetClientConnId(c.GetConnId())
			buf, err := dp.Pack(msg)
			if err != nil {
				log.Errorf("NiuniuRoom.sendToOthers base_net.DataPack.Pack error: %v", err)
				return
			}

			c.Send(buf)
		}
	}

	for _, c := range room.mapBystanders {
		if c.GetUid() != uid {
			msg.SetClientConnId(c.GetConnId())
			buf, err := dp.Pack(msg)
			if err != nil {
				log.Errorf("NiuniuRoom.sendToOthers base_net.DataPack.Pack error: %v", err)
				return
			}

			c.Send(buf)

		}
	}
}

func (room *Room) broadcastRoomStatus() {
	room.broadcast(msgIdProto.MsgId_s2cNiuniuBroadcastRoomStatus,
		&niuniuProto.S2CNiuniuBroadcastRoomStatus{RoomStatus: room.roomStatus,
			CountDown: uint32(room.cfg.MapWaitTime[uint32(room.roomStatus)])})
}

// 检测下注情况,代替没下注的闲家下注
func (room *Room) checkBet() {
	for uid, _ := range room.mapPlayers {
		if _, find := room.mapBet[uid]; !find && uid != room.bankerId {
			room.mapBet[uid] = niuniuConfig.Bet1
			replyPbMsg := &niuniuProto.S2CNiuniuBet{}
			replyPbMsg.Bet = niuniuConfig.Bet1
			replyPbMsg.Uid = uid
			room.broadcast(msgIdProto.MsgId_s2cNiuniuBet, replyPbMsg)

		}
	}
}

// 向房间内所有玩家发4张明牌
func (room *Room) sendKnownCard(cnt uint32) {
	for _, c := range room.mapPlayers {
		if c != nil && len(room.cardHeap) >= 5 {
			c.AppendHandCard(room.cardHeap[:cnt]...)
			c.SendPbMsg(msgIdProto.MsgId_s2cNiuniuSendKnownCard,
				&niuniuProto.S2CNiuniuSendKnownCard{Cards: room.cardHeap[:cnt]})

			room.cardHeap = room.cardHeap[cnt:]
		}
	}
}

func (room *Room) sendOneUnknownCard() {
	for _, c := range room.mapPlayers {
		if c != nil && len(room.cardHeap) >= 1 {
			c.AppendHandCard(room.cardHeap[0])
			c.SendPbMsg(msgIdProto.MsgId_s2cNiuniuSendOneUnknownCard,
				&niuniuProto.S2CNiuniuSendOneUnknownCard{Card: room.cardHeap[0]})

			room.cardHeap = room.cardHeap[1:]
		}
	}

	for _, c := range room.mapBystanders {
		c.SendPbMsg(msgIdProto.MsgId_s2cNiuniuSendOneUnknownCardWatch, nil)
	}

	// 转发给游戏中途入座的玩家
	for _, c := range room.seat {
		if c != nil {
			if _, find := room.mapPlayers[c.GetUid()]; find == false {
				c.SendPbMsg(msgIdProto.MsgId_s2cNiuniuSendOneUnknownCardWatch, nil)
			}
		}
	}
}

func (room *Room) changeRoomStatus(roomStatus niuniuProto.NiuniuRoomStatus, timestamp int64) {
	room.roomStatus = roomStatus
	room.roomStatusTimestamp = timestamp
	room.broadcastRoomStatus()
}

func (room *Room) isAllPlayerShowCard() bool {
	return len(room.mapPlayers) == len(room.mapShowCard)
}

// 游戏结束后,重新准备开始游戏
func (room *Room) reReady(timestamp int64) {
	for _, c := range room.mapPlayers {
		c.ReReady()
	}

	room.bankerId = 0
	room.status.settled = false
	room.status.broadcastBankerCards = false
	room.cardHeap = utility.GetPokerHeap()
	room.mapRobZhuang = map[uint32]uint32{}
	room.mapWinOrLose = map[uint32]int32{}
	room.mapBet = map[uint32]uint32{}
	room.mapShowCard = map[uint32]*niuniuProto.S2CNiuniuShowCard{}

	//加速模式
	if room.cfg.TotalNumberOfGame != room.numberOfGame && room.cfg.Faster {
		for _, v := range room.mapPlayers {
			v.Prepare()
		}
	} else {
		room.mapPlayers = map[uint32]IClient{}
	}

	room.changeRoomStatus(niuniuProto.NiuniuRoomStatus_niuniuRoomStatusReady, timestamp)
}

func (room *Room) haveNext() bool {
	if room.roomStatus == niuniuProto.NiuniuRoomStatus_niuniuRoomStatusSettle {
		if room.cfg.TotalNumberOfGame == room.numberOfGame {
			return false
		} else {
			return true
		}
	}

	return true
}

// 清除离线的客户端
func (room *Room) clearOfflineClient() {
	// 先对座位上的人遍历
	// 因为座位上的人就包括了游戏者和旁观者
	for seatIdx, cli := range room.seat {
		if cli != nil && cli.IsOffline() {
			uid := cli.GetUid()
			strUid := cli.GetUidString()
			p := &niuniuProto.S2CBroadcastLeaveNiuniuRoom{}
			p.RoomId = room.roomId
			p.SeatIndex = int32(seatIdx)
			room.removePlayer(cli)
			// 必须调用完 room.RemovePlayer 以后,才能调用 room.getBystanderNum
			p.BystanderNum = room.getBystanderNum()
			room.sendToOthers(uid, msgIdProto.MsgId_s2cBroadcastLeaveNiuniuRoom, p)
			redisOpt.DelNiuniuPlayerRoomIdAndSeatIdx(strUid, room.roomId, uint32(seatIdx))
			log.Debugf("结算完成后, 发现 uid: %d 已经离线, 从 redis 中清除掉这个玩家", uid)
		}
	}

	for _, cli := range room.mapBystanders {
		if cli.IsOffline() {
			uid := cli.GetUid()
			p := &niuniuProto.S2CBroadcastLeaveNiuniuRoom{}
			p.RoomId = room.roomId
			p.SeatIndex = int32(cli.GetSeatIndex())
			room.removePlayer(cli)
			// 必须调用完 room.RemovePlayer 以后,才能调用 room.getBystanderNum
			p.BystanderNum = room.getBystanderNum()
			room.sendToOthers(uid, msgIdProto.MsgId_s2cBroadcastLeaveNiuniuRoom, p)
			log.Debugf("结算完成后, 发现 uid: %d 已经离线, 从 redis 中清除掉这个玩家(居然在旁观者集合中)", uid)
		}
	}
}

// map[座位号] - IClient
func (room *Room) getPlayerInSeat() map[uint32]IClient {
	m := map[uint32]IClient{}
	for seatIdx, client := range room.seat {
		if client != nil {
			m[uint32(seatIdx)] = client
		}
	}

	return m
}

func (room *Room) playerIsInSeat(c IClient) bool {
	for _, client := range room.seat {
		if client != nil && client.GetUid() == c.GetUid() {
			return true
		}
	}

	return false
}

// 座位序号, 是否入座成功
func (room *Room) haveASeat(c IClient) (uint32, bool) {
	if c.GetSeatIndex() == -1 {
		for seatIdx, client := range room.seat {
			// 座位没有被占用,可以入座
			if client == nil {
				room.seat[seatIdx] = c
				c.SetSeatIndex(seatIdx)
				// 从旁观者集合中移除
				delete(room.mapBystanders, c.GetUid())
				return uint32(seatIdx), true
			}
		}
	}

	return 0, false
}

func (room *Room) playerInRoom(c IClient) bool {
	uid := c.GetUid()
	// 这里只对座位上的人和旁观者进行遍历
	// 因为座位上的人就包括了游戏者
	for _, cli := range room.seat {
		if cli != nil && cli.GetUid() == uid {
			return true
		}
	}

	for _, cli := range room.mapBystanders {
		if cli.GetUid() == uid {
			return true
		}
	}

	return false
}

func (room *Room) IsPlaying(uid uint32) bool {
	room.lock.RLock()
	defer room.lock.RUnlock()
	_, find := room.mapPlayers[uid]
	return find
}

// 生成断线重连的快照
func (room *Room) makeReconnectSnap(c IClient) *niuniuProto.ReconnectSnap {
	rcSnap := &niuniuProto.ReconnectSnap{}
	// 已经准备的玩家 uid 列表
	rcSnap.Prepper = room.getPrepperId()
	rcSnap.MapRobZhuang = room.mapRobZhuang
	// 自己的4张明牌
	rcSnap.KnownCards = c.GetHandCard()
	// 正在进行游戏的玩家 uid, 如果当前是抢庄,下注,亮牌,结算状态,下面这个字段才有意义
	for uid, _ := range room.mapPlayers {
		rcSnap.PlayerUids = append(rcSnap.PlayerUids, uid)
	}
	// 下注信息, uid - 下注倍数, 如果当前是下注, 亮牌, 结算状态,下面这个字段才有意义
	rcSnap.MapBet = room.mapBet
	// 庄家 uid, 如果当前是抢庄,下注,亮牌,结算状态,下面这个字段才有意义
	rcSnap.BankerId = room.bankerId
	// 亮牌信息, 如果当前是亮牌,结算状态,下面这个字段才有意义
	rcSnap.MapShowCard = room.mapShowCard
	// uid - 输赢金币, 结算状态下,下面这个字段才有意义
	rcSnap.MapWinOrLose = room.mapWinOrLose
	// 亮牌阶段, 获取自己的第 5 张暗牌
	if room.roomStatus == niuniuProto.NiuniuRoomStatus_niuniuRoomStatusPlay {
		handCards := c.GetHandCard()
		if len(handCards) == 5 {
			rcSnap.UnknownCard = handCards[4]
		}
	}

	rcSnap.MapPlayerGold, _ = redisOpt.GetSomePlayersGold(rcSnap.PlayerUids)

	return rcSnap
}
