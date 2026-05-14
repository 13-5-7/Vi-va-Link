package repository

import (
	"context"
	"errors"
	"time"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

// FindByCode 招待コードを検索する（未使用のもののみ）
func (r *InviteRepository) FindByCode(ctx context.Context, code string) (*InviteCode, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// MarkUsed 招待コードを使用済みにし、使用者のIDと company_id を返す
func (r *InviteRepository) MarkUsed(ctx context.Context, code string, userID uuid.UUID) (uuid.UUID, error) {
	log.Println("-----repoxitory/invite_repo.go MarkUsed called-----")

	var companyID uuid.UUID
	err := r.pool.QueryRow(ctx,
		`UPDATE invite_codes
		 SET used_by = $1, used_at = NOW()
		 WHERE code = $2 AND used_by IS NULL
		 RETURNING company_id`,
		userID, code,
	).Scan(&companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrInvalidInviteCode
		}
		return uuid.Nil, err
	}
	return companyID, nil
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
