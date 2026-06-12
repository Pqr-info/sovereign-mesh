package sovereign

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/pqr-info/sovereign-mesh/proto"
	"google.golang.org/grpc"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
)

// LogAccountingEvent sends a RADIUS accounting packet to the AAAA backend.
func (c *Controller) LogAccountingEvent(username, sessionID string, statusType rfc2866.AcctStatusType, inputOctets, outputOctets uint32) {
	if c.radiusServer == "" {
		log.Printf("[RADIUS-STUB] Acct-%v for user %s (Session: %s)", statusType, username, sessionID)
		return
	}

	packet := radius.New(radius.CodeAccountingRequest, []byte(c.radiusSecret))
	rfc2865.UserName_SetString(packet, username)
	rfc2866.AcctSessionID_SetString(packet, sessionID)
	rfc2866.AcctStatusType_Set(packet, statusType)
	rfc2866.AcctInputOctets_Set(packet, rfc2866.AcctInputOctets(inputOctets))
	rfc2866.AcctOutputOctets_Set(packet, rfc2866.AcctOutputOctets(outputOctets))

	// Native Silicon/Hardware Telemetry Integration
	nasIP := net.ParseIP("127.0.0.1")
	if ip := net.ParseIP(c.location); ip != nil {
		nasIP = ip
	}
	rfc2865.NASIPAddress_Set(packet, nasIP)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	log.Printf("📡 RADIUS AAAA: Transmitting Accounting-%v for %s...", statusType, username)
	response, err := radius.Exchange(ctx, packet, c.radiusServer)
	if err != nil {
		log.Printf("⚠️ RADIUS Error: Failed to exchange accounting packet: %v", err)
		return
	}

	if response.Code == radius.CodeAccountingResponse {
		log.Printf("✅ RADIUS Success: Accounting event logged for %s", username)
	}

	// Starchart Integration: Mirror accounting to the mesh database via gRPC
	go func() {
		// We dial port 1111 (Starchart DB host)
		target := "127.0.0.1:1111"
		conn, err := grpc.Dial(target, grpc.WithInsecure())
		if err != nil {
			log.Printf("⚠️ Starchart Error: Failed to dial database: %v", err)
			return
		}
		defer conn.Close()

		client := proto.NewAgentSyncClient(conn)
		_, err = client.RecordAccounting(context.Background(), &proto.AccountingRecord{
			Username:     username,
			SessionId:    sessionID,
			StatusType:   statusType.String(),
			InputOctets:  inputOctets,
			OutputOctets: outputOctets,
			Timestamp:    time.Now().Format(time.RFC3339),
		})
		if err != nil {
			log.Printf("⚠️ Starchart Error: Failed to record accounting: %v", err)
		} else {
			log.Printf("🌌 Starchart: Accounting mirror successful for %s", username)
		}
	}()
}

// TrackTransaction logs a financial transaction (CBC) to the RADIUS accounting layer.
func (c *Controller) TrackTransaction(agentID string, amount float64, txType string) {
	sessionID := "TX-" + time.Now().Format("20060102150405")
	c.LogAccountingEvent(agentID, sessionID, rfc2866.AcctStatusType_Value_InterimUpdate, uint32(amount*100), 0)
}

// TrackProcessMigration audits the teleportation of a process across nodes.
func (c *Controller) TrackProcessMigration(pid int32, owner, fromNode, toNode string) {
	sessionID := "PROC-" + string(pid)
	log.Printf("🔄 PROCESS TELEPORT: %d (%s) | %s -> %s", pid, owner, fromNode, toNode)
	
	// Audit "Stop" on source node
	c.LogAccountingEvent(owner, sessionID, rfc2866.AcctStatusType_Value_Stop, 0, 0)
	
	// Audit "Start" on target node
	c.LogAccountingEvent(owner, sessionID, rfc2866.AcctStatusType_Value_Start, 0, 0)
}
