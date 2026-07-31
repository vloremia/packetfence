package payment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/inverse-inc/packetfence/hospitality/internal/models"
)

var (
	ErrInvalidUpgrade = errors.New("invalid plan upgrade")
	ErrPaymentFailed  = errors.New("payment failed")
)

type Store interface {
	GetSession(context.Context, string) (*models.GuestSession, error)
	GetPlan(context.Context, string) (*models.WiFiPlan, error)
	GetPaymentByIdempotencyKey(context.Context, string, string) (*models.Payment, error)
	CreatePayment(context.Context, *models.Payment) (*models.Payment, error)
	CreateEntitlement(context.Context, *models.WiFiEntitlement) (*models.WiFiEntitlement, error)
}

type Folio interface {
	PostChargeToFolio(context.Context, string, int, string) (string, error)
}

type Service struct {
	Store Store
	Folio Folio
	Now   func() time.Time
}

type UpgradeRequest struct {
	PropertyID     string
	SessionID      string
	PlanID         string
	IdempotencyKey string
	PaymentMethod  string
}

type Result struct {
	Payment     *models.Payment         `json:"payment"`
	Entitlement *models.WiFiEntitlement `json:"entitlement"`
}

func (s *Service) Upgrade(ctx context.Context, req UpgradeRequest) (*Result, error) {
	if req.PropertyID == "" || req.SessionID == "" || req.PlanID == "" || req.IdempotencyKey == "" {
		return nil, ErrInvalidUpgrade
	}
	if existing, err := s.Store.GetPaymentByIdempotencyKey(ctx, req.PropertyID, req.IdempotencyKey); err == nil {
		return &Result{Payment: existing}, nil
	}
	session, err := s.Store.GetSession(ctx, req.SessionID)
	if err != nil || session.PropertyID != req.PropertyID || session.Status != "active" {
		return nil, ErrInvalidUpgrade
	}
	plan, err := s.Store.GetPlan(ctx, req.PlanID)
	if err != nil || plan.PropertyID != req.PropertyID || plan.PriceCents < 0 {
		return nil, ErrInvalidUpgrade
	}
	if plan.PriceCents == 0 && !plan.IsComplimentary {
		return nil, ErrInvalidUpgrade
	}
	method := req.PaymentMethod
	if method == "" {
		method = "folio"
	}
	if method != "folio" {
		return nil, ErrPaymentFailed
	}
	reservationID := ""
	if session.ReservationID != nil {
		reservationID = *session.ReservationID
	}
	folioRef, err := s.Folio.PostChargeToFolio(ctx, reservationID, plan.PriceCents, "Wi-Fi plan: "+plan.Name)
	if err != nil {
		return nil, ErrPaymentFailed
	}
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	expires := session.ExpiresAt
	if plan.SessionDurationMin != nil && now.Add(time.Duration(*plan.SessionDurationMin)*time.Minute).Before(expires) {
		expires = now.Add(time.Duration(*plan.SessionDurationMin) * time.Minute)
	}
	payment, err := s.Store.CreatePayment(ctx, &models.Payment{PropertyID: req.PropertyID, GuestSessionID: &session.ID, AmountCents: plan.PriceCents, Currency: plan.Currency, Status: "completed", PaymentMethod: method, FolioChargeID: folioRef, IdempotencyKey: req.IdempotencyKey})
	if err != nil {
		return nil, fmt.Errorf("saving payment: %w", err)
	}
	entitlement, err := s.Store.CreateEntitlement(ctx, &models.WiFiEntitlement{GuestSessionID: &session.ID, WifiPlanID: plan.ID, PropertyID: req.PropertyID, GuestProfileID: session.GuestProfileID, Status: "active", PriceCents: plan.PriceCents, Currency: plan.Currency, Source: "upgrade", PaymentID: &payment.ID, StartsAt: now, ExpiresAt: expires})
	if err != nil {
		return nil, fmt.Errorf("saving entitlement: %w", err)
	}
	return &Result{Payment: payment, Entitlement: entitlement}, nil
}

func (s *Service) ActivateComplimentary(ctx context.Context, req UpgradeRequest) (*Result, error) {
	if req.PropertyID == "" || req.SessionID == "" || req.PlanID == "" || req.IdempotencyKey == "" {
		return nil, ErrInvalidUpgrade
	}
	session, err := s.Store.GetSession(ctx, req.SessionID)
	if err != nil || session.PropertyID != req.PropertyID || session.Status != "active" {
		return nil, ErrInvalidUpgrade
	}
	plan, err := s.Store.GetPlan(ctx, req.PlanID)
	if err != nil || plan.PropertyID != req.PropertyID || !plan.IsComplimentary {
		return nil, ErrInvalidUpgrade
	}
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	entitlement, err := s.Store.CreateEntitlement(ctx, &models.WiFiEntitlement{GuestSessionID: &session.ID, WifiPlanID: plan.ID, PropertyID: req.PropertyID, GuestProfileID: session.GuestProfileID, Status: "active", PriceCents: 0, Currency: plan.Currency, Source: "complimentary", StartsAt: now, ExpiresAt: session.ExpiresAt})
	if err != nil {
		return nil, err
	}
	return &Result{Entitlement: entitlement}, nil
}

func (s *Service) ValidatePayment(_ context.Context, _ *models.Payment) error { return nil }

func (s *Service) String() string { return "payment" }

func (s *Service) ensure() error {
	if s.Store == nil || s.Folio == nil {
		return errors.New("payment service dependencies are not configured")
	}
	return nil
}
