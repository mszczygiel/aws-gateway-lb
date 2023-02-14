package handler

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type PayloadModifyFun func([]byte) []byte

type Packet struct {
	modified     bool
	packet       gopacket.Packet
	packetLayers []gopacket.Layer
}

func NewPacket(data []byte) (*Packet, error) {
	p := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)
	if p == nil {
		return nil, errors.New("invalid packet")
	}

	packetLayers := p.Layers()

	if len(packetLayers) < 4 {
		return nil, errors.New("packet has too few layers")
	}

	if packetLayers[0].LayerType() != layers.LayerTypeIPv4 || packetLayers[1].LayerType() != layers.LayerTypeUDP || packetLayers[2].LayerType() != layers.LayerTypeGeneve {
		return nil, errors.New("unexpected layers")
	}

	packet := &Packet{
		packet:       p,
		packetLayers: p.Layers(),
	}

	return packet, nil
}

func (p *Packet) ModifyUDP(f PayloadModifyFun) error {
	ip := p.insideIPv4()
	udp := p.insideUDP()
	if ip != nil && udp != nil {
		if payload, ok := p.packetLayers[len(p.packetLayers)-1].(*gopacket.Payload); ok {
			p.modified = true
			err := udp.SetNetworkLayerForChecksum(ip)
			if err != nil {
				return fmt.Errorf("failed SetNetworkLayerForChecksum: %w", err)
			}
			p.packetLayers[len(p.packetLayers)-1] = gopacket.Payload(f(payload.Payload()))
		}
	}
	return nil
}

func (p *Packet) ICMPSeq() *int {
	if len(p.packetLayers) >= 5 {
		if p.packetLayers[4].LayerType() == layers.LayerTypeICMPv4 {
			icmp := p.packetLayers[4].(*layers.ICMPv4)
			s := int(icmp.Seq)
			return &s
		}
	}

	return nil
}

func (p *Packet) SrcPort() int {
	if udp := p.insideUDP(); udp != nil {
		return int(udp.SrcPort)
	}
	return 0
}
func (p *Packet) DstPort() int {
	if udp := p.insideUDP(); udp != nil {
		return int(udp.DstPort)
	}
	return 0
}

func (p *Packet) PayloadContains(s string) bool {
	ip := p.insideIPv4()
	udp := p.insideUDP()
	if ip != nil && udp != nil {
		if payload, ok := p.packetLayers[len(p.packetLayers)-1].(*gopacket.Payload); ok {
			payloadStr := string(payload.Payload())
			return strings.Contains(payloadStr, s)
		}
	}

	return false
}

func (p *Packet) SwapSrcDstIPv4() {
	ip := p.packetLayers[0].(*layers.IPv4)
	dst := ip.DstIP
	ip.DstIP = ip.SrcIP
	ip.SrcIP = dst
}

func (p *Packet) Serialize() ([]byte, error) {
	buf := gopacket.NewSerializeBuffer()
	for i := len(p.packetLayers) - 1; i >= 0; i-- {
		if layer, ok := p.packetLayers[i].(gopacket.SerializableLayer); ok {
			var opts gopacket.SerializeOptions
			if p.modified && (i == p.insideUDPLayerIdx() || i == p.insideIPLayerIdx()) {
				opts = gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
			} else {
				opts = gopacket.SerializeOptions{FixLengths: true}
			}
			err := layer.SerializeTo(buf, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to serialize layer: %w", err)
			}
			buf.PushLayer(layer.LayerType())
		} else if layer, ok := p.packetLayers[i].(*layers.Geneve); ok {
			bytes, err := buf.PrependBytes(len(layer.Contents))
			if err != nil {
				log.Printf("failed to prepend geneve bytes: %v", err)
			}
			copy(bytes, layer.Contents)
		} else {
			return nil, fmt.Errorf("layer of unknown type: %v", p.packetLayers[i].LayerType())
		}
	}
	return buf.Bytes(), nil
}

func (p *Packet) insideUDP() *layers.UDP {
	insideUDPLayerIdx := p.insideUDPLayerIdx()
	if udp, ok := p.packetLayers[insideUDPLayerIdx].(*layers.UDP); ok {
		return udp
	}
	return nil
}

func (p *Packet) insideIPv4() *layers.IPv4 {
	insideIPLayerIdx := p.insideIPLayerIdx()
	if ip, ok := p.packetLayers[insideIPLayerIdx].(*layers.IPv4); ok {
		return ip
	}
	return nil
}

func (p *Packet) insideUDPLayerIdx() int {
	return len(p.packetLayers) - 2
}

func (p *Packet) insideIPLayerIdx() int {
	return len(p.packetLayers) - 3
}
