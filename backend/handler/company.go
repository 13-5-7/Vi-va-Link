package handler

import (
	"net/http"
	"log"

	"github.com/bus-logistics/backend/repository"
	"github.com/bus-logistics/backend/utils"
	"github.com/google/uuid"
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

// GetMyCompany 実行ユーザー（Operator）が所属する会社の情報を取得する
// GetMyCompany GET /api/v1/companies/me
func (h *CompanyHandler) GetMyCompany(c echo.Context) error {
	log.Println("-----handler/company.go GetMyCompany called-----")

	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "missing user_id")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user_id")
	}

	user, err := h.userRepo.FindByID(c.Request().Context(), userID)
	if err != nil || user == nil || user.CompanyID == nil {
		return utils.NewAppError(http.StatusNotFound, "NOT_FOUND", "company not found")
	}

	company, err := h.repo.FindByID(c.Request().Context(), *user.CompanyID)
	if err != nil || company == nil {
		return utils.NewAppError(http.StatusNotFound, "NOT_FOUND", "company not found")
	}
	return c.JSON(http.StatusOK, company)
}

// UpdateStorage 操作者が所属する会社の荷物置き場情報を更新する
// PATCH /api/v1/companies/me/storage
func (h *CompanyHandler) UpdateStorage(c echo.Context) error {
	log.Println("-----handler/company.go UpdateStorage called-----")

	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "missing user_id")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user_id")
	}

	user, err := h.userRepo.FindByID(c.Request().Context(), userID)
	if err != nil || user == nil || user.CompanyID == nil {
		return utils.NewAppError(http.StatusNotFound, "NOT_FOUND", "company not found")
	}

	var req struct {
		StorageImageURL    string `json:"storage_image_url"`
		StorageDescription string `json:"storage_description"`
	}
	if err := c.Bind(&req); err != nil {
		return utils.NewAppError(http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	if err := h.repo.UpdateStorage(c.Request().Context(), *user.CompanyID, req.StorageImageURL, req.StorageDescription); err != nil {
		return utils.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}

	company, _ := h.repo.FindByID(c.Request().Context(), *user.CompanyID)
	return c.JSON(http.StatusOK, company)
}
