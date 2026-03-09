package main

import (
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const port = 42069

func main() {
	server, err := Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, request request.Request) {
	var body []byte

	switch request.RequestLine.RequestTarget {
	case "/":
		body = []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
		w.WriteStatusLine(200)

	case "/yourproblem":
		body = []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
		w.WriteStatusLine(400)

		w.WriteBody([]byte(body))
	case "/myproblem":
		body = []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
		w.WriteStatusLine(500)
	}

	h := headers.NewHeaders()
	h.Set("Connection", "close")
	h.Set("Content-Length", strconv.Itoa(len(body)))
	h.Set("Content-Type", "text/html")
	w.WriteHeaders(h)

	w.WriteBody([]byte(body))
}
