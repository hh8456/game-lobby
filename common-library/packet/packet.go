package packet

// 逻辑包的包头
type PacketHeader struct {
	clientConnId uint32 //
	id           uint32 // 消息 id
	length       uint32 // data 的长度
}

const (
	// PacketHeader 的大小
	PacketHeaderSize     uint32 = 12
	C2SPacketMaxSize_16K uint32 = 16384                                   // 收到的客户端逻辑包最大长度, 包含包头
	PacketMaxSize_4MB    uint32 = 4 * 1024 * 1024                         // 服务器之间一次通信最多传送 4MB 的数据
	C2SPacketMaxBodySize uint32 = C2SPacketMaxSize_16K - PacketHeaderSize // 客户端逻辑包体的最大长度,去除包头
	BufSize_128byte      uint32 = 128
	BufSize_256byte      uint32 = 256
	BufSize_512byte      uint32 = 512
	BufSize_1024byte     uint32 = 1024
	BufSize_2K           uint32 = 2048
	BufSize_4K           uint32 = 4096
	BufSize_8K           uint32 = 8192
	BufSize_10K          uint32 = 10240
	BufSize_1MB          uint32 = 1024 * 1024
	BufSize_4MB          uint32 = 4 * 1024 * 1024
)

// 逻辑包结构体
type logicPacket struct {
	header PacketHeader
	data   []byte // 二进制消息
}
