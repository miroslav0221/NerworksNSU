package ServerTCP

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
)

const sizeint64 = 8

type FileInfo struct {
	sizeFile  int64
	sizeChunk int
	filename  string
	fd        *os.File
}

type Server struct {
	listenAddr   string
	listener     net.Listener
	quit         chan struct{}
	sizeNameFile int
	file         *FileInfo

	//msg        chan []byte
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr:   listenAddr,
		listener:     nil,
		quit:         make(chan struct{}),
		sizeNameFile: 1024,
		file:         nil,
		//msg:        make(chan []byte, 100),
	}
}

func NewFileInfo(filename string, fd *os.File, size int64) *FileInfo {
	return &FileInfo{
		filename:  filename,
		sizeChunk: 32 * 1024,
		fd:        fd,
		sizeFile:  size,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer listener.Close()
	s.listener = listener

	go s.accept()
	<-s.quit
	//close(s.msg)
	return nil
}

func (s *Server) accept() {
	for {
		connection, err := s.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection", err.Error())
			continue
		}
		fmt.Printf("Connection accepted : %s\n", connection.RemoteAddr().String())
		go s.reading(connection)
	}
}

func (s *Server) initFileStat(connection net.Conn) error {
	filename := s.readNameFile(connection)
	if filename == "" {
		return errors.New("failed read filename") //пытаться снова?
	}

	size := s.readSize(connection)
	if size == -1 {
		return errors.New("failed read size")
	}

	file, _ := os.Open(filename)

	fileInfo := NewFileInfo(filename, file, size)

	s.file = fileInfo

	return nil
}

func (s *Server) readNameFile(connection net.Conn) string {
	buff := make([]byte, s.sizeNameFile)
	n, err := connection.Read(buff)
	if err != nil {
		fmt.Println("Error reading name file", err.Error())
		return ""
	}
	fmt.Printf("n :\n", n)

	if n != s.sizeNameFile {
		fmt.Println("Error count bytes in namefile", n)
	}
	filename := string(bytes.TrimRight(buff, "\x00"))
	fmt.Println(filename)
	return filename
}

func (s *Server) readSize(connection net.Conn) int64 {
	buff := make([]byte, 8)
	n, err := connection.Read(buff)
	if n != sizeint64 || err != nil {
		return -1
	}
	size := int64(binary.BigEndian.Uint64(buff))
	return size
}

func (s *Server) reading(connection net.Conn) {
	defer connection.Close()

	for {

	}
}

func (s *Server) writingFile() {

}
