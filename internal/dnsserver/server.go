package dnsserver

import (
	"log"
	"sync"

	"github.com/miekg/dns"
)

// Server struct
type Server struct {
	addr     string
	resolver Resolver
}

// New server
func New(addr string, resolver Resolver) *Server {
	return &Server{
		addr:     addr,
		resolver: resolver,
	}
}

// Handler to handle things
func (s *Server) Handler(writer dns.ResponseWriter, message *dns.Msg) {

}

// Serve server
func (s *Server) Serve() {
	s.classicServe()
}

func (s *Server) classicServe() {
	log.Println("Starting server in classic mode...")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := dns.ListenAndServe(s.addr, "udp", dns.HandlerFunc(s.Handler))
		log.Fatalf("Failed to serve via UDP: %v", err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := dns.ListenAndServe(s.addr, "tcp", dns.HandlerFunc(s.Handler))
		log.Fatalf("Failed to serve via TCP: %v", err)
	}()

	wg.Wait()
}
