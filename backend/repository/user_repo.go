package repository

import (
	"context"
	"errors"
	"log"
	"fmt"

	"github.com/bus-logistics/backend/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	log.Println("-----UserRepository called-----")
	if pool == nil {
		log.Fatal("pool is required for UserRepository")
	}
	return &UserRepository{pool: pool}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	log.Printf("-----FindByEmail called: email=%s-----", email)

	const query = `SELECT id, email, password_hash, role, company_id, created_at FROM users WHERE email = $1`

	var u model.User
    err := r.pool.QueryRow(ctx, query, email).Scan(
        &u.ID, 
        &u.Email, 
        &u.PasswordHash, 
        &u.Role, 
        &u.CompanyID, 
        &u.CreatedAt,
    )

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("UserRepository.FindByEmail (email: %s): %w", email, err)
	}
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, email, passwordHash string, role model.Role) (*model.User, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// SetCompanyID はユーザーの company_id を更新する
func (r *UserRepository) SetCompanyID(ctx context.Context, userID uuid.UUID, companyID uuid.UUID) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// Delete はユーザーを削除する（登録ロールバック用）
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// UpdatePassword はユーザーのパスワードハッシュを更新する
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}
