package pf

import (
	"context"
	"time"
)

type Identity struct {
	PID string
	MAC string
}

type Gateway interface {
	CreateGuestIdentity(ctx context.Context, propertyID, mac, pid, role string, expiresAt time.Time) (*Identity, error)
	AssignRole(ctx context.Context, mac, role string) error
	GrantAccess(ctx context.Context, mac string) error
	RevokeAccess(ctx context.Context, mac string) error
	SetExpiration(ctx context.Context, mac string, expiresAt time.Time) error
	DisconnectSession(ctx context.Context, mac string) error
}
