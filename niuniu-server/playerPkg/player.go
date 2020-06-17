package playerPkg

import (
	"servers/base-library/base_net"
	"servers/common-library/connData"
	"servers/common-library/function"
	"servers/common-library/gateServer"
	"servers/common-library/log"
	"servers/common-library/proto/commonProto"
	"servers/common-library/proto/errorCodeProto"
	"servers/common-library/proto/msgIdProto"
	"servers/common-library/redisKeyPrefix"
	"servers/common-library/redisOpt"
	"servers/niuniu-server/roomPkg"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/hh8456/go-common/redisObj"
)

// Player 所有的方法,只能在 room 的单线程中调用,这样就可以不用加锁了
type Player struct {
	lock               sync.RWMutex
	connId             int64 // gate 上的连接号
	gateId             uint32
	uid                uint32
	wxid               string
	strUid             string
	chanConnData       chan *connData.ConnData
	isRun              bool
	seatIndex          int32
	room               *roomPkg.Room
	roomId             uint32
	handCards          []*commonProto.PokerCard
	getGate            func(uint32) *gateServer.GateServer
	lastAliveTimestamp int64 // 上次 ping 的时间
	prepare            int32 // 是否点击了准备
	offline            int32 // 是否离线
	chanTimer          chan int64
	lockClose          sync.RWMutex
	isClosed           bool
	chanCloseSignal    chan struct{}
}

func NewPlayer(connId int64, gateId, uid uint32, wxid string,
	getGate func(uint32) *gateServer.GateServer) *Player {
	return &Player{seatIndex: -1, connId: connId, gateId: gateId,
		uid: uid, strUid: strconv.Itoa(int(uid)), wxid: wxid,
		chanConnData:       make(chan *connData.ConnData, 200),
		getGate:            getGate,
		lastAliveTimestamp: time.Now().Unix(),
		chanTimer:          make(chan int64, 10),
		chanCloseSignal:    make(chan struct{}, 10),
	}
}

// 游戏结束后,重新准备开始游戏
func (p *Player) ReReady() {
	p.lock.Lock()
	p.handCards = p.handCards[:0]
	p.lock.Unlock()
	atomic.StoreInt32(&p.prepare, 0)
}

func (p *Player) UnPrepare() {
	atomic.StoreInt32(&p.prepare, 0)
}

func (p *Player) Prepare() {
	atomic.StoreInt32(&p.prepare, 1)
}

func (p *Player) IsPrepare() bool {
	return 1 == atomic.LoadInt32(&p.prepare)
}

// BaseServer.time 每30秒驱动一次
func (p *Player) Timer(timestamp int64) {
	select {
	case p.chanTimer <- timestamp:

	default:
	}
}

