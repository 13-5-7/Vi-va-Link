package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	//"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InviteCode struct {
	ID        uuid.UUID
	Code      string
	CompanyID uuid.UUID
	UsedBy    *uuid.UUID
	UsedAt    *time.Time
	CreatedAt time.Time
}

type InviteRepository struct {
	pool *pgxpool.Pool
}

func NewInviteRepository(pool *pgxpool.Pool) *InviteRepository {
	return &InviteRepository{pool: pool}
}

// FindByCode は招待コードを検索する（未使用のもののみ）
func (r *InviteRepository) FindByCode(ctx context.Context, code string) (*InviteCode, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// MarkUsed は招待コードを使用済みにし、使用者のIDと company_id を返す
func (r *InviteRepository) MarkUsed(ctx context.Context, code string, userID uuid.UUID) (uuid.UUID, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

var ErrInvalidInviteCode = errors.New("invalid or already used invite code")

// Create は新しい招待コードを作成する（管理者用）
func (r *InviteRepository) Create(ctx context.Context, code string, companyID uuid.UUID) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// List は招待コード一覧を返す。companyID が nil の場合は全件返す
func (r *InviteRepository) List(ctx context.Context, companyID *uuid.UUID) ([]InviteCode, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}
