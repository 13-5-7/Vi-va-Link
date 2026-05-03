package service

import (
	"context"
	//"crypto/rand"
	//"encoding/hex"
	"errors"
	//"log"
	"time"

	"github.com/bus-logistics/backend/repository"
	//"golang.org/x/crypto/bcrypt"
)

var (
	ErrResetTokenInvalid = errors.New("invalid or expired reset token")
	ErrUserNotFound      = errors.New("user not found")
)

type PasswordResetService struct {
	userRepo      *repository.UserRepository
	resetRepo     *repository.PasswordResetRepository
	tokenExpiry   time.Duration
}

func NewPasswordResetService(
	userRepo *repository.UserRepository,
	resetRepo *repository.PasswordResetRepository,
) *PasswordResetService {
	return &PasswordResetService{
		userRepo:    userRepo,
		resetRepo:   resetRepo,
		tokenExpiry: 1 * time.Hour, // トークン有効期限: 1時間
	}
}

// RequestReset はパスワードリセットトークンを発行する
// セキュリティ上、ユーザーが存在しない場合も同じレスポンスを返す（ユーザー存在確認攻撃を防ぐ）
func (s *PasswordResetService) RequestReset(ctx context.Context, email string) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// ResetPassword はトークンを検証してパスワードを更新する
func (s *PasswordResetService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// generateSecureToken は暗号学的に安全なランダムトークンを生成する
func generateSecureToken(length int) (string, error) { //nolint:unused
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}
