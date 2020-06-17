package client

import (
	"servers/base-library/base_net"
	"servers/common-library/function"
	"servers/common-library/log"
	"servers/common-library/proto/lobbyProto"
	"servers/common-library/proto/msgIdProto"
	"servers/iface"
	"sync"

	"github.com/hh8456/go-common/snowflake"
)

type Client struct {
	*base_net.Connection
	lock sync.RWMutex
	// XXX 如果 gameServer 宕机, 玩家不需要重新登录,但需要退出游戏并重新进入
	gameServerId       uint32 // 如果正在游戏,所在哪个 game server id, 如果没在游戏,则 gameServerId 为 0
	isRun              bool
	kick               bool //是否被服务器 kick 掉的
	deleteClientConnId func(connId int64)
	cbGetBackServer    func(uint32) iface.IBackEndServer
}

func NewClient(c *base_net.Socket, sendBufSize, recvBufSize uint32,
	deleteClientConnId func(int64),
	cbGetBackServer func(uint32) iface.IBackEndServer) *Client {

	connId := snowflake.GetSnowflakeId()
	conn := base_net.NewConnection(c, connId, sendBufSize, recvBufSize)
	conn.SetProperty("connId", connId)

	return &Client{Connection: conn, deleteClientConnId: deleteClientConnId,
		cbGetBackServer: cbGetBackServer}
}

func (c *Client) Kick() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.kick {
		c.kick = true
		c.Close()
		log.Debugf("设置 connId: %d kick = true", c.ConnId())
	}
}

func (c *Client) Run() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.isRun {
		return
	}

	c.isRun = true
	c.Connection.Run()
	go c.forwardLoop()
}

// 把客户端发来的数据转发给后端服务器
func (c *Client) forwardLoop() {
	defer func() {
		function.Catch()
		log.Debugf("Client forwardLoop 协程退出")
	}()

	for {
		select {
		// 取得客户端发来的数据
		case binData := <-c.GetChanRecvMsg():
			dp := base_net.NewDataPack()
			msgId := dp.UnpackMsgId(binData)
			recvLen, b := dp.UnpackMsgLen(binData)
			if (b == true) && (recvLen < c.GetMaxReadSize()) {

				dp.SetClientConnId(binData, c.ConnId())
				log.Println("收到消息:", msgId)
				// 转发给 lobby
				if msgId > uint32(msgIdProto.MsgId_startId) &&
					msgId < uint32(msgIdProto.MsgId_lobbyEndId) {
					// TODO 这里应该用一致性哈希获得 lobbyServerId
					lobbyServerId := uint32(301)
					lobbyServer := c.cbGetBackServer(lobbyServerId)
					if lobbyServer != nil {
						lobbyServer.Send(binData)
					}

					if msgId == uint32(msgIdProto.MsgId_c2sWxLogin) {
						c2sPbMsg := &lobbyProto.C2SWxLogin{}
						if function.ProtoUnmarshal(binData[dp.GetHeadLen():],
							c2sPbMsg, "lobbyProto.C2SWxLogin") {
							log.Debugf("收到客户端 connid: %d, wxid: %s 登录消息: msgIdProto.MsgId_c2sWxLogin",
								c.ConnId(), c2sPbMsg.Wxid)
						}
					}
				}

				// 转发给牛牛服务器
				if msgId > uint32(msgIdProto.MsgId_niuniuStartId) &&
					msgId < uint32(msgIdProto.MsgId_niuniuEndId) {
					log.Println("转发给牛牛服务器:", msgId)
					// 这里应该是 401; date: 20.5.21
					niuniuServerId := c.GetGameServerId()
					niuniuServer := c.cbGetBackServer(niuniuServerId)
					if niuniuServer != nil {
						niuniuServer.Send(binData)
					}
				}

			} else {
				log.Errorf("gate, 解包错误, 关闭客户端连接, "+
					"包长度 %d 可能超过最大长度 %d", recvLen, c.GetMaxReadSize())
				c.Close()
				return
			}

		case <-c.GetExitSignal():
			// 要把退出房间的消息发给牛牛服务器
			log.Debugf("gate 感知到玩家 connid: %d 断线", c.ConnId())
			niuniuServerId := c.GetGameServerId()
			niuniuServer := c.cbGetBackServer(niuniuServerId)
			if niuniuServer != nil {
				dp := base_net.NewDataPack()
				msgId := msgIdProto.MsgId_gate2NiuniuClientDisconnect
				buf, err := dp.Pack(base_net.NewMsgPackage(c.ConnId(), uint32(msgId), nil))
				if err != nil {
					log.Errorf("NewDataPack.Pack 打包失败, error: %v: pid = %s",
						err, msgId.String())
					return
				}

				c.deleteClientConnId(c.ConnId())
				log.Debugf("gate 感知到玩家 connid: %d 断线, 向牛牛服务器发起退出房间的通知", c.ConnId())
				niuniuServer.Send(buf)
			}

			return
		}
	}
}

func (c *Client) SetGameServerId(gameServerId uint32) {
	c.lock.Lock()
	c.gameServerId = gameServerId
	c.lock.Unlock()
}

func (c *Client) GetGameServerId() uint32 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.gameServerId
}
