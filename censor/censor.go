package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
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

func swapSrcDstIPv4(layer *layers.IPv4) {
	dst := layer.DstIP
	layer.DstIP = layer.SrcIP
	layer.SrcIP = dst
}
func swapSrcDstUDP(layer *layers.UDP) {
	dst := layer.DstPort
	layer.DstPort = layer.SrcPort
	layer.SrcPort = dst
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("specify local address")
	}
	go listenHealthCheck(8080)

	laddr := net.ParseIP(os.Args[1])

	fmt.Println("Welcome to Censor")
	conn, err := net.ListenIP("ip4:17", &net.IPAddr{IP: laddr})
	if err != nil {
		log.Panicf("failed to start listening on %v. Error: %v", laddr, err)
	}

	log.Printf("local addr: %v", conn.LocalAddr())

	for {
		buffer := make([]byte, 8500)
		length, err := conn.Read(buffer)
		if err != nil {
			log.Panicf("failed to read UDP message %v", err)
		}
		if ipLayer := gopacket.NewPacket(buffer[:length], layers.LayerTypeIPv4, gopacket.Default); ipLayer != nil {
			log.Printf("layer: %v", ipLayer)
			packetLayers := ipLayer.Layers()
			if len(packetLayers) < 4 {
				log.Printf("packet has too few layers %v", ipLayer)
				continue
			}
			if packetLayers[0].LayerType() != layers.LayerTypeIPv4 || packetLayers[1].LayerType() != layers.LayerTypeUDP || packetLayers[2].LayerType() != layers.LayerTypeGeneve {
				log.Printf("packet layers are unsupported: %v", ipLayer)
				continue
			}
			udp := packetLayers[1].(*layers.UDP)
			if udp.DstPort != 6081 {
				log.Printf("unsupported UDP destination port %v: %v", udp.DstPort, ipLayer)
				continue
			}
			ip := packetLayers[0].(*layers.IPv4)
			swapSrcDstIPv4(ip)
			swapSrcDstUDP(udp)
			geneve := packetLayers[2].(*layers.Geneve)
			buf := gopacket.NewSerializeBuffer()
			opts := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
			var notModifiedLayers []gopacket.SerializableLayer
			for _, layer := range packetLayers[2:] {
				serLayer, ok := layer.(gopacket.SerializableLayer)
				if !ok {
					if !(layer.LayerType() == layers.LayerTypeGeneve) {
						log.Printf("layer not serializable %v: %v", layer.LayerType(), layer)
						continue
					}
				}
				notModifiedLayers = append(notModifiedLayers, serLayer)
			}
			gopacket.SerializeLayers(buf, opts, notModifiedLayers...)
			geneveBytes, err := buf.PrependBytes(len(geneve.Contents))
			if err != nil {
				log.Printf("failed to prepend geneve bytes: %v", err)
				continue
			}
			copy(geneveBytes, geneve.Contents)
			buf.PushLayer(geneve.LayerType())
			udp.SerializeTo(buf, opts)
			buf.PushLayer(udp.LayerType())
			ip.SerializeTo(buf, opts)
			buf.PushLayer(ip.LayerType())

			response := buf.Bytes()
			raddr := net.UDPAddr{IP: ip.DstIP, Port: int(udp.DstPort)}
			if err != nil {
				log.Printf("cannot dial %v: %v", raddr, err)
				continue
			}
			// respondTo, err := net.DialUDP("udp4", &net.UDPAddr{IP: laddr, Port: 6081}, &raddr)
			respondTo, err := net.DialIP("ip4:17", &net.IPAddr{IP: ip.SrcIP}, &net.IPAddr{IP: ip.DstIP})
			if err != nil {
				log.Printf("failed to Dial respondTo: %v", err)
				continue
			}
			written, err := respondTo.Write(response)
			respondTo.Close()
			if err != nil {
				log.Printf("failed to write response bytes: %v. %v, %v", err, ip, udp)

			} else {
				log.Printf("written %v response bytes to %v. source bytes: %v", written, raddr, length)
			}
		} else {
			log.Printf("packet not an IPv4 packet")
			continue
		}

		// packet, err := geneve.CreatePacket(length, buffer)
		// log.Printf("packet: %v", packet.String())
		// if err != nil {
		// 	log.Printf("ignoring packet due to unrecognized payload: %v", err)
		// 	continue
		// }

		// log.Printf("Got packet (from %v): %v -> %v F: %v", conn.RemoteAddr(), packet.SourceIP(), packet.DestinationIP(), packet.TcpHeaderFlags())
		// log.Printf("received length %v", length)

		// if !packet.HasPayload() {
		// todo connection caching?
		// conn, err := net.DialUDP("udp4", &addr, raddr)
		// if err != nil {
		// 	log.Printf("failed to connect for sending a response %v", err)
		// 	continue
		// }
		// written, _, err := conn.WriteMsgUDP(buffer[:length], nil, raddr)
		// if err != nil {
		// log.Printf("failed to send response packet %v", err)
		// continue
		// }
		// log.Printf("written %v bytes", written)

		// }

	}

}
