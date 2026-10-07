package migrate

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var files embed.FS

func Up(ctx context.Context, db *sql.DB) error {
	return run(ctx, db, func() error { return goose.UpContext(ctx, db, ".") })
}

func Down(ctx context.Context, db *sql.DB) error {
	return run(ctx, db, func() error { return goose.DownContext(ctx, db, ".") })
}

func DownToZero(ctx context.Context, db *sql.DB) error {
	return run(ctx, db, func() error { return goose.DownToContext(ctx, db, ".", 0) })
}

func run(ctx context.Context, db *sql.DB, action func() error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := goose.SetDialect("mysql"); err != nil {
		return fmt.Errorf("configure migration dialect: %w", err)
	}
	goose.SetBaseFS(files)
	if err := action(); err != nil {
		return fmt.Errorf("run database migration: %w", err)
	}
	return nil
}
