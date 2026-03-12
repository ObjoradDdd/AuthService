package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	ConnString      string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type Storage struct {
	db *sql.DB
}

func New(ctx context.Context, cfg Config) (*Storage, error) {
	db, err := sql.Open("postgres", cfg.ConnString)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия БД: %w", err)
	}

	slog.Info(cfg.ConnString)

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ошибка пинга БД: %w", err)
	}

	slog.Info("DB connection successful", "driver", "postgres")

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	slog.Info("Closing DB connection")
	return s.db.Close()
}
