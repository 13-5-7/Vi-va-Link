package middleware

import (
	"github.com/labstack/echo/v4"
)

// SecurityHeaders はセキュリティ関連のレスポンスヘッダーを設定するミドルウェア
func SecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			h := c.Response().Header()

			// X-Content-Type-Options: MIMEスニッフィング攻撃を防ぐ
			h.Set("X-Content-Type-Options", "nosniff")

			// X-Frame-Options: クリックジャッキング攻撃を防ぐ
			h.Set("X-Frame-Options", "DENY")

			// X-XSS-Protection: 古いブラウザ向けXSSフィルター（現代ブラウザはCSPで対応）
			h.Set("X-XSS-Protection", "1; mode=block")

			// Referrer-Policy: リファラー情報の漏洩を制限
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions-Policy: 不要なブラウザ機能を無効化
			h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(self)")

			return next(c)
		}
	}
}
