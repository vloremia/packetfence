package pms

import (
	"context"
	"errors"
	"net/http"
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

type Contact struct {
	Email  string
	Mobile string
}
type ReservationUpdate struct {
	ReservationID string
	Status        string
	CheckOutAt    time.Time
}

type Connector interface {
	ValidateConfiguration(ctx context.Context) error
	ValidateGuestStay(ctx context.Context, propertyID, roomNumber, lastName string) (*GuestStay, error)
	FindReservation(ctx context.Context, propertyID, reservationID string) (*GuestStay, error)
	FindInHouseGuest(ctx context.Context, propertyID, roomNumber string) (*GuestStay, error)
	GetGuestProfile(ctx context.Context, guestID string) (*GuestStay, error)
	GetRoomStatus(ctx context.Context, propertyID, roomNumber string) (string, error)
	GetCheckInDate(ctx context.Context, reservationID string) (time.Time, error)
	GetCheckOutDate(ctx context.Context, reservationID string) (time.Time, error)
	GetReservationStatus(ctx context.Context, reservationID string) (string, error)
	PostChargeToFolio(ctx context.Context, reservationID string, amountCents int, description string) (string, error)
	UpdateGuestContactDetails(ctx context.Context, guestID string, contact Contact) error
	SubscribeToReservationUpdates(ctx context.Context, propertyID string) (<-chan ReservationUpdate, error)
	HandleWebhook(ctx context.Context, payload []byte, headers http.Header) (*ReservationUpdate, error)
	HealthCheck(ctx context.Context) error
}
