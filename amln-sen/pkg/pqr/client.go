package pqr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Session struct {
	BaseURL string
	AgentID string
	Client  *http.Client
}

func NewSession(baseURL, agentID string) *Session {
	return &Session{
		BaseURL: baseURL,
		AgentID: agentID,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// -------------------------------
// PQR Memory Structures
// -------------------------------

type MemoryPayload struct {
	MemoryType string                 `json:"memory_type"`
	Data       map[string]interface{} `json:"data"`
}

type TicketResponse struct {
	TicketID string `json:"ticket_id"`
}

type MemoryResponse struct {
	Memory map[string]interface{} `json:"memory"`
}

// -------------------------------
// Create a new memory ticket
// -------------------------------

func (s *Session) CreateMemory(ctx context.Context, title string, data map[string]interface{}) (string, error) {
	url := fmt.Sprintf("%s/REST/2.0/ticket", s.BaseURL)

	payload := map[string]interface{}{
		"title":            title,
		"creator_agent_id": s.AgentID,
		"data":             data,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out TicketResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}

	return out.TicketID, nil
}

// -------------------------------
// Store memory into an existing ticket
// -------------------------------

func (s *Session) StoreMemory(ctx context.Context, ticketID string, memoryType string, data map[string]interface{}) error {
	url := fmt.Sprintf("%s/REST/2.0/agent/%s/memory/%s", s.BaseURL, s.AgentID, ticketID)

	payload := MemoryPayload{
		MemoryType: memoryType,
		Data:       data,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	_, err := s.Client.Do(req)
	return err
}

// -------------------------------
// Recall memory from a ticket
// -------------------------------

func (s *Session) RecallMemory(ctx context.Context, ticketID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/REST/2.0/agent/%s/memory/%s", s.BaseURL, s.AgentID, ticketID)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out MemoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return out.Memory, nil
}

// -------------------------------
// Get all memory tickets for this agent
// -------------------------------

func (s *Session) GetAllMemories(ctx context.Context) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/REST/2.0/agent/%s/context", s.BaseURL, s.AgentID)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return out, nil
}
