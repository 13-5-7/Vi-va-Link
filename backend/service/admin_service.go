package service

import (
	"context"
	//"time"

	"github.com/bus-logistics/backend/repository"
	"github.com/google/uuid"
)

// InviteCodeInfo は招待コード情報（管理者API用）
type InviteCodeInfo struct {
	ID        uuid.UUID  `json:"id"`
	Code      string     `json:"code"`
	CompanyID uuid.UUID  `json:"company_id"`
	UsedBy    *uuid.UUID `json:"used_by"`
	UsedAt    *string    `json:"used_at"`
	CreatedAt string     `json:"created_at"`
}

type AdminService struct {
	inviteRepo *repository.InviteRepository
}

func NewAdminService(inviteRepo *repository.InviteRepository) *AdminService {
	return &AdminService{inviteRepo: inviteRepo}
}

// IssueInviteCode は指定バス会社向けの招待コードを発行する
func (s *AdminService) IssueInviteCode(ctx context.Context, companyID uuid.UUID) (string, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// ListInviteCodes は招待コード一覧を返す
func (s *AdminService) ListInviteCodes(ctx context.Context, companyID *uuid.UUID) ([]InviteCodeInfo, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}
