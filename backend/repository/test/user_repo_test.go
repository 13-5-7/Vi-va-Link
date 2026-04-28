package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bus-logistics/backend/model"
	"github.com/google/uuid"
)

func TestUserRepo_FindByEmail_Found(t *testing.T) {
	uid := uuid.New()
	repo := &MockUserRepo{
		FindByEmailFunc: func(_ context.Context, email string) (*model.User, error) {
			if email != "test@example.com" {
				return nil, nil
			}
			return &model.User{ID: uid, Email: email, Role: model.RoleShipper}, nil
		},
	}

	u, err := repo.FindByEmail(context.Background(), "test@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if u == nil {
		t.Fatal("want user, got nil")
	}
	if u.ID != uid {
		t.Errorf("want %v, got %v", uid, u.ID)
	}
	if u.Email != "test@example.com" {
		t.Errorf("want test@example.com, got %s", u.Email)
	}
}

func TestUserRepo_FindByEmail_NotFound(t *testing.T) {
	repo := &MockUserRepo{
		FindByEmailFunc: func(_ context.Context, _ string) (*model.User, error) {
			return nil, nil
		},
	}

	u, err := repo.FindByEmail(context.Background(), "notexist@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if u != nil {
		t.Errorf("want nil, got %v", u)
	}
}

func TestUserRepo_FindByEmail_Error(t *testing.T) {
	repo := &MockUserRepo{
		FindByEmailFunc: func(_ context.Context, _ string) (*model.User, error) {
			return nil, errors.New("db error")
		},
	}

	_, err := repo.FindByEmail(context.Background(), "test@example.com")
	if err == nil {
		t.Error("want error, got nil")
	}
}

func TestUserRepo_Create_Success(t *testing.T) {
	uid := uuid.New()
	repo := &MockUserRepo{
		CreateFunc: func(_ context.Context, email, hash string, role model.Role) (*model.User, error) {
			return &model.User{ID: uid, Email: email, PasswordHash: hash, Role: role}, nil
		},
	}

	u, err := repo.Create(context.Background(), "new@example.com", "$2a$hash", model.RoleBusOperator)
	if err != nil {
		t.Fatal(err)
	}
	if u.Email != "new@example.com" {
		t.Errorf("want new@example.com, got %s", u.Email)
	}
	if u.Role != model.RoleBusOperator {
		t.Errorf("want bus_operator, got %v", u.Role)
	}
	if u.PasswordHash != "$2a$hash" {
		t.Errorf("want hash stored, got %s", u.PasswordHash)
	}
}

func TestUserRepo_Create_DuplicateEmail_Error(t *testing.T) {
	repo := &MockUserRepo{
		CreateFunc: func(_ context.Context, _ string, _ string, _ model.Role) (*model.User, error) {
			return nil, errors.New("duplicate key value violates unique constraint")
		},
	}

	_, err := repo.Create(context.Background(), "dup@example.com", "hash", model.RoleShipper)
	if err == nil {
		t.Error("want error for duplicate email, got nil")
	}
}

func TestUserRepo_FindByID_Found(t *testing.T) {
	uid := uuid.New()
	repo := &MockUserRepo{
		FindByIDFunc: func(_ context.Context, id uuid.UUID) (*model.User, error) {
			if id != uid {
				return nil, nil
			}
			return &model.User{ID: uid, Email: "found@example.com", Role: model.RoleShipper}, nil
		},
	}

	u, err := repo.FindByID(context.Background(), uid)
	if err != nil {
		t.Fatal(err)
	}
	if u == nil {
		t.Fatal("want user, got nil")
	}
	if u.ID != uid {
		t.Errorf("want %v, got %v", uid, u.ID)
	}
}

func TestUserRepo_FindByID_NotFound(t *testing.T) {
	repo := &MockUserRepo{
		FindByIDFunc: func(_ context.Context, _ uuid.UUID) (*model.User, error) {
			return nil, nil
		},
	}

	u, err := repo.FindByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if u != nil {
		t.Errorf("want nil, got %v", u)
	}
}
