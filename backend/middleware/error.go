package middleware

import (
	"net/http"
	"log"

	"github.com/bus-logistics/backend/utils"
	"github.com/labstack/echo/v4"
)

func CustomErrorHandler(err error, c echo.Context) {
	log.Println("----error CustomErrorHandler called-----")

    // デフォルト値（予期せぬエラー用）
    code := http.StatusInternalServerError
    errorCode := "INTERNAL_ERROR"
    message := "Internal Server Error"

    // AppError (ビジネスエラー) かどうかを判定
    if ae, ok := err.(*utils.AppError); ok {
        code = ae.StatusCode
        errorCode = ae.ErrorCode
        message = ae.Message
    } else if he, ok := err.(*echo.HTTPError); ok {
        // Echoが投げる標準エラー（404, 405など）の判定
        code = he.Code
        if m, ok := he.Message.(string); ok {
            message = m
        }
        
        // ステータスコードに応じたエラーコードの割り当て
        switch code {
        case http.StatusNotFound:
            errorCode = "NOT_FOUND"
        case http.StatusUnauthorized:
            errorCode = "UNAUTHORIZED"
        case http.StatusBadRequest:
            errorCode = "BAD_REQUEST"
        case http.StatusForbidden:
            errorCode = "FORBIDDEN"
        }
    }

    // ログ出力
    if code >= 500 {
        log.Printf("[ERROR] 500系エラー発生: %v", err)
    }

    // レスポンス送信
    if !c.Response().Committed {
        c.JSON(code, map[string]any{
            "error": map[string]string{
                "code":    errorCode,
                "message": message,
            },
        })
    }
}