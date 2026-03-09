package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type StatusCode = int
type WriterStatus = int

const (
	WriterStatusStatusLine WriterStatus = iota
	WriterStatusHeaders
	WriterStatusBody
	WriterStatusDone
)

type Writer struct {
	writer io.Writer
	status WriterStatus
}

func (w *Writer) Write(p []byte) (int, error) {
	w.writer.Write(p)
	return len(p), nil
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.status != WriterStatusStatusLine {
		return fmt.Errorf("Unexpected status '%v' of writer when trying to write request line", w.status)
	}
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

	w.status = WriterStatusHeaders
	return nil
}
func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.status != WriterStatusHeaders {
		return fmt.Errorf("Unexpected status '%v' of writer when trying to write request line", w.status)
	}
	for k, v := range headers {
		fmt.Fprintf(w, "%s: %s\r\n", k, v)
	}

	// add CRLF between headers and body
	w.Write([]byte("\r\n"))

	w.status = WriterStatusBody
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.status != WriterStatusBody {
		return 0, fmt.Errorf("Unexpected status '%v' of writer when trying to write request line", w.status)
	}
	w.writer.Write(p)

	w.status = WriterStatusDone
	return len(p), nil
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		status: WriterStatusStatusLine,
	}
}

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)
