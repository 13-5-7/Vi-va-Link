package handler

import (
	"errors"
	"net/http"
	"regexp"
	"log"

	"github.com/bus-logistics/backend/model"
	"github.com/bus-logistics/backend/service"
	"github.com/bus-logistics/backend/utils"
	"github.com/labstack/echo/v4"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type AuthHandler struct {
	authService AuthServiceInterface
}

func NewAuthHandler(authService AuthServiceInterface) *AuthHandler {
	log.Println("-----NewAuthHandler called-----")
	if authService == nil {
		log.Fatal("authService is required for AuthHandler")
	}
	return &AuthHandler{authService: authService}
}

type registerRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	Role       string `json:"role"`
	InviteCode string `json:"invite_code"` // bus_operator 登録時に必須
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func errResponse(code, message string) map[string]any { //nolint:unused
	return map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	}
}

// Register 招待コードを使用してオペレーターユーザーを新規登録する
// 前提：有効な招待コードが事前に発行されている必要あり
func (h *AuthHandler) Register(c echo.Context) error {
	log.Println("-----handler/auth.go Register called-----")

	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return utils.NewAppError(http.StatusBadRequest, "BAD_REQUEST", "invalid request body") 
	}

	if utils.IsEmpty(req.Email) || utils.IsEmpty(req.Password) {
		return utils.NewAppError(http.StatusBadRequest, "VALIDATION_ERROR", "email and password are required") 
	}
	if !emailRegex.MatchString(req.Email) {
		return utils.NewAppError(http.StatusBadRequest, "VALIDATION_ERROR", "invalid email format") 
	}
	if len(req.Password) < 8 {
		return utils.NewAppError(http.StatusBadRequest, "VALIDATION_ERROR", "password must be at least 8 characters") 
	}
	if req.Role != string(model.RoleBusOperator) && req.Role != string(model.RoleShipper) {
		return utils.NewAppError(http.StatusBadRequest, "VALIDATION_ERROR", "role must be 'bus_operator' or 'shipper'") 
	}

	user, err := h.authService.Register(c.Request().Context(), service.RegisterRequest{
		Email:      req.Email,
		Password:   req.Password,
		Role:       model.Role(req.Role),
		InviteCode: req.InviteCode,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailAlreadyExists):
			return utils.NewAppError(http.StatusConflict, "EMAIL_ALREADY_EXISTS", "email already exists") 
		case errors.Is(err, service.ErrInvalidInviteCode):
			return utils.NewAppError(http.StatusBadRequest, "INVALID_INVITE_CODE", "招待コードが無効または使用済みです") 
		default:
			return utils.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error") 
		}
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
	})
}

// Login はユーザー認証を行い、JWTトークンを返却します
func (h *AuthHandler) Login(c echo.Context) error {
	log.Println("-----handler/auth.go Login called-----")
	
	var req loginRequest
	// リクエストボディのバインド
	if err := c.Bind(&req); err != nil {
		return utils.NewAppError(http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	// 入力値チェック
	if utils.IsEmpty(req.Email) || utils.IsEmpty(req.Password) || utils.IsEmpty(req.Role) {
		return utils.NewAppError(http.StatusUnprocessableEntity, "VALIDATION_ERROR", "missing required fields")
	}

	// メールアドレス形式チェック
	if !emailRegex.MatchString(req.Email) {
		return utils.NewAppError(http.StatusUnprocessableEntity, "INVALID_EMAIL_FORMAT", "invalid email format")
	}

	// 認証処理の実行
	resp, err := h.authService.Login(c.Request().Context(), service.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	})

	// エラーハンドリング
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return utils.NewAppError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
		}
		return utils.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}

	// 成功時はトークン情報を返却
	return c.JSON(http.StatusOK, map[string]any{
		"token":   resp.Token,
		"user_id": resp.UserID,
		"role":    resp.Role,
	})
}
