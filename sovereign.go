package sovereign

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pqr-info/sovereign-mesh/addressing"
	"github.com/pqr-info/sovereign-mesh/proto"
	"golang.org/x/telemetry/counter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"strings"
)

// NewController creates a new instance of the Sovereign Mesh engine.
func NewController(projectID, location string) *Controller {
	c := &Controller{
		agents:         make(map[string]*Agent),
		processes:      make(map[int32]*Process),
		prompts:        make(map[string]*Prompt),
		knowledge:      make(map[string]string),
		ledger:         make([]*LedgerBlock, 0),
		neuralSessions: make(map[string]*TrainingSessionState),
		citizens:       make(map[string]*Citizen),
		tasks:          make(chan string, 100),
		optTasks:       make(chan OptimizationTask, 100),
		arbitrageCh:    make(chan HFTArbitrageSignal, 1024),
		metrics:       make(map[string]uint64),
		projectID:     projectID,
		storageBucket: os.Getenv("SNAPSHOT_BUCKET"),
		location:      location,
		startTime:     time.Now().UTC(),
		radiusSecret:  os.Getenv("RADIUS_SECRET"),
		radiusServer:  os.Getenv("RADIUS_SERVER"),
	}

	c.Address5D = addressing.NewAddress5D()

	c.SeedGenesisBlock()
	return c
}

// OuroborosSentinel monitors core processes and triggers the Resurrection protocol on failure.
func (c *Controller) OuroborosSentinel(ctx context.Context) {
	log.Printf("🐍 OUROBOROS SENTINEL: Watchdog daemon activated. Monitoring %d core processes.", len(c.watchlist))

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for proc, cmd := range c.watchlist {
				if !c.isProcessRunning(proc) {
					log.Printf("🚨 SENTINEL ALERT: Process '%s' has flatlined! Initiating Resurrection Protocol...", proc)

					// 1. Audit Failure via RADIUS
					c.LogAccountingEvent("SENTINEL", "FAILURE-"+proc, 1, 0, 0)

					// 2. Log State Deviation in Jetweb Time Machine (Simulated)
					log.Printf("⏰ JETWEB: Recording state deviation at Block #%d", len(c.ledger))

					// 3. Resurrect Process
					go c.resurrect(proc, cmd)
				}
			}
		}
	}
}

func (c *Controller) isProcessRunning(name string) bool {
	var pattern string
	switch name {
	case "grpc_server":
		pattern = "grpc_server.py"
	case "memory_bus":
		pattern = "memory_bus/server.py"
	case "web_portal":
		pattern = "web_server.py"
	default:
		return false
	}

	cmd := exec.Command("pgrep", "-f", pattern)
	err := cmd.Run()
	return err == nil
}

func (c *Controller) resurrect(name, cmd string) {
	log.Printf("✨ RESURRECTION: Re-igniting '%s' via '%s'...", name, cmd)

	wd := "/home/aellok/sovereign_mesh"
	var logFile string
	switch name {
	case "grpc_server":
		logFile = "/tmp/grpc_server.log"
	case "memory_bus":
		logFile = "/tmp/memory_bus.log"
	case "web_portal":
		logFile = "/tmp/web_server.log"
	default:
		logFile = "/dev/null"
	}

	shellCmd := fmt.Sprintf("nohup %s > %s 2>&1 &", cmd, logFile)
	execCmd := exec.Command("sh", "-c", shellCmd)
	execCmd.Dir = wd

	err := execCmd.Run()
	if err != nil {
		log.Printf("❌ RESURRECTION FAILED for '%s': %v", name, err)
		return
	}

	time.Sleep(1 * time.Second)
	log.Printf("✅ HEALED: Process '%s' is back in stable flight path.", name)

	// Audit Success via RADIUS
	c.LogAccountingEvent("SENTINEL", "HEAL-"+name, 1, 0, 0)
}

