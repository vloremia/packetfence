package payment

import (
	"context"
	"errors"
	"sync"
)

type MockFolio struct {
	mu      sync.Mutex
	Charges map[string]string
	Fail    bool
}

func NewMockFolio() *MockFolio { return &MockFolio{Charges: make(map[string]string)} }

func (m *MockFolio) PostChargeToFolio(_ context.Context, reservationID string, _ int, _ string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Fail {
		return "", errors.New("mock folio failure")
	}
	reference := "folio-" + reservationID
	m.Charges[reservationID] = reference
	return reference, nil
}
