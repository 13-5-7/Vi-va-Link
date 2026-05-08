package handler

import (
	"net/http"
	"log"

	"github.com/bus-logistics/backend/repository"
	"github.com/bus-logistics/backend/utils"
	//"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type CompanyHandler struct {
	repo    *repository.CompanyRepository
	userRepo *repository.UserRepository
	pool    *pgxpool.Pool
}

func NewCompanyHandler(repo *repository.CompanyRepository, userRepo *repository.UserRepository, pool *pgxpool.Pool) *CompanyHandler {
	return &CompanyHandler{repo: repo, userRepo: userRepo, pool: pool}
}

// List バス会社一覧を全件取得して返却する
// GET /api/v1/companies
func (h *CompanyHandler) List(c echo.Context) error {
	log.Println("-----handler/company.go List called-----")

	companies, err := h.repo.List(c.Request().Context())
	if err != nil {
		// データベース接続失敗などの予期せぬエラー
		return utils.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
	return c.JSON(http.StatusOK, map[string]any{"companies": companies})
}

// GetMyCompany GET /api/v1/companies/me  (Operator自身の会社情報)
func (h *CompanyHandler) GetMyCompany(c echo.Context) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// UpdateStorage PATCH /api/v1/companies/me/storage  (荷物置き場情報更新)
func (h *CompanyHandler) UpdateStorage(c echo.Context) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}
