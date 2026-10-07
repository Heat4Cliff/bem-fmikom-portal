package security

import "testing"

func TestPasswordHashAndPolicy(t *testing.T) {
	password := "correct horse battery"
	hash, err := HashPassword(password, 4)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == password || !VerifyPassword(hash, password) || VerifyPassword(hash, "wrong password value") {
		t.Fatal("password hash verification failed")
	}
	if err := ValidatePassword("short"); err == nil {
		t.Fatal("short password should fail")
	}
	if err := ValidatePassword("1234567890123456789012345678901234567890123456789012345678901234567890123"); err == nil {
		t.Fatal("password over bcrypt's byte limit should fail")
	}
}
