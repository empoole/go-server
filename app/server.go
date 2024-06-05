package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type httpRequest struct {
	method string
	Headers map[string]string
	path string
	Body string
}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)

	if(err != nil) {
		fmt.Println("Error reading from connection buffer: ", err.Error())
		os.Exit(1)
	}

	request := new(httpRequest)
	request.parseRequest(string(buffer))

	if(request.path == "/") {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if (request.path[0:6] == "/echo/") {
		echo := request.path[6:]
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echo), echo)
		conn.Write([]byte(response))
	} else if (request.path == "/user-agent") {
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(request.Headers["User-Agent"]), request.Headers["User-Agent"])
		conn.Write(([]byte(response)))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

func (req *httpRequest) parseRequest(requestString string) *httpRequest {
	requestParts := strings.Split(string(requestString), "\r\n")

	req.method = strings.Split(requestParts[0], " ")[0]
	req.path = strings.Split(requestParts[0], " ")[1]
	req.Headers = make(map[string]string)

	for i := 1; i < len(requestParts); i++ {
		if requestParts[i] == "" {
			req.Body = requestParts[i+1]
			break
		}
		header := strings.Split(requestParts[i], ": ")
		req.Headers[header[0]] = header[1]
	}

	return req
}