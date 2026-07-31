package mock

import (
	"context"
	"strings"

	"github.com/inverse-inc/packetfence/hospitality/internal/services/pms"
)

type Connector struct {
	Stays   []pms.GuestStay
	Charges map[string]string
}

func New(stays []pms.GuestStay) *Connector {
	return &Connector{Stays: stays, Charges: make(map[string]string)}
}

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

func (c *Connector) PostChargeToFolio(_ context.Context, reservationID string, _ int, _ string) (string, error) {
	reference := "mock-folio-" + reservationID
	c.Charges[reservationID] = reference
	return reference, nil
}

func (c *Connector) HealthCheck(context.Context) error { return nil }
