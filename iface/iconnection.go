package iface

import (
	"servers/common-library/proto/errorCodeProto"
	"servers/common-library/proto/msgIdProto"

	"github.com/gogo/protobuf/proto"
)

type IConnection interface {
	Send([]byte)
	SendPbMsg(connId int64, pid msgIdProto.MsgId, pbMsg proto.Message)
	SendPbBuf(connId int64, pid msgIdProto.MsgId, msg []byte)
	SendMessage(IMessage)
	SendErrorCode(clientConnId int64, errCode errorCodeProto.ErrorCode)
	GetProperty(key string) interface{}
	SetProperty(key string, value interface{})
}
