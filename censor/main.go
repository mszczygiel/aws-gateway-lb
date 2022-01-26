package main

import (
	"fmt"
	"log"
	"mszczygiel.com/censor/handler"
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

var handlers = []handler.PacketHandler{
	handler.DropUDPPacketsOnPortContainingPayload(3000, "drop me"),
	handler.ReplacePayload("weakly typed", "strongly typed"),
	handler.PassEveryNICMPPackets(5),
}

func main() {
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
		recomputeChecksum := false
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
			if len(packetLayers) >= 5 {
				if packetLayers[4].LayerType() == layers.LayerTypeICMPv4 {
					icmp := packetLayers[4].(*layers.ICMPv4)
					if icmp.Seq%5 != 0 {
						log.Printf("DROP ICMP %v %v", icmp.TypeCode, icmp.Seq)
						continue
					}
				}
			}
			if packetLayers[0].LayerType() != layers.LayerTypeIPv4 || packetLayers[1].LayerType() != layers.LayerTypeUDP || packetLayers[2].LayerType() != layers.LayerTypeGeneve {
				continue
			}
			log.Print(packet.String())
			ip := packetLayers[0].(*layers.IPv4)
			swapSrcDstIPv4(ip)
			insideIPLayerIdx := len(packetLayers) - 3
			insideUDPLayerIdx := len(packetLayers) - 2
			if insideIP, ok := packetLayers[insideIPLayerIdx].(*layers.IPv4); ok {
				if insideUDP, ok := packetLayers[insideUDPLayerIdx].(*layers.UDP); ok {
					if insideUDP.SrcPort == CHAT_PORT || insideUDP.DstPort == CHAT_PORT {
						if payload, ok := packetLayers[len(packetLayers)-1].(*gopacket.Payload); ok {
							payloadStr := string(payload.Payload())
							if strings.Contains(strings.ToLower(payloadStr), "drop me") {
								log.Printf("Dropping message")
								continue
							}
							payloadStr = strings.ReplaceAll(payloadStr, "weakly typed", "strongly typed")
							recomputeChecksum = true
							insideUDP.SetNetworkLayerForChecksum(insideIP)
							packetLayers[len(packetLayers)-1] = gopacket.Payload([]byte(payloadStr))
						}
					}
				}
			}

			handleResult := handler.Handle(&packet, handlers)

			switch handleResult {
			case handler.DROP:
				continue
			case handler.MODIFIED:
				recomputeChecksum = true
			default:
			}

			buf := gopacket.NewSerializeBuffer()
			for i := len(packetLayers) - 1; i >= 0; i-- {
				if layer, ok := packetLayers[i].(gopacket.SerializableLayer); ok {
					var opts gopacket.SerializeOptions
					if recomputeChecksum && (i == insideUDPLayerIdx || i == insideIPLayerIdx) {
						opts = gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
					} else {
						opts = gopacket.SerializeOptions{FixLengths: true}
					}
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
