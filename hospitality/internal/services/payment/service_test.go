package payment

import (
	"context"
	"testing"
	"time"

	"github.com/inverse-inc/packetfence/hospitality/internal/models"
)

type store struct {
	session     *models.GuestSession
	plan        *models.WiFiPlan
	payment     *models.Payment
	entitlement *models.WiFiEntitlement
}

func (s *store) GetSession(context.Context, string) (*models.GuestSession, error) {
	return s.session, nil
}
func (s *store) GetPlan(context.Context, string) (*models.WiFiPlan, error) { return s.plan, nil }
func (s *store) GetPaymentByIdempotencyKey(context.Context, string, string) (*models.Payment, error) {
	if s.payment == nil {
		return nil, errMissing
	}
	return s.payment, nil
}
func (s *store) CreatePayment(_ context.Context, p *models.Payment) (*models.Payment, error) {
	p.ID = "payment-1"
	s.payment = p
	return p, nil
}
func (s *store) CreateEntitlement(_ context.Context, e *models.WiFiEntitlement) (*models.WiFiEntitlement, error) {
	e.ID = "entitlement-1"
	s.entitlement = e
	return e, nil
}

var errMissing = &missingError{}

type missingError struct{}

func (*missingError) Error() string { return "missing" }

func TestUpgradeActivatesOnlyAfterFolioCharge(t *testing.T) {
	now := time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC)
	db := &store{session: &models.GuestSession{ID: "session-1", PropertyID: "property-1", Status: "active", ExpiresAt: now.Add(24 * time.Hour), ReservationID: stringPtr("reservation-1")}, plan: &models.WiFiPlan{ID: "plan-1", PropertyID: "property-1", Name: "Premium", PriceCents: 990, Currency: "USD"}}
	folio := NewMockFolio()
	svc := &Service{Store: db, Folio: folio, Now: func() time.Time { return now }}
	result, err := svc.Upgrade(context.Background(), UpgradeRequest{PropertyID: "property-1", SessionID: "session-1", PlanID: "plan-1", IdempotencyKey: "idem-1"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Payment.Status != "completed" || result.Entitlement.Status != "active" {
		t.Fatal("upgrade was not activated")
	}
}

func TestFailedFolioDoesNotCreatePayment(t *testing.T) {
	db := &store{session: &models.GuestSession{ID: "session-1", PropertyID: "property-1", Status: "active"}, plan: &models.WiFiPlan{ID: "plan-1", PropertyID: "property-1", PriceCents: 990}}
	folio := NewMockFolio()
	folio.Fail = true
	svc := &Service{Store: db, Folio: folio}
	if _, err := svc.Upgrade(context.Background(), UpgradeRequest{PropertyID: "property-1", SessionID: "session-1", PlanID: "plan-1", IdempotencyKey: "idem-1"}); err != ErrPaymentFailed {
		t.Fatalf("got %v", err)
	}
	if db.payment != nil {
		t.Fatal("payment created after failed folio charge")
	}
}

func stringPtr(v string) *string { return &v }
