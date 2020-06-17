package function

import (
	"bufio"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"servers/common-library/packet"
	"time"
	"unsafe"

	log "github.com/sirupsen/logrus"

	//"github.com/golang/protobuf/proto"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
)

func Catch() {
	if v := recover(); v != nil {
		backtrace(v)
	}
}

func CatchWithInfo(info string) {
	if v := recover(); v != nil {
		log.Infof("catch panic info:%s", info)
		backtrace(v)
	}
}

func backtrace(message interface{}) {
	//fmt.Fprintf(os.Stderr, "Traceback[%s] (most recent call last):\n", time.Now())
	log.Errorf("Traceback[%s] (most recent call last):\n", time.Now())
	for i := 0; ; i++ {
		pc, file, line, ok := runtime.Caller(i + 1)
		if !ok {
			break
		}
		//fmt.Fprintf(os.Stderr, "% 3d. %s() %s:%d\n", i, runtime.FuncForPC(pc).Name(), file, line)
		log.Errorf("% 3d. %s() %s:%d\n", i, runtime.FuncForPC(pc).Name(), file, line)
	}
	//fmt.Fprintf(os.Stderr, "%v\n", message)
	log.Errorf("%v\n", message)
}

func ByteString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func Get_external() {
	resp, err := http.Get("http://baidu.com")
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
}

func Get_internal() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("error: " + err.Error())
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				os.Stdout.WriteString(ipnet.IP.String() + "\n")
			}
		}
	}
}

func Wait(buff []byte, sendChannel chan []byte) (int, bool) {
	msg, ok := <-sendChannel
	// 如果 sendChannel 通道关闭了,那么 ok 就是 false
	if !ok {
		return 0, false
	}

	copy(buff, msg)
	index := len(msg)

LOOP:
	// 如果余下的缓冲区还能够容纳逻辑包, 就继续从 channel 中取数据
	for len(buff[index:]) >= int(packet.C2SPacketMaxSize_16K) {
		select {
		case msg, ok := <-sendChannel:
			// 如果 sendChannel 通道关闭了,那么 ok 就是 false
			if !ok {
				return index, false
			}
			index += copy(buff[index:], msg)

			// 如果没有数据可取,那么就退出循环
		default:
			break LOOP
		}
	}

	return index, true
}

func SnappyEncodeMessage(value proto.Message) []byte {
	buf, err := proto.Marshal(value)

	if err != nil {
		log.Errorf("proto marshal failed! Message: %+v, err: %v",
			value, err)
		return nil
	}

	return snappy.Encode(nil, buf)
}

func SnappyDecodeMessage(value []byte, result proto.Message) error {
	dst, err := snappy.Decode(nil, value)

	if err != nil {
		log.Errorf("snappy decode failed! Message: %+v", result)
		return err
	}

	return proto.Unmarshal(dst, result)
}

func ReadPacket2(conn net.Conn, reader *bufio.Reader) ([]byte, error) {
	conn.SetReadDeadline(time.Now().Add(time.Minute * 2))
	headBuf := make([]byte, packet.PacketHeaderSize)
	// 先读包头
	if _, err := io.ReadFull(reader, headBuf[:packet.PacketHeaderSize]); err != nil {
		return nil, err
	}

	msgId := binary.BigEndian.Uint32(headBuf[0:])
	dataBodyLen := binary.BigEndian.Uint32(headBuf[4:])

	if dataBodyLen <= (1024*1024 - packet.PacketHeaderSize) {
		conn.SetReadDeadline(time.Now().Add(time.Minute * 2))
		buf := make([]byte, packet.PacketHeaderSize+dataBodyLen)
		copy(buf, headBuf)
		_, err := io.ReadFull(reader, buf[packet.PacketHeaderSize:packet.PacketHeaderSize+dataBodyLen])
		if err != nil {
			return nil, err
		}
		return buf, nil
	}

	return nil, fmt.Errorf("recv packet(%v) reach max length:%v", msgId, dataBodyLen)

}

// 这里要换成内存池对象, 返回 int, []byte, error
func ReadPacket_(reader *bufio.Reader, buf []byte) (uint32, error) {
	// 先读包头
	if _, err := io.ReadFull(reader, buf[:packet.PacketHeaderSize]); err != nil {
		return 0, err
	}

	//msgId := binary.BigEndian.Uint32(buf[0:])
	dataBodyLen := binary.BigEndian.Uint32(buf[4:])

	if dataBodyLen <= (1024*1024 - packet.PacketHeaderSize) {
		_, err := io.ReadFull(reader, buf[packet.PacketHeaderSize:packet.PacketHeaderSize+dataBodyLen])
		if err != nil {
			return 0, err
		} else {
			return packet.PacketHeaderSize + dataBodyLen, nil
		}
	} else {
		log.Error("对端发来的数据超过了 1M 字节")
	}

	return 0, nil
}

func ProtoUnmarshal(buf []byte, pb proto.Message, pbMsgName string) bool {
	err := proto.Unmarshal(buf, pb)
	if err == nil {
		return true
	}

	log.Errorf("proto.Unmarshal 出错, error: %s, 消息体名字 %s", err.Error(), pbMsgName)

	return false
}

func ProtoMarshal(pb proto.Message, pbMsgName string) ([]byte, bool) {
	buf, err := proto.Marshal(pb)
	if err == nil {
		return buf, true
	}

	log.Errorf("proto.Marshal 出错, error: %s, 消息体名字 %s", err.Error(), pbMsgName)
	return nil, false
}

func Must(i interface{}, e error) interface{} {
	if e != nil {
		panic(e)
	}
	return i
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

//func Pb2Bytes(cmdId msg_id.NetMsgId, pb proto.Message) ([]byte, bool) {
//msg, err := proto.Marshal(pb)
//if err == nil {
//msgLen := len(msg)
//buf := make([]byte, packet.PacketHeaderSize+msgLen)
//binary.BigEndian.PutUint32(buf, uint32(cmdId))
//binary.BigEndian.PutUint32(buf[4:], uint32(msgLen))
//copy(buf[packet.PacketHeaderSize:], msg)
//return buf, true
//} else {
//return nil, false
//}
//}

// 生成 32 位 MD5
func Md5_32bit(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
