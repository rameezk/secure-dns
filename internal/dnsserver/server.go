package dnsserver

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"sync"

	"github.com/miekg/dns"
)

var newID chan uint16

func init() {
	newID = make(chan uint16, 100)

	go func() {
		var id uint16
		var err error

		for {
			err = binary.Read(rand.Reader, binary.LittleEndian, &id)
			if err != nil {
				panic(fmt.Sprintf("error creating id: %v", err))
			}

			newID <- id
		}
	}()
}

// Server struct
type Server struct {
	Addr     string
	resolver Resolver
}

// New server
func New(addr string, resolver Resolver) *Server {
	return &Server{
		Addr:     addr,
		resolver: resolver,
	}
}

// Handler to handle things
func (s *Server) Handler(writer dns.ResponseWriter, message *dns.Msg) {
	log.Println("Attempting to handle query")

	log.Printf("from:%v     id:%v", writer.RemoteAddr(), message.Id)

	if len(message.Question) != 1 {
		dns.HandleFailed(writer, message)
		return
	}

	oldid := message.Id
	message.Id = <-newID

	log.Printf("Querying question: %v", message.Question[0].Name)
	fromUp, err := s.resolver.Query(message)
	if err != nil {
		log.Printf("Failed to resolve this query: %v", err)
		message.Id = oldid
		dns.HandleFailed(writer, message)
		return
	}

	fromUp.Id = oldid
	writer.WriteMsg(fromUp)
}

// Serve server
func (s *Server) Serve() {
	err := s.resolver.Init()
	if err != nil {
		log.Fatalf("Error initialising resolver: %v", err)
	}

	go s.resolver.Maintain()

	s.classicServe()
}

func (s *Server) classicServe() {
	log.Println("Starting server in classic mode...")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := dns.ListenAndServe(s.Addr, "udp", dns.HandlerFunc(s.Handler))
		log.Fatalf("Failed to serve via UDP: %v", err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := dns.ListenAndServe(s.Addr, "tcp", dns.HandlerFunc(s.Handler))
		log.Fatalf("Failed to serve via TCP: %v", err)
	}()

	wg.Wait()
}
