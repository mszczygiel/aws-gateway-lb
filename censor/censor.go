package main

import (
	"fmt"
	"log"
	"net"

	"mszczygiel.com/censor/geneve"
)

func listenHealthCheck(port int) {
	addr := net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: port,
	}
	listener, err := net.ListenTCP("tcp4", &addr)
	if err != nil {
		log.Panicf("failed to start listening for health check: %v", err)
	}

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}
		log.Printf("accepted health check connection from %v", conn.RemoteAddr())

		conn.Close()
	}
}

func main() {
	// if len(os.Args) != 2 {
	// 	log.Fatal("specify local address")
	// }
	go listenHealthCheck(8080)

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
	fmt.Printf("Local address: %v", conn.LocalAddr())

	for {
		buffer := make([]byte, 8500)
		oob := make([]byte, 8500)
		length, _, _, raddr, err := conn.ReadMsgUDP(buffer, oob)
		if err != nil {
			log.Panicf("failed to read UDP message %v", err)
		}
		packet, err := geneve.CreatePacket(length, buffer)
		if err != nil {
			log.Printf("ignoring packet due to unrecognized payload: %v", err)
			continue
		}

		log.Printf("Got packet (from %v): %v -> %v F: %v", raddr, packet.SourceIP(), packet.DestinationIP(), packet.TcpHeaderFlags())
		log.Printf("received length %v", length)

		// if !packet.HasPayload() {
		// todo connection caching?
		// conn, err := net.DialUDP("udp4", &addr, raddr)
		if err != nil {
			log.Printf("failed to connect for sending a response %v", err)
			continue
		}
		written, _, err := conn.WriteMsgUDP(buffer[:length], nil, raddr)
		if err != nil {
			log.Printf("failed to send response packet %v", err)
			continue
		}
		log.Printf("written %v bytes", written)

		// }

	}

}
