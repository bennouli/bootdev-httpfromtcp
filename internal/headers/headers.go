package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

const crlf = "\r\n"
const tcharWhiteList = "!#$%&'*+-.^_`|~"

/*
**
* field-line = "<field-name><:>_____<field-value>______"
 */
func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}

	// empty line, headers are done, consume the CRLF
	if idx == 0 {
		return 2, true, nil
	}

	delimiterIndex := bytes.Index(data[:idx], []byte(":"))
	if delimiterIndex < 1 {
		return 0, false, fmt.Errorf("Malformed header, missing delimiter %s", string(data))
	}
	// fmt.Println("Headers.parse parsing:", string(data[:delimiterIndex]))

	fieldName := strings.TrimLeft(string(data[:delimiterIndex]), " \t")

	for _, c := range fieldName {

		// is A-Z
		if c >= 65 && c <= 90 {
			continue
		}

		// is a-z
		if c >= 97 && c <= 122 {
			continue
		}

		// is digit
		if c >= 48 && c <= 57 {
			continue
		}

		// is special
		if strings.ContainsRune(tcharWhiteList, c) == true {
			continue
		}

		return 0, false, fmt.Errorf("Malformed header, invalid character in field-name %s", string(data))
	}

	if len(fieldName) < 1 {
		return 0, false, fmt.Errorf("Malformed header, empty field-name %s", string(data))
	}

	// ensure no trailing white space after field name
	if fieldName[len(fieldName)-1] == ' ' || fieldName[len(fieldName)-1] == '\t' {
		return 0, false, fmt.Errorf("Malformed header, trailing white space after field-name %s", string(data))
	}

	fieldValue := strings.Trim(string(data[delimiterIndex+1:idx]), " \t")

	// fmt.Println("Parsed:", idx+2, strings.ReplaceAll(string(data[:idx]), "\r\n", "_"))
	h.Set(fieldName, fieldValue)

	return idx + 2, false, nil

}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	v, ok := h[key]
	if ok {
		value = strings.Join([]string{
			v,
			value,
		}, ", ")
	}
	h[key] = value
}

func (h Headers) Get(key string) (string, bool) {
	v, ok := h[strings.ToLower(key)]
	return v, ok
}

func NewHeaders() Headers {
	headers := Headers{}
	return headers
}
