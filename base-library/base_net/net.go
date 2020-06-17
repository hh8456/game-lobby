package base_net

import (
	"net"
	"time"
)

// Listener 监听类
type Listener struct {
	// address 监听地址
	address string
	// callback 接受连接后的回调
	callback func(conn net.Conn)
	// isClosed 是否已关闭的标志
	isClosed bool
	// listener net库的listener
	listener net.Listener
}

// CreateListener 创建Listener对象
func CreateListener(addr string, callback func(conn net.Conn)) *Listener {
	return &Listener{
		address:  addr,
		callback: callback,
	}
}

// Start 开始监听
func (l *Listener) Start() error {
	ln, err := net.Listen("tcp", l.address)
	if err != nil {
		return err
	}
	defer ln.Close()
	l.listener = ln
	var tempDelay time.Duration // how long to sleep on accept failure

	for {
		conn, err := ln.Accept()
		if l.isClosed {
			break
		}
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			return err
		}
		tempDelay = 0
		l.callback(conn)
	}

	return nil
}

// Close 关闭监听
func (l *Listener) Close() error {
	l.isClosed = true
	return l.listener.Close()
}

func ConnectSocket(addr string, maxReadSize uint32) (*Socket, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return CreateSocket(conn, maxReadSize), nil
}
