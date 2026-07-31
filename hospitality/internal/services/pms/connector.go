package pms

import (
	"context"
	"errors"
	"time"
)

var ErrGuestStayNotFound = errors.New("guest stay not found")

type GuestStay struct {
	PropertyID    string
	ReservationID string
	GuestID       string
	RoomNumber    string
	FirstName     string
	LastName      string
	CheckInAt     time.Time
	CheckOutAt    time.Time
	Status        string
	VIP           bool
}

type Connector interface {
	ValidateGuestStay(ctx context.Context, propertyID, roomNumber, lastName string) (*GuestStay, error)
	PostChargeToFolio(ctx context.Context, reservationID string, amountCents int, description string) (string, error)
	HealthCheck(ctx context.Context) error
}
