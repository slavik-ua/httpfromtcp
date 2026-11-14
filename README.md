# Go HTTP Server (From Scratch)

A lightweight, custom HTTP/1.1 server implementation written in Go. This project bypasses the standard `net/http` server libraries to interact directly with TCP connections using the `net` package, providing a deep dive into how HTTP requests, responses, chunked encoding, and trailers work under the hood.

## 🚀 Features

*   **Zero-Dependency Server Logic**: Built on top of raw TCP sockets (`net` package).
*   **Custom Request Parsing**: Manually parses HTTP methods, targets, and headers.
*   **State-Aware Response Writer**: Enforces HTTP protocol order (Status Line -> Headers -> Body).
*   **Chunked Transfer Encoding**: Supports streaming responses.
*   **HTTP Trailers**: Implements trailer headers (e.g., for post-body checksums).
*   **Route Handling**: Basic router with support for static files, HTML responses, and proxying.
*   **Graceful Shutdown**: Handles OS signals (`SIGINT`, `SIGTERM`) to close listeners properly.

## 📂 Project Structure

```text
.
├── main.go                # Application entry point and route definitions
├── assets/
│   └── vim.mp4            # Sample video file for testing
└── internal/
    ├── headers/           # Custom header management
    ├── request/           # HTTP Request parser
    ├── response/          # HTTP Response writer & state machine
    └── server/            # TCP listener and connection handling
```

## 🛠️ Installation & Running
* Clone the repository (or ensure your files are organized as shown above).
* Add Assets: Create an assets folder and place a file named vim.mp4 inside it (required for the /video endpoint).
* Run the server:

```bash
go run ./cmd/httpserver
```

The server will start on port 42069.

```text
2026/02/23 10:00:00 Server started on port 42069
```

## 🔗 Endpoints

The [main.go](cmd/httpserver/main.go) file includes several demonstration handlers:
| Method | Request Target | Response Type | Behavior |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | `text/html` | **200 OK**: Returns a simple HTML success page. |
| `GET` | `/yourproblem` | `text/html` | **400 Bad Request**: Returns an HTML error page indicating client error. |
| `GET` | `/myproblem` | `text/html` | **500 Internal Server Error**: Returns an HTML error page indicating server error. |
| `GET` | `/video` | `video/mp4` | **Stream**: Streams `assets/vim.mp4` binary data. |
| `GET` | `/httpbin/{path}` | `chunked` | **Proxy**: Forwards request to `httpbin.org`. Streams response, calculates SHA256, and sends hash in **Trailers**. |

## 🧠 Architecture Highlights
### 1. TCP Listener (`internal/server`)

Instead of `http.ListenAndServe`, this project creates a `net.Listener`. It accepts incoming TCP connections and spawns a goroutine for each connection (`go s.handle(conn)`).

### 2. Request Parsing (`internal/request`)

The server reads directly from the `net.Conn`. It implements a state machine (`init` -> `headers` -> `body` -> `done`) to parse the raw byte stream into a structured `Request` object.

### 3. Safe Response Writer (`internal/response`)

The `Writer` struct ensures protocol compliance. It tracks the state of the response to prevent common errors (like writing headers after the body has started).

*   **States**: `StatusLine`, `Headers`, `Body`.
*   **Trailers**: Special support for writing HTTP trailers after a chunked body is finished.

### 4. The Proxy Handler (`/httpbin/`)

This is the most complex handler in the demo. It demonstrates:

1.  Initiating a client request to `httpbin.org`.
2.  Setting `Transfer-Encoding: chunked`.
3.  Reading the upstream body in a loop.
4.  Hashing the content (SHA256) while streaming.
5.  Writing the hash to the `Trailers` after the body is complete.

## ⚠️ Disclaimer

This is an educational project designed to understand the HTTP specification. It is not intended for production use. It lacks security hardening (timeouts, size limits, etc.) provided by the standard net/http library.