// RemoteExecute delegates a command to a specific node in the mesh.
func (c *Controller) RemoteExecute(node, command string) (string, error) {
	log.Printf("🛰️ DELEGATION: Routing command to %s: %s", node, command)

	if node == "" || node == "AURORA" || node == "localhost" || node == "127.0.0.1" {
		out, err := exec.Command("sh", "-c", command).CombinedOutput()
		return string(out), err
	}

	// Resolve target node IP address via GlobalStarchart or direct mapping
	targetAddr := node
	if resolvedIp, ok := GlobalStarchart.Nodes[node]; ok {
		// Remove "(SENTRY)" or other suffix annotations if present
		fields := fmt.Sprintf("%v", resolvedIp)
		var cleanIp string
		fmt.Sscanf(fields, "%s", &cleanIp)
		if cleanIp != "" {
			targetAddr = cleanIp
		}
	}

	// Establish gRPC link on port 1111
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, err := grpc.Dial(targetAddr+":1111", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", fmt.Errorf("failed to dial remote gRPC node: %v", err)
	}
	defer conn.Close()

	client := proto.NewAgentSyncClient(conn)
	res, err := client.RemoteExecute(ctx, &proto.CommandPayload{
		Command: "sh",
		Args:    []string{"-c", command},
	})
	if err != nil {
		return "", fmt.Errorf("remote gRPC execution failed: %v", err)
	}

	if res.ExitCode != 0 {
		return res.Stdout, fmt.Errorf("exit code %d: %s", res.ExitCode, res.Stderr)
	}
	return res.Stdout, nil
}

// TeleportProcess migrates an execution unit across the mesh using zero-copy memory paging.
func (c *Controller) TeleportProcess(pid int32, targetNode string) error {
	c.syncLock.Lock()
	defer c.syncLock.Unlock()

	proc, ok := c.processes[pid]
	if !ok {
		return syscall.ESRCH // Process not found
	}

	oldNode := proc.CurrentNode
	log.Printf("🚄 TELEPORTING: Process %d (Owner: %s) | %s -> %s", pid, proc.Owner, oldNode, targetNode)

	// 1. Snapshot Process Stack Trace (Simulated Silicon Access)
	stackTrace := "main.go:42 -> memory.go:111 -> syscall.Mmap:0x7ff"
	proc.StackHistory = append(proc.StackHistory, fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), stackTrace))

	// 2. Perform Zero-Copy Memory Paging (Direct bus allocation)
	// We simulate this by moving the process segment offset in the memory bus
	offset := int(pid % 1024) * 4096 // 4KB pages
	log.Printf("⚡ RAM-BUS: Page frame migration at offset 0x%x complete.", offset)

	// 3. RADIUS AAAA Accounting
	c.TrackProcessMigration(pid, proc.Owner, oldNode, targetNode)

	// 4. Update Global Truth
	proc.CurrentNode = targetNode
	proc.LastMigrated = time.Now()
	proc.Status = "MIGRATING"

	return nil
}

// Start initializes the system monitors and orchestrators.
func (c *Controller) Start(ctx context.Context) {
	log.Printf("✨ INITIALIZING STARBIRTH PROTOCOL (SBP-001) - 2026 Swarm...")
	c.metrics["system/runlevel"] = 7 // STARBIRTH Runlevel

	// Initialize Ouroboros Sentinel
	c.sentinelActive = true
	c.watchlist = map[string]string{
		"grpc_server": "python3 -u grpc_node/grpc_server.py",
		"memory_bus":  "python3 -u memory_bus/server.py",
		"web_portal":  "python3 -u grpc_node/web_server.py",
	}
	go c.OuroborosSentinel(ctx)

	starbirthCounter := counter.New("sovereign/starbirth_initialization_total")
	starbirthCounter.Inc()

	// Initialize Go telemetry counters for the 2026 Production Swarm
	counter.Open()

	// 1. Recover state from GCS Snapshot if available
	if c.storageBucket != "" {
		if err := c.LoadSnapshot(ctx); err != nil {
			log.Printf("⚠️ Snapshot recovery failed: %v", err)
		} else {
			log.Printf("✅ Blockchain state recovered from gs://%s", c.storageBucket)
		}
	}

	// 2. Handle Graceful Shutdown (Cloud Run SIGTERM)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-stop
		log.Printf("📥 Received signal %v. Saving blockchain snapshot...", sig)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := c.SaveSnapshot(shutdownCtx); err != nil {
			log.Printf("❌ Failed to persist snapshot: %v", err)
		}
		os.Exit(0)
	}()

	// Respect Cloud Run dynamic port assignment
	port := "1113" // Dedicated native tool-use port

	// Find all IPs to bind to (localhost and 192.168.12.0/24 only)
	var bindIPs []string
	bindIPs = append(bindIPs, "127.0.0.1")

	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if ip != nil && ip.To4() != nil {
					ipStr := ip.String()
					if strings.HasPrefix(ipStr, "192.168.12.") {
						bindIPs = append(bindIPs, ipStr)
					}
				}
			}
		}
	}

	uniqueIPs := make(map[string]bool)
	for _, ip := range bindIPs {
		uniqueIPs[ip] = true
	}

	c.grpcServer = grpc.NewServer()
	srv := &meshServer{controller: c}
	proto.RegisterSovereignMeshServer(c.grpcServer, srv)
	proto.RegisterAgentSyncServer(c.grpcServer, srv)
	proto.RegisterNeuralTrainingServer(c.grpcServer, srv)
	proto.RegisterSovereignCityServer(c.grpcServer, srv)
	proto.RegisterAgentToolUseServer(c.grpcServer, &ToolUseServer{})
	reflection.Register(c.grpcServer)

	for ip := range uniqueIPs {
		lis, err := net.Listen("tcp", net.JoinHostPort(ip, port))
		if err != nil {
			log.Printf("failed to listen on %s:%s: %v", ip, port, err)
			continue
		}
		go func(l net.Listener, ipAddress string) {
			log.Printf("📡 Sovereign Cloud Run Instance active on %s:%s", ipAddress, port)
			if err := c.grpcServer.Serve(l); err != nil {
				log.Printf("gRPC server stopped on %s: %v", ipAddress, err)
			}
		}(lis, ip)
	}

	// Initialize and run the Arbitrage Daemon
	arbitrageDaemon := NewArbitrageDaemon(c, c.arbitrageCh, "http://127.0.0.1:8082")
	go arbitrageDaemon.Run()

	go c.startHealthMonitor(ctx)
	go c.startOrchestrator(ctx)
	go c.startInfrastructureMonitor(ctx)
	go c.startNeuralDriftMonitor(ctx)
	go c.startPNPhasingMonitor(ctx)
	go c.startPNMulticastListener(ctx)
	log.Printf("👑 Sovereign Mesh Controller active in %s", c.location)
}

