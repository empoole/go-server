package main

import (
	"bytes"
	"compress/gzip"
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

	if(request.path == "/") {
		response = buildResponse(200, request.getEncoding(), "", "")
	} else if (strings.Contains(request.path, "/echo/")) {
		response = handleEcho(request)
	} else if (request.path == "/user-agent") {
		response = buildResponse(200, request.getEncoding(), "text/plain", request.Headers["user-agent"])
	} else if (strings.Contains(request.path, "/files/") && request.method == "GET") {
		response = handleGetFile(request)
	} else if (strings.Contains(request.path, "/files/") && request.method == "POST") {
		response = handlePostFile(request)
	} else {
		response = buildResponse(404, request.getEncoding(), "", "")
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

func (request *httpRequest) getEncoding() string {
	encoding := ""
	// This project currently only accepts gzip encoding
	if (strings.Contains(request.Headers["accept-encoding"], "gzip")) {
		encoding = "gzip"
	}
	return encoding
}

func gzipString(data string) string {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(data)); err != nil {
		fmt.Println("Error encoding request body: ", err.Error())
		os.Exit(1)
	}
	if err := gz.Close(); err != nil {
		fmt.Println("Error closing gzip writer: ", err.Error())
		os.Exit(1)
	}
	return string(b.Bytes())
}

func handleEcho(request *httpRequest) string {
	echo := strings.TrimPrefix(request.path, "/echo/")
	encoding := request.getEncoding()
	if(encoding == "gzip") {
		echo = gzipString(echo)
	}
	return buildResponse(200, encoding, "text/plain", echo)
}

func handleGetFile(request *httpRequest) string {
	dir := os.Args[2]
	fileName := strings.TrimPrefix(request.path, "/files/")
	encoding := request.getEncoding()
	data, err := os.ReadFile(dir + fileName)
	response := ""
	if err != nil {
		response = buildResponse(404, encoding, "", "")
	} else {
		response = buildResponse(200, encoding, "application/octet-stream", string(data))
	}
	return response
}

func handlePostFile(request *httpRequest) string {
	dir := os.Args[2]
	fileName := strings.TrimPrefix(request.path, "/files/")
	contents := bytes.Trim([]byte(request.Body), "\x00")
	err := os.WriteFile(dir + fileName, contents, 0644)
	encoding := request.getEncoding()
	response := ""
	if err != nil {
		response = buildResponse(404, encoding, "", "")
	} else {
		response = buildResponse(201, encoding, "", "")
	}
	return response
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