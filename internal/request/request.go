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

	for req.state != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request, in state: %d, read %d bytes on EOF", req.state, numBytesRead)
				}
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead

		numBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed
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

		totalBytesParsed += bytesParsed

		if bytesParsed == 0 {
			break
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
		contentLenStr, ok := r.Headers.Get("content-length")
		if !ok {
			r.state = requestStateDone
			return len(data), nil
		}

		contentLen, err := strconv.Atoi(contentLenStr)
		if err != nil {
			return 0, fmt.Errorf("Malformed Content-Length %s", err)
		}
		r.Body = append(r.Body, data...)

		if len(r.Body) > contentLen {
			return 0, fmt.Errorf("Body exceeds content-length")
		}

		if len(r.Body) == contentLen {
			r.state = requestStateDone
		}

		return len(data), nil

	}

	return 0, nil
}
