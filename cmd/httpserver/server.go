package main

import (
	"bufio"
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type Server struct {
	status   atomic.Bool
	listener net.Listener
}

func Serve(port int) (*Server, error) {

	listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	server := Server{listener: listener, status: atomic.Bool{}}
	server.status.Store(true)

	go server.Listen()

	return &server, nil
}

func (s *Server) Close() {
	s.listener.Close()
}

func (s *Server) Listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.status.Load() == false {
				return
			}
			log.Fatalf("Could not accept connection %v", err)
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	_, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Println("not parsed?")
		if s.status.Load() == false {
			fmt.Println("not running ... ??")
			return
		}
		log.Fatalf("Could not read request %v", err)
	}

	w := bufio.NewWriter(conn)
	fmt.Fprintln(w, "HTTP/1.1 200 OK")
	fmt.Fprintln(w, "Content-Type: text/plain")
	fmt.Fprintln(w, "Connection: closed")
	fmt.Fprintln(w, "Content-Length: 13")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Hello World!")
	err = w.Flush()
	if err != nil && s.status.Load() == true {
		log.Fatalf("Could not read request %v", err)
	}

}
