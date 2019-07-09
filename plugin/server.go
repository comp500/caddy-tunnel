package plugin

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/caddyserver/caddy/caddyhttp/httpserver"
)

// Server is a httpserver.Handler that handles TCP/UDP tunneling requests.
type Server struct {
	NextHandler   httpserver.Handler
	RequestPath   string // Temporary, remove when configuration system done
	Upstream      string
	UpstreamProto string
}

// Serves HTTP requests for tunneling. See httpserver.Handler.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method == http.MethodGet {
		if httpserver.Path(r.URL.Path).Matches(s.RequestPath) {
			fmt.Print("Got a http request") // TODO: Remove

			flusher, ok := w.(http.Flusher)
			if !ok {
				// TODO: evaluate whether this is fatal? do ResponseWriters without Flusher automatically flush?
				return http.StatusInternalServerError, errors.New("ResponseWriter does not implement Flusher")
			}

			if r.ProtoAtLeast(2, 0) { // HTTP2 and above supports streaming
				s.handleConnection(r.Body, w)
			} else if r.ProtoAtLeast(1, 0) { // HTTP1 requires a hijacker
				hijacker, ok := w.(http.Hijacker)
				if !ok {
					return http.StatusInternalServerError, errors.New("ResponseWriter does not implement Hijacker")
				}

				// Successful connection, must be sent and flushed before Hijack
				w.WriteHeader(200)
				flusher.Flush() // Flush the connection

				// Hijack connection
				// Ignore returned bufio.Reader, as client will only give more data after first response
				clientConn, _, err := hijacker.Hijack()
				if err != nil {
					return http.StatusInternalServerError, errors.New("failed to hijack: " + err.Error())
				}

				s.handleConnection(clientConn, clientConn) // net.Conn implements both interfaces
			} else {
				return http.StatusInternalServerError, errors.New("Invalid HTTP protocol in request")
			}

			return 0, nil
		}
	}
	return s.NextHandler.ServeHTTP(w, r)
}

type parseReturn struct {
	isControl bool
	Data      []byte
}

// Handles TCP/UDP connection within HTTP
func (s Server) handleConnection(r io.ReadCloser, w io.Writer) {
	proto := s.UpstreamProto
	if len(proto) == 0 {
		proto = "tcp"
	}
	conn, err := net.Dial(proto, s.Upstream)
	if err != nil {
		fmt.Print(err)
		r.Close() // Can we close the writer properly?
		return
	}

	done := make(chan struct{})
	done2 := make(chan struct{})
	go copyLog(conn, r, done)
	go copyLog(w, conn, done2)

	<-done2
}

func copyLog(dst io.Writer, src io.Reader, done chan struct{}) {
	i, err := io.Copy(dst, src)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Written %v", i)
	close(done)
}
