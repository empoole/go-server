package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	
	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)

	if(err != nil) {
		fmt.Println("Error reading from connection buffer: ", err.Error())
		os.Exit(1)
	}

	request_parts := strings.Split(string(buffer), "\r\n")
	request_path := strings.Split(request_parts[0], " ")[1]

	if(request_path == "/") {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if (request_path[0:6] == "/echo/") {
		echo := request_path[6:]
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echo), echo)
		conn.Write([]byte(response))		
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}

}