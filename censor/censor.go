package main

import (
	"fmt"
	"net"

	"mszczygiel.com/censor/geneve"
)

func main() {
	addr := net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 6081,
	}
	fmt.Println("Welcome to Censor")
	conn, err := net.ListenUDP("udp4", &addr)
	if err != nil {
		fmt.Printf("failed to start listening on %v. Error: %v", addr, err)
		panic(err)
	}

	s := make([]byte, 10)
	var arr *[10]byte = (*[10]byte)(s[0:10])
	arr[2] = 5
	fmt.Printf("%v", arr)
	fmt.Printf("%v", s)

	for {
		buffer := make([]byte, 8500)
		oob := make([]byte, 8500)
		length, _, _, _, err := conn.ReadMsgUDP(buffer, oob)
		if err != nil {
			panic(err)
		}
		packet, _ := geneve.ParsePacket(length, buffer)

		fmt.Println(packet.String())
		fmt.Println()
	}

}
