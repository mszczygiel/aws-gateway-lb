package handler

import (
	"context"
	"fmt"
	"log"
	"strings"

	"golang.org/x/sys/unix"
)

const chatPort = 3000

func Start(ctx context.Context) error {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_RAW, unix.IPPROTO_UDP)
	if err != nil {
		return fmt.Errorf("failed to create a RAW socket: %w", err)
	}
	defer unix.Close(fd)

	err = unix.SetsockoptInt(fd, unix.IPPROTO_IP, unix.IP_HDRINCL, 1)
	if err != nil {
		return fmt.Errorf("failed to set IP_HDRINCL flag: %w", err)
	}

	return run(ctx, fd)
}

func run(ctx context.Context, fd int) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			buffer := make([]byte, 8500)
			length, raddr, err := unix.Recvfrom(fd, buffer, 0)
			if err != nil {
				log.Printf("failed to read UDP message %v", err)
				continue
			}
			packet, err := NewPacket(buffer[:length])
			if err != nil {
				log.Printf("Failed to create a packet: %s", err)
				continue
			}
			log.Print(packet.String())
			packet.SwapSrcDstIPv4()

			// skip every 5th ICMP packet
			if icmpSeq := packet.ICMPSeq(); icmpSeq != nil {
				if *icmpSeq%5 == 0 {
					continue
				}
			}

			if packet.SrcPort() == chatPort || packet.DstPort() == chatPort {

				// drop, if payload contains string "drop me"
				if packet.PayloadContains("drop me") {
					continue
				}

				err = packet.ModifyUDP(replaceWeaklyTyped)
				if err != nil {
					log.Printf("failed to modify a UDP packet: %s", err)
					continue
				}
			}

			response, err := packet.Serialize()
			if err != nil {
				log.Printf("failed to serialize packet: %s", err)
				continue
			}
			err = unix.Sendto(fd, response, 0, raddr)
			if err != nil {
				log.Printf("failed to write response: %v", err)
			}
		}
	}
}

func replaceWeaklyTyped(b []byte) []byte {
	s := string(b)
	return []byte(strings.ReplaceAll(s, "weakly typed", "strongly typed"))
}
