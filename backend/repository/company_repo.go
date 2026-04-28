package repository

import (
	"context"

	"github.com/bus-logistics/backend/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CompanyRepository struct {
	pool *pgxpool.Pool
}

func NewCompanyRepository(pool *pgxpool.Pool) *CompanyRepository {
	return &CompanyRepository{pool: pool}
}

func (r *CompanyRepository) List(ctx context.Context) ([]model.BusCompany, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

func (r *CompanyRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.BusCompany, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// UpdateStorage は荷物置き場の画像URLと説明を更新する
func (r *CompanyRepository) UpdateStorage(ctx context.Context, id uuid.UUID, imageURL, description string) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}
