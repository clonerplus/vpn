package models

import "time"

type Plan struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	DurationDays  int       `json:"duration_days"`
	DataLimitGB   float64   `json:"data_limit_gb"`
	Price         float64   `json:"price"`
	CreatedAt     time.Time `json:"created_at"`
}

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type Config struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	Protocol      string     `json:"protocol"`
	ConfigJSON    string     `json:"config_json"`
	ExpiresAt     *time.Time `json:"expires_at"`
	DataLimitBytes *int64    `json:"data_limit_bytes"`
	DataUsedBytes  int64     `json:"data_used_bytes"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
}

type ConfigCreateRequest struct {
	UserID       int64   `json:"user_id"`
	Protocol     string  `json:"protocol"`
	ConfigJSON   string  `json:"config_json"`
	DurationDays *int    `json:"duration_days"`
	DataLimitGB  *float64 `json:"data_limit_gb"`
}

type ConfigUpdateRequest struct {
	DurationDays *int      `json:"duration_days"`
	DataLimitGB  *float64  `json:"data_limit_gb"`
	IsActive     *bool     `json:"is_active"`
}

type Subscription struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	PlanID    int64      `json:"plan_id"`
	StartsAt  time.Time  `json:"starts_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	DataUsed  int64      `json:"data_used"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
}

type SubscriptionCreateRequest struct {
	UserID int64 `json:"user_id"`
	PlanID int64 `json:"plan_id"`
}

type UsageStats struct {
	ConfigID     int64   `json:"config_id"`
	DataUsedGB   float64 `json:"data_used_gb"`
	DataLimitGB  *float64 `json:"data_limit_gb"`
	UsagePercent float64 `json:"usage_percent"`
	ExpiresAt    *time.Time `json:"expires_at"`
	IsActive     bool    `json:"is_active"`
}

type ProxyUsageUpdate struct {
	ConfigID   int64 `json:"config_id"`
	BytesIn    int64 `json:"bytes_in"`
	BytesOut   int64 `json:"bytes_out"`
}
