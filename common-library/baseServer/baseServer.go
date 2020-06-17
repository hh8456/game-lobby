package baseServer

import (
	"fmt"
	"net"
	"servers/base-library/base_net"
	"servers/common-library/connData"
	"servers/common-library/function"
	"servers/common-library/gateServer"
	"servers/common-library/log"
	"servers/common-library/packet"
	"servers/iface"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

const (
	circle = 30
)

var runingTime int64

type slot struct {
	m map[string]struct{} // wxid - struct{}
}

func newSlot() *slot {
	return &slot{m: map[string]struct{}{}}
}

func (s *slot) add(wxid string) {
	s.m[wxid] = struct{}{}
}

func (s *slot) delete(wxid string) {
	delete(s.m, wxid)
}

// 大厅,游戏服都属于后端服务器,都需要和 gate 连接并接受 player 的登录
// BaseServer 是大厅,游戏服共同的 "父类"
type BaseServer struct {
	lockGate      sync.RWMutex
	mapGate       map[uint32]*gateServer.GateServer // gateServerId - GateServer
	lockPlayer    sync.RWMutex
	mapPlayer     map[string]iface.IPlayer // wxid - Player
	mapConnid     map[int64]string         // connId - wxid
	mapWxidConnid map[string]int64         // wxid - connId
	DB            *gorm.DB
	playerExpires int64 // player 的缓存时间
	slots         []*slot
	mapIndicator  map[string]*slot // 用来跟踪某个对象在哪个槽位
}

// TODO 这里的形式参数需要修改为配置信息
func New(dbStr string, playerExpires int64) *BaseServer {
	//db, err := gorm.Open("mysql", "dev:dev123@tcp(192.168.0.155)/games?charset=utf8mb4&parseTime=True&loc=Local")
	db, err := gorm.Open("mysql", dbStr)
	if err != nil {
		str := fmt.Sprintf("连接数据库错误: %v\n", err)
		panic(str)
	}

	db.SingularTable(true)
	db.DB().SetMaxIdleConns(20)
	db.DB().SetMaxOpenConns(50)

	s := &BaseServer{mapGate: make(map[uint32]*gateServer.GateServer, 10),
		mapPlayer:     make(map[string]iface.IPlayer, 20000), // 两万并发
		mapConnid:     make(map[int64]string, 20000),
		mapWxidConnid: make(map[string]int64, 20000),
		DB:            db, playerExpires: playerExpires,
		mapIndicator: make(map[string]*slot, 2000),
	}

	for i := 0; i < circle; i++ {
		s.slots = append(s.slots, newSlot())
	}

	go s.timer()
	return s
}

// 定时器是用来删除僵尸 player 的
func (s *BaseServer) timer() {
	defer function.Catch()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {

		select {
		case <-ticker.C:
			timestamp := time.Now().Unix()
			slotIndex := atomic.LoadInt64(&runingTime) % circle
			slot := s.slots[slotIndex]
			//log.Debugf("baseServer 定时器正在检查槽位 %d ", slotIndex)
			s.lockPlayer.Lock()
			//log.Debugf("baseServer 定时器上槽位 %d 的玩家: %v", slotIndex, slot.m)
			for wxid, _ := range slot.m {
				if player, find := s.mapPlayer[wxid]; find {
					if timestamp-player.GetLastAliveTimestamp() < s.playerExpires {
						log.Debugf("baseServer 时间轮定时器半分钟触发一次, wxid: %s, connId: %d", wxid, s.mapWxidConnid[wxid])
						player.Timer(timestamp)
					} else {
						player.Close()
						log.Debugf("player 长时间没发送心跳包, baseServer 底层模块将其关闭, 调用 player.Close 完毕")
						delete(s.mapPlayer, wxid)
						delete(s.mapIndicator, wxid)
						slot.delete(wxid)
						connId, find := s.mapWxidConnid[wxid]
						if find {
							delete(s.mapConnid, connId)
							delete(s.mapWxidConnid, wxid)
							log.Debugf("baseServer 在玩家心跳超时后, 清理了 mapConnid, mapPlayer, mapWxidConnid 中的信息")
						} else {
							log.Errorf("baseServer 底层库没通过 wxid %s 找到 connId, "+
								"表明 mapWxidConnid, mapConnid, mapPlayer 信息未同步 ", wxid)
						}
					}
				} else {
					log.Errorf("baseServer 底层库的时间轮算法中, 发现 BaseServer.slots 和 BaseServer.mapPlayer 信息不同步")
				}
			}

			check := len(s.mapConnid) == len(s.mapPlayer) && len(s.mapConnid) == len(s.mapWxidConnid)
			if check == false {
				log.Errorf("baseServer 底层库的时间轮算法中, 发现 mapConnid, mapPlayer, mapWxidConnid 信息不同步")
			}

			s.lockPlayer.Unlock()
			atomic.AddInt64(&runingTime, 1)
		}
	}
}

func (s *BaseServer) storeGate(gateServerId uint32,
	gateServer *gateServer.GateServer) bool {
	s.lockGate.Lock()
	defer s.lockGate.Unlock()

	if _, find := s.mapGate[gateServerId]; find {
		return false
	}

	s.mapGate[gateServerId] = gateServer
	return true
}

func (s *BaseServer) GetGate(gateServerId uint32) *gateServer.GateServer {
	s.lockGate.RLock()
	defer s.lockGate.RUnlock()

	if s, find := s.mapGate[gateServerId]; find {
		return s
	}

	return nil
}

func (s *BaseServer) removeGate(gateServerId uint32) {
	s.lockGate.Lock()
	defer s.lockGate.Unlock()
	delete(s.mapGate, gateServerId)
}

func (s *BaseServer) ListenTcpGate(addr string,
	outFunc func(*connData.ConnData)) {

	defer function.Catch()

	sendBufSize := uint32(20000)
	recvBufSize := uint32(20000)

	l := base_net.CreateListener(addr, func(conn net.Conn) {
		g := gateServer.NewGateServer(
			base_net.CreateSocket(conn, packet.PacketMaxSize_4MB),
			sendBufSize, recvBufSize, s.storeGate, s.removeGate, outFunc)

		g.Run()
	})

	for {
		err := l.Start()
		if err != nil {
			log.Errorf("lobby listen error: %v, addr: %s", err, addr)
		}

		time.Sleep(time.Second)
	}
}

// player.connId 是在 gate-server 上由雪花算法保证了唯一性
// 这里假设 player.connId 不可能重复,所以不做容错性处理
func (s *BaseServer) StorePlayer(wxid string, player iface.IPlayer, newConnId, oldConnId int64) {
	s.lockPlayer.Lock()
	defer s.lockPlayer.Unlock()
	delete(s.mapConnid, oldConnId)
	s.mapConnid[newConnId] = wxid
	s.mapPlayer[wxid] = player
	s.mapWxidConnid[wxid] = newConnId

	if slot, find := s.mapIndicator[wxid]; find {
		slot.delete(wxid)
	}

	// 确定时间轮定时器的槽位
	slotIndex := atomic.LoadInt64(&runingTime)%circle - 1
	if slotIndex < 0 {
		slotIndex = circle - 1
	}

	slot := s.slots[slotIndex]
	slot.add(wxid)
	s.mapIndicator[wxid] = slot

	log.Debugf("baseServer.StorePlayer 保存 wxid: %s, 分配的时间轮定时器的槽位号: %d", wxid, slotIndex)
}

func (s *BaseServer) FindPlayer(wxid string) iface.IPlayer {
	s.lockPlayer.RLock()
	defer s.lockPlayer.RUnlock()
	return s.mapPlayer[wxid]
}

func (s *BaseServer) FindPlayerByConnId(connId int64) iface.IPlayer {
	s.lockPlayer.RLock()
	defer s.lockPlayer.RUnlock()
	wxid := s.mapConnid[connId]
	return s.mapPlayer[wxid]
}
