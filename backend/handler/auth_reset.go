package handler

import (
	"context"
	//"errors"
	//"net/http"

	//"github.com/bus-logistics/backend/repository"
	//"github.com/bus-logistics/backend/service"
	"github.com/labstack/echo/v4"
)

// PasswordResetServiceInterface はPasswordResetHandlerが依存するサービスのインターフェース
type PasswordResetServiceInterface interface {
	RequestReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}

type PasswordResetHandler struct {
	resetService PasswordResetServiceInterface
}

func NewPasswordResetHandler(resetService PasswordResetServiceInterface) *PasswordResetHandler {
	return &PasswordResetHandler{resetService: resetService}
}

// RequestReset POST /api/v1/auth/password-reset/request
// パスワードリセットトークンを発行する（メール送信は現在ログ出力）
func (h *PasswordResetHandler) RequestReset(c echo.Context) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// ResetPassword POST /api/v1/auth/password-reset/confirm
// トークンを検証してパスワードを更新する
func (h *PasswordResetHandler) ResetPassword(c echo.Context) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}
