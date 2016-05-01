package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

var port string

func init() {
	flag.StringVar(&port, "port", "9999", "Port to run server on")
	if !strings.Contains(port, ":") {
		port = fmt.Sprintf(":%s", port)
	}
}

func main() {
	flag.Parse()
	l, err := net.Listen("tcp", "localhost"+port)
	if err != nil {
		log.Fatalf("Error listening on %s, %v", port, err)
	}
	log.Printf("Listening on %s", l.Addr())
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Unable to accept conn. %v", err)
		}
		go handleConnection(conn)
		log.Printf("Handling new connection: %s", conn.RemoteAddr())
	}
}

func handleConnection(c net.Conn) {
	defer c.Close()
	for {
		buf := make([]byte, 1)
		n, err := c.Read(buf)
		if err != nil && err != io.EOF {
			log.Printf("Unable to read bytes, %v", err)
		}
		if n > 0 {
			log.Printf("Read %d bytes, %v", n, buf)
			log.Printf("Message is %d", buf[0])
		}
	}
}
