package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"proj/internal/request"
	"proj/internal/response"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestProxyStreamsBodyAndTrailers(t *testing.T) {
	for _, payload := range [][]byte{nil, bytes.Repeat([]byte{0, 1, '\r', '\n', 0xff, 'x'}, 400)} {
		t.Run(fmt.Sprintf("%d_bytes", len(payload)), func(t *testing.T) {
			original := http.DefaultTransport
			t.Cleanup(func() { http.DefaultTransport = original })
			http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.String() != "https://httpbin.org/bytes/2400" {
					t.Fatalf("unexpected upstream URL: %s", req.URL)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(bytes.NewReader(payload)),
				}, nil
			})

			var wire bytes.Buffer
			handler(response.NewWriter(&wire), &request.Request{
				RequestLine: request.RequestLine{Method: "GET", RequestTarget: "/httpbin/bytes/2400"},
			})
			reader := bufio.NewReader(&wire)
			resp, err := http.ReadResponse(reader, nil)
			if err != nil {
				t.Fatalf("invalid HTTP response: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("unexpected status: %d", resp.StatusCode)
			}
			if len(resp.TransferEncoding) != 1 || resp.TransferEncoding[0] != "chunked" {
				t.Fatalf("unexpected transfer encoding: %v", resp.TransferEncoding)
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("invalid chunked body or trailers: %v", err)
			}
			if !bytes.Equal(body, payload) {
				t.Fatal("proxy changed the response body")
			}
			if got, want := resp.Trailer.Get("X-Content-Sha256"), fmt.Sprintf("%x", sha256.Sum256(payload)); got != want {
				t.Fatalf("checksum trailer = %q, want %q", got, want)
			}
			if got, want := resp.Trailer.Get("X-Content-Length"), fmt.Sprint(len(payload)); got != want {
				t.Fatalf("length trailer = %q, want %q", got, want)
			}
			if reader.Buffered() != 0 || wire.Len() != 0 {
				t.Fatal("unexpected bytes after the HTTP response")
			}
		})
	}
}

func TestProxyUpstreamFailure(t *testing.T) {
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })
	http.DefaultTransport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("upstream unavailable")
	})
	var wire bytes.Buffer
	handler(response.NewWriter(&wire), &request.Request{
		RequestLine: request.RequestLine{Method: "GET", RequestTarget: "/httpbin/get"},
	})
	resp, err := http.ReadResponse(bufio.NewReader(&wire), nil)
	if err != nil {
		t.Fatalf("invalid error response: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("unexpected error status: %d", resp.StatusCode)
	}
	if _, err := io.ReadAll(resp.Body); err != nil {
		t.Fatalf("incomplete error response: %v", err)
	}
}
