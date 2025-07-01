package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
)

const (
	// DNS server to forward requests to
	upstreamDNS = "8.8.8.8:53"
	// Local port to listen on
	localPort = ":53"
)

// DNSHandler handles DNS requests
type DNSHandler struct {
	upstream string
}

// ServeDNS implements the dns.Handler interface
func (h *DNSHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	// Log the domain being queried
	for _, question := range r.Question {
		domain := strings.TrimSuffix(question.Name, ".")
		fmt.Printf("DNS Query: %s (Type: %s) from %s\n", domain, dns.TypeToString[question.Qtype], w.RemoteAddr())
	}

	// Create a client to forward the request with timeout
	client := &dns.Client{
		Timeout: 10 * time.Second,
	}

	// Forward the request to upstream DNS
	resp, _, err := client.Exchange(r, h.upstream)
	if err != nil {
		log.Printf("Error forwarding DNS request: %v", err)
		// Create a proper error response instead of using HandleFailed
		response := new(dns.Msg)
		response.SetReply(r)
		response.Rcode = dns.RcodeServerFailure
		w.WriteMsg(response)
		return
	}

	// Set the response
	w.WriteMsg(resp)
}

func main() {
	// Check if running as root (required for port 53)
	if os.Geteuid() != 0 {
		log.Fatal("This program must be run as root to bind to port 53")
	}

	// Create DNS handler
	handler := &DNSHandler{
		upstream: upstreamDNS,
	}

	// Create DNS server for UDP
	udpServer := &dns.Server{
		Addr:    localPort,
		Net:     "udp",
		Handler: handler,
	}

	// Create DNS server for TCP
	tcpServer := &dns.Server{
		Addr:    localPort,
		Net:     "tcp",
		Handler: handler,
	}

	fmt.Printf("Starting DNS forwarder on port 53\n")
	fmt.Printf("Forwarding requests to: %s\n", upstreamDNS)
	fmt.Printf("Logging all domain queries to stdout\n")
	fmt.Println("Press Ctrl+C to stop")

	// Start UDP server in a goroutine
	go func() {
		if err := udpServer.ListenAndServe(); err != nil {
			log.Printf("UDP server error: %v", err)
		}
	}()

	// Start TCP server in main goroutine
	if err := tcpServer.ListenAndServe(); err != nil {
		log.Fatalf("TCP server error: %v", err)
	}
}
