package response

import (
	"fmt"
	"io"

	"proj/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var content []byte
	switch statusCode {
	case StatusOK:
		content = []byte("HTTP/1.1 200 OK\r\n")
	case StatusBadRequest:
		content = []byte("HTTP/1.1 400 Bad Request\r\n")
	case StatusInternalServerError:
		content = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	default:
		return fmt.Errorf("Unrecognized status code")
	}

	_, err := w.Write(content)
	return err
}

func GetDefaultHeaders(contentLen int) *headers.Headers {
	newHeaders := headers.NewHeaders()
	newHeaders.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	newHeaders.Set("Connection", "close")
	newHeaders.Set("Content-Type", "text/plain")

	return newHeaders
}

func WriteHeaders(w io.Writer, headers *headers.Headers) error {
	b := []byte{}
	for k, v := range headers.GetAllValues() {
		b = fmt.Appendf(b, "%s: %s\r\n", k, v)
	}
	b = fmt.Appendf(b, "\r\n")

	_, err := w.Write(b)
	return err
}
