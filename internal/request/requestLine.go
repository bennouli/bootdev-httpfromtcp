package request

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}

	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)

	if err != nil {
		return nil, 0, err
	}

	return requestLine, idx + 2, nil
}

func requestLineFromString(text string) (*RequestLine, error) {
	parts := strings.Split(text, " ")

	if len(parts) != 3 {
		return nil, errors.New("Invalid Structure")
	}

	method := parts[0]

	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return nil, errors.New("Invalid Method")
		}
	}

	requestTarget := parts[1]

	versionParts := strings.Split(parts[2], "/")
	if versionParts[0] != "HTTP" || len(versionParts) != 2 {
		return nil, fmt.Errorf("Malformed version %s", text)
	}

	if versionParts[1] != "1.1" {
		return nil, errors.New("Unsupported HttpVersion")
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   versionParts[1],
	}, nil
}
