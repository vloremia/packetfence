package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/inverse-inc/packetfence/hospitality/internal/api"
	"github.com/inverse-inc/packetfence/hospitality/internal/config"
	"github.com/inverse-inc/packetfence/hospitality/internal/db"
	"github.com/inverse-inc/packetfence/hospitality/internal/services/guestauth"
	"github.com/inverse-inc/packetfence/hospitality/internal/services/payment"
	"github.com/inverse-inc/packetfence/hospitality/internal/services/pf"
	pfmock "github.com/inverse-inc/packetfence/hospitality/internal/services/pf/mock"
	pfrest "github.com/inverse-inc/packetfence/hospitality/internal/services/pf/rest"
	pmsmock "github.com/inverse-inc/packetfence/hospitality/internal/services/pms/mock"
	"github.com/inverse-inc/packetfence/hospitality/internal/services/session"
	"github.com/inverse-inc/packetfence/hospitality/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	database, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	if err := database.Migrate(ctx); err != nil {
		log.Fatal(err)
	}

	data := store.New(database)
	var packetFence pf.Gateway = pfmock.New()
	if cfg.PacketFenceAPIURL != "" {
		packetFence = pfrest.New(cfg.PacketFenceAPIURL, cfg.PacketFenceToken)
	}
	pms := pmsmock.New(nil)
	guestAuth := &guestauth.Service{Store: data, PMS: pms, PacketFence: packetFence, HashSecret: cfg.SecretKey, GracePeriod: cfg.GracePeriod}
	sessions := &session.Service{Store: data, PacketFence: packetFence}
	folio := payment.NewMockFolio()
	payments := &payment.Service{Store: data, Folio: folio}
	server := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: (&api.Router{Store: data, GuestAuth: guestAuth, Sessions: sessions, Payments: payments, AdminAPIKey: os.Getenv("HOSPITALITY_ADMIN_API_KEY"), RateLimit: cfg.RateLimitPerMinute}).Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	go reconcile(ctx, data, sessions)
	log.Printf("hospitality service listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func reconcile(ctx context.Context, data *store.Store, sessions *session.Service) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		properties, err := data.ListAllProperties(ctx)
		if err != nil {
			log.Printf("hospitality reconciliation failed: %v", err)
			continue
		}
		for _, property := range properties {
			if _, err := sessions.Reconcile(ctx, property.ID, time.Now().UTC()); err != nil {
				log.Printf("session reconciliation failed for property %s: %v", property.ID, err)
			}
			events, err := data.ListEvents(ctx, property.ID)
			if err != nil {
				continue
			}
			for i := range events {
				if events[i].Status != "completed" && !events[i].EndsAt.After(time.Now().UTC()) {
					events[i].Status = "completed"
					_ = data.UpdateEvent(ctx, &events[i])
				}
			}
		}
	}
}
