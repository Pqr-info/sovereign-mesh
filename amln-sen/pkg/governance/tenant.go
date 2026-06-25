package governance

import (
	"errors"
	"sync"
	"time"
)

type TenantState int

const (
	Created TenantState = iota
	Active
	Suspended
	Deleted
)

type TenantPlan string

const (
	PlanFree       TenantPlan = "Free"
	PlanStandard   TenantPlan = "Standard"
	PlanEnterprise TenantPlan = "Enterprise"
)

// Tenant represents a TLS-27 managed client space
type Tenant struct {
	TenantID               string       `json:"tenant_id"`
	OwnerAddress           string       `json:"owner_address"` // 81-char address
	Plan                   TenantPlan   `json:"plan"`
	State                  TenantState  `json:"state"`
	Go27Balance            float64      `json:"go27_balance"`
	ComputeCyclesRemaining int64        `json:"compute_cycles_remaining"`
	MemoryWindowExpiration time.Time    `json:"memory_window_expiration"`
	PQRNamespace           string       `json:"pqr_namespace"`
	CreatedAt              time.Time    `json:"created_at"`
	mu                     sync.Mutex
}

// TenantLifecycleManager coordinates TLS-27 lifecycles
type TenantLifecycleManager struct {
	Mu      sync.RWMutex
	Tenants map[string]*Tenant `json:"tenants"`
}

func NewTenantLifecycleManager() *TenantLifecycleManager {
	return &TenantLifecycleManager{
		Tenants: make(map[string]*Tenant),
	}
}

// CreateTenant initializes a new tenant context under TLS-27 rules
func (m *TenantLifecycleManager) CreateTenant(id string, ownerAddr string, plan TenantPlan) (*Tenant, error) {
	if len(ownerAddr) != 81 {
		return nil, errors.New("owner address must be exactly 81 characters")
	}

	m.Mu.Lock()
	defer m.Mu.Unlock()

	if _, exists := m.Tenants[id]; exists {
		return nil, errors.New("tenant ID already exists")
	}

	t := &Tenant{
		TenantID:               id,
		OwnerAddress:           ownerAddr,
		Plan:                   plan,
		State:                  Created,
		Go27Balance:            0.0,
		ComputeCyclesRemaining: 0,
		MemoryWindowExpiration: time.Now().UTC(),
		PQRNamespace:           "pqr-ns-" + id,
		CreatedAt:              time.Now().UTC(),
	}

	if plan == PlanFree {
		// Free Tier has 7 days retention
		t.MemoryWindowExpiration = time.Now().UTC().AddDate(0, 0, 7)
		t.State = Active
	}

	m.Tenants[id] = t
	return t, nil
}

// PurchaseGo27 implements the default service rate ($0.81) exchange
func (t *Tenant) PurchaseGo27(amount float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Go27Balance += amount
	// $0.81 yields 9 compute cycles and extends memory window by 27 days
	multipliers := mathFloor(amount / 0.81)
	if multipliers > 0 {
		t.ComputeCyclesRemaining += int64(multipliers * 9)
		if t.MemoryWindowExpiration.Before(time.Now().UTC()) {
			t.MemoryWindowExpiration = time.Now().UTC()
		}
		t.MemoryWindowExpiration = t.MemoryWindowExpiration.AddDate(0, 0, int(multipliers*27))
		t.State = Active
	}
}

// ConsumeCycle subtracts compute iterations
func (t *Tenant) ConsumeCycle() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Plan == PlanFree {
		// Free Tier is subject to minimal runtimes
		return true
	}

	if t.State != Active {
		return false
	}

	if t.ComputeCyclesRemaining > 0 {
		t.ComputeCyclesRemaining--
		return true
	}

	t.State = Suspended
	return false
}

// Suspend pauses operation
func (t *Tenant) Suspend() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.State = Suspended
}

// Delete prunes all resources
func (m *TenantLifecycleManager) DeleteTenant(id string) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	t, exists := m.Tenants[id]
	if !exists {
		return errors.New("tenant not found")
	}

	t.mu.Lock()
	t.State = Deleted
	t.Go27Balance = 0.0
	t.ComputeCyclesRemaining = 0
	t.mu.Unlock()

	delete(m.Tenants, id)
	return nil
}

func mathFloor(val float64) float64 {
	return float64(int(val))
}
