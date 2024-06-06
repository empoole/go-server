package main

import (
	"bytes"
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

	response := ""

	encoding := ""
	if (strings.Contains(request.Headers["accept-encoding"], "gzip")) {
		encoding = "gzip"
	}

	if(request.path == "/") {
		response = buildResponse(200, encoding, "", "")
	} else if (strings.Contains(request.path, "/echo/")) {
		echo := strings.TrimPrefix(request.path, "/echo/")
		response = buildResponse(200, encoding, "text/plain", echo)
	} else if (request.path == "/user-agent") {
		response = buildResponse(200, encoding, "text/plain", request.Headers["user-agent"])
	} else if (strings.Contains(request.path, "/files/") && request.method == "GET") {
		dir := os.Args[2]
		fileName := strings.TrimPrefix(request.path, "/files/")
		data, err := os.ReadFile(dir + fileName)
		if err != nil {
			response = buildResponse(404, encoding, "", "")
		} else {
			response = buildResponse(200, encoding, "application/octet-stream", string(data))
		}
	} else if (strings.Contains(request.path, "/files/") && request.method == "POST") {
		dir := os.Args[2]
		fileName := strings.TrimPrefix(request.path, "/files/")
		contents := bytes.Trim([]byte(request.Body), "\x00")
		err := os.WriteFile(dir + fileName, contents, 0644)
		if err != nil {
			response = buildResponse(404, encoding, "", "")
		} else {
			response = buildResponse(201, encoding, "", "")
		}
	} else {
		response = buildResponse(404, encoding, "", "")
	}

	conn.Write(([]byte(response)))
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
		req.Headers[strings.ToLower(header[0])] = header[1]
	}

	return req
}

func buildResponse(code int, encoding string, contentType string, content string) string {
	responseCodes := map[int]string{
		200: "OK",
		201: "Created",
		404: "Not Found",
	}

	responseString := "HTTP/1.1 "
	responseString += fmt.Sprintf("%d %s", code, responseCodes[code])
	responseString += "\r\n"
	if(encoding != "") {
		responseString += "Content-Encoding: " + encoding
		responseString += "\r\n"
	}
	if (contentType != "") {
		responseString += "Content-Type: " + contentType
		responseString += "\r\n"
	}
	if (content != "") {
		responseString += fmt.Sprintf("Content-Length: %d", len(content))
		responseString += "\r\n"
	}
	responseString += "\r\n"
	if (content != "") {
		responseString += content
	}

	return responseString
}