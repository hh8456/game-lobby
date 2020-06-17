package base_net

import (
	"context"
	"servers/common-library/function"
	"servers/common-library/log"
	"servers/common-library/proto/errorCodeProto"
	"servers/common-library/proto/msgIdProto"
	"servers/iface"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/sirupsen/logrus"
)

// XXX 可以通过 Connection 获取 gateid
type Connection struct {
	*Socket
	connId      int64 // 连接id, 全局唯一
	lock        sync.RWMutex
	isClosed    bool
	ctx         context.Context
	cancel      context.CancelFunc
	chanSendMsg chan []byte            // 保存将要发给 socket 的数据
	chanRecvMsg chan []byte            // 保存从 socket 读到的数据
	mapProperty map[string]interface{} // XXX 包含了 gateid - uint32
}

func NewConnection(s *Socket, connId int64, sendBufSize, recvBufSize uint32) *Connection {
	ctx, cancel := context.WithCancel(context.Background())
	return &Connection{Socket: s, connId: connId,
		ctx: ctx, cancel: cancel,
		chanSendMsg: make(chan []byte, sendBufSize),
		chanRecvMsg: make(chan []byte, recvBufSize),
		mapProperty: map[string]interface{}{}}
}

func (c *Connection) SetProperty(key string, value interface{}) {
	c.mapProperty[key] = value
}

func (c *Connection) GetProperty(key string) interface{} {
	if v, find := c.mapProperty[key]; find {
		return v
	}

	return nil
}

func (c *Connection) ConnId() int64 {
	return c.connId
}

func (c *Connection) Run() {
	go c.recvLoop()
	go c.sendLoop()
}

// 用于通知外部协程关闭
func (c *Connection) GetExitSignal() <-chan struct{} {
	return c.ctx.Done()
}

// 外部协程从这里获取收到的数据
func (c *Connection) GetChanRecvMsg() <-chan []byte {
	return c.chanRecvMsg
}

// recvLoop 和 sendLoop 其中一方退出时, 会触发另外一方退出
func (c *Connection) recvLoop() {
	var connId int64
	v := c.GetProperty("connId")
	if v != nil {
		if value, ok := v.(int64); ok {
			connId = value
		}
	}

	logrus.Debugf("Connection.recvLoop 启动, connId: %d", connId)
	defer func() {
		function.Catch()
		logrus.Debugf("Connection.recvLoop 退出, 调用 "+
			"Connection.Close, connId: %d", connId)
		c.Close()
	}()

	for {
		binData, err := c.ReadOne()
		if err != nil {
			logrus.Debugf("Connection.recvLoop 退出, socket "+
				"read error: %v, connId: %d", err, connId)
			return
		}

		select {
		case c.chanRecvMsg <- binData:

		default:
			logrus.Errorf("Connection.chanRecvMsg 通道容量满, 导致消息溢出")

		}
	}
}

// recvLoop 和 sendLoop 其中一方退出时, 会触发另外一方退出
func (c *Connection) sendLoop() {
	var connId int64
	v := c.GetProperty("connId")
	if v != nil {
		if value, ok := v.(int64); ok {
			connId = value
		}
	}

	logrus.Debugf("Connection.sendLoop 启动, connId: %d", connId)
	defer func() {
		function.Catch()
		logrus.Debugf("Connection.sendLoop 退出, connId: %d", connId)
	}()

	for {
		select {
		case binData, ok := <-c.chanSendMsg:
			if ok {
				if len(binData) > 4096 {
					logrus.Debugf("sendLoop 中发送较大的逻辑包, 长度: %d", len(binData))
				}

				c.Socket.Send(binData)
			} else {
				logrus.Debugf("base_net/Connection sendLoop 中感知到 close(Connection.chanSendMsg) ")
				c.Socket.Close()
				return
			}
		}
	}
}

// Connection.Send 是把 binData 投递到发送缓冲区
// 如果需要直接发送, 就使用 Connection.Socket.Send
func (c *Connection) Send(binData []byte) {
	// 加锁是为了防止写 closed c.chanSendMsg 崩溃
	// go test bench 的并发压测显示, 不加锁 40 ns/op, 加读锁后 62 ns/op
	c.lock.RLock()
	if c.isClosed == false {
		select {
		case c.chanSendMsg <- binData:

		default:
			logrus.Errorf("Connection.chanSendMsg 通道容量( %d  )满了", len(c.chanSendMsg))
		}

	}
	c.lock.RUnlock()
}

func (s *Connection) SendErrorCode(clientConnId int64, errCode errorCodeProto.ErrorCode) {
	s2cErrorCode := &errorCodeProto.S2CErrorCode{}
	s2cErrorCode.Code = int32(errCode)
	s.SendPbMsg(clientConnId, msgIdProto.MsgId_s2cErrorCode, s2cErrorCode)
}

// Connection.SendPbMsg 是把数据打包并投递到发送缓冲区
// 如果需要直接发送, 就使用 Connection.Socket.SendPbMsg
func (s *Connection) SendPbMsg(clientConnId int64, pid msgIdProto.MsgId, pbMsg proto.Message) {
	if pbMsg != nil {
		msg, err := proto.Marshal(pbMsg)
		if err != nil {
			log.Errorf("proto.Marshal 序列化失败: pid = %s", pid.String())
			return
		}
		s.SendPbBuf(clientConnId, pid, msg)
	} else {
		s.SendPbBuf(clientConnId, pid, nil)
	}
}

// Connection.SendPbBuf 是把数据打包并投递到发送缓冲区
// 如果需要直接发送, 就使用 Connection.Socket.SendPbBuf
func (c *Connection) SendPbBuf(clientConnId int64, pid msgIdProto.MsgId, msg []byte) {
	dp := NewDataPack()
	buf, err := dp.Pack(NewMsgPackage(clientConnId, uint32(pid), msg))
	if err != nil {
		log.Errorf("NewDataPack.Pack 打包失败, error: %v: pid = %s", err, pid.String())
		return
	}

	//log.Debugf("向客户端发送消息 id: %s", pid.String())
	c.Send(buf)
}

func (c *Connection) SendMessage(message iface.IMessage) {
	dp := NewDataPack()
	buf, err := dp.Pack(message)
	if err != nil {
		log.Errorf("Connection.SendMessage NewDataPack.Pack 打包失败, error: %v: ", err)
		return
	}

	c.Send(buf)
}

// 1
// 主动调用 Connection.Close 会触发 Connection.sendLoop 协程退出,
// Connection.sendLoop 会关闭 Connection.Socket, 触发 Connection.recvLoop 协程退出

// 2
// 如果 Connection.recvLoop 协程感知到对端关闭连接,那么就会先退出, 并调用 Connection.Close
// Connection.Close 会触发 Connection.sendLoop 协程退出
func (c *Connection) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.isClosed == true {
		return
	}

	c.isClosed = true
	close(c.chanSendMsg)
	log.Debugf("Connection.Close 函数中,调用 Connection.cancel ")
	c.cancel()
	log.Debugf("Connection.Close 函数中,调用 Connection.cancel 完毕")
}
