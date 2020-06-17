package iface

type IPlayer interface {
	SetGateId(gateId uint32)
	GetGateId() uint32
	GetUid() uint32
	SetConnId(connId int64)
	GetConnId() int64
	SetLastAliveTimestamp(timestamp int64)
	GetLastAliveTimestamp() int64
	Close()
	Timer(timestamp int64)
}
