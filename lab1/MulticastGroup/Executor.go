package MulticastGroup

import (
	"fmt"
	"time"
)

const delay = 3

func Execute() {
	address, port := ParseArguments()

	if address == "" && port == -1 {
		fmt.Printf("Usage: multicast-group <multicast address> <port>\n")
		return
	}

	multicastGroup, err := NewMulticastGroup(address)
	if err != nil {
		fmt.Printf("Failed to create MulticastGroup: %v\n", err)
		return
	}

	defer multicastGroup.Disconnect()

	go multicastGroup.UpdatingTime()
	go multicastGroup.CheckingAlliveIp()

	err = multicastGroup.Connect(port)
	if err != nil {
		fmt.Printf("Failed to connect to MulticastGroup: %v\n", err)
		return
	}

	fmt.Printf("Multicast group: %s\n", multicastGroup.GetAddress())

	go multicastGroup.ReceiveMessage()
	time.Sleep(delay * time.Second)
	go multicastGroup.SendingMessageToGroup()

	fmt.Println("Press Enter to exit...")
	var input string
	fmt.Scanln(&input)
}
