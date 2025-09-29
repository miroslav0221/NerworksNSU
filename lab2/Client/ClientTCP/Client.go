package ClientTCP

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
)

type FileInfo struct {
	sizeFile  int64
	sizeChunk int
	filename  string
	fd        *os.File
}
type Client struct {
	serverAddr string
	conn       net.Conn
	chunkFiles chan []byte
	file       *FileInfo
}

func NewFileInfo(filename string, fd *os.File, size int64) *FileInfo {
	return &FileInfo{
		filename:  filename,
		sizeChunk: 32 * 1024,
		fd:        fd,
		sizeFile:  size,
	}
}

func NewClient(serverAddr string) *Client {
	return &Client{serverAddr: serverAddr,
		conn:       nil,
		file:       nil,
		chunkFiles: make(chan []byte, 100),
	}
}

func (c *Client) Connect() error {
	conn, err := net.Dial("tcp", c.serverAddr)
	if err != nil {
		fmt.Println("Failed to connect to server")
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) Close() {
	err := c.conn.Close()
	if err != nil {
		fmt.Println("Failed to close connection")
	}
}

func (c *Client) ReadingFile() {
	defer c.file.fd.Close()
	var countBytes int64
	countBytes = 0

	for countBytes < c.file.sizeFile {
		buf := make([]byte, c.file.sizeChunk)
		n, err := c.file.fd.Read(buf)
		if err != nil {
			println("Failed to read file")
			continue
		}
		c.chunkFiles <- buf[:n]
		countBytes += int64(n)
	}
}

func (c *Client) initFileStat(namefile string) {
	file, err := os.Open(namefile)
	if err != nil {
		fmt.Println("Failed to open file")
		return
	}
	stat, _ := file.Stat()
	size := stat.Size()
	fileInfo := NewFileInfo(namefile, file, size)
	c.file = fileInfo
}

func (c *Client) sendNamefile() error {
	buf := make([]byte, 4096)

	copy(buf, []byte(c.file.filename))
	n, err := c.conn.Write(buf)
	if err != nil {
		fmt.Println("Failed to sead file name")
		return err
	}
	if n != len(buf) {
		fmt.Println("Failed to send file name")
		err := c.sendNamefile()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) sendSize() error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(c.file.sizeFile))
	_, err := c.conn.Write(buf)
	if err != nil {
		fmt.Println("Failed to send file size")
		return err
	}
	return nil
}

func (c *Client) SendingFile(namefile string) {
	c.initFileStat(namefile)

	err := c.sendNamefile()
	if err != nil {
		return
	}

	err = c.sendSize()
	if err != nil {
		return
	}

	go c.ReadingFile()

	for chunk := range c.chunkFiles {
		totalSent := 0
		for totalSent < len(chunk) {
			n, err := c.conn.Write(chunk[totalSent:])
			if err != nil {
				fmt.Println("Failed to send chunk:", err)
				return
			}
			totalSent += n
		}
		fmt.Printf("Wrote %d bytes\n", totalSent)
	}

}
