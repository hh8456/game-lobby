package connData

import (
	"servers/iface"
)

type ConnData struct {
	//Conn    iface.IConnection
	iface.IConnection
	BinData []byte
}
