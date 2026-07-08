package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/miekg/dns"
)

// --- ANSI Colors for logs ---
const (
	Cyan   = "\033[96m"
	Green  = "\033[92m"
	Gold   = "\033[93m"
	Red    = "\033[91m"
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Purple = "\033[95m"
)

var (
	cfAccountID   = os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	cfNamespaceID = os.Getenv("CLOUDFLARE_KV_NAMESPACE_ID")
	cfAPIToken    = os.Getenv("CLOUDFLARE_API_TOKEN")
	upstreamDNS   = "1.1.1.1:53"
)

// ResolveLocalFallback loads from local mesh_hosts.json
func ResolveLocalFallback(key string) (string, string) {
	filePath := "/home/theal/sovereign-27/mesh_hosts.json"
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", ""
	}
	var hosts map[string]string
	if err := json.Unmarshal(data, &hosts); err != nil {
		return "", ""
	}

	// Normalize key
	key = strings.TrimSuffix(strings.ToLower(key), ".")
	
	// Map friendly TLD names to KV JSON keys
	switch key {
	case "gcp.pqr":
		return "A", "34.135.83.69"
	case "grpc.pqr":
		if val, ok := hosts["trycloudflare_grpc_url"]; ok {
			return "CNAME", strings.TrimPrefix(val, "https://")
		}
	case "ssh.pqr":
		if val, ok := hosts["trycloudflare_ssh_url"]; ok {
			return "CNAME", strings.TrimPrefix(val, "https://")
		}
	case "web.pqr":
		if val, ok := hosts["trycloudflare_web_url"]; ok {
			return "CNAME", strings.TrimPrefix(val, "https://")
		}
	case "api.pqr":
		if val, ok := hosts["trycloudflare_api_url"]; ok {
			return "CNAME", strings.TrimPrefix(val, "https://")
		}
	}
	return "", ""
}

// ResolveFromCloudflareKV queries Cloudflare KV REST API
func ResolveFromCloudflareKV(key string) (string, string) {
	if cfAccountID == "" || cfNamespaceID == "" || cfAPIToken == "" {
		// Fallback to local file if credentials missing
		return ResolveLocalFallback(key)
	}

	normalizedKey := strings.TrimSuffix(strings.ToLower(key), ".")
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/storage/kv/namespaces/%s/values/%s",
		cfAccountID, cfNamespaceID, normalizedKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", ""
	}
	req.Header.Set("Authorization", "Bearer "+cfAPIToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		// Fallback to local
		return ResolveLocalFallback(key)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ResolveLocalFallback(key)
	}

	val := string(body)
	if strings.Contains(val, ".") {
		// If it has IP structure, it's A record, otherwise CNAME
		if net.ParseIP(val) != nil {
			return "A", val
		}
		return "CNAME", val
	}
	return "", ""
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	if len(r.Question) == 0 {
		w.WriteMsg(m)
		return
	}

	q := r.Question[0]
	name := q.Name

	// Only intercept .pqr zone
	if strings.HasSuffix(strings.ToLower(name), ".pqr.") {
		log.Printf("%s[RESOLVER] Intercepted .pqr query: %s %s%s", Cyan, name, dns.TypeToString[q.Qtype], Reset)

		recType, recVal := ResolveFromCloudflareKV(name)
		if recType != "" {
			log.Printf("%s[RESOLVER] Resolved: %s -> %s (%s)%s", Green, name, recVal, recType, Reset)
			
			// Build Answer
			if recType == "A" && q.Qtype == dns.TypeA {
				rr, err := dns.NewRR(fmt.Sprintf("%s 30 IN A %s", name, recVal))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
			} else if recType == "CNAME" && (q.Qtype == dns.TypeCNAME || q.Qtype == dns.TypeA) {
				// Make sure CNAME ends with a dot
				cnameVal := recVal
				if !strings.HasSuffix(cnameVal, ".") {
					cnameVal += "."
				}
				rr, err := dns.NewRR(fmt.Sprintf("%s 30 IN CNAME %s", name, cnameVal))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
			}
		} else {
			log.Printf("%s[RESOLVER] No record found for %s%s", Red, name, Reset)
			m.Rcode = dns.RcodeNameError
		}
		w.WriteMsg(m)
		return
	}

	// Forward non-pqr queries to upstream recursive DNS resolver
	c := new(dns.Client)
	c.Net = "udp"
	c.Timeout = 3 * time.Second
	in, _, err := c.Exchange(r, upstreamDNS)
	if err != nil {
		log.Printf("%s[UPSTREAM-ERROR] Failed to query %s: %v%s", Red, name, err, Reset)
		dns.HandleFailed(w, r)
		return
	}
	w.WriteMsg(in)
}

func main() {
	port := os.Getenv("DNS_PORT")
	if port == "" {
		port = "8053"
	}

	dns.HandleFunc(".", handleDNSRequest)

	server := &dns.Server{Addr: ":" + port, Net: "udp"}

	log.Printf("%s======================================================================%s", Gold, Reset)
	log.Printf("%s%s          SOVEREIGN MESH - INTERNAL .PQR GO DNS RESOLVER%s", Bold, Purple, Reset)
	log.Printf("%s======================================================================%s", Gold, Reset)
	log.Printf("📡 DNS server listening on UDP :%s (Upstream: %s)", port, upstreamDNS)
	if cfAccountID != "" && cfNamespaceID != "" {
		log.Printf("⚡ Mode: Cloudflare KV Connected (Namespace: %s)", cfNamespaceID)
	} else {
		log.Printf("⚡ Mode: Mock Local Fallback (mesh_hosts.json)")
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start DNS server: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down DNS server...")
	server.Shutdown()
}
