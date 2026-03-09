package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type StatusCode = int

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) {
	statusText := ""

	switch statusCode {
	case 200:
		statusText = "OK"
	case 400:
		statusText = "Bad Request"
	case 500:
		statusText = "Internal Server Error"
	}

	fmt.Fprintf(w, "HTTP/1.1 %v %s\r\n", statusCode, statusText)
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()

	h.Set("content-length", strconv.Itoa(contentLen))
	h.Set("connection", "close")
	h.Set("content-type", "text/plain")

	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		fmt.Fprintf(w, "%s: %s\r\n", k, v)
	}

	// add CRLF between headers and body
	fmt.Fprintln(w)

	return nil
}