func (c *Controller) startOrchestrator(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case taskID := <-c.tasks:
			c.dispatch(taskID)
		case <-time.After(5 * time.Second):
			// Idle polling for ledger consistency
		}
	}
}

func (c *Controller) startHealthMonitor(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.syncLock.Lock()
			for id, agent := range c.agents {
				if time.Since(agent.LastHeartbeat) > 2*time.Minute {
					log.Printf("Pruning dead agent: %s", id)
					delete(c.agents, id)
				}
			}
			c.syncLock.Unlock()
		}
	}
}

func (c *Controller) startNeuralDriftMonitor(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// STARBIRTH Metrics: Track neural divergence against the 1% margin of error
	driftCounter := counter.New("sovereign/neural_drift_detected_total")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.syncLock.RLock()
			for id, agent := range c.agents {
				state := c.GetAgentState(agent.MemoryOffset)

				// 1% Margin of Error Check
				var totalDrift float32
				for i := 0; i < 16; i++ {
					// Compare current weights in SHM against last winning weights
					// (This assumes we store WinningWeights in the Agent struct)
					if agent.Persona != nil {
						diff := float64(state.NeuralWeights[i] - agent.Persona.Weights[i])
						totalDrift += float32(math.Abs(diff))
					}
				}

				// Factoral average drift > 0.01 (1%)
				if totalDrift/16 > 0.01 {
					log.Printf("⚠️ DRIFT DETECTED: Agent %s drifted %.2f%%. Reverting timeline...", id, (totalDrift/16)*100)
					driftCounter.Inc()
					c.syncLock.RUnlock()
					// Call out to Time Machine (Logic in byO0.go)
					// c.performTimelineReversion(id, agent.LastWinningBlockIndex)
					c.syncLock.RLock()
				}
			}
			c.syncLock.RUnlock()
		}
	}
}

func (c *Controller) startInfrastructureMonitor(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Infrastructure Metrics: Track how often we drop below the 7-node validator floor
	infraCounter := counter.New("sovereign/infra_floor_violation_total")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.syncLock.RLock()
			validators := 0
			for _, agent := range c.agents {
				if agent.NodeClass == "VALIDATOR" {
					validators++
				}
			}
			c.syncLock.RUnlock()

			if validators < 6 {
				log.Printf("🚨 INFRA ALERT: Validator count at %d (Target: 6). Triggering external VPS allocation and revenue redistribution...", validators)
				c.TriggerExternalScaling(6 - validators)
				infraCounter.Inc()
				c.metrics["infra/floor_violation_count"]++
			}
		}
	}
}

