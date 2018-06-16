package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
)

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	fmt.Println("Client is WIP, not implemented yet.")

	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Listening on port 8080")
	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(c net.Conn) {
			// Shut down the connection.
			defer c.Close()

			u, err := url.Parse("http://localhost/test")
			if err != nil {
				log.Fatal(err)
			}
			req := &http.Request{
				Method:     "GET",
				URL:        u,
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
				Header:     make(http.Header),
				Host:       u.Host,
			}

			conn, err := net.Dial("tcp", "localhost:80")
			if err != nil {
				// handle error
				log.Fatal(err)
			}
			defer conn.Close()

			err = req.Write(conn)
			if err != nil {
				// handle error
				log.Fatal(err)
			}
			_, err = http.ReadResponse(bufio.NewReader(conn), req)
			if err != nil {
				// handle error
				log.Fatal(err)
			}

			done := make(chan struct{})
			done2 := make(chan struct{})
			go copyLog(c, conn, done)
			go copyLog(conn, c, done)
			log.Print("running")

			<-done2

		}(conn)
	}
}

func copyLog(dst io.Writer, src io.Reader, done chan struct{}) {
	i, err := io.Copy(dst, src)
	if err != nil {
		log.Print(err)
	}
	log.Printf("Written %v", i)
	close(done)
}
