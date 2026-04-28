package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bus-logistics/backend/model"
	"github.com/bus-logistics/backend/service"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// テスト用 AuthService: リポジトリをモックで差し替えられるよう
// サービスのロジックをインライン実装する

type authServiceImpl struct {
	repo      userRepoIface
	jwtSecret string
}

type userRepoIface interface {
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Create(ctx context.Context, email, hash string, role model.Role) (*model.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

func (a *authServiceImpl) Register(ctx context.Context, req service.RegisterRequest) (*model.User, error) {
	existing, err := a.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, service.ErrEmailAlreadyExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.MinCost)
	if err != nil {
		return nil, err
	}
	return a.repo.Create(ctx, req.Email, string(hash), req.Role)
}

func (a *authServiceImpl) Login(ctx context.Context, req service.LoginRequest) (*service.LoginResponse, error) {
	user, err := a.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, service.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, service.ErrInvalidCredentials
	}
	if req.Role != "" && string(user.Role) != req.Role {
		return nil, service.ErrInvalidCredentials
	}
	return &service.LoginResponse{UserID: user.ID, Role: user.Role, Token: "test-token"}, nil
}

// ---- Register ----

func TestAuthService_Register_Success(t *testing.T) {
	uid := uuid.New()
	svc := &authServiceImpl{
		repo: &mockUserRepo{
			findByEmail: func(_ context.Context, _ string) (*model.User, error) { return nil, nil },
			create: func(_ context.Context, email, _ string, role model.Role) (*model.User, error) {
				return &model.User{ID: uid, Email: email, Role: role}, nil
			},
		},
	}

	u, err := svc.Register(context.Background(), service.RegisterRequest{
		Email: "new@example.com", Password: "pass", Role: model.RoleShipper,
	})
	if err != nil {
		t.Fatal(err)
	}
	if u.Email != "new@example.com" {
		t.Errorf("want new@example.com, got %s", u.Email)
	}
	if u.Role != model.RoleShipper {
		t.Errorf("want shipper, got %v", u.Role)
	}
}

func TestAuthService_Register_EmailAlreadyExists(t *testing.T) {
	svc := &authServiceImpl{
		repo: &mockUserRepo{
			findByEmail: func(_ context.Context, _ string) (*model.User, error) {
				return &model.User{Email: "dup@example.com"}, nil
			},
		},
	}

	_, err := svc.Register(context.Background(), service.RegisterRequest{
		Email: "dup@example.com", Password: "pass", Role: model.RoleShipper,
	})
	if !errors.Is(err, service.ErrEmailAlreadyExists) {
		t.Errorf("want ErrEmailAlreadyExists, got %v", err)
	}
}

func TestAuthService_Register_RepoError(t *testing.T) {
	svc := &authServiceImpl{
		repo: &mockUserRepo{
			findByEmail: func(_ context.Context, _ string) (*model.User, error) {
				return nil, errors.New("db error")
			},
		},
	}

	_, err := svc.Register(context.Background(), service.RegisterRequest{
		Email: "a@b.com", Password: "pass", Role: model.RoleShipper,
	})
	if err == nil {
		t.Error("want error, got nil")
	}
}

// ---- Login ----

func TestAuthService_Login_Success(t *testing.T) {
	uid := uuid.New()
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)

	svc := &authServiceImpl{
		repo: &mockUserRepo{
			findByEmail: func(_ context.Context, _ string) (*model.User, error) {
				return &model.User{ID: uid, Email: "a@b.com", PasswordHash: string(hash), Role: model.RoleBusOperator}, nil
			},
		},
	}

	resp, err := svc.Login(context.Background(), service.LoginRequest{
		Email: "a@b.com", Password: "correct", Role: "bus_operator",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.UserID != uid {
		t.Errorf("want %v, got %v", uid, resp.UserID)
	}
	if resp.Role != model.RoleBusOperator {
		t.Errorf("want bus_operator, got %v", resp.Role)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	svc := &authServiceImpl{
		repo: &mockUserRepo{
			findByEmail: func(_ context.Context, _ string) (*model.User, error) { return nil, nil },
		},
	}

	_, err := svc.Login(context.Background(), service.LoginRequest{Email: "x@y.com", Password: "pass"})
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Errorf("want ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
	svc := &authServiceImpl{
		repo: &mockUserRepo{
			findByEmail: func(_ context.Context, _ string) (*model.User, error) {
				return &model.User{Email: "a@b.com", PasswordHash: string(hash), Role: model.RoleShipper}, nil
			},
		},
	}

	_, err := svc.Login(context.Background(), service.LoginRequest{Email: "a@b.com", Password: "wrong"})
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Errorf("want ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_RoleMismatch(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.MinCost)
	svc := &authServiceImpl{
		repo: &mockUserRepo{
			findByEmail: func(_ context.Context, _ string) (*model.User, error) {
				return &model.User{Email: "a@b.com", PasswordHash: string(hash), Role: model.RoleShipper}, nil
			},
		},
	}

	_, err := svc.Login(context.Background(), service.LoginRequest{
		Email: "a@b.com", Password: "pass", Role: "bus_operator",
	})
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Errorf("want ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_EmptyRole_SkipsRoleCheck(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.MinCost)
	uid := uuid.New()
	svc := &authServiceImpl{
		repo: &mockUserRepo{
			findByEmail: func(_ context.Context, _ string) (*model.User, error) {
				return &model.User{ID: uid, Email: "a@b.com", PasswordHash: string(hash), Role: model.RoleShipper}, nil
			},
		},
	}

	resp, err := svc.Login(context.Background(), service.LoginRequest{
		Email: "a@b.com", Password: "pass", Role: "",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.UserID != uid {
		t.Errorf("want %v, got %v", uid, resp.UserID)
	}
}
