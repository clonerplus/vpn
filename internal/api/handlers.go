package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/clonerplus/vpn-manager/internal/db"
	"github.com/clonerplus/vpn-manager/internal/models"
	"github.com/gorilla/mux"
)

type Handler struct {
	store *db.Store
}

func NewHandler(store *db.Store) *Handler {
	return &Handler{store: store}
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	jsonResponse(w, status, map[string]string{"error": msg})
}

func parseID(r *http.Request, key string) (int64, error) {
	return strconv.ParseInt(mux.Vars(r)[key], 10, 64)
}

// --- Users ---

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" {
		jsonError(w, 400, "username required")
		return
	}

	user, err := h.store.CreateUser(req.Username)
	if err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonResponse(w, 201, user)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.ListUsers()
	if err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonResponse(w, 200, users)
}

// --- Plans ---

func (h *Handler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string  `json:"name"`
		DurationDays int     `json:"duration_days"`
		DataLimitGB  float64 `json:"data_limit_gb"`
		Price        float64 `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		jsonError(w, 400, "name, duration_days, data_limit_gb required")
		return
	}

	plan, err := h.store.CreatePlan(req.Name, req.DurationDays, req.DataLimitGB, req.Price)
	if err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonResponse(w, 201, plan)
}

func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.store.ListPlans()
	if err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonResponse(w, 200, plans)
}

func (h *Handler) DeletePlan(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		jsonError(w, 400, "invalid id")
		return
	}
	if err := h.store.DeletePlan(id); err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonResponse(w, 200, map[string]string{"status": "deleted"})
}

// --- Configs ---

func (h *Handler) CreateConfig(w http.ResponseWriter, r *http.Request) {
	var req models.ConfigCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Protocol == "" || req.ConfigJSON == "" {
		jsonError(w, 400, "protocol and config_json required")
		return
	}

	config, err := h.store.CreateConfig(req.UserID, req.Protocol, req.ConfigJSON, req.DurationDays, req.DataLimitGB)
	if err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonResponse(w, 201, config)
}

func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		jsonError(w, 400, "invalid id")
		return
	}

	config, err := h.store.GetConfig(id)
	if err != nil {
		jsonError(w, 404, "config not found")
		return
	}
	jsonResponse(w, 200, config)
}

func (h *Handler) ListConfigs(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	var userID int64
	if userIDStr != "" {
		userID, _ = strconv.ParseInt(userIDStr, 10, 64)
	}

	configs, err := h.store.ListConfigs(userID)
	if err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonResponse(w, 200, configs)
}

func (h *Handler) ValidateConfig(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		jsonError(w, 400, "invalid id")
		return
	}

	config, err := h.store.ValidateConfig(id)
	if err != nil {
		jsonResponse(w, 403, map[string]interface{}{
			"valid":  false,
			"reason": err.Error(),
		})
		return
	}

	stats, _ := h.store.GetConfigStats(id)
	jsonResponse(w, 200, map[string]interface{}{
		"valid":  true,
		"config": config,
		"stats":  stats,
	})
}

func (h *Handler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		jsonError(w, 400, "invalid id")
		return
	}

	var req models.ConfigUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid request body")
		return
	}

	if err := h.store.UpdateConfig(id, req); err != nil {
		jsonError(w, 500, err.Error())
		return
	}

	config, _ := h.store.GetConfig(id)
	jsonResponse(w, 200, config)
}

func (h *Handler) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		jsonError(w, 400, "invalid id")
		return
	}
	if err := h.store.DeleteConfig(id); err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonResponse(w, 200, map[string]string{"status": "deleted"})
}

func (h *Handler) GetConfigStats(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		jsonError(w, 400, "invalid id")
		return
	}

	stats, err := h.store.GetConfigStats(id)
	if err != nil {
		jsonError(w, 404, "config not found")
		return
	}
	jsonResponse(w, 200, stats)
}

// --- Usage update (called by proxy) ---

func (h *Handler) UpdateUsage(w http.ResponseWriter, r *http.Request) {
	var updates []models.ProxyUsageUpdate
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		jsonError(w, 400, "invalid request body")
		return
	}

	for _, u := range updates {
		total := u.BytesIn + u.BytesOut
		if total > 0 {
			h.store.UpdateUsage(u.ConfigID, total)
		}
	}
	jsonResponse(w, 200, map[string]string{"status": "updated"})
}

// --- Subscriptions ---

func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req models.SubscriptionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "user_id and plan_id required")
		return
	}

	sub, err := h.store.CreateSubscription(req.UserID, req.PlanID)
	if err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonResponse(w, 201, sub)
}

func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	var userID int64
	if userIDStr != "" {
		userID, _ = strconv.ParseInt(userIDStr, 10, 64)
	}

	subs, err := h.store.ListSubscriptions(userID)
	if err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonResponse(w, 200, subs)
}

// --- Cleanup ---

func (h *Handler) Cleanup(w http.ResponseWriter, r *http.Request) {
	expired, _ := h.store.DeactivateExpired()
	overLimit, _ := h.store.DeactivateOverLimit()
	expiredSubs, _ := h.store.DeactivateExpiredSubscriptions()

	jsonResponse(w, 200, map[string]int64{
		"expired_configs":          expired,
		"over_limit_configs":       overLimit,
		"expired_subscriptions":    expiredSubs,
	})
}

// --- Health ---

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, 200, map[string]string{"status": "ok"})
}
