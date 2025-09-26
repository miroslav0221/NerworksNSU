package MulticastGroup

import (
	"fmt"
	"net"
	"os"
)

func ParseArguments() string {
	args := os.Args[1:]
	if len(args) < 1 {
		return ""
	}

	address := args[0]

	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Printf("Invalid address format: %v\n", err)
		return ""
	}

	if !addr.IP.IsMulticast() {
		fmt.Printf("Address is not multicast: %s\n", address)
		return ""
	}

	return address
}
