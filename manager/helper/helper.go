package helper

import (
	"strings"
)

// Helper to split "ip:port"
func SplitAddr(addr string) (ip, port string) {
	// Handles IPv4, IPv6, and edge cases
	lastColon := strings.LastIndex(addr, ":")
	if lastColon == -1 {
		return addr, ""
	}
	ip = addr[:lastColon]
	port = addr[lastColon+1:]
	return
}
