package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"time"
)

// Signer handles ECDSA signing of cognitive vectors.
type Signer struct {
	privateKey *ecdsa.PrivateKey
	publicKey  ecdsa.PublicKey
}

// NewSigner generates a new ECDSA P-256 keypair.
func NewSigner() (*Signer, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return &Signer{
		privateKey: priv,
		publicKey:  priv.PublicKey,
	}, nil
}

// ------------------------------------------------------------
// CognitiveEnvelope is the signed payload returned by the API.
// ------------------------------------------------------------

type CognitiveEnvelope struct {
	NodeID    string    `json:"node_id"`
	Timestamp time.Time `json:"timestamp"`
	Vector    []float64 `json:"vector"`
	Alpha     float64   `json:"alpha"`
	Signature string    `json:"signature"`
	PubKey    string    `json:"pubkey"`
}

// ------------------------------------------------------------
// SignCognition signs Ck + α + NodeID + timestamp.
// ------------------------------------------------------------

func (s *Signer) SignCognition(nodeID string, Ck []float64, alpha float64) (*CognitiveEnvelope, error) {
	env := CognitiveEnvelope{
		NodeID:    nodeID,
		Timestamp: time.Now().UTC(),
		Vector:    Ck,
		Alpha:     alpha,
	}

	// Serialize envelope without signature
	raw, err := json.Marshal(struct {
		NodeID    string    `json:"node_id"`
		Timestamp time.Time `json:"timestamp"`
		Vector    []float64 `json:"vector"`
		Alpha     float64   `json:"alpha"`
	}{
		NodeID:    env.NodeID,
		Timestamp: env.Timestamp,
		Vector:    env.Vector,
		Alpha:     env.Alpha,
	})
	if err != nil {
		return nil, err
	}

	// Hash the payload
	hash := sha256.Sum256(raw)

	// Sign the hash
	r, sVal, err := ecdsa.Sign(rand.Reader, s.privateKey, hash[:])
	if err != nil {
		return nil, err
	}

	// Encode signature as base64(r||s)
	sigBytes, _ := json.Marshal([]string{
		base64.StdEncoding.EncodeToString(r.Bytes()),
		base64.StdEncoding.EncodeToString(sVal.Bytes()),
	})
	env.Signature = base64.StdEncoding.EncodeToString(sigBytes)

	// Encode public key
	pubBytes, _ := json.Marshal(struct {
		X string `json:"x"`
		Y string `json:"y"`
	}{
		X: base64.StdEncoding.EncodeToString(s.publicKey.X.Bytes()),
		Y: base64.StdEncoding.EncodeToString(s.publicKey.Y.Bytes()),
	})
	env.PubKey = base64.StdEncoding.EncodeToString(pubBytes)

	return &env, nil
}
