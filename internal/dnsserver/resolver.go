package dnsserver

import (
	"github.com/miekg/dns"
)

// Resolver is the interface
type Resolver interface {
	// Init
	Init() error

	// Query
	Query(message *dns.Msg) (*dns.Msg, error)
}
