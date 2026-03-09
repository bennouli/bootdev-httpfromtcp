package main

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type Server struct {
	status   atomic.Bool
	listener net.Listener
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
		handler:  handler,
	}
	go server.listen()

	return server, nil
}

func (s *Server) Close() {
	s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.status.Load() {
				return
			}
			log.Printf("Error accepting connection %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		if s.status.Load() == false {
			fmt.Println("not running ... ??")
			return
		}
		return
	}

	handlerBuffer := bytes.NewBuffer([]byte{})

	handlerError := s.handler(handlerBuffer, *req)
	if handlerError != nil {
		handlerError.WriteTo(conn)
		return
	}

	response.WriteStatusLine(conn, 200)
	headers := response.GetDefaultHeaders(handlerBuffer.Len())
	response.WriteHeaders(conn, headers)
	conn.Write(handlerBuffer.Bytes())
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (err HandlerError) WriteTo(w io.Writer) {
	messageBytes := []byte(err.Message)

	response.WriteStatusLine(w, err.StatusCode)

	headers := response.GetDefaultHeaders(len(messageBytes))
	response.WriteHeaders(w, headers)

	w.Write(messageBytes)
}

type Handler func(w io.Writer, request request.Request) *HandlerError
