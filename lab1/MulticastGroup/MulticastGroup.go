package MulticastGroup

import (
	"errors"
	"fmt"
	"net"
	"time"
)

const timeDeadIP = 10
const sizeBuf = 1024

type MulticastGroup struct {
	address     *net.UDPAddr
	connIn      *net.UDPConn
	connOut     *net.UDPConn
	addresses   map[string]int
	currentTime int
	iface       *net.Interface
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

func (multicastGroup *MulticastGroup) Connect() error {
	var network string
	var connO *net.UDPConn
	var err error

	if multicastGroup.address.IP.To4() != nil {
		network = "udp4"
		connO, err = net.DialUDP(network, nil, multicastGroup.address)
		if err != nil {
			return fmt.Errorf("failed to dial UDP4: %w", err)
		}
	} else {
		network = "udp6"
		interfaces, err := net.Interfaces()
		if err != nil {
			return fmt.Errorf("error getting interfaces: %w", err)
		}

		for _, iface := range interfaces {
			if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagMulticast == 0 {
				continue
			}

			addrs, _ := iface.Addrs()
			for _, addr := range addrs {
				ip, _, _ := net.ParseCIDR(addr.String())
				if ip.To16() != nil && ip.To4() == nil {
					localAddr := &net.UDPAddr{
						IP:   net.IPv6unspecified,
						Zone: iface.Name,
					}
					connO, err = net.DialUDP(network, localAddr, multicastGroup.address)
					if err != nil {
						continue
					}
					multicastGroup.iface = &iface
					fmt.Printf("Using interface %s for IPv6 multicast\n", iface.Name)
					goto FOUND
				}
			}
		}
	FOUND:
		if connO == nil {
			return errors.New("no suitable IPv6 interface found for multicast")
		}
	}

	connI, err := net.ListenMulticastUDP(network, multicastGroup.iface, multicastGroup.address)
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
			fmt.Printf("Failed to read from UDP: %v\n", err)
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
				fmt.Printf("%s\n", address)
			}
		}
	}
}

func (multicastGroup *MulticastGroup) UpdatingTime() {
	start := time.Now()
	for {
		time.Sleep(time.Second * 1)
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
		time.Sleep(time.Second * 5)
		err := multicastGroup.SendMessage("I am alive\n")
		if err != nil {
			return
		}
	}
}
