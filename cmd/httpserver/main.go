package main

import (
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
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

	if strings.HasPrefix(request.RequestLine.RequestTarget, "/httpbin") {
		parts := strings.Split(request.RequestLine.RequestTarget, "/")
		leaf := parts[len(parts)-1]

		// DETERMINE ROUTE
		url := "https://httpbin.org"
		if leaf == "html" {
			url += "/html"
		} else {
			_, err := strconv.Atoi(leaf)
			if err != nil {
				handleError(w, 400) // TODO: 404
				return
			}
			url += "/stream/" + leaf
		}

		resp, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			handleError(w, 500)
			return
		}
		err = w.WriteStatusLine(200)
		if err != nil {
			fmt.Println(err)
			return
		}

		h := headers.NewHeaders()
		trailers := headers.NewHeaders()

		// STATUS LINE + HEADERS
		switch leaf {
		case "html":
			h.Set("Host", "httpbin.org")
			h.Set("Content-Type", "text/html")
			h.Set("Transfer-Encoding", "chunked")
			h.Set("Trailer", http.CanonicalHeaderKey("X-Content-SHA256")+", "+http.CanonicalHeaderKey("X-Content-Length"))
			err = w.WriteHeaders(h)
			if err != nil {
				fmt.Println(err)
				return
			}

		default:
			h.Set("Host", "httpbin.org")
			h.Set("Content-Type", "application/json")
			h.Set("Transfer-Encoding", "chunked")

			err = w.WriteHeaders(h)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		// BODY
		totalBytesRead := 0
		b := make([]byte, 1024)
		done := false
		for !done {
			n, err := resp.Body.Read(b)
			totalBytesRead += n
			fmt.Println(n)
			if err != nil && err != io.EOF {
				fmt.Println("unexpected", err)
				w.WriteChunkedBodyDone()
				done = true
				break
			}

			body = append(body, b[:n]...)
			w.WriteChunkedBody(b[:n])

			if err == io.EOF {
				w.WriteChunkedBodyDone()
				done = true
			}
		}

		// TRAILERS
		if leaf == "html" {
			trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", sha256.Sum256(body)))
			trailers.Set("X-Content-Length", strconv.Itoa(totalBytesRead))
			fmt.Println(trailers)
		}

		err = w.WriteTrailers(trailers)
		if err != nil {
			fmt.Println("Error while writing trailers", err)
		}

		return
	}

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
		handleError(w, 500)
	case "/myproblem":
		handleError(w, 400)
	}

	h := headers.NewHeaders()
	h.Set("Connection", "close")
	h.Set("Content-Length", strconv.Itoa(len(body)))
	h.Set("Content-Type", "text/html")
	w.WriteHeaders(h)

	w.WriteBody([]byte(body))
}

func handleError(w *response.Writer, statusCode response.StatusCode) {
	w.WriteStatusLine(statusCode)
	switch statusCode {
	case 400:
		w.WriteStatusLine(400)
		w.WriteBody([]byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`))

	case 500:
		w.WriteStatusLine(500)
		w.WriteBody([]byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`))
	}
}
