package MulticastGroup

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"
)

const (
	timeDeadIP      = 10
	sizeBuf         = 1024
	timeSendMessage = 5
)

type MulticastGroup struct {
	address     *net.UDPAddr
	connIn      *net.UDPConn
	connOut     *net.UDPConn
	network     string
	addresses   map[string]int
	currentTime int
	iface       *net.Interface
}

type InterfaceInfo struct {
	Interface *net.Interface
	Type      string
}

func (multicastGroup *MulticastGroup) GetInConn() *net.UDPConn {
	return multicastGroup.connIn
}

func (multicastGroup *MulticastGroup) GetOutConn() *net.UDPConn {
	return multicastGroup.connIn
}

func (multicastGroup *MulticastGroup) GetAddress() string {
	return multicastGroup.address.String()
}

func NewMulticastGroup(address string) (*MulticastGroup, error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Printf("Failed to resolve UDP address: %v\n", err)
		return nil, err
	}
	m := make(map[string]int)
	return &MulticastGroup{
		address:     addr,
		connIn:      nil,
		connOut:     nil,
		addresses:   m,
		currentTime: 0,
		iface:       nil,
	}, nil
}

func (multicastGroup *MulticastGroup) getInterfaces() ([]InterfaceInfo, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var suitable []InterfaceInfo

	for _, ifi := range interfaces {
		if ifi.Flags&(net.FlagUp|net.FlagMulticast) != (net.FlagUp|net.FlagMulticast) ||
			ifi.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := ifi.Addrs()
		if err != nil {
			continue
		}

		var hasIPv4, hasIPv6 bool

		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok || ipnet.IP.IsLoopback() {
				continue
			}

			if ipnet.IP.To4() != nil {
				hasIPv4 = true
			} else {
				hasIPv6 = true
			}
		}

		var ifiType string
		var isCompatible bool

		if multicastGroup.network == "udp4" && hasIPv4 {
			isCompatible = true
			ifiType = "IPv4"
		} else if multicastGroup.network == "udp6" && hasIPv6 {
			isCompatible = true
			ifiType = "IPv6"
		}

		if hasIPv4 && hasIPv6 {
			ifiType = "IPv4/IPv6"
		}

		if isCompatible {
			suitable = append(suitable, InterfaceInfo{
				Interface: &ifi,
				Type:      ifiType,
			})
		}
	}

	sort.Slice(suitable, func(i, j int) bool {
		return suitable[i].Interface.Name < suitable[j].Interface.Name
	})

	return suitable, nil
}

func (multicastGroup *MulticastGroup) selectMulticastInterface() (*net.Interface, error) {
	interfaces, err := multicastGroup.getInterfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get available interfaces: %v", err)
	}

	if len(interfaces) == 0 {
		return nil, fmt.Errorf("no suitable multicast interfaces found for %s", multicastGroup.network)
	}

	if len(interfaces) == 1 {
		selected := interfaces[0].Interface
		fmt.Printf("Automatically selected interface: %s\n", selected.Name)
		fmt.Printf("Press Enter to continue...")
		return selected, nil
	}

	fmt.Println("=== Network Interface Selection ===")
	fmt.Printf("Protocol: %s | Multicast Address: %s\n", multicastGroup.network, multicastGroup.address.String())
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	fmt.Printf("%-5s %-20s %-15s\n", "№", "Interface", "Type")
	fmt.Println(strings.Repeat("-", 50))

	for i, ifaceInfo := range interfaces {
		fmt.Printf("%-5d %-20s %-15s\n",
			i+1, ifaceInfo.Interface.Name, ifaceInfo.Type)
	}

	fmt.Println(strings.Repeat("-", 50))

	var selectedInterface *net.Interface
	selectedInterface = interfaces[0].Interface
	fmt.Printf("Selected Interface: %s\n", selectedInterface.Name)

	return selectedInterface, nil
}

func (multicastGroup *MulticastGroup) Connect(port int) error {
	var connO *net.UDPConn
	var err error

	if multicastGroup.address.IP.To4() != nil {
		multicastGroup.network = "udp4"
	} else {
		multicastGroup.network = "udp6"
	}

	ifi, err := multicastGroup.selectMulticastInterface()
	if err != nil {
		fmt.Printf("Failed to select multicast interface: %v", err)
		return err
	}

	multicastGroup.iface = ifi

	if multicastGroup.network == "udp6" && ifi != nil {
		multicastGroup.address.Zone = ifi.Name
	}

	if port == -1 {

		connO, err = net.DialUDP(multicastGroup.network, nil, multicastGroup.address)

	} else {

		localAddr := &net.UDPAddr{
			IP:   net.IPv4zero,
			Port: port,
		}
		connO, err = net.DialUDP(multicastGroup.network, localAddr, multicastGroup.address)
		if err != nil {
			fmt.Printf("Failed to connect to multicast interface: %v", err)
			return err
		}
	}

	connI, err := net.ListenMulticastUDP(multicastGroup.network, multicastGroup.iface, multicastGroup.address)
	if err != nil {
		return fmt.Errorf("failed to listen multicast: %w", err)
	}

	connI.SetReadBuffer(sizeBuf)

	multicastGroup.connIn = connI
	multicastGroup.connOut = connO
	multicastGroup.addresses[multicastGroup.connOut.LocalAddr().String()] = multicastGroup.currentTime

	return nil
}

func (multicastGroup *MulticastGroup) Disconnect() error {
	if multicastGroup.connIn != nil {
		_ = multicastGroup.connIn.Close()
	}
	if multicastGroup.connOut != nil {
		_ = multicastGroup.connOut.Close()
	}
	return nil
}

func (multicastGroup *MulticastGroup) ReceiveMessage() {
	if multicastGroup.connIn == nil {
		fmt.Printf("MulticastGroup connection is nil\n")
		return
	}

	buffer := make([]byte, sizeBuf)
	fmt.Printf("MulticastGroup receiving message...\n")

	for {
		_, srcAddr, err := multicastGroup.connIn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("Failed to read from UDP")
			return
		}
		multicastGroup.addresses[srcAddr.String()] = multicastGroup.currentTime
	}
}

func (multicastGroup *MulticastGroup) CheckingAlliveIp() {
	for {
		time.Sleep(timeDeadIP * time.Second)
		fmt.Printf("--------------------------------------\n")
		fmt.Printf("Alive ip-addresses in multicast group\n")
		for address, value := range multicastGroup.addresses {
			if multicastGroup.currentTime-value < timeDeadIP {
				fmt.Printf("%s  ✅\n", address)
			}
		}
	}
}

func (multicastGroup *MulticastGroup) UpdatingTime() {
	start := time.Now()
	for {
		time.Sleep(time.Second)
		multicastGroup.currentTime = int(time.Since(start).Seconds())
	}
}

func (multicastGroup *MulticastGroup) SendMessage(message string) error {
	if multicastGroup.connOut == nil {
		return errors.New("connection failed")
	}
	_, err := multicastGroup.connOut.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write to UDP: %w", err)
	}
	return nil
}

func (multicastGroup *MulticastGroup) SendingMessageToGroup() {
	for {
		time.Sleep(time.Second * timeSendMessage)
		err := multicastGroup.SendMessage("I am alive\n")
		if err != nil {
			return
		}
	}
}
