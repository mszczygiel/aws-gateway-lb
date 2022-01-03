package main

import (
	"log"
	"net"
	"time"
)

func SpamUdp() {
	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 6081})
	if err != nil {
		log.Panicf("failed to dial UDP spammer: %v", err)
	}
	for {
		_, err = conn.Write([]byte("Ping"))
		if err != nil {
			log.Printf("failed to spam UDP: %v", err)
		} else {
			log.Printf("written UDP packet")
		}
		time.Sleep(1 * time.Second)
	}
}
