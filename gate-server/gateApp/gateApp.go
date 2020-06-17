package gateApp

import (
	"net"
	"servers/base-library/base_net"
	"servers/common-library/config"
	"servers/common-library/function"
	"servers/common-library/log"
	"servers/common-library/packet"
	"servers/gate-server/gateApp/backEndServer"
	"servers/gate-server/gateApp/client"
	"servers/iface"
	"sync"
	"time"
)

type gateApp struct {
	localGateServerId uint32 // gate server id
	lockBackEndServer sync.RWMutex
	mapBackEndServer  map[uint32]*backEndServer.BackEndServer // 后端服务器唯一id - *backEndServer.BackEndServer
	// XXX 这里考虑用 10 个 lock 和 map,
	lockClient sync.RWMutex
	mapClient  map[int64]*client.Client // connId - *client.Client
	// XXX 需要从配置中读取 lobby 和其他后端服务器的 id - ip addr 并存放到 mapBackEndServerAddr
	mapBackEndServerAddr map[uint32]string // remoteServerId - ip addr
}

func CreateGateApp() *gateApp {
	err := config.InitSConfig("../config/server_config.yml")
	if err != nil {
		log.Error(err)
	}

	p := &gateApp{
		// TODO localGateServerId 的值要从配置中来
		localGateServerId:    1,
		mapBackEndServer:     make(map[uint32]*backEndServer.BackEndServer, 10),
		mapClient:            make(map[int64]*client.Client, 1000),
		mapBackEndServerAddr: make(map[uint32]string, 10),
	}

	// 大厅
	//p.mapBackEndServerAddr[301] = "192.168.0.155:3301"
	p.mapBackEndServerAddr[301] = config.Cfg.LobbyAddr
	// 牛牛服务器
	//p.mapBackEndServerAddr[401] = "192.168.0.184:3401"
	p.mapBackEndServerAddr[401] = config.Cfg.NiuNIuAddr

	return p
}

func (g *gateApp) getBackServer(remoteServerId uint32) iface.IBackEndServer {
	g.lockBackEndServer.RLock()
	defer g.lockBackEndServer.RUnlock()
	backEndServer, find := g.mapBackEndServer[remoteServerId]
	if find {
		return backEndServer
	}

	return nil
}

func (g *gateApp) kickClient(connId int64) {
	g.lockClient.Lock()
	client, find := g.mapClient[connId]
	if find {
		delete(g.mapClient, connId)
	}
	g.lockClient.Unlock()

	if find {
		client.Kick()
	}
}

func (g *gateApp) deleteClientConnId(connId int64) {
	g.lockClient.Lock()
	delete(g.mapClient, connId)
	g.lockClient.Unlock()
}

func (g *gateApp) Run(listenAddr string) {
	if g.localGateServerId == 0 {
		return
	}

	// XXX 这里要从 json 配置获取后端服务器列表
	for backEndSrvId, addr := range g.mapBackEndServerAddr {
		if backEndSrvId > 0 {
			go g.keepAliveForBackEndSrv(backEndSrvId, addr)
		}
	}

	time.Sleep(2 * time.Second)
	go g.listenTcpClient(listenAddr)
	go g.timer()
}

func (g *gateApp) timer() {
	defer function.Catch()
	backEndServerList := make([]*backEndServer.BackEndServer, 0, 20)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	i := uint32(0)
	for {
		select {
		case <-ticker.C:
			i++
			if i%20 == 0 {
				backEndServerList = backEndServerList[:0]
				g.lockBackEndServer.RLock()
				for _, backEndServer := range g.mapBackEndServer {
					backEndServerList = append(backEndServerList, backEndServer)
				}
				g.lockBackEndServer.RUnlock()

				for _, backEndServer := range backEndServerList {
					backEndServer.Ping()
				}
			}

		}
	}
}

// 监听客户端的 tcp 连接
func (g *gateApp) listenTcpClient(addr string) {
	defer function.Catch()

	sendBufSize := uint32(200)
	recvBufSize := uint32(200)
	l := base_net.CreateListener(addr, func(conn net.Conn) {
		c := client.NewClient(base_net.CreateSocket(conn,
			packet.C2SPacketMaxSize_16K), sendBufSize, recvBufSize,
			g.deleteClientConnId, g.getBackServer)
		g.lockClient.Lock()
		g.mapClient[c.ConnId()] = c
		g.lockClient.Unlock()

		c.Run()
	})

	for {
		err := l.Start()
		if err != nil {
			log.Errorf("listen connection error: %v, addr: %s", err, addr)
		}
		time.Sleep(time.Second)
	}
}

func (g *gateApp) connectBackEndServer(remoteServerId uint32, addr string) (bool, *backEndServer.BackEndServer) {
	defer function.Catch()
	socket, err := base_net.ConnectSocket(addr, uint32(packet.PacketMaxSize_4MB))
	if err != nil {
		log.Errorf("gate server 连接后端服务器 addr: %s 错误: %v", addr, err)
		return false, nil
	}

	sendBufSize := uint32(20000)
	recvBufSize := uint32(20000)
	s := backEndServer.NewBackEndServer(socket, g.localGateServerId,
		remoteServerId, sendBufSize, recvBufSize,
		g.disconnectBackEndSrv, g.sendToClient,
		g.setClientGameSrvId, g.sendToBackEndServer, g.kickClient)

	return true, s
}

func (g *gateApp) sendToClient(binData []byte) {
	dp := base_net.DataPack{}
	clientConnId := dp.UnpackClientConnId(binData)
	g.lockClient.RLock()
	client, find := g.mapClient[clientConnId]
	g.lockClient.RUnlock()
	if find {
		client.Send(binData)
	}
}

func (g *gateApp) setClientGameSrvId(clientConnId int64, gameServerId uint32) {
	g.lockClient.RLock()
	client, find := g.mapClient[clientConnId]
	g.lockClient.RUnlock()
	if find {
		client.SetGameServerId(gameServerId)
	}
}

// 和后端服务器之间要断线重连
func (g *gateApp) keepAliveForBackEndSrv(remoteServerId uint32, addr string) {

	for {
		g.lockBackEndServer.RLock()
		_, find := g.mapBackEndServer[remoteServerId]
		g.lockBackEndServer.RUnlock()
		if find {
			time.Sleep(time.Second * 10)
			continue
		}

		bucc, backEndServer := g.connectBackEndServer(remoteServerId, addr)
		if bucc {
			log.Debug("gate 连接 后端服务器 成功")
			g.lockBackEndServer.Lock()
			g.mapBackEndServer[remoteServerId] = backEndServer
			g.lockBackEndServer.Unlock()

			go backEndServer.Run()
		} else {
			log.Debug("gate 断线重连 后端服务器 失败, 10秒钟后会继续尝试重连")
			time.Sleep(time.Second * 10)
			continue
		}
	}
}

func (g *gateApp) disconnectBackEndSrv(remoteServerId uint32) {
	log.Debugf("gate 和 后端服务器 %d 断开连接", remoteServerId)
	g.lockBackEndServer.Lock()
	delete(g.mapBackEndServer, remoteServerId)
	g.lockBackEndServer.Unlock()
}

func (g *gateApp) sendToBackEndServer(remoteServerId uint32, binData []byte) {
	g.lockBackEndServer.RLock()
	backEndServer, find := g.mapBackEndServer[remoteServerId]
	g.lockBackEndServer.RUnlock()
	if find {
		backEndServer.Send(binData)
	}
}
