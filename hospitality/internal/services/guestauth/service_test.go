package guestauth

import (
	"context"
	"testing"
	"time"

	"github.com/inverse-inc/packetfence/hospitality/internal/models"
	pfmock "github.com/inverse-inc/packetfence/hospitality/internal/services/pf/mock"
	"github.com/inverse-inc/packetfence/hospitality/internal/services/pms"
	pmsmock "github.com/inverse-inc/packetfence/hospitality/internal/services/pms/mock"
)

type memoryStore struct {
	profile *models.GuestProfile
	res     *models.Reservation
	sess    *models.GuestSession
}

func (m *memoryStore) UpsertGuestProfile(_ context.Context, p *models.GuestProfile) (*models.GuestProfile, error) {
	p.ID = "profile-1"
	m.profile = p
	return p, nil
}
func (m *memoryStore) UpsertReservation(_ context.Context, r *models.Reservation) (*models.Reservation, error) {
	r.ID = "reservation-1"
	m.res = r
	return r, nil
}
func (m *memoryStore) CreateSession(_ context.Context, s *models.GuestSession) (*models.GuestSession, error) {
	s.ID = "session-1"
	m.sess = s
	return s, nil
}

func TestAuthenticateNormalizesAndExpiresAtCheckout(t *testing.T) {
	now := time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC)
	pmsConn := pmsmock.New([]pms.GuestStay{{
		PropertyID:    "p1",
		ReservationID: "r1", GuestID: "g1", RoomNumber: "101", FirstName: "Ada", LastName: "O'Connor",
		CheckInAt: now.Add(-24 * time.Hour), CheckOutAt: now.Add(24 * time.Hour), Status: "checked_in",
	}})
	store := &memoryStore{}
	pfConn := pfmock.New()
	svc := &Service{Store: store, PMS: pmsConn, PacketFence: pfConn, HashSecret: "test", GracePeriod: time.Hour, Now: func() time.Time { return now }}
	result, err := svc.Authenticate(context.Background(), Request{PropertyID: "p1", RoomNumber: "101", LastName: "  O'CONNOR ", MACAddress: "aa:bb:cc:dd:ee:ff", PlanRole: "guest"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Session.ExpiresAt != now.Add(25*time.Hour) {
		t.Fatalf("unexpected expiry: %s", result.Session.ExpiresAt)
	}
	if !pfConn.Identities["aa:bb:cc:dd:ee:ff"].Granted {
		t.Fatal("access was not granted")
	}
}

func TestAuthenticateUsesGenericError(t *testing.T) {
	svc := &Service{PMS: pmsmock.New(nil), Store: &memoryStore{}, PacketFence: pfmock.New(), GracePeriod: time.Hour}
	if _, err := svc.Authenticate(context.Background(), Request{PropertyID: "p1", RoomNumber: "999", LastName: "Unknown", MACAddress: "aa:bb:cc:dd:ee:ff"}); err != ErrInvalidGuestDetails {
		t.Fatalf("expected generic validation error, got %v", err)
	}
}
