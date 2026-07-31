package mock

import (
	"context"
	"sync"
	"time"

	"github.com/inverse-inc/packetfence/hospitality/internal/services/pf"
)

type IdentityState struct {
	pf.Identity
	Role      string
	ExpiresAt time.Time
	Granted   bool
}

type Gateway struct {
	mu         sync.Mutex
	Identities map[string]IdentityState
}

func New() *Gateway { return &Gateway{Identities: make(map[string]IdentityState)} }

func (g *Gateway) CreateGuestIdentity(_ context.Context, _ string, mac, pid, role string, expiresAt time.Time) (*pf.Identity, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Identities[mac] = IdentityState{Identity: pf.Identity{PID: pid, MAC: mac}, Role: role, ExpiresAt: expiresAt}
	identity := g.Identities[mac].Identity
	return &identity, nil
}

func (g *Gateway) AssignRole(_ context.Context, mac, role string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	state := g.Identities[mac]
	state.Role = role
	g.Identities[mac] = state
	return nil
}

func (g *Gateway) GrantAccess(_ context.Context, mac string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	state := g.Identities[mac]
	state.Granted = true
	g.Identities[mac] = state
	return nil
}

func (g *Gateway) RevokeAccess(_ context.Context, mac string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	state := g.Identities[mac]
	state.Granted = false
	g.Identities[mac] = state
	return nil
}

func (g *Gateway) SetExpiration(_ context.Context, mac string, expiresAt time.Time) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	state := g.Identities[mac]
	state.ExpiresAt = expiresAt
	g.Identities[mac] = state
	return nil
}

func (g *Gateway) DisconnectSession(_ context.Context, mac string) error {
	return g.RevokeAccess(context.Background(), mac)
}
