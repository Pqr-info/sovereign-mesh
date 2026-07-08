import os
import sys
import json
import urllib.request

# --- AESTHETIC COLORS ---
CYAN = "\033[96m"
GREEN = "\033[92m"
GOLD = "\033[93m"
RED = "\033[91m"
RESET = "\033[0m"
BOLD = "\033[1m"

def main():
    print(f"{GOLD}======================================================================{RESET}")
    print(f"{BOLD}{CYAN}          SOVEREIGN MESH - CLOUDFLARE DNS SYNC FOR JETWEB.US{RESET}")
    print(f"{GOLD}======================================================================{RESET}")

    # 1. Retrieve credentials from environment
    api_key = os.getenv("CLOUDFLARE_API_KEY")
    email = os.getenv("CLOUDFLARE_EMAIL")

    if not api_key or not email:
        print(f"{RED}ERROR: CLOUDFLARE_API_KEY and CLOUDFLARE_EMAIL environment variables must be set.{RESET}")
        print("Usage:")
        print(f"  {BOLD}CLOUDFLARE_API_KEY=\"your_key\" CLOUDFLARE_EMAIL=\"your_email\" python3 update_jetweb_dns.py{RESET}")
        sys.exit(1)

    domain = "jetweb.us"
    print(f"⚙️ Querying Cloudflare Zone ID for {BOLD}{domain}{RESET}...")

    # 2. Fetch Zone ID
    zone_url = f"https://api.cloudflare.com/client/v4/zones?name={domain}"
    req = urllib.request.Request(zone_url)
    req.add_header("X-Auth-Key", api_key)
    req.add_header("X-Auth-Email", email)
    req.add_header("Content-Type", "application/json")

    try:
        with urllib.request.urlopen(req) as resp:
            data = json.loads(resp.read().decode())
            if not data.get("success") or not data.get("result"):
                print(f"{RED}ERROR: Failed to fetch Zone ID. Response: {data}{RESET}")
                sys.exit(1)
            zone_id = data["result"][0]["id"]
            print(f"✅ Found Zone ID: {BOLD}{zone_id}{RESET}")
    except Exception as e:
        print(f"{RED}ERROR: Failed to connect to Cloudflare API: {e}{RESET}")
        sys.exit(1)

    # 3. Define the records to map
    gcp_ip = "34.135.83.69"
    records = [
        {"type": "A", "name": f"gcp.{domain}", "content": gcp_ip, "proxied": False},
        {"type": "CNAME", "name": f"grpc.{domain}", "content": "skill-relatively-trading-refine.trycloudflare.com", "proxied": False},
        {"type": "CNAME", "name": f"ssh.{domain}", "content": "division-jaguar-renew-casio.trycloudflare.com", "proxied": False},
        {"type": "CNAME", "name": f"web.{domain}", "content": "explorer-serve-enrollment-prospective.trycloudflare.com", "proxied": False},
    ]

    for record in records:
        rec_name = record["name"]
        rec_type = record["type"]
        rec_content = record["content"]
        proxied = record["proxied"]

        print(f"\n🌐 Synchronizing {rec_type} record: {BOLD}{rec_name}{RESET} -> {BOLD}{rec_content}{RESET}...")

        # A. Check if record already exists
        list_url = f"https://api.cloudflare.com/client/v4/zones/{zone_id}/dns_records?name={rec_name}&type={rec_type}"
        req_list = urllib.request.Request(list_url)
        req_list.add_header("X-Auth-Key", api_key)
        req_list.add_header("X-Auth-Email", email)
        req_list.add_header("Content-Type", "application/json")

        record_id = None
        try:
            with urllib.request.urlopen(req_list) as resp:
                list_data = json.loads(resp.read().decode())
                if list_data.get("result"):
                    record_id = list_data["result"][0]["id"]
                    print(f"🔍 Existing record found with ID: {record_id}")
        except Exception as e:
            print(f"{RED}Warning: Failed to query existing record: {e}{RESET}")

        # B. Create or Update record
        payload = {
            "type": rec_type,
            "name": rec_name,
            "content": rec_content,
            "ttl": 1,
            "proxied": proxied
        }
        body = json.dumps(payload).encode("utf-8")

        if record_id:
            # Update (PUT)
            upsert_url = f"https://api.cloudflare.com/client/v4/zones/{zone_id}/dns_records/{record_id}"
            method = "PUT"
        else:
            # Create (POST)
            upsert_url = f"https://api.cloudflare.com/client/v4/zones/{zone_id}/dns_records"
            method = "POST"

        req_upsert = urllib.request.Request(upsert_url, data=body, method=method)
        req_upsert.add_header("X-Auth-Key", api_key)
        req_upsert.add_header("X-Auth-Email", email)
        req_upsert.add_header("Content-Type", "application/json")

        try:
            with urllib.request.urlopen(req_upsert) as resp:
                upsert_data = json.loads(resp.read().decode())
                if upsert_data.get("success"):
                    print(f"{GREEN}✓ SUCCESS: {rec_name} synchronized successfully.{RESET}")
                else:
                    print(f"{RED}❌ FAILED: {upsert_data}{RESET}")
        except Exception as e:
            print(f"{RED}❌ FAILED: Connection error: {e}{RESET}")

    print(f"\n{GOLD}======================================================================{RESET}")
    print(f"{GREEN}{BOLD}🎉 DNS SYNCHRONIZATION RUN COMPLETE!{RESET}")
    print(f"{GOLD}======================================================================{RESET}")

if __name__ == "__main__":
    main()
