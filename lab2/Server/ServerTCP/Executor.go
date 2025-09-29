package ServerTCP

import "fmt"

func Execute() {
	server := NewServer(":9000")

	err := server.Start()
	if err != nil {
		fmt.Println(err)
		return
	}
}
