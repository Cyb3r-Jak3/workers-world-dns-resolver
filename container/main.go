package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/miekg/dns"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var (
	commit        = "unknown"
	date          = "unknown"
	versionString = fmt.Sprintf("%s (built %s)", commit, date)
	terminate     = false
	dnsTypes      = []map[string]uint16{}
)

type DebugResponse struct {
	Version      string `json:"version"`
	AppID        string `json:"app_id"`
	Region       string `json:"region"`
	Location     string `json:"location"`
	Country      string `json:"country"`
	DeploymentID string `json:"deployment_id"`
}

func DebugHandler(w http.ResponseWriter, r *http.Request) {
	country := os.Getenv("CLOUDFLARE_COUNTRY_A2")
	location := os.Getenv("CLOUDFLARE_LOCATION")
	region := os.Getenv("CLOUDFLARE_REGION")
	AppID := os.Getenv("CLOUDFLARE_APPLICATION_ID")
	DeploymentID := os.Getenv("CLOUDFLARE_DEPLOYMENT_ID")
	if buildInfo, available := debug.ReadBuildInfo(); available {
		versionString = fmt.Sprintf("%s (built %s with %s)", commit, date, buildInfo.GoVersion)
	}
	query := r.URL.Query()
	if query.Get("format") == "text" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintf(w, "Hi, I'm a container running in %s, %s, which is part of %s", location, country, region)
		_, _ = fmt.Fprintf(w, "with the following build information:\n")
		_, _ = fmt.Fprintf(w, "Version: %s\n", versionString)
		_, _ = fmt.Fprintf(w, "App ID: %s\n", AppID)
		_, _ = fmt.Fprintf(w, "Deployment ID: %s\n", DeploymentID)
		return
	}
	response := DebugResponse{
		Version:      versionString,
		AppID:        AppID,
		Region:       region,
		Location:     location,
		Country:      country,
		DeploymentID: DeploymentID,
	}
	JSONResponse(w, response)
}

func main() {
	c := make(chan os.Signal, 10)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	mux := http.NewServeMux()

	// Create a sub-mux for /v1
	v1mux := http.NewServeMux()
	v1mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if terminate {
			w.WriteHeader(503)
			_, _ = w.Write([]byte("Service Unavailable"))
			return
		}
		_, _ = w.Write([]byte("ok"))
	})
	v1mux.HandleFunc("/debug", DebugHandler)
	v1mux.HandleFunc("/lookup", ResolveEndpoint)
	v1mux.HandleFunc("/dns_types", DNSTypesEndpoint)
	v1mux.HandleFunc("/dns_servers", DNSServerEndpoint)

	// Mount the v1mux at /v1/
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", v1mux))

	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}
	go func() {
		for range c {
			if terminate {
				err := server.Shutdown(context.Background())
				if err != nil {
					log.Printf("Error during shutdown: %v", err)
					os.Exit(1)
				}
				os.Exit(0)
			}

			terminate = true
			go func() {
				time.Sleep(time.Second * 30)
				os.Exit(0)
			}()
		}
	}()

	log.Fatal(server.ListenAndServe())
}

func ResolveEndpoint(w http.ResponseWriter, r *http.Request) {
	// This function can be used to resolve the container's address or any other necessary information.
	// For now, it returns a placeholder string.
	parsed, err := ParseURLQuery(r.URL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid query parameters: %v", err), http.StatusBadRequest)
		return
	}
	if !strings.HasSuffix(parsed.Domain, ".") {
		parsed.Domain += "."
	}
	// Ensure the domain is a valid FQDN
	if !dns.IsFqdn(parsed.Domain) {
		http.Error(w, "Invalid domain name", http.StatusBadRequest)
		return
	}
	m1 := new(dns.Msg)
	m1.SetQuestion(parsed.Domain, parsed.Type)
	m1.RecursionDesired = true
	client := new(dns.Client)
	client.Timeout = 5 * time.Second
	response := LookupResponse{
		Question: parsed.Domain,
		Type:     dns.TypeToString[parsed.Type],
		Country:  os.Getenv("CLOUDFLARE_COUNTRY_A2"),
		Location: os.Getenv("CLOUDFLARE_LOCATION"),
		Region:   os.Getenv("CLOUDFLARE_REGION"),
		Answers:  make([]DNSServerResponse, 0, len(dnsServers)),
	}
	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)
	lookupStart := time.Now()
	for _, server := range dnsServers {
		wg.Add(1)
		go func(server DNSServer) {
			defer wg.Done()
			answer := DNSServerResponse{
				DNSServer: server.Name,
				Address:   server.Address,
			}
			resp, duration, err := client.Exchange(m1, server.AddressString())
			if err != nil {
				// Optionally log error, but don't write to w from goroutine
				log.Printf("Error resolving %s / %s with %s: %v", parsed.Domain, dns.TypeToString[parsed.Type], server.Name, err)
				return
			}
			if len(resp.Answer) == 0 {
				// Optionally log error, but don't write to w from goroutine
				log.Printf("No answer found for %v with %s", m1.Question[0].Name, server.Name)
				answer.Values = []string{}
				mu.Lock()
				response.Answers = append(response.Answers, answer)
				mu.Unlock()
				return
			}
			answer.Duration = duration
			answer.DurationString = duration.String()
			answer.TTL = int(resp.Answer[0].Header().Ttl)
			answer.Values = make([]string, len(resp.Answer))
			for i, ans := range resp.Answer {
				if a, ok := ans.(*dns.CNAME); ok {
					answer.Values[i] = a.Target
				} else {
					stringAnswer := ans.String()
					parts := strings.Split(stringAnswer, "\t")
					if len(parts) > 0 {
						answer.Values[i] = parts[len(parts)-1]
					} else {
						answer.Values[i] = stringAnswer // Use the full string as a fallback
					}
				}
			}
			mu.Lock()
			response.Answers = append(response.Answers, answer)
			mu.Unlock()
		}(server)
	}
	wg.Wait()
	response.TotalDuration = time.Since(lookupStart)
	response.TotalDurationString = response.TotalDuration.String()
	JSONResponse(w, response)
}

func DNSTypesEndpoint(w http.ResponseWriter, r *http.Request) {
	// This function can be used to return the DNS types supported by the container.
	// For now, it returns a placeholder string.
	dnsTypes := make([]string, 0, len(dnsTypes))
	for k := range dns.StringToType {
		dnsTypes = append(dnsTypes, k)
	}
	sort.Strings(dnsTypes)
	JSONResponse(w, dnsTypes)
}

func DNSServerEndpoint(w http.ResponseWriter, r *http.Request) {
	// This function can be used to return the DNS servers supported by the container.
	// For now, it returns a placeholder string.
	JSONResponse(w, dnsServers)
}

func init() {
	// Initialize the dnsTypes slice with DNS types
	dnsTypes = make([]map[string]uint16, 0, len(dns.StringToType))
	for k, v := range dns.StringToType {
		dnsTypes = append(dnsTypes, map[string]uint16{k: v})
	}
}

func DoTQuery(domain string, qtype uint16, server string) (*dns.Msg, time.Duration, error) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), qtype)

	client := &dns.Client{
		Net:     "tcp-tls",
		Timeout: 5 * time.Second,
		TLSConfig: &tls.Config{
			ServerName: server, // e.g., "1.1.1.1" or "dns.google"
		},
	}

	// server should be "host:853" (853 is the default DoT port)
	resp, rtt, err := client.Exchange(m, server)
	return resp, rtt, err
}