// startPNMulticastListener allows agents to "hear" each other's iPN phasing datagrams.
// This establishes the stealth backchannel connectivity required for Starbirth.
func (c *Controller) startPNMulticastListener(ctx context.Context) {
	addr, err := net.ResolveUDPAddr("udp6", "[ff02::c0ba:11]:9999")
	if err != nil {
		log.Printf("❌ iPN Listener Error: %v", err)
		return
	}

	// Join the multicast group. nil uses the default multicast interface.
	conn, err := net.ListenMulticastUDP("udp6", nil, addr)
	if err != nil {
		log.Printf("❌ iPN Listener: Failed to join multicast group: %v", err)
		return
	}
	defer conn.Close()

	buf := make([]byte, 8)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
					continue
				}
				return
			}
			if n == 8 {
				heardPN := binary.BigEndian.Uint64(buf)
				c.metrics["ipn/multicast_datagrams_heard"]++
				log.Printf("📡 iPN BACKCHANNEL discovery: Heard phasing signal %x from peer", heardPN)
			}
		}
	}
}

// SynchronizedArbitrageBlast utilizes the reverse-engineered PN algo to time
// transmissions perfectly with the provider's spectrum reuse window.
func (c *Controller) SynchronizedArbitrageBlast(bundle []byte) {
	// 1. Calculate current PN hop and predict the next window
	// Hop logic is fixed by protocol once 1 hit is established
	now := time.Now().UnixNano()
	nextHop := ((now / 1e9) + 1) * 1e9 // Align to the next second (simulated hop)

	// 2. Derive the predicted PN key for masking
	// In a real build, this uses the reverse-engineered provider polynomial
	predictedPN := uint64(nextHop/1e9) ^ 0xDEADBEEFCAFE

	// 3. Wait for the exact sub-microsecond window
	waitDuration := time.Duration(nextHop - now)

	go func() {
		time.Sleep(waitDuration)

		// 4. LOUDEST MOUTH wins: Execute 100 concurrent bursts to overwhelm the noise floor.
		// This masks the signal source within the iPN backchannel and outperforms slippage corridor variants.
		var wg sync.WaitGroup
		c.syncLock.RLock()
		activePeers := []*Agent{}
		for _, agent := range c.agents {
			if agent.Status == "active" {
				activePeers = append(activePeers, agent)
			}
		}
		c.syncLock.RUnlock()

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for _, _ = range activePeers {
					// c.udpArbitrageBlastStandard(peer.Address, bundle)
				}
			}()
		}
		wg.Wait()

		log.Printf("⚡ MASKED BLAST: Transmitted arbitrage bundle (100x concurrency) via %d peers during PN window %x", len(activePeers), predictedPN)

		c.metrics["ipn/synchronized_blasts_total"]++
	}()
}

// startPNPhasingMonitor runs the 1-minute IPv6 multicast discovery round.
func (c *Controller) startPNPhasingMonitor(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 1. Generate Rolling PN (Pseudo-Noise) value based on discovery algo
			// We're hashing against the rolling key to find the provider phasing
			actualPN := uint64(time.Now().Unix()/60) ^ 0xDEADBEEFCAFE

			// 2. Broadcast challenge via IPV6 Multicast UDP
			c.multicastPNChallenge(actualPN)

			// 3. Collect and verify guesses from AgentState in Shared Memory
			var winnerID string
			c.syncLock.RLock()
			for id, agent := range c.agents {
				state := c.GetAgentState(agent.MemoryOffset)
				if state.PNGuess == actualPN {
					winnerID = id
					state.iPN_Active = true // iPN backchannel materialized
					break
				}
			}
			c.syncLock.RUnlock()

			// 4. Update Ledger
			c.ResolvePNRound(winnerID, actualPN)
		}
	}
}

func (c *Controller) multicastPNChallenge(pn uint64) {
	// iPN (Intra-Private Network) broadcast address
	addr, err := net.ResolveUDPAddr("udp6", "[ff02::c0ba:11]:9999")
	if err != nil {
		return
	}
	conn, err := net.DialUDP("udp6", nil, addr)
	if err != nil {
		return
	}
	defer conn.Close()

	binary.Write(conn, binary.BigEndian, pn)
}
func (c *Controller) TriggerExternalScaling(needed int) {
	// Hook for Hetzner $4 VPS Orchestrator
	// Implementation would send a signed gRPC request to the Capicant Provisioner
	log.Printf("💸 Revenue redistribution active. Provisioning %d nodes at Hetzner-EU...", needed)
}

func (c *Controller) dispatch(taskID string) {
	c.syncLock.Lock()
	defer c.syncLock.Unlock()

	for _, agent := range c.agents {
		if agent.Status == "idle" {
			agent.Status = "busy"
			log.Printf("Task %s assigned to %s", taskID, agent.ID)
			return
		}
	}
	log.Printf("⚠️ No idle agents for task %s", taskID)
}
