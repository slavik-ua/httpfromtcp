package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers struct {
	headers map[string]string
}

var rn = []byte("\r\n")

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func (h *Headers) Get(name string) string {
	return h.headers[strings.ToLower(name)]
}

func (h *Headers) Set(name, value string) {
	name = strings.ToLower(name)
	if v, ok := h.headers[name]; !ok {
		h.headers[name] = value
	} else {
		h.headers[name] = fmt.Sprintf("%s, %s", v, value)
	}
}

func (h *Headers) Replace(name, value string) {
	name = strings.ToLower(name)
	h.headers[name] = value
}

func (h *Headers) Delete(name string) {
	name = strings.ToLower(name)
	delete(h.headers, name)
}

func (h *Headers) GetAllValues() map[string]string {
	return h.headers
}

var WRONG_HEADER = fmt.Errorf("Wrong Header Line")
var EMPTY_HEADER = fmt.Errorf("Empty Header")

func isToken(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	for _, c := range data {
		found := false

		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			found = true
		}

		switch c {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			found = true
		}

		if !found {
			return false
		}
	}

	return true
}

func parseHeader(data []byte) (string, string, error) {
	parts := bytes.SplitN(data, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", EMPTY_HEADER
	}
	name := parts[0]
	value := bytes.TrimSpace(parts[1])

	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", WRONG_HEADER
	}

	return string(name), string(value), nil
}

// Host: localhost:42069\r\n\r\n
func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false

	for {
		idx := bytes.Index(data[read:], rn)
		if idx == -1 {
			return read, false, nil
		}

		if idx == 0 {
			read += len(rn)
			done = true
			break
		}

		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}
		if !isToken([]byte(name)) {
			return 0, false, WRONG_HEADER
		}
		h.Set(name, value)
		read += idx + len(rn)
	}

	return read, done, nil
}
