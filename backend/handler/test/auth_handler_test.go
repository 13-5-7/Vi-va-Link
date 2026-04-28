package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bus-logistics/backend/handler"
	"github.com/bus-logistics/backend/model"
	"github.com/bus-logistics/backend/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func newEcho() *echo.Echo { return echo.New() }

func makeRequest(method, path, body string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return req, httptest.NewRecorder()
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode error: %v\nbody: %s", err, rec.Body.String())
	}
	return m
}

func errCode(t *testing.T, body map[string]any) string {
	t.Helper()
	errMap, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatal("response has no 'error' object")
	}
	return errMap["code"].(string)
}

// ---- Register ----

func TestRegister_ValidationError_EmptyEmail(t *testing.T) {
	e := newEcho()
	h := handler.NewAuthHandler(&MockAuthService{})

	req, rec := makeRequest(http.MethodPost, "/", `{"email":"","password":"pass","role":"bus_operator"}`)
	c := e.NewContext(req, rec)

	if err := h.Register(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "VALIDATION_ERROR" {
		t.Errorf("want VALIDATION_ERROR, got %s", code)
	}
}

func TestRegister_ValidationError_EmptyPassword(t *testing.T) {
	e := newEcho()
	h := handler.NewAuthHandler(&MockAuthService{})

	req, rec := makeRequest(http.MethodPost, "/", `{"email":"a@b.com","password":"","role":"bus_operator"}`)
	c := e.NewContext(req, rec)

	if err := h.Register(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
}

func TestRegister_ValidationError_InvalidRole(t *testing.T) {
	e := newEcho()
	h := handler.NewAuthHandler(&MockAuthService{})

	req, rec := makeRequest(http.MethodPost, "/", `{"email":"a@b.com","password":"pass","role":"admin"}`)
	c := e.NewContext(req, rec)

	if err := h.Register(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "VALIDATION_ERROR" {
		t.Errorf("want VALIDATION_ERROR, got %s", code)
	}
}

func TestRegister_Conflict_EmailAlreadyExists(t *testing.T) {
	e := newEcho()
	h := handler.NewAuthHandler(&MockAuthService{
		RegisterFunc: func(_ context.Context, _ service.RegisterRequest) (*model.User, error) {
			return nil, service.ErrEmailAlreadyExists
		},
	})

	req, rec := makeRequest(http.MethodPost, "/", `{"email":"a@example.com","password":"password123","role":"shipper"}`)
	c := e.NewContext(req, rec)

	if err := h.Register(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("want 409, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "EMAIL_ALREADY_EXISTS" {
		t.Errorf("want EMAIL_ALREADY_EXISTS, got %s", code)
	}
}

func TestRegister_Success(t *testing.T) {
	uid := uuid.New()
	e := newEcho()
	h := handler.NewAuthHandler(&MockAuthService{
		RegisterFunc: func(_ context.Context, req service.RegisterRequest) (*model.User, error) {
			return &model.User{
				ID:    uid,
				Email: req.Email,
				Role:  req.Role,
			}, nil
		},
	})

	req, rec := makeRequest(http.MethodPost, "/", `{"email":"a@example.com","password":"password123","role":"bus_operator"}`)
	c := e.NewContext(req, rec)

	if err := h.Register(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("want 201, got %d", rec.Code)
	}
	body := decodeBody(t, rec)
	if body["email"] != "a@example.com" {
		t.Errorf("want email a@example.com, got %v", body["email"])
	}
}

// ---- Login ----

func TestLogin_ValidationError_EmptyEmail(t *testing.T) {
	e := newEcho()
	h := handler.NewAuthHandler(&MockAuthService{})

	req, rec := makeRequest(http.MethodPost, "/", `{"email":"","password":"pass"}`)
	c := e.NewContext(req, rec)

	if err := h.Login(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
}

func TestLogin_ValidationError_EmptyPassword(t *testing.T) {
	e := newEcho()
	h := handler.NewAuthHandler(&MockAuthService{})

	req, rec := makeRequest(http.MethodPost, "/", `{"email":"a@b.com","password":""}`)
	c := e.NewContext(req, rec)

	if err := h.Login(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
}

func TestLogin_Unauthorized_InvalidCredentials(t *testing.T) {
	e := newEcho()
	h := handler.NewAuthHandler(&MockAuthService{
		LoginFunc: func(_ context.Context, _ service.LoginRequest) (*service.LoginResponse, error) {
			return nil, service.ErrInvalidCredentials
		},
	})

	req, rec := makeRequest(http.MethodPost, "/", `{"email":"a@b.com","password":"wrong"}`)
	c := e.NewContext(req, rec)

	if err := h.Login(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "INVALID_CREDENTIALS" {
		t.Errorf("want INVALID_CREDENTIALS, got %s", code)
	}
}

func TestLogin_Success(t *testing.T) {
	uid := uuid.New()
	e := newEcho()
	h := handler.NewAuthHandler(&MockAuthService{
		LoginFunc: func(_ context.Context, _ service.LoginRequest) (*service.LoginResponse, error) {
			return &service.LoginResponse{
				Token:  "jwt-token",
				UserID: uid,
				Role:   model.RoleBusOperator,
			}, nil
		},
	})

	req, rec := makeRequest(http.MethodPost, "/", `{"email":"a@b.com","password":"pass","role":"bus_operator"}`)
	c := e.NewContext(req, rec)

	if err := h.Login(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
	body := decodeBody(t, rec)
	if body["token"] != "jwt-token" {
		t.Errorf("want token jwt-token, got %v", body["token"])
	}
	if body["role"] != string(model.RoleBusOperator) {
		t.Errorf("want role bus_operator, got %v", body["role"])
	}
}

// ---- errResponse format ----

func TestErrResponse_HasCodeAndMessage(t *testing.T) {
	e := newEcho()
	h := handler.NewAuthHandler(&MockAuthService{})

	req, rec := makeRequest(http.MethodPost, "/", `{"email":"","password":""}`)
	c := e.NewContext(req, rec)
	_ = h.Login(c)

	body := decodeBody(t, rec)
	errMap, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatal("response must have 'error' object")
	}
	if _, ok := errMap["code"]; !ok {
		t.Error("error object must have 'code'")
	}
	if _, ok := errMap["message"]; !ok {
		t.Error("error object must have 'message'")
	}
}

// time パッケージの使用（コンパイルエラー回避）
var _ = time.Now

// ---- Register: メール形式・パスワード長バリデーション ----

func TestRegister_ValidationError_InvalidEmailFormat(t *testing.T) {
	e := newEcho()
	h := handler.NewAuthHandler(&MockAuthService{})

	req, rec := makeRequest(http.MethodPost, "/", `{"email":"notanemail","password":"password123","role":"shipper"}`)
	c := e.NewContext(req, rec)

	if err := h.Register(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "VALIDATION_ERROR" {
		t.Errorf("want VALIDATION_ERROR, got %s", code)
	}
}

func TestRegister_ValidationError_PasswordTooShort(t *testing.T) {
	e := newEcho()
	h := handler.NewAuthHandler(&MockAuthService{})

	req, rec := makeRequest(http.MethodPost, "/", `{"email":"a@b.com","password":"short","role":"shipper"}`)
	c := e.NewContext(req, rec)

	if err := h.Register(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "VALIDATION_ERROR" {
		t.Errorf("want VALIDATION_ERROR, got %s", code)
	}
}

func TestRegister_Success_ValidEmail8CharPassword(t *testing.T) {
	uid := uuid.New()
	e := newEcho()
	h := handler.NewAuthHandler(&MockAuthService{
		RegisterFunc: func(_ context.Context, req service.RegisterRequest) (*model.User, error) {
			return &model.User{ID: uid, Email: req.Email, Role: req.Role}, nil
		},
	})

	req, rec := makeRequest(http.MethodPost, "/", `{"email":"valid@example.com","password":"password1","role":"shipper"}`)
	c := e.NewContext(req, rec)

	if err := h.Register(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("want 201, got %d\nbody: %s", rec.Code, rec.Body.String())
	}
}
