//go:build integration

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	custommiddleware "github.com/bus-logistics/backend/middleware"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func TestIntegrationFlow(t *testing.T) {
	t.Skip("integration test: requires running database")
	// 認証 → スケジュール登録 → 予約 → 追跡の一連のフローをテスト
}

// setupTestServer は httptest.NewServer でテスト用 Echo サーバーを起動する。
// 実際の DB には接続しない（統合テストは t.Skip でスキップ）。
func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	e := echo.New()
	e.HTTPErrorHandler = custommiddleware.CustomErrorHandler
	e.Use(middleware.Recover())

	// ヘルスチェックのみ登録（DB不要）
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	return httptest.NewServer(e)
}

// TestHealthEndpoint はヘルスチェックエンドポイントが正常に応答することを確認する。
// このテストは DB 不要で実行できる。
func TestHealthEndpoint(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
}

// TestIntegrationAuthToTracking は認証 → スケジュール登録 → 予約 → 追跡の
// 一連のフローを検証する統合テスト。実行には稼働中の DB が必要。
func TestIntegrationAuthToTracking(t *testing.T) {
	t.Skip("integration test: requires running database")

	// --- 1. オペレーター登録 ---
	operatorEmail := fmt.Sprintf("operator-%d@example.com", time.Now().UnixNano())
	operatorToken := registerAndLogin(t, operatorEmail, "password123", "bus_operator")

	// --- 2. 荷主登録 ---
	shipperEmail := fmt.Sprintf("shipper-%d@example.com", time.Now().UnixNano())
	shipperToken := registerAndLogin(t, shipperEmail, "password123", "shipper")

	// --- 3. スケジュール登録（オペレーター） ---
	scheduleID := createSchedule(t, operatorToken)

	// --- 4. 予約（荷主） ---
	trackingNumber := createBooking(t, shipperToken, scheduleID)

	// --- 5. 追跡（認証不要） ---
	trackPackage(t, trackingNumber)
}

// 以下はヘルパー関数（実際の HTTP クライアントを使用）

func registerAndLogin(t *testing.T, email, password, role string) string {
	t.Helper()
	baseURL := "http://localhost:8080"

	// 登録
	regBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
		"role":     role,
	})
	resp, err := http.Post(baseURL+"/api/v1/auth/register", "application/json", bytes.NewReader(regBody))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d", resp.StatusCode)
	}

	// ログイン
	loginBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	resp2, err := http.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("login: expected 200, got %d", resp2.StatusCode)
	}

	var loginResp map[string]any
	if err := json.NewDecoder(resp2.Body).Decode(&loginResp); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	token, ok := loginResp["token"].(string)
	if !ok || token == "" {
		t.Fatal("token not found in login response")
	}
	return token
}

func createSchedule(t *testing.T, token string) string {
	t.Helper()
	baseURL := "http://localhost:8080"

	body, _ := json.Marshal(map[string]any{
		"origin_lat":    35.6812,
		"origin_lng":    139.7671,
		"origin_name":   "東京駅",
		"dest_lat":      34.6937,
		"dest_lng":      135.5023,
		"dest_name":     "大阪駅",
		"depart_at":     time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"arrive_at":     time.Now().Add(30 * time.Hour).Format(time.RFC3339),
		"max_weight_kg": 100.0,
		"max_size_cm":   200.0,
	})

	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/schedules", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create schedule failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create schedule: expected 201, got %d", resp.StatusCode)
	}

	var scheduleResp map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&scheduleResp); err != nil {
		t.Fatalf("decode schedule response: %v", err)
	}
	id, ok := scheduleResp["id"].(string)
	if !ok || id == "" {
		t.Fatal("schedule id not found in response")
	}
	return id
}

func createBooking(t *testing.T, token, scheduleID string) string {
	t.Helper()
	baseURL := "http://localhost:8080"

	body, _ := json.Marshal(map[string]any{
		"schedule_id":     scheduleID,
		"weight_kg":       10.0,
		"size_cm":         50.0,
		"content_desc":    "テスト荷物",
		"recipient_name":  "山田太郎",
		"recipient_phone": "090-0000-0000",
		"recipient_addr":  "大阪府大阪市",
	})

	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/bookings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create booking failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create booking: expected 201, got %d", resp.StatusCode)
	}

	var bookingResp map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&bookingResp); err != nil {
		t.Fatalf("decode booking response: %v", err)
	}
	tn, ok := bookingResp["tracking_number"].(string)
	if !ok || tn == "" {
		t.Fatal("tracking_number not found in booking response")
	}
	return tn
}

func trackPackage(t *testing.T, trackingNumber string) {
	t.Helper()
	baseURL := "http://localhost:8080"

	resp, err := http.Get(baseURL + "/api/v1/tracking/" + trackingNumber)
	if err != nil {
		t.Fatalf("track package failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("track package: expected 200, got %d", resp.StatusCode)
	}

	var trackResp map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&trackResp); err != nil {
		t.Fatalf("decode tracking response: %v", err)
	}
	if trackResp["tracking_number"] != trackingNumber {
		t.Errorf("expected tracking_number=%q, got %q", trackingNumber, trackResp["tracking_number"])
	}
	if trackResp["status"] == nil {
		t.Error("status field missing in tracking response")
	}
}
