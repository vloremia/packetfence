package mock

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/inverse-inc/packetfence/hospitality/internal/services/pms"
)

type Connector struct {
	Stays   []pms.GuestStay
	Charges map[string]string
}

func New(stays []pms.GuestStay) *Connector {
	return &Connector{Stays: stays, Charges: make(map[string]string)}
}

func (c *Connector) ValidateConfiguration(context.Context) error { return nil }

func (c *Connector) ValidateGuestStay(_ context.Context, propertyID, roomNumber, lastName string) (*pms.GuestStay, error) {
	for i := range c.Stays {
		stay := &c.Stays[i]
		if stay.PropertyID == propertyID && stay.Status == "checked_in" && stay.RoomNumber == roomNumber &&
			strings.EqualFold(stay.LastName, lastName) {
			copy := *stay
			return &copy, nil
		}
	}
	return nil, pms.ErrGuestStayNotFound
}

func (c *Connector) FindReservation(_ context.Context, _ string, reservationID string) (*pms.GuestStay, error) {
	for i := range c.Stays {
		if c.Stays[i].ReservationID == reservationID {
			item := c.Stays[i]
			return &item, nil
		}
	}
	return nil, pms.ErrGuestStayNotFound
}

func (c *Connector) FindInHouseGuest(_ context.Context, propertyID, roomNumber string) (*pms.GuestStay, error) {
	for i := range c.Stays {
		if c.Stays[i].PropertyID == propertyID && c.Stays[i].RoomNumber == roomNumber && c.Stays[i].Status == "checked_in" {
			item := c.Stays[i]
			return &item, nil
		}
	}
	return nil, pms.ErrGuestStayNotFound
}

func (c *Connector) GetGuestProfile(_ context.Context, guestID string) (*pms.GuestStay, error) {
	for i := range c.Stays {
		if c.Stays[i].GuestID == guestID {
			item := c.Stays[i]
			return &item, nil
		}
	}
	return nil, pms.ErrGuestStayNotFound
}

func (c *Connector) GetRoomStatus(_ context.Context, propertyID, roomNumber string) (string, error) {
	if _, err := c.FindInHouseGuest(context.Background(), propertyID, roomNumber); err == nil {
		return "occupied", nil
	}
	return "available", nil
}
func (c *Connector) GetCheckInDate(_ context.Context, reservationID string) (time.Time, error) {
	item, err := c.FindReservation(context.Background(), "", reservationID)
	if err != nil {
		return time.Time{}, err
	}
	return item.CheckInAt, nil
}
func (c *Connector) GetCheckOutDate(_ context.Context, reservationID string) (time.Time, error) {
	item, err := c.FindReservation(context.Background(), "", reservationID)
	if err != nil {
		return time.Time{}, err
	}
	return item.CheckOutAt, nil
}
func (c *Connector) GetReservationStatus(_ context.Context, reservationID string) (string, error) {
	item, err := c.FindReservation(context.Background(), "", reservationID)
	if err != nil {
		return "", err
	}
	return item.Status, nil
}

func (c *Connector) PostChargeToFolio(_ context.Context, reservationID string, _ int, _ string) (string, error) {
	reference := "mock-folio-" + reservationID
	c.Charges[reservationID] = reference
	return reference, nil
}

func (*Connector) UpdateGuestContactDetails(context.Context, string, pms.Contact) error { return nil }
func (*Connector) SubscribeToReservationUpdates(context.Context, string) (<-chan pms.ReservationUpdate, error) {
	return make(chan pms.ReservationUpdate), nil
}
func (*Connector) HandleWebhook(context.Context, []byte, http.Header) (*pms.ReservationUpdate, error) {
	return nil, nil
}

func (c *Connector) HealthCheck(context.Context) error { return nil }
