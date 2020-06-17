package niuniuApp

import (
	"servers/common-library/baseServer"
	"servers/common-library/config"
	"servers/niuniu-server/niuniuConfig"
	"servers/niuniu-server/roomPkg"
	"sync"
	"time"
)

const playerExpires = 7200 // player 的缓存时间是半小时
//const playerExpires = 60 // player 的缓存时间是1分钟

type niuniuApp struct {
	*baseServer.BaseServer
	lock sync.RWMutex
}

func CreateNiuniuApp() *niuniuApp {
	// TODO dbStr 应该修改为来自配置
	dbStr := config.Cfg.MysqlAddr
	//dbStr := "dev:dev123@tcp(192.168.0.155)/games?charset=utf8mb4&parseTime=True&loc=Local"

	p := &niuniuApp{
		BaseServer: baseServer.New(dbStr, playerExpires),
	}

	// TODO 这里要根据配置生成房间并写入 redis
	roomIds := []uint32{}
	for i := 0; i < 10; i++ {
		roomIds = append(roomIds, uint32(i)+1)
	}

	// TODO 写redis, 设置房间信息的功能,要放在管理后台
	niuniuConfig.SetRoomConfig(roomIds)
	if false == roomPkg.CreateRoom(401, roomIds, p.DB) {
		// 暂停几秒,等待写入日志到磁盘
		time.Sleep(time.Second)
		panic("由于写 redis 导致启动失败")
	}

	return p
}

func (n *niuniuApp) Run(addr string) {
	go n.ListenTcpGate(addr, n.dispathMsgToPlayer)
	// 房间定时器,驱动游戏进行
	go n.roomTimer()
}

// 用于房间定时器
func (n *niuniuApp) roomTimer() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			timestamp := time.Now().Unix()
			roomPkg.Timer(timestamp)
		}
	}
}
