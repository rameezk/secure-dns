package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/rameezk/secure-dns/internal/dnsserver"
	// "github.com/rameezk/secure-dns/internal/httpresolver"
)

var (
	dnsListenAddr = flag.String("listen_addr", ":53", "address to listen on")
)

func main() {
	flag.Parse()

	log.Printf("Initialising dns server...")

	var wg sync.WaitGroup

	var resolver dnsserver.Resolver
	// resolver = httpresolver.New(upstream)

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
