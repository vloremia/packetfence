package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/inverse-inc/packetfence/hospitality/internal/services/pf"
)

type Gateway struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func New(baseURL, token string) *Gateway {
	return &Gateway{BaseURL: strings.TrimRight(baseURL, "/"), Token: token, HTTPClient: &http.Client{Timeout: 10 * time.Second}}
}

func (g *Gateway) request(ctx context.Context, method, path string, body any, out any) error {
	var payload *bytes.Reader
	if body == nil {
		payload = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, g.BaseURL+path, payload)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if g.Token != "" {
		req.Header.Set("Authorization", "Bearer "+g.Token)
	}
	res, err := g.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("PacketFence API returned status %d", res.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(res.Body).Decode(out)
	}
	return nil
}

func (g *Gateway) CreateGuestIdentity(ctx context.Context, propertyID, mac, pid, role string, expiresAt time.Time) (*pf.Identity, error) {
	var out struct {
		MAC string `json:"mac"`
		PID string `json:"pid"`
	}
	err := g.request(ctx, http.MethodPost, "/api/v1/nodes", map[string]any{"mac": mac, "pid": pid, "category": role, "unregdate": expiresAt.UTC().Format(time.RFC3339), "property_id": propertyID}, &out)
	if err != nil {
		return nil, err
	}
	if out.MAC == "" {
		out.MAC = mac
	}
	if out.PID == "" {
		out.PID = pid
	}
	return &pf.Identity{MAC: out.MAC, PID: out.PID}, nil
}

func (g *Gateway) AssignRole(ctx context.Context, mac, role string) error {
	return g.request(ctx, http.MethodPatch, "/api/v1/node/"+mac, map[string]any{"category": role}, nil)
}
func (g *Gateway) GrantAccess(ctx context.Context, mac string) error {
	return g.request(ctx, http.MethodPost, "/api/v1/node/"+mac+"/register", nil, nil)
}
func (g *Gateway) RevokeAccess(ctx context.Context, mac string) error {
	return g.request(ctx, http.MethodPost, "/api/v1/node/"+mac+"/deregister", nil, nil)
}
func (g *Gateway) SetExpiration(ctx context.Context, mac string, expiresAt time.Time) error {
	return g.request(ctx, http.MethodPatch, "/api/v1/node/"+mac, map[string]any{"unregdate": expiresAt.UTC().Format(time.RFC3339)}, nil)
}
func (g *Gateway) DisconnectSession(ctx context.Context, mac string) error {
	return g.request(ctx, http.MethodPost, "/api/v1/node/"+mac+"/disconnect", nil, nil)
}
