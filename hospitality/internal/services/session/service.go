package session

import (
	"context"
	"fmt"
	"time"

	"github.com/inverse-inc/packetfence/hospitality/internal/models"
	"github.com/inverse-inc/packetfence/hospitality/internal/services/pf"
)

type Store interface {
	GetSession(context.Context, string) (*models.GuestSession, error)
	UpdateSession(context.Context, *models.GuestSession) error
	ListSessionsExpiringBefore(context.Context, string, time.Time) ([]models.GuestSession, error)
}

type Service struct {
	Store       Store
	PacketFence pf.Gateway
	Now         func() time.Time
}

func (s *Service) Revoke(ctx context.Context, id string) (*models.GuestSession, error) {
	item, err := s.Store.GetSession(ctx, id)
	if err != nil {
		return nil, err
	}
	if item.Status != "revoked" {
		if err := s.PacketFence.RevokeAccess(ctx, item.PacketFenceMAC); err != nil {
			return nil, fmt.Errorf("revoking PacketFence access: %w", err)
		}
		if err := s.PacketFence.DisconnectSession(ctx, item.PacketFenceMAC); err != nil {
			return nil, fmt.Errorf("disconnecting PacketFence session: %w", err)
		}
		item.Status = "revoked"
		now := time.Now().UTC()
		if s.Now != nil {
			now = s.Now().UTC()
		}
		item.EndedAt = &now
		if err := s.Store.UpdateSession(ctx, item); err != nil {
			return nil, err
		}
	}
	return item, nil
}

func (s *Service) Reconcile(ctx context.Context, propertyID string, before time.Time) (int, error) {
	items, err := s.Store.ListSessionsExpiringBefore(ctx, propertyID, before)
	if err != nil {
		return 0, err
	}
	count := 0
	for i := range items {
		if _, err := s.Revoke(ctx, items[i].ID); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
