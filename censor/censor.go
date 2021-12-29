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

	for {
		buffer := make([]byte, 8500)
		oob := make([]byte, 8500)
		length, _, _, _, err := conn.ReadMsgUDP(buffer, oob)
		if err != nil {
			panic(err)
		}
		packet, _ := geneve.CreatePacket(length, buffer)

		fmt.Println(packet.String())
		fmt.Println()
	}

}
