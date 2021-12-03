package main

import (
	"crypto/tls"
	"log"
	"net"
)

// Server is a TCP server that takes an incoming request and sends it to another
// server, proxying the response back to the client.
type Server struct {
	// TCP address to listen on
	Addr string

	// TCP address of target server
	Target string

	// ModifyRequest is an optional function that modifies the request from a client to the target server.
	ModifyRequest func(b *[]byte)

	// ModifyResponse is an optional function that modifies the response from the target server.
	ModifyResponse func(b *[]byte)

	// TLS configuration to listen on.
	TLSConfig *tls.Config

	// TLS configuration for the proxy if needed to connect to the target server with TLS protocol.
	// If nil, TCP protocol is used.
	TLSConfigTarget *tls.Config
}

// ListenAndServe listens on the TCP network address laddr and then handle packets
// on incoming connections.
func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	return s.serve(listener)
}

// ListenAndServeTLS acts identically to ListenAndServe, except that it uses TLS
// protocol. Additionally, files containing a certificate and matching private key
// for the server must be provided if neither the Server's TLSConfig.Certificates nor
// TLSConfig.GetCertificate are populated.
func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
	configHasCert := len(s.TLSConfig.Certificates) > 0 || s.TLSConfig.GetCertificate != nil
	if !configHasCert || certFile != "" || keyFile != "" {
		var err error
		s.TLSConfig.Certificates = make([]tls.Certificate, 1)
		s.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return err
		}
	}
	listener, err := tls.Listen("tcp", s.Addr, s.TLSConfig)
	if err != nil {
		return err
	}
	return s.serve(listener)
}

func (s *Server) serve(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	// connects to target server
	var rconn net.Conn
	var err error
	if s.TLSConfigTarget == nil {
		rconn, err = net.Dial("tcp", s.Target)
	} else {
		rconn, err = tls.Dial("tcp", s.Target, s.TLSConfigTarget)
	}
	if err != nil {
		return
	}

	// write to dst what it reads from src
	var pipe = func(src, dst net.Conn, filter func(b *[]byte)) {
		defer func() {
			conn.Close()
			rconn.Close()
		}()

		buff := make([]byte, 65535)
		for {
			n, err := src.Read(buff)
			if err != nil {
				log.Println(err)
				return
			}
			b := buff[:n]

			if filter != nil {
				filter(&b)
			}

			_, err = dst.Write(b)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}

	go pipe(conn, rconn, s.ModifyRequest)
	go pipe(rconn, conn, s.ModifyResponse)
}
