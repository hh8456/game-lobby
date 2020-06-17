package base_net

type Message struct {
	ClientConnId int64  // gate 生成, 客户端的连接号
	Id           uint32 // 消息 id
	DataLen      uint32 // 消息长度
	Data         []byte // 消息内容
}

func NewMsgPackage(clientConnId int64, id uint32, data []byte) *Message {
	if data != nil {
		return &Message{Id: id,
			ClientConnId: clientConnId, DataLen: uint32(len(data)), Data: data}
	}

	return &Message{Id: id,
		ClientConnId: clientConnId, DataLen: uint32(len(data)), Data: []byte{}}
}

func (msg *Message) GetDataLen() uint32 {
	return msg.DataLen
}

func (msg *Message) GetMsgId() uint32 {
	return msg.Id
}

func (msg *Message) GetData() []byte {
	return msg.Data
}

func (msg *Message) GetClientConnId() int64 {
	return msg.ClientConnId
}

func (msg *Message) SetMsgId(msgId uint32) {
	msg.Id = msgId
}

func (msg *Message) SetData(data []byte) {
	msg.Data = data
}

func (msg *Message) SetClientConnId(clientConnId int64) {
	msg.ClientConnId = clientConnId
}
