package geneve

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var data = []byte{
	8, 0, 8, 0,
	0, 0, 0, 0,
	1, 8, 1, 2,
	54, 47, 198, 60,
	166, 242, 138, 65,
	1, 8, 2, 2,
	0, 0, 0, 0,
	0, 0, 0, 0,
	1, 8, 3, 1,
	137, 130, 221, 57,
	69, 0, 0, 60,
	172, 59, 64, 0,
	254, 6, 76, 27,
	192, 168, 1, 10,
	192, 168, 2, 10,
	229, 182, 11, 184,
	45, 234, 209, 153,
	0, 0, 0, 0,
	160, 2, 105, 3,
	76, 80, 0, 0,
	2, 4, 33, 197,
	4, 2, 8, 10,
	188, 80, 68, 243,
	0, 0, 0, 0,
	1, 3, 3, 7,
}

func TestCreatePacket(t *testing.T) {
	packet, _ := CreatePacket(len(data), data)

	assert.Equal(t, TcpHeaderFlags(2), packet.TcpHeaderFlags())

}
