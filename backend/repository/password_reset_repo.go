package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	//"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTokenNotFound = errors.New("password reset token not found or expired")

type PasswordResetToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	UsedAt    *time.Time
}

type PasswordResetRepository struct {
	pool *pgxpool.Pool
}

func NewPasswordResetRepository(pool *pgxpool.Pool) *PasswordResetRepository {
	return &PasswordResetRepository{pool: pool}
}

// Create は新しいパスワードリセットトークンを作成する
func (r *PasswordResetRepository) Create(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// FindValidToken は有効なトークンを検索する（未使用・期限内）
func (r *PasswordResetRepository) FindValidToken(ctx context.Context, token string) (*PasswordResetToken, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// MarkUsed はトークンを使用済みにする
func (r *PasswordResetRepository) MarkUsed(ctx context.Context, token string) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}
