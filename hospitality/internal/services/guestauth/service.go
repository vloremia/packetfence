package guestauth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/inverse-inc/packetfence/hospitality/internal/models"
	"github.com/inverse-inc/packetfence/hospitality/internal/services/pf"
	"github.com/inverse-inc/packetfence/hospitality/internal/services/pms"
	"github.com/inverse-inc/packetfence/hospitality/internal/util/names"
	"github.com/inverse-inc/packetfence/hospitality/internal/util/security"
)

var ErrInvalidGuestDetails = errors.New("guest details could not be validated")

type Store interface {
	UpsertGuestProfile(context.Context, *models.GuestProfile) (*models.GuestProfile, error)
	UpsertReservation(context.Context, *models.Reservation) (*models.Reservation, error)
	CreateSession(context.Context, *models.GuestSession) (*models.GuestSession, error)
}

type Service struct {
	Store       Store
	PMS         pms.Connector
	PacketFence pf.Gateway
	HashSecret  string
	GracePeriod time.Duration
	Now         func() time.Time
}

type Request struct {
	PropertyID string
	RoomNumber string
	LastName   string
	MACAddress string
	PlanID     *string
	PlanRole   string
	MaxDevices int
}

type Result struct {
	Session     *models.GuestSession
	Reservation *models.Reservation
	Profile     *models.GuestProfile
}

func (s *Service) Authenticate(ctx context.Context, req Request) (*Result, error) {
	if req.PropertyID == "" || req.RoomNumber == "" || req.LastName == "" || req.MACAddress == "" {
		return nil, ErrInvalidGuestDetails
	}
	normalized := names.Normalize(req.LastName)
	if normalized == "" {
		return nil, ErrInvalidGuestDetails
	}
	stay, err := s.PMS.ValidateGuestStay(ctx, req.PropertyID, req.RoomNumber, normalized)
	if err != nil {
		return nil, ErrInvalidGuestDetails
	}
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	expires := stay.CheckOutAt.Add(s.GracePeriod)
	if !expires.After(now) {
		return nil, ErrInvalidGuestDetails
	}
	lastNameHash := security.LookupHash(s.HashSecret, normalized)
	profile, err := s.Store.UpsertGuestProfile(ctx, &models.GuestProfile{
		PropertyID:       req.PropertyID,
		PMSGuestID:       stay.GuestID,
		FirstName:        stay.FirstName,
		LastNameHash:     lastNameHash,
		Source:           "room",
		FirstConnectedAt: &now,
		LastConnectedAt:  &now,
	})
	if err != nil {
		return nil, fmt.Errorf("saving guest profile: %w", err)
	}
	reservation, err := s.Store.UpsertReservation(ctx, &models.Reservation{
		PropertyID:       req.PropertyID,
		PMSReservationID: stay.ReservationID,
		PMSGuestID:       stay.GuestID,
		GuestProfileID:   &profile.ID,
		RoomNumber:       req.RoomNumber,
		GuestNameHash:    lastNameHash,
		LastNameHash:     lastNameHash,
		FirstName:        stay.FirstName,
		CheckInAt:        stay.CheckInAt,
		CheckOutAt:       stay.CheckOutAt,
		Status:           stay.Status,
		VIPStatus:        stay.VIP,
		LastSyncedAt:     &now,
	})
	if err != nil {
		return nil, fmt.Errorf("saving reservation: %w", err)
	}
	pid := "hotel-" + profile.ID
	identity, err := s.PacketFence.CreateGuestIdentity(ctx, req.PropertyID, req.MACAddress, pid, req.PlanRole, expires)
	if err != nil {
		return nil, fmt.Errorf("creating PacketFence identity: %w", err)
	}
	if err := s.PacketFence.GrantAccess(ctx, identity.MAC); err != nil {
		return nil, fmt.Errorf("granting PacketFence access: %w", err)
	}
	maxDevices := req.MaxDevices
	if maxDevices <= 0 {
		maxDevices = 2
	}
	session, err := s.Store.CreateSession(ctx, &models.GuestSession{
		PropertyID:     req.PropertyID,
		ReservationID:  &reservation.ID,
		GuestProfileID: &profile.ID,
		MACAddress:     req.MACAddress,
		PacketFenceMAC: identity.MAC,
		PacketFencePID: identity.PID,
		WifiPlanID:     req.PlanID,
		Status:         "active",
		ExpiresAt:      expires,
		StartedAt:      now,
		AuthMethod:     "room",
		DeviceCount:    1,
		MaxDevices:     maxDevices,
	})
	if err != nil {
		_ = s.PacketFence.RevokeAccess(ctx, identity.MAC)
		return nil, fmt.Errorf("saving guest session: %w", err)
	}
	return &Result{Session: session, Reservation: reservation, Profile: profile}, nil
}
