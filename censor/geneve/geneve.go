package geneve

import (
	"fmt"
	"strings"
)

type TcpHeaderFlags byte

const (
	TCP_FLAG_FIN = 1 << iota
	TCP_FLAG_SYN = 1 << iota
	TCP_FLAG_RST = 1 << iota
	TCP_FLAG_PSH = 1 << iota
	TCP_FLAG_ACK = 1 << iota
	TCP_FLAG_URG = 1 << iota
	TCP_FLAG_ECE = 1 << iota
	TCP_FLAG_CWR = 1 << iota
)

const (
	OUTER_GENEVE_OPTIONS_OFFSET = 8
	IP_HEADER_OFFSET            = 32 + OUTER_GENEVE_OPTIONS_OFFSET
	TCP_HEADER_OFFSET           = 20 + IP_HEADER_OFFSET
	TCP_HEADER_FLAGS_OFFSET     = 13 + TCP_HEADER_OFFSET
)

type Packet struct {
	data []byte
}

func CreatePacket(length int, data []byte) (packet Packet, err error) {
	packet = Packet{
		data: data[:length],
	}
	return
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (packet *Packet) String() string {
	i := 0
	var builder strings.Builder
	for i < len(packet.data) {
		size := min(4, len(packet.data)-i)
		d := packet.data[i : i+size]
		builder.WriteString(fmt.Sprintf("%v\n", d))
		i += size
	}
	return builder.String()
}

func (packet *Packet) IpHeaderTotalLength() int {
	// todo
	return 0
}

func (packet *Packet) GetPayload() []byte {
	// todo
	return []byte{}
}

func (packet *Packet) SetPayload(payload []byte) {
	// todo
}

func (packet *Packet) TcpHeaderFlags() TcpHeaderFlags {
	return TcpHeaderFlags(packet.data[TCP_HEADER_FLAGS_OFFSET])
}
