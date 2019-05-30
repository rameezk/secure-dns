package httpsresolver

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/miekg/dns"
	"github.com/rameezk/secure-dns/internal/dnsjson"
	"github.com/rameezk/secure-dns/internal/dnsserver"
)

type httpsResolver struct {
	Upstream *url.URL
	client   *http.Client
	mode     string
}

// New creates a new HTTPS resolver
func New(upstream *url.URL) *httpsResolver {
	log.Println("Creating a new HTTPS resolver")
	return &httpsResolver{
		Upstream: upstream,
	}
}

func (resolver *httpsResolver) Init() error {
	log.Println("Initialising HTTPS resolver")
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	resolver.client = &http.Client{
		Timeout:   4 * time.Second,
		Transport: transport,
	}

	// Load certs here

	return nil
}

func (resolver *httpsResolver) Maintain() {
}

func (resolver *httpsResolver) Query(request *dns.Msg) (*dns.Msg, error) {
	log.Println("Resolving DNS query")
	return resolver.queryJSON(request)
}

func (resolver *httpsResolver) queryJSON(request *dns.Msg) (*dns.Msg, error) {
	// Only answer single-question queries.
	// In practice, these are all we get, and almost no server supports
	// multi-question requests anyway.
	log.Println("Resolving DNS query using JSON")

	if len(request.Question) != 1 {
		return nil, fmt.Errorf("multi-question query")
	}

	question := request.Question[0]
	// Only answer IN-class queries, which are the ones used in practice.
	if question.Qclass != dns.ClassINET {
		return nil, fmt.Errorf("query class != IN")
	}

	// Build the query and send the request.
	url := *resolver.Upstream
	vs := url.Query()
	vs.Set("name", question.Name)
	vs.Set("type", dns.TypeToString[question.Qtype])
	url.RawQuery = vs.Encode()

	hr, err := resolver.client.Get(url.String())
	if err != nil {
		return nil, fmt.Errorf("GET failed: %v", err)
	}
	defer hr.Body.Close()

	if hr.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Response status: %s", hr.Status)
	}

	// Read the HTTPS response, and parse the JSON.
	body, err := ioutil.ReadAll(io.LimitReader(hr.Body, 64*1024))
	if err != nil {
		return nil, fmt.Errorf("Failed to read body: %v", err)
	}

	jr := &dnsjson.Response{}
	err = json.Unmarshal(body, jr)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshall: %v", err)
	}

	if len(jr.Question) != 1 {
		return nil, fmt.Errorf("Wrong number of questions in the response")
	}

	// Build the DNS response.
	resp := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id:       request.Id,
			Response: true,
			Opcode:   request.Opcode,
			Rcode:    jr.Status,

			Truncated:          jr.TC,
			RecursionDesired:   jr.RD,
			RecursionAvailable: jr.RA,
			AuthenticatedData:  jr.AD,
			CheckingDisabled:   jr.CD,
		},
		Question: []dns.Question{
			{
				Name:   jr.Question[0].Name,
				Qtype:  jr.Question[0].Type,
				Qclass: dns.ClassINET,
			}},
	}

	for _, answer := range jr.Answer {
		// TODO: This "works" but is quite hacky. Is there a better way,
		// without doing lots of data parsing?
		s := fmt.Sprintf("%s %d IN %s %s",
			answer.Name, answer.TTL,
			dns.TypeToString[answer.Type], answer.Data)
		rr, err := dns.NewRR(s)
		if err != nil {
			return nil, fmt.Errorf("Error parsing answer: %v", err)
		}

		resp.Answer = append(resp.Answer, rr)
	}

	return resp, nil
}

// Compile-time check that the implementation matches the interface.
var _ dnsserver.Resolver = &httpsResolver{}
