package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/config"
	"github.com/go-sql-driver/mysql"
)

func Open(ctx context.Context, cfg config.Config) (*sql.DB, error) {
	return OpenWithSettings(ctx, config.DatabaseSettings{URL: cfg.DatabaseURL, MaxOpenConns: cfg.DBMaxOpenConns, MaxIdleConns: cfg.DBMaxIdleConns, ConnMaxLifetime: cfg.DBConnMaxLife})
}

func OpenWithSettings(ctx context.Context, settings config.DatabaseSettings) (*sql.DB, error) {
	dsn, err := mysql.ParseDSN(settings.URL)
	if err != nil {
		return nil, fmt.Errorf("parse MySQL DSN")
	}
	dsn.ParseTime = true
	dsn.Loc = time.UTC
	if dsn.Params == nil {
		dsn.Params = make(map[string]string)
	}
	dsn.Params["time_zone"] = "'+00:00'"
	dsn.Collation = "utf8mb4_unicode_ci"
	db, err := sql.Open("mysql", dsn.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("open MySQL database: %w", err)
	}
	db.SetMaxOpenConns(settings.MaxOpenConns)
	db.SetMaxIdleConns(settings.MaxIdleConns)
	db.SetConnMaxLifetime(settings.ConnMaxLifetime)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connect to MySQL database: %w", err)
	}
	return db, nil
}
