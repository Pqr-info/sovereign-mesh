package substrate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInstantiateAgentRPC(t *testing.T) {
	pdbID := [4]byte{'Z', 'E', 'T', 'A'}
	sigHash := [32]byte{1, 2, 3}
	dialectFamily := [16]byte{4, 5, 6}
	sequence := []byte("MEEPQSDPSVEPPLSQETFSDLWKLLPENNVLSPLP")

	// Set up mock HTTP JSON-RPC server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content-type, got %s", r.Header.Get("Content-Type"))
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		if payload["method"] != "author_submitExtrinsic" {
			t.Errorf("expected method author_submitExtrinsic, got %v", payload["method"])
		}

		params, ok := payload["params"].([]interface{})
		if !ok || len(params) != 1 {
			t.Fatalf("expected params array with 1 element, got %v", payload["params"])
		}

		paramStr, ok := params[0].(string)
		if !ok {
			t.Fatalf("expected param to be string, got %T", params[0])
		}

		expectedHex := fmt.Sprintf("0x%x%x%x%x", pdbID, sigHash, dialectFamily, sequence)
		if paramStr != expectedHex {
			t.Errorf("expected param %s, got %s", expectedHex, paramStr)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"0x123456","id":1}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.InstantiateAgent(context.Background(), pdbID, sigHash, dialectFamily, sequence)
	if err != nil {
		t.Fatalf("failed to instantiate agent: %v", err)
	}
}

// TestOfflineFallback verifies that the client falls back gracefully when the server is offline
func TestOfflineFallback(t *testing.T) {
	client := NewClient("http://localhost:12345") // Unreachable port
	err := client.InstantiateAgent(context.Background(), [4]byte{}, [32]byte{}, [16]byte{}, []byte{})
	if err != nil {
		t.Fatalf("expected no error on offline fallback, got %v", err)
	}
}
