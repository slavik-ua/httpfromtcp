package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"proj/internal/request"
	"proj/internal/server"
)

const port = 42069

func handler(w io.Writer, req *request.Request) *server.HandlerError {
	target := req.RequestLine.RequestTarget
	if target == "/yourproblem" {
		return &server.HandlerError{
			StatusCode: 400,
			Message:    []byte("Your problem is not my problem\n"),
		}
	}
	if target == "/myproblem" {
		return &server.HandlerError{
			StatusCode: 500,
			Message:    []byte("Woopsie, my bad\n"),
		}
	}

	w.Write([]byte("All good frfr\n"))

	return nil
}

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server stopped")
}
