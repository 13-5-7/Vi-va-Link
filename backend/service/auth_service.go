package service

import (
	"context"
	"errors"
	"time"
	"log"

	"github.com/bus-logistics/backend/model"
	"github.com/bus-logistics/backend/repository"
	"github.com/bus-logistics/backend/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidInviteCode  = errors.New("invalid or already used invite code")
)

type AuthService struct {
	userRepo   *repository.UserRepository
	inviteRepo *repository.InviteRepository
	jwtSecret  []byte
}

func NewAuthService(userRepo *repository.UserRepository, inviteRepo *repository.InviteRepository, jwtSecret string) *AuthService {
	log.Println("-----NewAuthService called-----")
	if userRepo == nil || inviteRepo == nil || jwtSecret == "" {
		log.Fatal("required parameter is missing for AuthService")
	}
	return &AuthService{userRepo: userRepo, inviteRepo: inviteRepo, jwtSecret: []byte(jwtSecret)}
}

type RegisterRequest struct {
	Email      string
	Password   string
	Role       model.Role
	InviteCode string // bus_operator 登録時に必須
}

type LoginRequest struct {
	Email    string
	Password string
	Role     string // ログイン画面のロールと一致するか検証に使用
}

type LoginResponse struct {
	Token  string
	UserID uuid.UUID
	Role   model.Role
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*model.User, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	log.Println("-----handler/auth_service.go Login called-----")

	// ロールの存在確認
	if !model.IsValidRole(req.Role) {
        return nil, ErrInvalidCredentials
    }

	// ユーザーの存在確認
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if utils.IsEmpty(user) {
		return nil, ErrInvalidCredentials
	}

	// パスワード検証（bcryptによるハッシュ比較）
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 認可チェック：指定されたロールでログイン可能か
	if string(user.Role) != req.Role {
		return nil, ErrInvalidCredentials
	}

	// トークンの生成
	signed, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:  signed,
		UserID: user.ID,
		Role:   user.Role,
	}, nil
}

func (s *AuthService) generateToken(user *model.User) (string, error) {
	// JWTトークンの生成（有効期限: 24時間）
    claims := jwt.MapClaims{
        "user_id": user.ID.String(),
        "role":    string(user.Role),
        "exp":     time.Now().Add(24 * time.Hour).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.jwtSecret)
}