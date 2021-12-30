package main

import (
	"fmt"
	"log"
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
		length, _, _, raddr, err := conn.ReadMsgUDP(buffer, oob)
		if err != nil {
			log.Panicf("failed to read UDP message %v", err)
		}
		packet, err := geneve.CreatePacket(length, buffer)
		if err != nil {
			log.Printf("ignoring packet due to unrecognized payload %v", err)
			continue
		}

		log.Printf("Got packet (from %v): %v -> %v F: %v", raddr, packet.SourceIP(), packet.DestinationIP(), packet.TcpHeaderFlags())

		// if !packet.HasPayload() {
		// todo connection caching?
		conn, err := net.DialUDP("udp4", nil, raddr)
		if err != nil {
			log.Printf("failed to connect for sending a response %v", err)
			continue
		}
		_, _, err = conn.WriteMsgUDP(packet.Data, nil, nil)
		if err != nil {
			log.Printf("failed to send response packet %v", err)
			continue
		}

		// }

	}

}
