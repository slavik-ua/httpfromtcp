package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"proj/internal/headers"
	"proj/internal/request"
	"proj/internal/response"
	"proj/internal/server"
)

const port = 42069

func handler(w *response.Writer, req *request.Request) {
	target := req.RequestLine.RequestTarget
	req_headers := headers.NewHeaders()

	if strings.HasPrefix(target, "/httpbin/") {
		value := strings.TrimPrefix(target, "/httpbin/")
		req_headers.Set("transfer-encoding", "chunked")
		req_headers.Set("trailer", "x-content-sha256")
		req_headers.Set("trailer", "x-content-length")

		resp, err := http.Get(fmt.Sprintf("https://httpbin.org/%s", value))
		if err != nil {
			return
		}
		defer resp.Body.Close()

		buf := make([]byte, 1024)
		hash := sha256.New()
		total := 0
		n, err := resp.Body.Read(buf)
		for n > 0 {
			_, err := w.WriteChunkedBody(buf[:n])
			if err != nil {
				return
			}
			hash.Write(buf[:n])
			total += n
			n, err = resp.Body.Read(buf)
		}
		w.WriteChunkedBodyDone()

		trailers := headers.NewHeaders()
		trailers.Set("x-contnet-sha256", fmt.Sprintf("%x", hash))
		trailers.Set("x-content-length", fmt.Sprintf("%d", total))
		w.WriteTrailers(trailers)

		return
	}

	var message []byte

	switch target {
	case "/yourproblem":
		w.WriteStatusLine(response.StatusBadRequest)
		req_headers.Set("content-type", "text/html")
		message = []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
`)

	case "/myproblem":
		w.WriteStatusLine(response.StatusInternalServerError)
		req_headers.Set("content-type", "text/html")
		message = []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
`)

	case "/":
		w.WriteStatusLine(response.StatusOK)
		req_headers.Set("content-type", "text/html")
		message = []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
`)
	case "/video":
		req_headers.Set("content-type", "video/mp4")

		var err error
		message, err = os.ReadFile("assets/vim.mp4")
		if err != nil {
			w.WriteStatusLine(response.StatusInternalServerError)
			return
		}
		w.WriteStatusLine(response.StatusOK)
	}

	req_headers.Set("content-length", fmt.Sprintf("%d", len(message)))
	w.WriteHeaders(req_headers)
	w.WriteBody(message)
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
