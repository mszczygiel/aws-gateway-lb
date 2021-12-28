package geneve

import (
	"fmt"
	"strings"
)

type OuterEthernetHeader struct {
	data *[18]byte
}

type OutherIPv4Header struct {
	data *[20]byte
}

type OuterUDPHeader struct {
	data *[8]byte
}

type GeneveHeader struct {
	header  *[8]byte
	options []byte
}

type InnerEthernetHeader struct {
	data *[8]byte
}

type Payload struct {
	ethertype *[2]byte
	data      []byte
}

type FrameCheckSequence struct {
	data *[4]byte
}

type Packet struct {
	data []byte
}

func ParsePacket(length int, data []byte) (packet Packet, err error) {
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
	fmt.Println(len(packet.data))
	for i < len(packet.data) {
		size := min(4, len(packet.data)-i)
		fmt.Println(size)
		d := packet.data[i : i+size]
		builder.WriteString(fmt.Sprintf("%v\n", d))
		i += size
	}
	return builder.String()
}
