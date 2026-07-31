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
	server := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: (&api.Router{Store: data, GuestAuth: guestAuth, Sessions: sessions, AdminAPIKey: os.Getenv("HOSPITALITY_ADMIN_API_KEY"), RateLimit: cfg.RateLimitPerMinute}).Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("hospitality service listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
