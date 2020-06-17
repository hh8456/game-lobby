package iface

type IMessage interface {
	GetDataLen() uint32
	GetMsgId() uint32
	GetData() []byte
	GetClientConnId() int64

	SetMsgId(uint32)
	SetData([]byte)
	SetClientConnId(int64)
}
