package player

import (
	"servers/base-library/base_net"
	"servers/common-library/connData"
	"servers/common-library/log"
	"servers/common-library/proto/commonProto"
	"servers/common-library/proto/msgIdProto"
	"servers/common-library/redisOpt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jinzhu/gorm"
)

var (
	mapFunc map[msgIdProto.MsgId]func(*Player, *connData.ConnData)
)

func init() {
	mapFunc = map[msgIdProto.MsgId]func(*Player, *connData.ConnData){}
}

type Player struct {
	lock               sync.RWMutex
	connId             int64 // gate 上的连接号
	gateId             uint32
	uid                uint32
	strUid             string
	wxid               string
	chanConnData       chan *connData.ConnData
	isRun              bool
	lastAliveTimestamp int64 // 上次 ping 的时间
	chanTimer          chan int64
	DB                 *gorm.DB
	lockClose          sync.RWMutex
	isClosed           bool
	chanCloseSignal    chan struct{}
}

func NewPlayer(connId int64, gateId, uid uint32, wxid string,
	db *gorm.DB) *Player {

	return &Player{connId: connId, gateId: gateId, uid: uid,
		strUid: strconv.Itoa(int(uid)), wxid: wxid,
		chanConnData:       make(chan *connData.ConnData, 200),
		lastAliveTimestamp: time.Now().Unix(),
		chanTimer:          make(chan int64, 10),
		DB:                 db,
		chanCloseSignal:    make(chan struct{}, 10),
	}
}

func AppendFunc(msgId msgIdProto.MsgId, f func(*Player, *connData.ConnData)) {
	mapFunc[msgId] = f
}

//BaseServer.timer 每分钟驱动一次
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
		defer log.Debugf("player uid: %d, connId: %d  Run 协程退出",
			p.GetUid(), p.GetConnId())
		for {
			select {
			case connData, ok := <-p.chanConnData:
				if ok {
					p.handle(connData)
				} else {
					return
				}

			case <-p.chanTimer:

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

func (p *Player) Handle(connData *connData.ConnData) {
	p.lockClose.RLock()
	if p.isClosed == false {
		select {
		case p.chanConnData <- connData:

		default:
		}
		p.lockClose.RUnlock()
	}
}

func (p *Player) handle(connData *connData.ConnData) {
	dp := base_net.DataPack{}
	msgId := dp.UnpackMsgId(connData.BinData)
	if f, find := mapFunc[msgIdProto.MsgId(msgId)]; find {
		f(p, connData)
	}
}

func (p *Player) Close() {
	select {
	case p.chanCloseSignal <- struct{}{}:

	default:

	}

}

func (p *Player) SetGateId(gateId uint32) {
	p.lock.Lock()
	p.gateId = gateId
	p.lock.Unlock()
}

func (p *Player) GetGateId() uint32 {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.gateId
}

func (p *Player) GetWxid() string {
	return p.wxid
}

func (p *Player) GetUidString() string {
	return p.strUid
}

func (p *Player) GetUid() uint32 {
	return p.uid
}

func (p *Player) SetConnId(connId int64) {
	atomic.StoreInt64(&p.connId, connId)
}

func (p *Player) GetConnId() int64 {
	return atomic.LoadInt64(&p.connId)
}

func (p *Player) GetPlayerBaseInfo() *commonProto.PlayerBaseInfo {
	playerBaseInfo, b := redisOpt.LoadPlayerBaseInfo(strconv.Itoa(int(p.uid)))

	if b {
		p := playerBaseInfo.ToPbMsg()

		// FIXME p.GameType 还未赋值, 等做自建房的时候再补充了
		p.RoomId = redisOpt.GetNiuniuPlayerRoomId(p.GetUid())

		return p
	}

	return nil
}

func (p *Player) GetLastAliveTimestamp() int64 {
	return atomic.LoadInt64(&p.lastAliveTimestamp)
}

func (p *Player) SetLastAliveTimestamp(timestamp int64) {
	atomic.StoreInt64(&p.lastAliveTimestamp, timestamp)
}