func (p *Player) Run() {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.isRun {
		return
	}

	p.isRun = true
	go func() {
		defer function.Catch()
		uid := p.GetUid()
		strUid := p.GetUidString()
		rdsPlayerRoomId := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuPlayerRoomId)
		rdsPlayerSeatIdx := redisObj.NewSessionWithPrefix(redisKeyPrefix.NiuniuSeatIndex)
		for {
			select {
			case connData, ok := <-p.chanConnData:
				if ok {
					p.handle(connData)
				} else {
					log.Debugf("Player.chanConnData 通道关闭, uid: %d 协程退出", uid)
					return
				}

				// BaseServer.time 每30秒驱动一次
			case <-p.chanTimer:
				strRoomId, e1 := rdsPlayerRoomId.Get(strUid)
				if e1 == nil {
					// 先查询是否有房间和座位号,如果有,就延长生命周期
					strSeatIndex, e2 := rdsPlayerSeatIdx.Get(strUid)
					if e2 == nil {
						// redis 中, niuniu_player_room_id:uid - uid, niuniu_seat_index:uid - uid, niuniu_player_room_id:roomid:niuniu_seat_index:seatIdx - uid
						// 这3个键值对的生命周期是一致的
						e3 := rdsPlayerRoomId.Expire(strUid, time.Minute)
						if e3 != nil {
							log.Errorf("BaseServer.time 每30秒驱动 Player.Run , uid: %d, 在 redis 中延长( uid - roomid )键值对 %s - %s 存活期出错,"+
								"error: %v", uid, rdsPlayerRoomId.GetPrefix()+":"+strUid, strRoomId, e3)
						} else {
							log.Debugf("BaseServer.time 每30秒驱动 Player.Run , uid: %d, 在 redis 中延长( uid - roomid )键值对 %s - %s 存活期成功",
								uid, rdsPlayerRoomId.GetPrefix()+":"+strUid, strRoomId)
						}

						e4 := rdsPlayerSeatIdx.Expire(strUid, time.Minute)
						if e4 != nil {
							log.Errorf("BaseServer.time 每30秒驱动 Player.Run , uid: %d, 在 redis 中延长( uid - seatIdx )键值对 %s - %s 存活期出错, "+
								"error: %v", uid, rdsPlayerSeatIdx.GetPrefix()+":"+strUid, strSeatIndex, e4)
						} else {
							log.Debugf("BaseServer.time 每30秒驱动 Player.Run , uid: %d, 在 redis 中延长( uid - seatIdx )键值对 %s - %s 存活期成功",
								uid, rdsPlayerSeatIdx.GetPrefix()+":"+strUid, strSeatIndex)
						}

						prefix := redisKeyPrefix.NiuniuPlayerInRoom + ":" + strRoomId + ":" + redisKeyPrefix.NiuniuSeatIndex
						rdsPlayerRoomIdAndSeatIdx := redisObj.NewSessionWithPrefix(prefix)
						e5 := rdsPlayerRoomIdAndSeatIdx.Expire(strSeatIndex, time.Minute)
						if e5 != nil {
							log.Errorf("BaseServer.time 每30秒驱动 Player.Run , uid: %d, 在 redis 中延长( roomid:seatIdx - uid ) 键值对 %s - uid 存活期出错, "+
								"error: %v", uid, prefix+":"+strSeatIndex, e5)
						} else {
							log.Debugf("BaseServer.time 每30秒驱动 Player.Run , uid: %d, 在 redis 中延长( roomid:seatIdx - uid ) 键值对 %s - uid 存活期成功",
								uid, prefix+":"+strSeatIndex)
						}

					} else {
						log.Errorf("BaseServer.time 每30秒驱动 Player.Run , uid: %d, 在 redis 中查询所在房间 id 成功, "+
							"但查询座位上的玩家出错: %v", uid, e2)
					}
				} else {
					log.Errorf("BaseServer.time 每30秒驱动 Player.Run , uid: %d, 在 redis 中查询所在房间 id 出错: %v",
						p.GetUid(), e1)
				}

			case <-p.chanCloseSignal:
				//wxid, _ := redisOpt.GetUidByWxid()
				log.Debugf("底层库 baseServer 调用了 Player.Close 方法, uid: %d", p.GetUid())
				p.lockClose.Lock()
				if p.isClosed == false {
					p.isClosed = true
					close(p.chanConnData)
				}
				p.lockClose.Unlock()
				log.Debugf("底层库 baseServer 调用了 Player.Close 方法, uid: %d 加锁完毕", p.GetUid())
			}
		}
	}()
}

// baseServer 底层会调用 Player.Close()
func (p *Player) Close() {
	select {
	case p.chanCloseSignal <- struct{}{}:

	default:

	}
}

func (p *Player) GetSeatIndex() int32 {
	return atomic.LoadInt32(&p.seatIndex)
}

func (p *Player) SetSeatIndex(seatIndex int) {
	atomic.StoreInt32(&p.seatIndex, int32(seatIndex))
}

func (p *Player) GetPlayerBaseInfo() *commonProto.PlayerBaseInfo {
	playerBaseInfo, b := redisOpt.LoadPlayerBaseInfo(strconv.Itoa(int(p.uid)))

	if b {
		return playerBaseInfo.ToPbMsg()
	}

	return nil
}

func (p *Player) GetUidString() string {
	return p.strUid
}

func (p *Player) GetUid() uint32 {
	return p.uid
}

