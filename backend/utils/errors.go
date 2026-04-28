package utils

type AppError struct {
    StatusCode int    // HTTPステータスコード
    ErrorCode  string // "VALIDATION_ERROR" などの識別子
    Message    string // ユーザー向けメッセージ
}

// error インターフェースを満たすためのメソッド
func (e *AppError) Error() string {
    return e.Message
}

// 頻出するエラーを楽に作るためのヘルパー
func NewAppError(status int, code string, msg string) *AppError {
    return &AppError{status, code, msg}
}