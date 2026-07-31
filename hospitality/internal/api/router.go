package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/inverse-inc/packetfence/hospitality/internal/models"
	"github.com/inverse-inc/packetfence/hospitality/internal/services/guestauth"
	"github.com/inverse-inc/packetfence/hospitality/internal/services/session"
	"github.com/inverse-inc/packetfence/hospitality/internal/store"
)

type Router struct {
	Store       *store.Store
	GuestAuth   *guestauth.Service
	Sessions    *session.Service
	AdminAPIKey string
	RateLimit   int
}

func (a *Router) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(requestID)
	r.Get("/health/live", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/health/ready", a.ready)
	limiter := newRateLimiter(a.RateLimit)
	r.Route("/api/v1", func(r chi.Router) {
		r.With(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if !limiter.allow(clientKey(req), time.Now().UTC()) {
					writeError(w, http.StatusTooManyRequests, "too many requests")
					return
				}
				next.ServeHTTP(w, req)
			})
		}).Post("/guest-auth/room", a.roomAuth)
		r.With(a.adminAuth).Get("/properties", a.listProperties)
		r.With(a.adminAuth).Post("/properties", a.createProperty)
		r.With(a.adminAuth).Get("/properties/{id}", a.getProperty)
		r.With(a.adminAuth).Get("/properties/{id}/wifi-plans", a.listPlans)
		r.With(a.adminAuth).Get("/guest-sessions", a.listSessions)
		r.With(a.adminAuth).Post("/guest-sessions/{id}/revoke", a.revokeSession)
	})
	return r
}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Correlation-ID")
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set("X-Correlation-ID", id)
		next.ServeHTTP(w, r)
	})
}

func (a *Router) adminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.AdminAPIKey == "" || r.Header.Get("X-API-Key") != a.AdminAPIKey {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *Router) ready(w http.ResponseWriter, r *http.Request) {
	if err := a.Store.DB.Pool.Ping(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

type roomAuthRequest struct {
	PropertyID string  `json:"property_id"`
	RoomNumber string  `json:"room_number"`
	LastName   string  `json:"last_name"`
	MACAddress string  `json:"mac_address"`
	PlanID     *string `json:"plan_id"`
	PlanRole   string  `json:"plan_role"`
	MaxDevices int     `json:"max_devices"`
}

func (a *Router) roomAuth(w http.ResponseWriter, r *http.Request) {
	var req roomAuthRequest
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&req) != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	result, err := a.GuestAuth.Authenticate(r.Context(), guestauth.Request{PropertyID: req.PropertyID, RoomNumber: req.RoomNumber, LastName: req.LastName, MACAddress: req.MACAddress, PlanID: req.PlanID, PlanRole: req.PlanRole, MaxDevices: req.MaxDevices})
	if errors.Is(err, guestauth.ErrInvalidGuestDetails) {
		writeError(w, http.StatusUnauthorized, "guest details could not be validated")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadGateway, "authentication service unavailable")
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (a *Router) listProperties(w http.ResponseWriter, r *http.Request) {
	org := r.Header.Get("X-Organisation-ID")
	if org == "" {
		writeError(w, http.StatusBadRequest, "X-Organisation-ID is required")
		return
	}
	items, err := a.Store.ListProperties(r.Context(), org)
	if err != nil {
		writeError(w, 500, "database error")
		return
	}
	writeJSON(w, 200, items)
}

func (a *Router) createProperty(w http.ResponseWriter, r *http.Request) {
	org := r.Header.Get("X-Organisation-ID")
	if org == "" {
		writeError(w, 400, "X-Organisation-ID is required")
		return
	}
	var p models.Property
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 32*1024)).Decode(&p) != nil {
		writeError(w, 400, "invalid request")
		return
	}
	p.OrganisationID = org
	item, err := a.Store.CreateProperty(r.Context(), org, &p)
	if err != nil {
		writeError(w, 500, "database error")
		return
	}
	writeJSON(w, 201, item)
}

func (a *Router) getProperty(w http.ResponseWriter, r *http.Request) {
	item, err := a.Store.GetProperty(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, 404, "not found")
		return
	}
	if org := r.Header.Get("X-Organisation-ID"); org == "" || item.OrganisationID != org {
		writeError(w, 404, "not found")
		return
	}
	writeJSON(w, 200, item)
}

func (a *Router) listPlans(w http.ResponseWriter, r *http.Request) {
	items, err := a.Store.ListPlans(r.Context(), chi.URLParam(r, "id"), false)
	if err != nil {
		writeError(w, 500, "database error")
		return
	}
	writeJSON(w, 200, items)
}
func (a *Router) listSessions(w http.ResponseWriter, r *http.Request) {
	propertyID := r.URL.Query().Get("property_id")
	if propertyID == "" {
		writeError(w, 400, "property_id is required")
		return
	}
	items, err := a.Store.ListSessions(r.Context(), propertyID, r.URL.Query().Get("status"))
	if err != nil {
		writeError(w, 500, "database error")
		return
	}
	writeJSON(w, 200, items)
}
func (a *Router) revokeSession(w http.ResponseWriter, r *http.Request) {
	item, err := a.Sessions.Revoke(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, 500, "could not revoke session")
		return
	}
	writeJSON(w, 200, item)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"message": message}})
}
