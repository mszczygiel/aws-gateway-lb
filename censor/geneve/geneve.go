package geneve

import (
	"errors"
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
	PROTOCOL_OFFSET             = 9 + IP_HEADER_OFFSET
	SOURCE_IP_OFFSET            = 12 + IP_HEADER_OFFSET
	DESTINATION_IP_OFFSET       = 16 + IP_HEADER_OFFSET
	TCP_HEADER_OFFSET           = 20 + IP_HEADER_OFFSET
	TCP_HEADER_FLAGS_OFFSET     = 13 + TCP_HEADER_OFFSET
	PROTOCOL_TCP                = 6
)

type Packet struct {
	Data []byte
}

func CreatePacket(length int, data []byte) (packet Packet, err error) {
	if length < TCP_HEADER_OFFSET {
		err = errors.New("packet too short")
		return
	}

	// if data[PROTOCOL_OFFSET] != PROTOCOL_TCP {
	// 	err = fmt.Errorf("not supported protocol: %v", data[PROTOCOL_OFFSET])
	// 	return
	// }

	tmpPacket := Packet{
		Data: data[:length],
	}
	totalLength := tmpPacket.IpHeaderTotalLength()
	if length < IP_HEADER_OFFSET+totalLength {
		err = errors.New("packet too short")
		return
	}

	packet = tmpPacket
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
	for i < len(packet.Data) {
		size := min(4, len(packet.Data)-i)
		d := packet.Data[i : i+size]
		builder.WriteString(fmt.Sprintf("%v\n", d))
		i += size
	}
	return builder.String()
}

func (packet *Packet) IpHeaderTotalLength() int {
	return int(packet.Data[IP_HEADER_OFFSET+3])
}

func (packet *Packet) GetPayload() []byte {
	// todo
	return []byte{}
}

func (packet *Packet) HasPayload() bool {
	return len(packet.GetPayload()) > 0
}

func (packet *Packet) SetPayload(payload []byte) {
	// todo
}

func (packet *Packet) SourceIP() []byte {
	return packet.Data[SOURCE_IP_OFFSET : SOURCE_IP_OFFSET+4]
}

func (packet *Packet) DestinationIP() []byte {
	return packet.Data[DESTINATION_IP_OFFSET : DESTINATION_IP_OFFSET+4]
}

func (packet *Packet) TcpHeaderFlags() TcpHeaderFlags {
	return TcpHeaderFlags(packet.Data[TCP_HEADER_FLAGS_OFFSET])
}
