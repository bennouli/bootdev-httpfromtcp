package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Accepted conntection from ", conn.RemoteAddr())

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for k, v := range req.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}
		fmt.Println("Body:")
		fmt.Println(string(req.Body))

		// fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
	}
}
