package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// AdminKeyAuth は X-Admin-Key ヘッダーによる簡易管理者認証ミドルウェア
// ADMIN_KEY が未設定の場合はすべてのリクエストを拒否する
func AdminKeyAuth(adminKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if adminKey == "" {
				return c.JSON(http.StatusServiceUnavailable, map[string]any{
					"error": map[string]string{
						"code":    "ADMIN_DISABLED",
						"message": "管理者機能は設定されていません",
					},
				})
			}

			key := c.Request().Header.Get("X-Admin-Key")
			if key == "" || key != adminKey {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"error": map[string]string{
						"code":    "UNAUTHORIZED",
						"message": "管理者キーが無効です",
					},
				})
			}

			return next(c)
		}
	}
}
