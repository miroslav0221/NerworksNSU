package MulticastGroup

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

func ParseArguments() (string, int) {
	args := os.Args[1:]
	if len(args) < 1 && len(args) > 2 {
		return "", -1
	}

	port := -1
	if len(args) == 2 {
		numberPort, err := strconv.Atoi(args[1])
		if err != nil {
			return "", -1
		}
		port = numberPort
	}
	address := args[0]

	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Printf("Invalid address format: %v\n", err)
		return "", -1
	}

	if !addr.IP.IsMulticast() {
		fmt.Printf("Address is not multicast: %s\n", address)
		return "", -1
	}

	return address, port
}
