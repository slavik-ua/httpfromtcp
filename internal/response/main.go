package response

import (
	"bytes"
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

type WriterState int

const (
	StatusLine WriterState = iota
	Headers
	Body
)

var WrongOrderError = fmt.Errorf("should write in order statusline, headers, body")

type Writer struct {
	writer io.Writer
	state  WriterState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		state:  StatusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != StatusLine {
		return WrongOrderError
	}
	w.state = Headers
	return WriteStatusLine(w.writer, statusCode)
}

func (w *Writer) WriteHeaders(headers *headers.Headers) error {
	if w.state != Headers {
		return WrongOrderError
	}
	w.state = Body
	return WriteHeaders(w.writer, headers)
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != Body {
		return 0, WrongOrderError
	}

	n, err := w.writer.Write(p)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	buf := bytes.NewBuffer([]byte{})
	buf.Write([]byte(fmt.Sprintf("%x", len(p)) + "\r\n"))
	buf.Write(p)
	buf.WriteString("\r\n")

	b := buf.Bytes()

	n, err := w.writer.Write(b)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	return w.writer.Write([]byte("0\r\n"))
}

func (w *Writer) WriteTrailers(h *headers.Headers) error {
	return WriteHeaders(w.writer, h)
}

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
