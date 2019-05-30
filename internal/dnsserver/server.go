package dnsserver

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"regexp"
	"strings"
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

	failoverUpstream string
	failoverDomains  []string
}

// New server
func New(addr string, resolver Resolver) *Server {
	return &Server{
		Addr:     addr,
		resolver: resolver,
	}
}

// SetFailover upstream server and list of domains
func (s *Server) SetFailover(upstream string, domains []string) {
	log.Printf("Setting failover upstream: %v", upstream)
	s.failoverUpstream = upstream
	s.failoverDomains = domains
}

// Handler to handle things
func (s *Server) Handler(writer dns.ResponseWriter, message *dns.Msg) {
	log.Println("Attempting to handle query")

	log.Printf("from:%v     id:%v", writer.RemoteAddr(), message.Id)

	if len(message.Question) != 1 {
		dns.HandleFailed(writer, message)
		return
	}

	// Check if we need to failover for this domain
	if s.shouldFailoverDomain(message.Question[0].Name, s.failoverDomains) {
		s.handleUDP(writer, message)
		return
	}

	s.handleHTTPS(writer, message)
}

func (s *Server) handleHTTPS(writer dns.ResponseWriter, message *dns.Msg) {
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
	log.Printf("Query resolved via HTTPS: %v", message.Question[0].Name)
}

func (s *Server) handleUDP(writer dns.ResponseWriter, message *dns.Msg) {
	u, err := dns.Exchange(message, s.failoverUpstream)
	if err == nil {
		writer.WriteMsg(u)
		log.Printf("Query resolved via UDP: %v", message.Question[0].Name)
	} else {
		log.Printf("Failed to resolve query via UDP: %v", err)
		dns.HandleFailed(writer, message)
	}
}

func (s *Server) shouldFailoverDomain(domain string, failoverDomains []string) bool {
	if (len(s.failoverDomains) > 0) && (s.failoverUpstream != "") {
		log.Printf("Checking if domain is in failover list: %v", domain)
		domainToQuery := strings.TrimSuffix(domain, ".")
		for _, domainToFailover := range s.failoverDomains {
			if matched, _ := regexp.MatchString(".*\\.?"+domainToFailover+".*", domainToQuery); matched {
				log.Printf("Domain matches failover list, will failover to UDP")
				return true
			}
		}
	}

	return false
}

// Serve server
func (s *Server) Serve() {
	err := s.resolver.Init()
	if err != nil {
		log.Fatalf("Error initialising resolver: %v", err)
	}

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
