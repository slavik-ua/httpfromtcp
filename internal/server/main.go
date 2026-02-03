package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"proj/internal/request"
	"proj/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    []byte
}

func (hErr *HandlerError) Write(w io.Writer) {
	response.WriteStatusLine(w, response.StatusCode(hErr.StatusCode))

	headers := response.GetDefaultHeaders(len(hErr.Message))
	response.WriteHeaders(w, headers)
	w.Write(hErr.Message)
}

type Server struct {
	port     int
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		handler:  handler,
		listener: listener,
	}

	go s.listen()

	return s, nil
}

func (s *Server) Close() error {
	s.listener.Close()
	s.closed.Store(true)
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("error: %v\n", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn io.ReadWriteCloser) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusBadRequest,
		}
		hErr.Write(conn)
		return
	}

	writer := response.NewWriter(conn)
	s.handler(writer, req)

	// buf := bytes.NewBuffer([]byte{})
	// hErr := s.handler(buf, req)
	// if hErr != nil {
	// 	hErr.Write(conn)
	// 	return
	// }

	// b := buf.Bytes()
	// response.WriteStatusLine(conn, response.StatusOK)
	// headers := response.GetDefaultHeaders(len(b))
	// response.WriteHeaders(conn, headers)
	// conn.Write(b)
}
