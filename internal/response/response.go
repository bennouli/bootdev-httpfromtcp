package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type StatusCode = int

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

type WriterStatus = int

const (
	WriterStatusStatusLine WriterStatus = iota
	WriterStatusHeaders
	WriterStatusBody
	WriterStatusTrailers
	WriterStatusDone
)

type Writer struct {
	writer io.Writer
	status WriterStatus
}

func (w *Writer) Write(p []byte) (int, error) {
	return w.writer.Write(p)
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
func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.status != WriterStatusHeaders {
		return fmt.Errorf("Unexpected status '%v' of writer when trying to write headers", w.status)
	}
	for k, v := range h {
		fmt.Fprintf(w, "%s: %s\r\n", k, v)
	}

	// add CRLF between headers and body
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	w.status = WriterStatusBody
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.status != WriterStatusBody {
		return 0, fmt.Errorf("Unexpected status '%v' of writer when trying to write body", w.status)
	}
	n, err := w.Write(p)

	w.status = WriterStatusTrailers

	return n, err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.status != WriterStatusBody {
		return 0, fmt.Errorf("Unexpected status '%v' of writer when trying to write body", w.status)
	}

	fmt.Fprintf(w, "%x\r\n", len(p))
	n, err := w.Write(append(p, '\r', '\n'))

	return n, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.status != WriterStatusBody {
		return 0, fmt.Errorf("Unexpected status '%v' of writer when trying to write body", w.status)
	}

	n, err := w.Write([]byte("0\r\n"))

	w.status = WriterStatusTrailers
	return n, err
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.status != WriterStatusTrailers {
		return fmt.Errorf("Unexpected status '%v' of writer when trying to write trailers", w.status)
	}
	fmt.Println("writing trailers")

	for k, v := range h {
		fmt.Printf("%s: %s\r\n", k, v)
		fmt.Fprintf(w, "%s: %s\r\n", k, v)
	}

	fmt.Println("closing CRLF")
	// add closing CRLF
	_, err := w.Write([]byte("\r\n"))

	w.status = WriterStatusDone
	return err
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		status: WriterStatusStatusLine,
	}
}
