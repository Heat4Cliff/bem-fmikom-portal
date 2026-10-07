package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

// LoadDevelopmentFiles reads repository-root and API-local dotenv files when
// present. Existing process environment always takes precedence.
func LoadDevelopmentFiles() error {
	err := godotenv.Load("../../.env", ".env")
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
