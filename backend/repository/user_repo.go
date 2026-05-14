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

// Create 新しいユーザーレコードを作成し、作成されたユーザーの情報をモデルとして返します。
func (r *UserRepository) Create(ctx context.Context, email, passwordHash string, role model.Role) (*model.User, error) {
	log.Println("-----handler/user_repo.go Create called-----")

	var u model.User
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3)
		 RETURNING id, email, password_hash, role, company_id, created_at`,
		email, passwordHash, role,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CompanyID, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByID ユーザマスタテーブルから指定されたIDに一致するユーザレコードを取得する
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	log.Println("-----repository/user_repo.go FindByID called")

	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, role, company_id, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CompanyID, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// SetCompanyID ユーザーの company_id を更新する
func (r *UserRepository) SetCompanyID(ctx context.Context, userID uuid.UUID, companyID uuid.UUID) error {
	log.Println("-----repository/user_repo.go SetCompanyID called")
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET company_id = $1 WHERE id = $2`,
		companyID, userID,
	)
	return err
}

// Delete ユーザーを削除する（登録ロールバック用）
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	log.Println("-----repository/user_repo.go Delete called")
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

// UpdatePassword はユーザーのパスワードハッシュを更新する
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}
