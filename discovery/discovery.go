package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type NeighborInfo struct {
	ID        string  // UMI / node ID
	Addr      string  // "udp6://[fe80::...]:9999" or "udp4://192.168.1.10:9999"
	Transport string  // "LAN_MULTICAST"
	Metric    float64 // initial metric (e.g. 1.0 for local LAN)
}

type DiscoveryConfig struct {
	OnNeighborDiscovered func(NeighborInfo)
	OnNeighborLost       func(id string)
}

type DiscoveryService struct {
	cfg      DiscoveryConfig
	mu       sync.RWMutex
	peers    map[string]NeighborInfo
	lastSeen map[string]time.Time
}

func NewDiscoveryService(cfg DiscoveryConfig) (*DiscoveryService, error) {
	return &DiscoveryService{
		cfg:      cfg,
		peers:    make(map[string]NeighborInfo),
		lastSeen: make(map[string]time.Time),
	}, nil
}

func (d *DiscoveryService) Start(ctx context.Context) error {
	go d.listenMulticast(ctx)
	go d.sendBeacons(ctx)
	go d.monitorExpiry(ctx)
	return nil
}

func (d *DiscoveryService) listenMulticast(ctx context.Context) {
	// 1) Try IPv6 first
	if err := d.listenUDP6(ctx, "[ff02::c0ba:11]:9999"); err != nil {
		log.Printf("[DISCOVERY] IPv6 multicast failed: %v. Falling back to IPv4...", err)
		// 2) Fallback to IPv4
		if err4 := d.listenUDP4(ctx, "239.0.0.57:9999"); err4 != nil {
			log.Printf("[DISCOVERY] IPv4 multicast failed: %v", err4)
		}
	}
}

func (d *DiscoveryService) listenUDP6(ctx context.Context, maddr string) error {
	pc, err := net.ListenPacket("udp6", "[::]:9999")
	if err != nil {
		return err
	}
	defer pc.Close()

	addr, err := net.ResolveUDPAddr("udp6", maddr)
	if err != nil {
		return err
	}

	conn, err := net.ListenMulticastUDP("udp6", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	buf := make([]byte, 2048)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			n, src, err := conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}
			d.handlePacket(buf[:n], src, "udp6")
		}
	}
}

func (d *DiscoveryService) listenUDP4(ctx context.Context, maddr string) error {
	addr, err := net.ResolveUDPAddr("udp4", maddr)
	if err != nil {
		return err
	}
	conn, err := net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	buf := make([]byte, 2048)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			n, src, err := conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}
			d.handlePacket(buf[:n], src, "udp4")
		}
	}
}

type beaconPayload struct {
	ID string `json:"id"`
}

func (d *DiscoveryService) handlePacket(data []byte, src net.Addr, proto string) {
	var p beaconPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return
	}
	if p.ID == "" {
		return
	}

	addr := fmt.Sprintf("%s://%s", proto, src.String())
	n := NeighborInfo{
		ID:        p.ID,
		Addr:      addr,
		Transport: "LAN_MULTICAST",
		Metric:    1.0,
	}

	d.mu.Lock()
	_, exists := d.peers[p.ID]
	d.peers[p.ID] = n
	d.lastSeen[p.ID] = time.Now()
	d.mu.Unlock()

	if !exists && d.cfg.OnNeighborDiscovered != nil {
		d.cfg.OnNeighborDiscovered(n)
	}
}

func (d *DiscoveryService) sendBeacons(ctx context.Context) {
	addr, err := net.ResolveUDPAddr("udp4", "239.0.0.57:9999")
	if err != nil {
		return
	}
	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			payload, _ := json.Marshal(beaconPayload{ID: "node-local"})
			conn.Write(payload)
		}
	}
}

func (d *DiscoveryService) monitorExpiry(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.mu.Lock()
			now := time.Now()
			for id, lastSeen := range d.lastSeen {
				if now.Sub(lastSeen) > 15*time.Second {
					delete(d.lastSeen, id)
					delete(d.peers, id)
					if d.cfg.OnNeighborLost != nil {
						go d.cfg.OnNeighborLost(id)
					}
				}
			}
			d.mu.Unlock()
		}
	}
}
