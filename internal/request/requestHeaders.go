package request

import (
	"bytes"
	"httpfromtcp/internal/headers"
)

func parseHeaders(r *Request, data []byte) (done bool, numBytesParsed int, err error) {
	numBytesParsed = 0
	lines := bytes.Split(data, []byte(crlf))

	if r.Headers == nil {
		r.Headers = headers.NewHeaders()
	}

	for _, line := range lines {
		n, done, err := r.Headers.Parse(line)

		if err != nil || done == true {
			return done, n, err
		}

		numBytesParsed += n
	}

	return done, numBytesParsed, nil
}
