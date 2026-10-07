package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/config"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/database"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/database/migrate"
)

func main() {
	if len(os.Args) != 2 || (os.Args[1] != "up" && os.Args[1] != "down") {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/migrate up|down")
		os.Exit(2)
	}
	if err := config.LoadDevelopmentFiles(); err != nil {
		fail(fmt.Errorf("load local environment files"))
	}
	settings, err := config.LoadDatabaseSettings()
	if err != nil {
		fail(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := database.OpenWithSettings(ctx, settings)
	if err != nil {
		fail(err)
	}
	defer db.Close()
	if os.Args[1] == "up" {
		err = migrate.Up(ctx, db)
	} else {
		err = migrate.Down(ctx, db)
	}
	if err != nil {
		fail(err)
	}
	fmt.Printf("migrations %s complete\n", os.Args[1])
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
