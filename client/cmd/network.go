package cmd

import (
	"fmt"
	"log"
	"net"
	"net/netip"
)

func getIPInSubnet(subnet string) (string, error) {
	log.Printf(
		"Parsing subnet: %s",
		subnet,
	)
	target, err := netip.ParsePrefix(subnet) // Parses IP written with CIDR notation
	if err != nil {
		return "", fmt.Errorf("invalid advertise subnet %q: %w", subnet, err)
	}
	target = target.Masked() // Normalizes something like 192.168.5.17/24 to 192.168.5.0/24

	if !target.Addr().Is4() {
		return "", fmt.Errorf("advertise subnet must be IPv4: %s", target)
	}

	interfaces, err := net.Interfaces() // Every network adapter
	// In general, lab PCs will have a "real internet" network adapter and an "lab LAN" adapter
	if err != nil {
		return "", fmt.Errorf("list network interfaces: %w", err)
	}

	var matches []netip.Addr

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 ||
			iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addresses, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, address := range addresses {
			localPrefix, err := netip.ParsePrefix(address.String())
			if err != nil {
				continue
			}

			ip := localPrefix.Addr().Unmap()
			if ip.Is4() && target.Contains(ip) {
				matches = append(matches, ip)
			}
			// Check if we have an IP address that matches
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no local address found in subnet %s", target)
	case 1:
		return matches[0].String(), nil
	default:
		return "", fmt.Errorf("multiple local addresses found in subnet %s: %v", target, matches)
	}
}
