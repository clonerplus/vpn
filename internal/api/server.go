package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/clonerplus/vpn-manager/internal/config"
	"github.com/clonerplus/vpn-manager/internal/db"
	"github.com/gorilla/mux"
)

func Run(cfg *config.Config, store *db.Store) {
	h := NewHandler(store)
	r := mux.NewRouter()

	// Health
	r.HandleFunc("/health", h.Health).Methods("GET")

	// Users
	r.HandleFunc("/api/users", h.ListUsers).Methods("GET")
	r.HandleFunc("/api/users", h.CreateUser).Methods("POST")

	// Plans
	r.HandleFunc("/api/plans", h.ListPlans).Methods("GET")
	r.HandleFunc("/api/plans", h.CreatePlan).Methods("POST")
	r.HandleFunc("/api/plans/{id}", h.DeletePlan).Methods("DELETE")

	// Configs
	r.HandleFunc("/api/configs", h.ListConfigs).Methods("GET")
	r.HandleFunc("/api/configs", h.CreateConfig).Methods("POST")
	r.HandleFunc("/api/configs/{id}", h.GetConfig).Methods("GET")
	r.HandleFunc("/api/configs/{id}", h.UpdateConfig).Methods("PATCH")
	r.HandleFunc("/api/configs/{id}", h.DeleteConfig).Methods("DELETE")
	r.HandleFunc("/api/configs/{id}/validate", h.ValidateConfig).Methods("GET")
	r.HandleFunc("/api/configs/{id}/stats", h.GetConfigStats).Methods("GET")

	// Usage (proxy callback)
	r.HandleFunc("/api/usage", h.UpdateUsage).Methods("POST")

	// Subscriptions
	r.HandleFunc("/api/subscriptions", h.ListSubscriptions).Methods("GET")
	r.HandleFunc("/api/subscriptions", h.CreateSubscription).Methods("POST")

	// Cleanup (cron endpoint)
	r.HandleFunc("/api/cleanup", h.Cleanup).Methods("POST")

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("vpn-manager listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
