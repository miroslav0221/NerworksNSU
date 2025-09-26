package MulticastGroup

import (
	"fmt"
	"time"
)

func Execute() {
	address := ParseArguments()

	if address == "" {
		fmt.Printf("Usage: multicast-group <multicast address>\n")
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

	err = multicastGroup.Connect()
	if err != nil {
		fmt.Printf("Failed to connect to MulticastGroup: %v\n", err)
		return
	}

	fmt.Printf("My address: %s\n", multicastGroup.GetOutConn().LocalAddr().String())
	fmt.Printf("Multicast group: %s\n", multicastGroup.GetAddress())

	go multicastGroup.ReceiveMessage()
	time.Sleep(3 * time.Second)
	go multicastGroup.SendingMessageToGroup()

	fmt.Println("Press Enter to exit...")
	var input string
	fmt.Scanln(&input)
}
