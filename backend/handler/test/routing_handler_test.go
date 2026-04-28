package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bus-logistics/backend/handler"
	"github.com/labstack/echo/v4"
)

func TestGetRoute_BadRequest_MissingParams(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		errKey string
	}{
		{"missing origin_lng", "origin_lat=35.0&dest_lng=135.0&dest_lat=34.0", "invalid origin_lng"},
		{"missing origin_lat", "origin_lng=139.0&dest_lng=135.0&dest_lat=34.0", "invalid origin_lat"},
		{"missing dest_lng", "origin_lng=139.0&origin_lat=35.0&dest_lat=34.0", "invalid dest_lng"},
		{"missing dest_lat", "origin_lng=139.0&origin_lat=35.0&dest_lng=135.0", "invalid dest_lat"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			// 存在しないURLを使ってOSRM呼び出しを失敗させる
			h := handler.NewRoutingHandler("http://localhost:0")

			req := httptest.NewRequest(http.MethodGet, "/?"+tt.query, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if err := h.GetRoute(c); err != nil {
				t.Fatal(err)
			}
			if rec.Code != http.StatusBadRequest {
				t.Errorf("want 400, got %d", rec.Code)
			}
			var body map[string]string
			json.Unmarshal(rec.Body.Bytes(), &body)
			if body["error"] != tt.errKey {
				t.Errorf("want error=%q, got %q", tt.errKey, body["error"])
			}
		})
	}
}

func TestGetRoute_Fallback_WhenOSRMUnavailable(t *testing.T) {
	e := echo.New()
	// 到達不能なURLを指定してフォールバックを発動させる
	h := handler.NewRoutingHandler("http://localhost:0")

	req := httptest.NewRequest(http.MethodGet,
		"/?origin_lng=139.7671&origin_lat=35.6812&dest_lng=135.5023&dest_lat=34.6937", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.GetRoute(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body["type"] != "LineString" {
		t.Errorf("want type=LineString (fallback), got %v", body["type"])
	}
	coords, ok := body["coordinates"].([]any)
	if !ok || len(coords) != 2 {
		t.Errorf("want 2 coordinates, got %v", body["coordinates"])
	}
}

func TestGetRoute_Fallback_CoordinatesOrder(t *testing.T) {
	e := echo.New()
	h := handler.NewRoutingHandler("http://localhost:0")

	// origin: lng=139.7671, lat=35.6812 / dest: lng=135.5023, lat=34.6937
	req := httptest.NewRequest(http.MethodGet,
		"/?origin_lng=139.7671&origin_lat=35.6812&dest_lng=135.5023&dest_lat=34.6937", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = h.GetRoute(c)

	var body map[string]any
	json.Unmarshal(rec.Body.Bytes(), &body)

	coords := body["coordinates"].([]any)
	origin := coords[0].([]any)
	dest := coords[1].([]any)

	// GeoJSON は [lng, lat] の順
	if origin[0].(float64) != 139.7671 {
		t.Errorf("origin lng want 139.7671, got %v", origin[0])
	}
	if origin[1].(float64) != 35.6812 {
		t.Errorf("origin lat want 35.6812, got %v", origin[1])
	}
	if dest[0].(float64) != 135.5023 {
		t.Errorf("dest lng want 135.5023, got %v", dest[0])
	}
	if dest[1].(float64) != 34.6937 {
		t.Errorf("dest lat want 34.6937, got %v", dest[1])
	}
}
