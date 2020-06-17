package base_net

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"servers/common-library/log"
	"servers/common-library/proto/msgIdProto"
	"time"

	"github.com/gogo/protobuf/proto"
)

const (
	defTimeoutDuration = time.Minute
)

var ErrOverMaxReadingSize = errors.New("over max reading size")

type Socket struct {
	conn                 net.Conn
	recvBuf              []byte
	reader               *bufio.Reader
	maxReadSize          uint32
	timeoutReadDuration  time.Duration
	timeoutWriteDuration time.Duration
}

func CreateSocket(conn net.Conn, maxReadSize uint32) *Socket {
	dp := NewDataPack()
	return &Socket{
		conn:                 conn,
		recvBuf:              make([]byte, dp.GetHeadLen()),
		reader:               bufio.NewReader(conn),
		maxReadSize:          maxReadSize,
		timeoutReadDuration:  defTimeoutDuration,
		timeoutWriteDuration: defTimeoutDuration,
	}
}

func (s *Socket) GetMaxReadSize() uint32 {
	return s.maxReadSize
}

func (s *Socket) Conn() net.Conn {
	return s.conn
}

func (s *Socket) ReadOne() ([]byte, error) {
	b, e := s.read()
	if e != nil {
		return nil, e
	}
	return b, nil
}

func (s *Socket) read() ([]byte, error) {
	s.conn.SetReadDeadline(time.Now().Add(s.timeoutReadDuration))
	if _, err := io.ReadFull(s.reader, s.recvBuf); err != nil {
		return nil, err
	}

	dp := NewDataPack()
	msgId := binary.BigEndian.Uint32(s.recvBuf[4:])
	msgLen, b := dp.UnpackMsgLen(s.recvBuf)
	if !b {
		return nil, fmt.Errorf("unpack msg head dataLen error,  msgid: %d", msgId)
	}

	if msgLen > s.maxReadSize {
		return nil, fmt.Errorf("read too large( len = %d ) data, msgid: %d",
			msgLen, msgId)
	}

	s.conn.SetReadDeadline(time.Now().Add(s.timeoutReadDuration))
	length := dp.GetHeadLen() + msgLen
	buf := make([]byte, length)

	copy(buf, s.recvBuf)
	if _, err := io.ReadFull(s.reader, buf[dp.GetHeadLen():length]); err != nil {
		return nil, err
	}

	return buf, nil
}

func (s *Socket) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

func (s *Socket) Close() {
	s.conn.Close()
}

//func (s *Socket) SendBuf(msg []byte) error {
func (s *Socket) Send(msg []byte) error {
	_, e := s.conn.Write(msg)
	return e
}

func (s *Socket) SendPbMsg(clientConnId int64, pid msgIdProto.MsgId, pbMsg proto.Message) error {
	if pbMsg != nil {
		msg, err := proto.Marshal(pbMsg)
		if err != nil {
			log.Errorf("序列化失败: pid = %s", pid.String())
			return err
		}

		return s.SendPbBuf(clientConnId, pid, msg)

	} else {
		return s.SendPbBuf(clientConnId, pid, nil)
	}
}

func (s *Socket) SendPbBuf(clientConnId int64, pid msgIdProto.MsgId, msg []byte) error {
	dp := NewDataPack()
	buf, err := dp.Pack(NewMsgPackage(clientConnId, uint32(pid), msg))
	if err != nil {
		log.Errorf("NewDataPack.Pack 打包失败, error: %v: pid = %s", err, pid.String())
		return err
	}

	_, e := s.conn.Write(buf)
	return e
}
