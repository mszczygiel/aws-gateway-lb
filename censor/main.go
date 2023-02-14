package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/mszczygiel/aws-gateway-lb/censor/handler"
)

func listenHealthCheck(ctx context.Context, port int) {
	addr := net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: port,
	}
	listener, err := net.ListenTCP("tcp4", &addr)
	if err != nil {
		log.Panicf("failed to start listening for health check: %v", err)
	}
	defer listener.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := listener.AcceptTCP()
			if err != nil {
				log.Printf("failed to accept connection: %v", err)
				continue
			}
			log.Printf("accepted health check connection from %v", conn.RemoteAddr())

			conn.Close()
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)
	go func() {
		<-stopChan
		log.Printf("stopping")
		cancel()
	}()
	go listenHealthCheck(ctx, 8080)

	fmt.Println(
		`Welcome to Censor.
	I'm supposed to run as a virtual appliance for AWS Gateway Load Balancer.
	I drop every 5th ICMP packet, UDP packets containing "drop me" in the payload.
	I also replace every "weakly typed" with "strongly typed"`)

	h := handler.New()
	err := h.Start(ctx)
	if err != nil {
		log.Panicf("failed to start the handler: %v", err)
	}

	h.Wait()
	log.Printf("DONE")

}
