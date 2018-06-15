package plugin

import (
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// Server is a httpserver.Handler that handles TCP/UDP tunneling requests.
type Server struct {
	NextHandler httpserver.Handler
	RequestPath string // Temporary, remove when configuration system done
	Upstream    string
}

// Serves HTTP requests for tunneling. See httpserver.Handler.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method == http.MethodGet {
		if httpserver.Path(r.URL.Path).Matches(s.RequestPath) {
			log.Print("Got a http request") // TODO: Remove

			flusher, ok := w.(http.Flusher)
			if !ok {
				// TODO: evaluate whether this is fatal?
				return http.StatusInternalServerError, errors.New("ResponseWriter does not implement Flusher")
			}

			if r.ProtoMajor == 1 { // HTTP1 requires a hijacker
				hijacker, ok := w.(http.Hijacker)
				if !ok {
					return http.StatusInternalServerError, errors.New("ResponseWriter does not implement Hijacker")
				}
				// Hijack connection
				// Ignore returned bufio.Reader, as client will only give more data after first response
				clientConn, _, err := hijacker.Hijack()
				if err != nil {
					return http.StatusInternalServerError, errors.New("failed to hijack: " + err.Error())
				}

				// Successful connection
				w.WriteHeader(200)
				flusher.Flush() // Flush the connection

				go s.handleConnection(clientConn, clientConn) // net.Conn implements both interfaces
			} else if r.ProtoMajor >= 2 { // HTTP2 and above supports streaming
				// Successful connection
				w.WriteHeader(200)
				flusher.Flush() // Flush the connection

				go s.handleConnection(r.Body, w)
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

}
