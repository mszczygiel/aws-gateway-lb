package handler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"golang.org/x/sys/unix"
)

const chatPort = 3000

type Handler struct {
	fd int
	mu sync.Mutex
	wg sync.WaitGroup
}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) Start(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.fd != 0 {
		return errors.New("already started")
	}

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_RAW, unix.IPPROTO_UDP)
	if err != nil {
		return fmt.Errorf("failed to create a RAW socket: %w", err)
	}

	err = unix.SetsockoptInt(fd, unix.IPPROTO_IP, unix.IP_HDRINCL, 1)
	if err != nil {
		return fmt.Errorf("failed to set IP_HDRINCL flag: %w", err)
	}

	h.fd = fd
	h.wg.Add(1)

	go h.run(ctx)

	return nil
}

func (h *Handler) run(ctx context.Context) {
	defer unix.Close(h.fd)
	defer h.wg.Done()

	buffer := make([]byte, 8500)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			length, raddr, err := unix.Recvfrom(h.fd, buffer, 0)
			if err != nil {
				log.Printf("failed to read UDP message %v", err)
				continue
			}
			packet, err := NewPacket(buffer[:length])
			if err != nil {
				log.Printf("Failed to create a packet: %s", err)
				continue
			}
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
			err = unix.Sendto(h.fd, response, 0, raddr)
			if err != nil {
				log.Printf("failed to write response: %v", err)
			} else {
				log.Printf("written %v response bytes. Source bytes: %v", len(response), length)
			}
		}
	}
}

func (h *Handler) Wait() {
	h.wg.Wait()
}

func replaceWeaklyTyped(b []byte) []byte {
	s := string(b)
	return []byte(strings.ReplaceAll(s, "weakly typed", "strongly typed"))
}
