// arbitrage.go
package sovereign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pqr-info/sovereign-mesh/addressing"
)

type ArbitrageDaemon struct {
	mesh       *Controller
	Inbound    <-chan HFTArbitrageSignal
	httpClient *http.Client
	pqrURL     string
}

type PQRCreateTicketRequest struct {
	Summary     string            `json:"summary"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type PQRCreateTicketResponse struct {
	ID string `json:"id"`
}

type PQRUpdateTicketRequest struct {
	Status string `json:"status"`
}

func GenerateSolvencyProof(sig HFTArbitrageSignal) ([]byte, error) {
	// stub: Pedersen + Bulletproofs via NPU in real impl
	return []byte("mock-proof"), nil
}

func IsProfitable(sig HFTArbitrageSignal) bool {
	// Use default parameters for traction check
	if !sig.ValidateTraction(sig.Price, 0.05, 10*time.Second) {
		return false
	}
	// add spread, fees, slippage logic here
	return true
}

func NewArbitrageDaemon(mesh *Controller, in <-chan HFTArbitrageSignal, pqrURL string) *ArbitrageDaemon {
	return &ArbitrageDaemon{
		mesh:       mesh,
		Inbound:    in,
		httpClient: &http.Client{Timeout: 3 * time.Second},
		pqrURL:     pqrURL,
	}
}

func (d *ArbitrageDaemon) Run() {
	for sig := range d.Inbound {
		d.handleSignal(sig)
	}
}

func (d *ArbitrageDaemon) handleSignal(sig HFTArbitrageSignal) {
	if !IsProfitable(sig) {
		return
	}

	if _, err := GenerateSolvencyProof(sig); err != nil {
		// optionally log
		return
	}

	// Resolve 5-D coordinate
	coord := d.mesh.Address5D.Resolve(fmt.Sprintf("%d", sig.AssetID), time.Unix(0, sig.Timestamp))

	ticketID, err := d.createTicket(sig, coord)
	if err != nil {
		return
	}

	_ = d.completeTicket(ticketID)
}

func (d *ArbitrageDaemon) createTicket(sig HFTArbitrageSignal, coord addressing.Address5DCoord) (string, error) {
	meta := map[string]string{
		"symbol": fmt.Sprintf("Asset:%d", sig.AssetID),
		"side":   fmt.Sprintf("%d", sig.Side),
		"5d":     coord.String(),
	}

	reqBody := PQRCreateTicketRequest{
		Summary: fmt.Sprintf("Arbitrage trade Asset:%d Seq:%d", sig.AssetID, sig.Sequence),
		Description: fmt.Sprintf(
			"Executed arbitrage: Seq:%d\nPrice: %.8f\nSize: %.8f\nSide: %d\nTimestamp: %d\n5D: %s",
			sig.Sequence, sig.Price, sig.Volume, sig.Side, sig.Timestamp, coord.String(),
		),
		Tags:     []string{"arbitrage", "hft", "compliance"},
		Metadata: meta,
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := d.pqrURL + "/REST/2.0/ticket"
	resp, err := d.httpClient.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("PQR create ticket status %d", resp.StatusCode)
	}

	var parsed PQRCreateTicketResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}

	return parsed.ID, nil
}

func (d *ArbitrageDaemon) completeTicket(id string) error {
	reqBody := PQRUpdateTicketRequest{Status: "COMPLETED"}
	b, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/REST/2.0/ticket/%s", d.pqrURL, id)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("PQR complete ticket status %d", resp.StatusCode)
	}
	return nil
}
