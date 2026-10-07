package domain

import "time"

type Role string

const (
	RoleSuperAdmin Role = "SUPER_ADMIN"
	RoleBPH        Role = "BPH"
	RoleKepala     Role = "KEPALA_BIDANG"
)

type AccountStatus string

const (
	StatusActive   AccountStatus = "ACTIVE"
	StatusInactive AccountStatus = "INACTIVE"
)

type Division struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type User struct {
	ID           int64         `json:"id"`
	Name         string        `json:"name"`
	Email        string        `json:"email"`
	PasswordHash string        `json:"-"`
	Role         Role          `json:"role"`
	DivisionID   *int64        `json:"-"`
	Division     *Division     `json:"division"`
	Permissions  []string      `json:"permissions,omitempty"`
	Status       AccountStatus `json:"-"`
	LastLoginAt  *time.Time    `json:"-"`
}

type Principal struct {
	UserID     int64
	Name       string
	Email      string
	Role       Role
	DivisionID *int64
	Status     AccountStatus
}

func (user User) Principal() Principal {
	return Principal{UserID: user.ID, Name: user.Name, Email: user.Email, Role: user.Role, DivisionID: cloneInt64(user.DivisionID), Status: user.Status}
}

func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
