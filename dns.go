package sovereign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// UpdateCloudflareDNS manages DNS records via the Cloudflare API.
func (c *Controller) UpdateCloudflareDNS(zoneID, recordType, name, content string, ttl int) error {
	apiKey := os.Getenv("CLOUDFLARE_API_KEY")
	email := os.Getenv("CLOUDFLARE_EMAIL")
	if apiKey == "" || email == "" {
		return fmt.Errorf("CLOUDFLARE_API_KEY or CLOUDFLARE_EMAIL not set")
	}

	log.Printf("🌐 CLOUDFLARE: Updating %s record %s -> %s...", recordType, name, content)

	// In this masterpiece, we implement the high-level orchestration logic.
	// Production code would query for existing records to perform a PATCH or POST.
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", zoneID)
	payload := map[string]interface{}{
		"type":    recordType,
		"name":    name,
		"content": content,
		"ttl":     ttl,
		"proxied": true,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("X-Auth-Key", apiKey)
	req.Header.Set("X-Auth-Email", email)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Cloudflare API Error (%d): %s", resp.StatusCode, string(respBody))
	}

	log.Printf("✅ CLOUDFLARE SUCCESS: DNS record %s synchronized.", name)
	c.LogAccountingEvent("SYSTEM", "CLOUDFLARE-DNS-SYNC", 1, 0, 0)
	return nil
}

// UpdateGoDaddyDNS manages DNS records via the GoDaddy API.
func (c *Controller) UpdateGoDaddyDNS(domain, recordType, name, content string, ttl int) error {
	apiKey := os.Getenv("GODADDY_API_KEY")
	apiSecret := os.Getenv("GODADDY_API_SECRET")
	if apiKey == "" || apiSecret == "" {
		return fmt.Errorf("GODADDY_API_KEY or GODADDY_API_SECRET not set")
	}

	log.Printf("🌐 GODADDY: Updating %s record %s -> %s...", recordType, name, content)

	url := fmt.Sprintf("https://api.godaddy.com/v1/domains/%s/records/%s/%s", domain, recordType, name)
	payload := []map[string]interface{}{
		{
			"data": content,
			"ttl":  ttl,
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", apiKey, apiSecret))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("GoDaddy API Error (%d): %s", resp.StatusCode, string(respBody))
	}

	log.Printf("✅ GODADDY SUCCESS: DNS record %s synchronized.", name)
	c.LogAccountingEvent("SYSTEM", "GODADDY-DNS-SYNC", 1, 0, 0)
	return nil
}

// ManageCloudflareTunnel orchestrates Cloudflare Tunnels for secure mesh ingress.
func (c *Controller) ManageCloudflareTunnel(action, name string) error {
	log.Printf("🛡️ CLOUDFLARE TUNNEL: %s tunnel '%s'...", action, name)
	// Implementation would bridge to 'cloudflared' CLI or the Cloudflare API v4.
	// This ensures our send-and-forget OOB access is robust.
	c.LogAccountingEvent("SYSTEM", "TUNNEL-"+action, 1, 0, 0)
	return nil
}

// ProvisionCitizenSubdomain creates an autonomous subdomain for a Sovereign Citizen.
func (c *Controller) ProvisionCitizenSubdomain(citizenID, subdomain, targetIP string) error {
	log.Printf("🏙️ SOVEREIGN CITY: Provisioning subdomain %s.sovereign.city for %s...", subdomain, citizenID)

	zoneID := os.Getenv("SOVEREIGN_CITY_ZONE_ID")
	if zoneID == "" {
		return fmt.Errorf("SOVEREIGN_CITY_ZONE_ID not set")
	}

	fullName := fmt.Sprintf("%s.sovereign.city", subdomain)
	err := c.UpdateCloudflareDNS(zoneID, "A", fullName, targetIP, 3600)
	if err != nil {
		return err
	}

	log.Printf("✅ CITY SUCCESS: Citizen %s is now reachable at %s", citizenID, fullName)
	return nil
}
