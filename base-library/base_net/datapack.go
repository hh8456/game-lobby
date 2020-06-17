package base_net

import (
	"bytes"
	"encoding/binary"
	"errors"
	"servers/iface"
)

type DataPack struct{}

func NewDataPack() *DataPack {
	return &DataPack{}
}

func (dp *DataPack) GetHeadLen() uint32 {
	return 16
}

func (dp *DataPack) Pack(msg iface.IMessage) ([]byte, error) {
	dataBuf := bytes.NewBuffer([]byte{})

	if err := binary.Write(dataBuf, binary.BigEndian, msg.GetClientConnId()); err != nil {
		return nil, err
	}

	if err := binary.Write(dataBuf, binary.BigEndian, msg.GetMsgId()); err != nil {
		return nil, err
	}

	if err := binary.Write(dataBuf, binary.BigEndian, msg.GetDataLen()); err != nil {
		return nil, err
	}

	if err := binary.Write(dataBuf, binary.BigEndian, msg.GetData()); err != nil {
		return nil, err
	}

	return dataBuf.Bytes(), nil
}

func (dp *DataPack) Unpack(binaryData []byte, maxDataLen uint32) (iface.IMessage, error) {
	dataBuf := bytes.NewReader(binaryData)

	msg := &Message{}

	if err := binary.Read(dataBuf, binary.BigEndian, &msg.ClientConnId); err != nil {
		return nil, err
	}

	if err := binary.Read(dataBuf, binary.BigEndian, &msg.Id); err != nil {
		return nil, err
	}

	if err := binary.Read(dataBuf, binary.BigEndian, &msg.DataLen); err != nil {
		return nil, err
	}

	if msg.DataLen < maxDataLen {
		if uint32(len(binaryData)) == dp.GetHeadLen()+msg.DataLen {
			msg.Data = binaryData[dp.GetHeadLen():]
			return msg, nil
		} else {
			return nil, errors.New("unpack fail, msgLen error")
		}
	} else {
		/*if maxDataLen < msg.DataLen */
		return nil, errors.New("too large msg data recieved")
	}
}

func (dp *DataPack) UnpackClientConnId(binaryData []byte) int64 {
	if uint32(len(binaryData)) < dp.GetHeadLen() {
		return 0
	}
	return int64(binary.BigEndian.Uint64(binaryData))
}

func (dp *DataPack) SetClientConnId(binaryData []byte, connId int64) {
	if uint32(len(binaryData)) >= dp.GetHeadLen() {
		binary.BigEndian.PutUint64(binaryData, uint64(connId))
	}
}

func (dp *DataPack) SetMsgId(binaryData []byte, msgId uint32) {
	if uint32(len(binaryData)) >= dp.GetHeadLen() {
		binary.BigEndian.PutUint32(binaryData[8:], msgId)
	}
}

func (dp *DataPack) UnpackMsgId(binaryData []byte) uint32 {
	if uint32(len(binaryData)) < dp.GetHeadLen() {
		return 0
	}

	return binary.BigEndian.Uint32(binaryData[8:])
}

func (dp *DataPack) SetMsgLen(binaryData []byte, msgLen uint32) {
	if uint32(len(binaryData)) >= dp.GetHeadLen() {
		binary.BigEndian.PutUint32(binaryData[12:], msgLen)
	}
}

func (dp *DataPack) UnpackMsgLen(binaryData []byte) (uint32, bool) {
	if uint32(len(binaryData)) < dp.GetHeadLen() {
		return 0, false
	}

	return binary.BigEndian.Uint32(binaryData[12:]), true
}
