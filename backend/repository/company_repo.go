package repository

import (
	"context"
	"log"

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

// List バス会社一覧を名称順で取得する
func (r *CompanyRepository) List(ctx context.Context) ([]model.BusCompany, error) {
	log.Println("-----repository/company_repo.go List called-----")

	// 画像URLがNULLの場合は空文字として取得
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, COALESCE(storage_image_url,''), storage_description, created_at
		 FROM bus_companies ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []model.BusCompany
	for rows.Next() {
		var c model.BusCompany
		if err := rows.Scan(&c.ID, &c.Name, &c.StorageImageURL, &c.StorageDescription, &c.CreatedAt); err != nil {
			return nil, err
		}
		companies = append(companies, c)
	}
	return companies, rows.Err()
}

// FindByID 会社マスタテーブルから指定されたIDに一致する会社レコードを取得する
func (r *CompanyRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.BusCompany, error) {
	log.Println("-----repository/company_repo.go FindByID called-----")

	var c model.BusCompany
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, COALESCE(storage_image_url,''), storage_description, created_at
		 FROM bus_companies WHERE id = $1`, id,
	).Scan(&c.ID, &c.Name, &c.StorageImageURL, &c.StorageDescription, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateStorage 荷物置き場の画像URLと説明を更新する
func (r *CompanyRepository) UpdateStorage(ctx context.Context, id uuid.UUID, imageURL, description string) error {
	log.Println("-----repository/company_repo.go UpdateStorage called-----")

	_, err := r.pool.Exec(ctx,
		`UPDATE bus_companies SET storage_image_url = $1, storage_description = $2 WHERE id = $3`,
		imageURL, description, id,
	)
	return err
}
