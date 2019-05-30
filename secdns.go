package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"

	"github.com/rameezk/secure-dns/internal/dnsserver"
	"github.com/rameezk/secure-dns/internal/httpsresolver"
)

var (
	dnsListenAddr = flag.String("listen_addr", ":53", "address to listen on")
	httpsUpstream = flag.String("https_upstream",
		"https://dns.google.com/resolve",
		"URL of upstream DNS-to-HTTP server")
)

func main() {
	flag.Parse()

	log.Printf("Initialising dns server...")
	fmt.Println("Hi")

	var wg sync.WaitGroup

	upstream, err := url.Parse(*httpsUpstream)
	if err != nil {
		log.Fatalf("Error connecting to upstream HTTPS server, did you specify the correct address? %v", err)
	}
	var resolver dnsserver.Resolver
	resolver = httpsresolver.New(upstream)

	server := dnsserver.New(*dnsListenAddr, resolver)

	wg.Add(1)
	go func() {
		defer wg.Done()
		server.Serve()
	}()

	wg.Wait()
}

// Usage Print usage
var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])

	flag.PrintDefaults()
}

// flag.Usage = func() {
// 	fmt.Fprintf(os.Stderr, "Usage: \n")
// }
