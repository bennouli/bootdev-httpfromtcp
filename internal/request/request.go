package request

import (
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type requestState = int

const crlf = "\r\n"
const bufferSize = 8
const (
	requestStateInitialized requestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{
		Headers: headers.NewHeaders(),
		state:   requestStateInitialized,
	}
	count := 0

	for req.state != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				req.state = requestStateDone
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead

		numBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}
		count++
		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed
	}

	contentLength := req.Headers.Get("content-length")
	if contentLength != "" {
		cl, err := strconv.Atoi(contentLength)
		if err != nil {
			return nil, err
		}
		if len(req.Body) < cl {
			return nil, fmt.Errorf("Body too short")
		}

	}

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0

	for r.state != requestStateDone {
		bytesParsed, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}

		if bytesParsed == 0 {
			return 0, nil
		}

		totalBytesParsed += bytesParsed
		if totalBytesParsed >= len(data[totalBytesParsed:]) {
			return totalBytesParsed, nil
		}

	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		requestLine, numBytesParsed, err := parseRequestLine(data)
		if err != nil {
			return numBytesParsed, err
		}
		if requestLine == nil {
			return 0, nil
		}

		r.state = requestStateParsingHeaders
		r.RequestLine = *requestLine
		return numBytesParsed, nil
	case requestStateParsingHeaders:

		numBytesParsed, done, err := r.Headers.Parse(data)
		if err != nil || done == true {
			r.state = requestStateParsingBody
			return numBytesParsed, err
		}

		return numBytesParsed, nil
	case requestStateParsingBody:
		contentLengthHeader := r.Headers.Get("content-length")
		var contentLength int = 0
		if contentLengthHeader != "" {
			parsedInt, err := strconv.Atoi(contentLengthHeader)
			if err != nil {
				return 0, fmt.Errorf("Malformed value for header 'Content-Length' %s", err)
			}
			contentLength = parsedInt

		}

		if contentLength == 0 || len(r.Body) == contentLength {
			r.state = requestStateDone
			return 0, nil
		}

		if len(r.Body) > contentLength {
			return 0, fmt.Errorf("Body exceeds content-length")
		}

		r.Body = append(r.Body, data...)

		return len(data), nil

	}

	return 0, nil
}
