package main

import (
	"context"
	"crypto/subtle"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/config"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/database"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/service"
	"golang.org/x/term"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name := flag.String("name", "", "initial SUPER_ADMIN name")
	email := flag.String("email", "", "initial SUPER_ADMIN email")
	flag.Parse()
	if *name == "" || *email == "" || !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("usage: go run ./cmd/create-admin -name \"Admin BEM\" -email admin@example.org (interactive terminal required)")
	}
	fmt.Fprint(os.Stderr, "Password (minimum 12 characters): ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return fmt.Errorf("read password from terminal")
	}
	defer wipe(password)
	fmt.Fprint(os.Stderr, "Confirm password: ")
	confirmation, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return fmt.Errorf("read password confirmation")
	}
	defer wipe(confirmation)
	if subtle.ConstantTimeCompare(password, confirmation) != 1 {
		return fmt.Errorf("passwords do not match")
	}
	if err := config.LoadDevelopmentFiles(); err != nil {
		return fmt.Errorf("load local environment files")
	}
	settings, err := config.LoadDatabaseSettings()
	if err != nil {
		return err
	}
	db, err := database.OpenWithSettings(context.Background(), settings)
	if err != nil {
		return err
	}
	defer db.Close()
	bcryptCost := 12
	if raw := os.Getenv("AUTH_BCRYPT_COST"); raw != "" {
		if _, err := fmt.Sscanf(raw, "%d", &bcryptCost); err != nil || bcryptCost < 4 || bcryptCost > 31 {
			return fmt.Errorf("AUTH_BCRYPT_COST must be between 4 and 31")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := service.CreateSuperAdmin(ctx, db, bcryptCost, *name, *email, string(password)); err != nil {
		if errors.Is(err, service.ErrConflict) {
			return fmt.Errorf("an account with that email already exists")
		}
		return err
	}
	fmt.Fprintln(os.Stdout, "initial SUPER_ADMIN account created")
	return nil
}

func wipe(value []byte) {
	for index := range value {
		value[index] = 0
	}
}
