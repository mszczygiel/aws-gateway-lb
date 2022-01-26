package handler

import "github.com/google/gopacket"

type PacketHandleResult = int

const (
	PASS     PacketHandleResult = iota
	DROP     PacketHandleResult = iota
	MODIFIED PacketHandleResult = iota
)

type PacketHandler = func(packet *gopacket.Packet) PacketHandleResult

func DropUDPPacketsOnPortContainingPayload(port int, payload string) PacketHandler {
	return func(packet *gopacket.Packet) PacketHandleResult {
		return PASS
	}
}

func PassEveryNICMPPackets(n int) PacketHandler {
	return func(packet *gopacket.Packet) PacketHandleResult {
		return PASS
	}
}

func ReplacePayload(s string, replacement string) PacketHandler {
	return func(packet *gopacket.Packet) PacketHandleResult {
		return PASS
	}
}

func Handle(packet *gopacket.Packet, handlers []PacketHandler) PacketHandleResult {
	result := PASS
	for _, handler := range handlers {
		handlerResult := handler(packet)
		if handlerResult == DROP {
			return DROP
		}

		if handlerResult == MODIFIED {
			result = MODIFIED
		}
	}
	return result
}
