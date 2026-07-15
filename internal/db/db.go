package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/clonerplus/vpn-manager/internal/models"
)

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func migrate(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS plans (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			duration_days INTEGER NOT NULL,
			data_limit_gb REAL NOT NULL,
			price REAL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			protocol TEXT NOT NULL,
			config_json TEXT NOT NULL,
			expires_at DATETIME,
			data_limit_bytes INTEGER,
			data_used_bytes INTEGER DEFAULT 0,
			is_active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS subscriptions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			plan_id INTEGER NOT NULL,
			starts_at DATETIME NOT NULL,
			expires_at DATETIME NOT NULL,
			data_used INTEGER DEFAULT 0,
			is_active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (plan_id) REFERENCES plans(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_configs_user_id ON configs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_configs_active ON configs(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_configs_expires ON configs(expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_user ON subscriptions(user_id)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("exec migration: %w", err)
		}
	}
	return nil
}

// --- Users ---

func (s *Store) CreateUser(username string) (*models.User, error) {
	res, err := s.db.Exec("INSERT INTO users (username) VALUES (?)", username)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &models.User{ID: id, Username: username, CreatedAt: time.Now()}, nil
}

func (s *Store) GetUser(id int64) (*models.User, error) {
	u := &models.User{}
	err := s.db.QueryRow("SELECT id, username, created_at FROM users WHERE id = ?", id).
		Scan(&u.ID, &u.Username, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) ListUsers() ([]models.User, error) {
	rows, err := s.db.Query("SELECT id, username, created_at FROM users ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// --- Plans ---

func (s *Store) CreatePlan(name string, durationDays int, dataLimitGB float64, price float64) (*models.Plan, error) {
	res, err := s.db.Exec(
		"INSERT INTO plans (name, duration_days, data_limit_gb, price) VALUES (?, ?, ?, ?)",
		name, durationDays, dataLimitGB, price,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &models.Plan{ID: id, Name: name, DurationDays: durationDays, DataLimitGB: dataLimitGB, Price: price, CreatedAt: time.Now()}, nil
}

func (s *Store) ListPlans() ([]models.Plan, error) {
	rows, err := s.db.Query("SELECT id, name, duration_days, data_limit_gb, price, created_at FROM plans ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []models.Plan
	for rows.Next() {
		var p models.Plan
		if err := rows.Scan(&p.ID, &p.Name, &p.DurationDays, &p.DataLimitGB, &p.Price, &p.CreatedAt); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, nil
}

func (s *Store) GetPlan(id int64) (*models.Plan, error) {
	p := &models.Plan{}
	err := s.db.QueryRow(
		"SELECT id, name, duration_days, data_limit_gb, price, created_at FROM plans WHERE id = ?", id,
	).Scan(&p.ID, &p.Name, &p.DurationDays, &p.DataLimitGB, &p.Price, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) DeletePlan(id int64) error {
	_, err := s.db.Exec("DELETE FROM plans WHERE id = ?", id)
	return err
}

// --- Configs ---

func (s *Store) CreateConfig(userID int64, protocol, configJSON string, durationDays *int, dataLimitGB *float64) (*models.Config, error) {
	var expiresAt *time.Time
	if durationDays != nil {
		t := time.Now().AddDate(0, 0, *durationDays)
		expiresAt = &t
	}

	var dataLimitBytes *int64
	if dataLimitGB != nil {
		b := int64(*dataLimitGB * 1024 * 1024 * 1024)
		dataLimitBytes = &b
	}

	res, err := s.db.Exec(
		`INSERT INTO configs (user_id, protocol, config_json, expires_at, data_limit_bytes)
		 VALUES (?, ?, ?, ?, ?)`,
		userID, protocol, configJSON, expiresAt, dataLimitBytes,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()

	return &models.Config{
		ID:             id,
		UserID:         userID,
		Protocol:       protocol,
		ConfigJSON:     configJSON,
		ExpiresAt:      expiresAt,
		DataLimitBytes: dataLimitBytes,
		DataUsedBytes:  0,
		IsActive:       true,
		CreatedAt:      time.Now(),
	}, nil
}

func (s *Store) GetConfig(id int64) (*models.Config, error) {
	c := &models.Config{}
	err := s.db.QueryRow(
		`SELECT id, user_id, protocol, config_json, expires_at, data_limit_bytes, data_used_bytes, is_active, created_at
		 FROM configs WHERE id = ?`, id,
	).Scan(&c.ID, &c.UserID, &c.Protocol, &c.ConfigJSON, &c.ExpiresAt, &c.DataLimitBytes, &c.DataUsedBytes, &c.IsActive, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Store) ListConfigs(userID int64) ([]models.Config, error) {
	var rows *sql.Rows
	var err error
	if userID > 0 {
		rows, err = s.db.Query(
			`SELECT id, user_id, protocol, config_json, expires_at, data_limit_bytes, data_used_bytes, is_active, created_at
			 FROM configs WHERE user_id = ? ORDER BY id`, userID,
		)
	} else {
		rows, err = s.db.Query(
			`SELECT id, user_id, protocol, config_json, expires_at, data_limit_bytes, data_used_bytes, is_active, created_at
			 FROM configs ORDER BY id`,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []models.Config
	for rows.Next() {
		var c models.Config
		if err := rows.Scan(&c.ID, &c.UserID, &c.Protocol, &c.ConfigJSON, &c.ExpiresAt, &c.DataLimitBytes, &c.DataUsedBytes, &c.IsActive, &c.CreatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, nil
}

func (s *Store) ValidateConfig(id int64) (*models.Config, error) {
	c, err := s.GetConfig(id)
	if err != nil {
		return nil, err
	}

	if !c.IsActive {
		return nil, fmt.Errorf("config is deactivated")
	}

	if c.ExpiresAt != nil && c.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("config has expired")
	}

	if c.DataLimitBytes != nil && c.DataUsedBytes >= *c.DataLimitBytes {
		return nil, fmt.Errorf("data limit exceeded")
	}

	return c, nil
}

func (s *Store) UpdateConfig(id int64, req models.ConfigUpdateRequest) error {
	if req.DurationDays != nil {
		t := time.Now().AddDate(0, 0, *req.DurationDays)
		if _, err := s.db.Exec("UPDATE configs SET expires_at = ? WHERE id = ?", t, id); err != nil {
			return err
		}
	}
	if req.DataLimitGB != nil {
		b := int64(*req.DataLimitGB * 1024 * 1024 * 1024)
		if _, err := s.db.Exec("UPDATE configs SET data_limit_bytes = ? WHERE id = ?", b, id); err != nil {
			return err
		}
	}
	if req.IsActive != nil {
		if _, err := s.db.Exec("UPDATE configs SET is_active = ? WHERE id = ?", *req.IsActive, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) DeleteConfig(id int64) error {
	_, err := s.db.Exec("DELETE FROM configs WHERE id = ?", id)
	return err
}

func (s *Store) UpdateUsage(configID int64, bytes int64) error {
	_, err := s.db.Exec(
		"UPDATE configs SET data_used_bytes = data_used_bytes + ? WHERE id = ?",
		bytes, configID,
	)
	return err
}

// --- Subscriptions ---

func (s *Store) CreateSubscription(userID, planID int64) (*models.Subscription, error) {
	plan, err := s.GetPlan(planID)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	now := time.Now()
	expiresAt := now.AddDate(0, 0, plan.DurationDays)

	res, err := s.db.Exec(
		`INSERT INTO subscriptions (user_id, plan_id, starts_at, expires_at)
		 VALUES (?, ?, ?, ?)`,
		userID, planID, now, expiresAt,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()

	return &models.Subscription{
		ID:        id,
		UserID:    userID,
		PlanID:    planID,
		StartsAt:  now,
		ExpiresAt: expiresAt,
		IsActive:  true,
		CreatedAt: now,
	}, nil
}

func (s *Store) ListSubscriptions(userID int64) ([]models.Subscription, error) {
	var rows *sql.Rows
	var err error
	if userID > 0 {
		rows, err = s.db.Query(
			`SELECT id, user_id, plan_id, starts_at, expires_at, data_used, is_active, created_at
			 FROM subscriptions WHERE user_id = ? ORDER BY id`, userID,
		)
	} else {
		rows, err = s.db.Query(
			`SELECT id, user_id, plan_id, starts_at, expires_at, data_used, is_active, created_at
			 FROM subscriptions ORDER BY id`,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []models.Subscription
	for rows.Next() {
		var sub models.Subscription
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.PlanID, &sub.StartsAt, &sub.ExpiresAt, &sub.DataUsed, &sub.IsActive, &sub.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

// --- Expired config cleanup ---

func (s *Store) DeactivateExpired() (int64, error) {
	res, err := s.db.Exec(
		`UPDATE configs SET is_active = 0
		 WHERE is_active = 1 AND expires_at IS NOT NULL AND expires_at < ?`,
		time.Now(),
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *Store) DeactivateOverLimit() (int64, error) {
	res, err := s.db.Exec(
		`UPDATE configs SET is_active = 0
		 WHERE is_active = 1 AND data_limit_bytes IS NOT NULL AND data_used_bytes >= data_limit_bytes`,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *Store) DeactivateExpiredSubscriptions() (int64, error) {
	res, err := s.db.Exec(
		`UPDATE subscriptions SET is_active = 0
		 WHERE is_active = 1 AND expires_at < ?`,
		time.Now(),
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// --- Usage stats ---

func (s *Store) GetConfigStats(configID int64) (*models.UsageStats, error) {
	c, err := s.GetConfig(configID)
	if err != nil {
		return nil, err
	}

	stats := &models.UsageStats{
		ConfigID:   c.ID,
		DataUsedGB: float64(c.DataUsedBytes) / (1024 * 1024 * 1024),
		ExpiresAt:  c.ExpiresAt,
		IsActive:   c.IsActive,
	}

	if c.DataLimitBytes != nil {
		limitGB := float64(*c.DataLimitBytes) / (1024 * 1024 * 1024)
		stats.DataLimitGB = &limitGB
		if *c.DataLimitBytes > 0 {
			stats.UsagePercent = float64(c.DataUsedBytes) / float64(*c.DataLimitBytes) * 100
		}
	}

	return stats, nil
}

// --- Proxy log sync ---

func (s *Store) GetConfigsByProtocol(protocol string) ([]models.Config, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, protocol, config_json, expires_at, data_limit_bytes, data_used_bytes, is_active, created_at
		 FROM configs WHERE protocol = ? AND is_active = 1`, protocol,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []models.Config
	for rows.Next() {
		var c models.Config
		if err := rows.Scan(&c.ID, &c.UserID, &c.Protocol, &c.ConfigJSON, &c.ExpiresAt, &c.DataLimitBytes, &c.DataUsedBytes, &c.IsActive, &c.CreatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, nil
}
