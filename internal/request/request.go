package request

import (
	"bytes"
	"fmt"
	"io"
	"proj/internal/headers"
	"strconv"
)

type parserState string

const (
	StateInit    parserState = "init"
	StateDone    parserState = "done"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
)

type Request struct {
	RequestLine    RequestLine
	Headers        *headers.Headers
	Body           []byte
	expectedLength int
	state          parserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
	}
}

var WrongRequest = fmt.Errorf("wrong request line")
var SEPARATOR = []byte("\r\n")

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, WrongRequest
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 {
		return nil, 0, WrongRequest
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}
	return rl, read, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		currentData := data[read:]

		switch r.state {
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n
			r.state = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				return 0, err
			}

			if n == 0 {
				break outer
			}

			read += n
			if done {
				r.state = StateBody
				str := r.Headers.Get("Content-Length")
				if str == "" {
					r.state = StateDone
					break outer
				}
				contentLenght, err := strconv.Atoi(str)
				if err != nil {
					return 0, fmt.Errorf("Wrong context-length")
				}
				r.expectedLength = contentLenght
			}

		case StateBody:
			if len(currentData) == 0 {
				break outer
			}

			remaining := min(r.expectedLength-len(r.Body), len(currentData))
			if len(currentData) > remaining {
				currentData = currentData[:remaining]
			}

			r.Body = append(r.Body, currentData...)
			read += remaining

			if r.expectedLength == len(r.Body) {
				r.state = StateDone
				break outer
			}

		case StateDone:
			break outer
		}
	}

	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += n

		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil
}
