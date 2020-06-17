package iface

type IBackEndServer interface {
	Send([]byte)
}
