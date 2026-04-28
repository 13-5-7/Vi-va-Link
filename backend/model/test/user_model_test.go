package model_test

import (
	"testing"

	"github.com/bus-logistics/backend/model"
)

// ---- Role 定数値 ----

func TestRole_Constants(t *testing.T) {
	tests := []struct {
		role model.Role
		want string
	}{
		{model.RoleBusOperator, "bus_operator"},
		{model.RoleShipper, "shipper"},
	}
	for _, tt := range tests {
		if string(tt.role) != tt.want {
			t.Errorf("want %q, got %q", tt.want, tt.role)
		}
	}
}

// ---- User 構造体フィールド ----

func TestUser_PasswordHashNotExposedInRole(t *testing.T) {
	u := model.User{
		Email:        "test@example.com",
		PasswordHash: "$2a$10$hashedvalue",
		Role:         model.RoleBusOperator,
	}
	if u.PasswordHash == "" {
		t.Error("PasswordHash should be stored")
	}
	if u.Role != model.RoleBusOperator {
		t.Errorf("want bus_operator, got %v", u.Role)
	}
}

func TestUser_RoleDistinct(t *testing.T) {
	if model.RoleBusOperator == model.RoleShipper {
		t.Error("RoleBusOperator and RoleShipper must be distinct")
	}
}