func (p *Player) GetWxid() string {
	return p.wxid
}

func (p *Player) AppendHandCard(cards ...*commonProto.PokerCard) {
	p.lock.Lock()
	p.handCards = append(p.handCards, cards...)
	p.lock.Unlock()
}

func (p *Player) GetHandCard() []*commonProto.PokerCard {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.handCards
}

func (p *Player) Handle(connData *connData.ConnData) {
	p.lockClose.RLock()
	if p.isClosed == false {
		select {
		case p.chanConnData <- connData:

		default:
		}
	}
	p.lockClose.RUnlock()
}

func (p *Player) handle(connData *connData.ConnData) {
	dp := &base_net.DataPack{}
	msgId := dp.UnpackMsgId(connData.BinData)

	if f, ok := mapPlayerLogicFunc[msgIdProto.MsgId(msgId)]; ok {
		f(p, connData)
	}
}

func (p *Player) SendErrorCode(errCode errorCodeProto.ErrorCode) {
	s2cErrorCode := &errorCodeProto.S2CErrorCode{}
	s2cErrorCode.Code = int32(errCode)
	p.SendPbMsg(msgIdProto.MsgId_s2cErrorCode, s2cErrorCode)
}

func (p *Player) Send(binData []byte) {
	getId := p.GetGateId()
	gateServer := p.getGate(getId)
	if gateServer != nil {
		gateServer.Send(binData)
	}
}

func (p *Player) SendPbMsg(msgId msgIdProto.MsgId, pbMsg proto.Message) {
	dp := base_net.NewDataPack()
	if pbMsg != nil {
		msg, err := proto.Marshal(pbMsg)
		if err != nil {
			log.Errorf("proto.Marshal 序列化失败: pid = %s", msgId.String())
			return
		}
		buf, err2 := dp.Pack(base_net.NewMsgPackage(p.connId, uint32(msgId), msg))
		if err2 != nil {
			log.Errorf("NewDataPack.Pack 打包失败, error: %v: pid = %s",
				err2, msgId.String())
			return
		}

		p.Send(buf)
	} else {
		buf, err2 := dp.Pack(base_net.NewMsgPackage(p.connId, uint32(msgId), nil))
		if err2 != nil {
			log.Errorf("NewDataPack.Pack 打包失败, error: %v: pid = %s",
				err2, msgId.String())
			return
		}

		p.Send(buf)
	}

}

// 是否正在房间中
func (p *Player) isInRoom() bool {
	return atomic.LoadUint32(&p.roomId) != 0
}

func (p *Player) SetRoomId(roomId uint32, room *roomPkg.Room) {
	p.lock.Lock()
	p.room = room
	p.lock.Unlock()
	atomic.StoreUint32(&p.roomId, roomId)
}

func (p *Player) GetRoomId() uint32 {
	return atomic.LoadUint32(&p.roomId)
}

func (p *Player) GetGateId() uint32 {
	return atomic.LoadUint32(&p.gateId)
}

func (p *Player) SetGateId(gateId uint32) {
	atomic.StoreUint32(&p.gateId, gateId)
}

func (p *Player) GetConnId() int64 {
	return atomic.LoadInt64(&p.connId)
}

func (p *Player) SetConnId(connId int64) {
	atomic.StoreInt64(&p.connId, connId)
}

func (p *Player) SetLastAliveTimestamp(timestamp int64) {
	atomic.StoreInt64(&p.lastAliveTimestamp, timestamp)
}

func (p *Player) GetLastAliveTimestamp() int64 {
	return atomic.LoadInt64(&p.lastAliveTimestamp)
}

func (p *Player) GetRoom() *roomPkg.Room {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.room
}

func (p *Player) Online() {
	atomic.StoreInt32(&p.offline, 0)
}

func (p *Player) Offline() {
	atomic.StoreInt32(&p.offline, 1)
}

func (p *Player) IsOnline() bool {
	return atomic.LoadInt32(&p.offline) == 0
}

func (p *Player) IsOffline() bool {
	return atomic.LoadInt32(&p.offline) == 1
}
