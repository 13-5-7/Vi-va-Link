package handler

import (
	"context"
	//"net/http"

	"github.com/bus-logistics/backend/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AdminInviteServiceInterface は管理者用招待コードサービスのインターフェース
type AdminInviteServiceInterface interface {
	IssueInviteCode(ctx context.Context, companyID uuid.UUID) (string, error)
	ListInviteCodes(ctx context.Context, companyID *uuid.UUID) ([]service.InviteCodeInfo, error)
}

type AdminHandler struct {
	inviteService AdminInviteServiceInterface
}

func NewAdminHandler(inviteService AdminInviteServiceInterface) *AdminHandler {
	return &AdminHandler{inviteService: inviteService}
}

// IssueInviteCode POST /api/v1/admin/invite-codes
// 指定バス会社向けの招待コードを発行する（管理者のみ）
func (h *AdminHandler) IssueInviteCode(c echo.Context) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// ListInviteCodes GET /api/v1/admin/invite-codes
// 招待コード一覧を取得する（管理者のみ）
func (h *AdminHandler) ListInviteCodes(c echo.Context) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}
