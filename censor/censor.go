package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/sys/unix"
)


const (
	CHAT_PORT = 3000
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
	defer listener.Close()

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

func swapSrcDstIPv4(layer *layers.IPv4) {
	dst := layer.DstIP
	layer.DstIP = layer.SrcIP
	layer.SrcIP = dst
}

func main() {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: 6081})
	if err != nil {
		log.Panicf("Failed to bind: %v", err)
	}
	defer conn.Close()
	go listenHealthCheck(8080)

	fmt.Println("Welcome to Censor")
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_RAW, unix.IPPROTO_UDP)
	if err != nil {
		log.Panicf("failed to create a RAW socket: %v", err)
	}
	defer unix.Close(fd)

	err = unix.SetsockoptInt(fd, unix.IPPROTO_IP, unix.IP_HDRINCL, 1)
	if err != nil {
		log.Panicf("failed to set IP_HDRINCL flag: %v", err)
	}

	for {
		buffer := make([]byte, 8500)
		length, raddr, err := unix.Recvfrom(fd, buffer, 0)
		if err != nil {
			log.Printf("failed to read UDP message %v", err)
			continue
		}
		if packet := gopacket.NewPacket(buffer[:length], layers.LayerTypeIPv4, gopacket.Default); packet != nil {
			packetLayers := packet.Layers()
			if len(packetLayers) < 4 {
				continue
			}
			if packetLayers[0].LayerType() != layers.LayerTypeIPv4 || packetLayers[1].LayerType() != layers.LayerTypeUDP || packetLayers[2].LayerType() != layers.LayerTypeGeneve {
				continue
			}
			log.Print(packet.String())
			ip := packetLayers[0].(*layers.IPv4)
			swapSrcDstIPv4(ip)
			ip.Checksum = 0
			if insideUDP, ok := packetLayers[len(packetLayers)-2].(*layers.UDP); ok {

				if insideUDP.SrcPort == CHAT_PORT || insideUDP.DstPort == CHAT_PORT {
					if payload, ok := packetLayers[len(packetLayers)-1].(*gopacket.Payload); ok {
						payloadStr := string(payload.Payload())
						if strings.Contains(strings.ToLower(payloadStr), "asap") {
							// drop ASAP messages
							log.Printf("Dropping ASAP message")
							continue
						}
					}
				}
			}
			buf := gopacket.NewSerializeBuffer()
			opts := gopacket.SerializeOptions{ComputeChecksums: false, FixLengths: false}
			for i := len(packetLayers) - 1; i >= 0; i-- {
				if layer, ok := packetLayers[i].(gopacket.SerializableLayer); ok {
					err := layer.SerializeTo(buf, opts)
					if err != nil {
						log.Printf("failed to serialize layer: %v", err)
					}
					buf.PushLayer(layer.LayerType())
				} else if layer, ok := packetLayers[i].(*layers.Geneve); ok {
					bytes, err := buf.PrependBytes(len(layer.Contents))
					if err != nil {
						log.Printf("failed to prepend geneve bytes: %v", err)
					}
					copy(bytes, layer.Contents)
				} else {
					log.Printf("layer of unknown type: %v", packetLayers[i].LayerType())
				}
			}
			response := buf.Bytes()
			err = unix.Sendto(fd, response, 0, raddr)
			if err != nil {
				log.Printf("failed to write response: %v", err)
			} else {
				log.Printf("written %v response bytes. Source bytes: %v", len(response), length)
			}
		} else {
			log.Printf("failed to create packet from bytes: %v", buffer[:length])
		}
	}

}
