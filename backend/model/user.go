package model

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleBusOperator Role = "bus_operator"
	RoleShipper     Role = "shipper"
)

type User struct {
	ID           uuid.UUID  `db:"id"`
	Email        string     `db:"email"`
	PasswordHash string     `db:"password_hash"`
	Role         Role       `db:"role"`
	CompanyID    *uuid.UUID `db:"company_id"` // bus_operator のみ使用
	CreatedAt    time.Time  `db:"created_at"`
}

func IsValidRole(role string) bool {
    switch Role(role) {
	case RoleBusOperator, RoleShipper:
		return true
    }
    return false
}