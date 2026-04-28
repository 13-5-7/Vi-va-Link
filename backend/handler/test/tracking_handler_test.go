package handler_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/bus-logistics/backend/handler"
	"github.com/bus-logistics/backend/model"
	"github.com/bus-logistics/backend/service"
	"github.com/google/uuid"
)

// ---- GetByTrackingNumber ----

func TestGetByTrackingNumber_NotFound(t *testing.T) {
	e := newEcho()
	h := handler.NewTrackingHandler(&MockTrackingService{
		GetByTrackingNumberFunc: func(_ context.Context, _ string) (*service.TrackingInfo, error) {
			return nil, service.ErrBookingNotFound
		},
	})

	req, rec := makeRequest(http.MethodGet, "/", "")
	c := e.NewContext(req, rec)
	c.SetParamNames("tracking_number")
	c.SetParamValues("TRK-NOTEXIST")

	if err := h.GetByTrackingNumber(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "NOT_FOUND" {
		t.Errorf("want NOT_FOUND, got %s", code)
	}
}

func TestGetByTrackingNumber_Success(t *testing.T) {
	e := newEcho()
	now := time.Now()
	h := handler.NewTrackingHandler(&MockTrackingService{
		GetByTrackingNumberFunc: func(_ context.Context, tn string) (*service.TrackingInfo, error) {
			return &service.TrackingInfo{
				Booking: &model.Booking{
					TrackingNumber:  tn,
					Status:          model.BookingStatusInTransit,
					StatusUpdatedAt: now,
				},
				Schedule: &model.Schedule{
					OriginName: "Tokyo Station",
					DestName:   "Osaka Station",
					DepartAt:   now,
				},
			}, nil
		},
	})

	req, rec := makeRequest(http.MethodGet, "/", "")
	c := e.NewContext(req, rec)
	c.SetParamNames("tracking_number")
	c.SetParamValues("TRK-ABC123")

	if err := h.GetByTrackingNumber(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
	body := decodeBody(t, rec)
	if body["tracking_number"] != "TRK-ABC123" {
		t.Errorf("want TRK-ABC123, got %v", body["tracking_number"])
	}
	if body["status"] != string(model.BookingStatusInTransit) {
		t.Errorf("want in_transit, got %v", body["status"])
	}
	schedule, ok := body["schedule"].(map[string]any)
	if !ok {
		t.Fatal("response must have 'schedule' object")
	}
	if schedule["origin_name"] != "Tokyo Station" {
		t.Errorf("want Tokyo Station, got %v", schedule["origin_name"])
	}
}

// ---- UpdateStatus ----

func TestUpdateStatus_BadRequest_InvalidBookingID(t *testing.T) {
	e := newEcho()
	h := handler.NewTrackingHandler(&MockTrackingService{})

	req, rec := makeRequest(http.MethodPatch, "/", `{"status":"loaded"}`)
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	if err := h.UpdateStatus(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
}

func TestUpdateStatus_Unauthorized_MissingUserID(t *testing.T) {
	e := newEcho()
	h := handler.NewTrackingHandler(&MockTrackingService{})

	req, rec := makeRequest(http.MethodPatch, "/", `{"status":"loaded"}`)
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	// user_id をセットしない

	if err := h.UpdateStatus(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rec.Code)
	}
}

func TestUpdateStatus_BadRequest_EmptyStatus(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	h := handler.NewTrackingHandler(&MockTrackingService{})

	req, rec := makeRequest(http.MethodPatch, "/", `{"status":""}`)
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	setUserID(c, operatorID.String())

	if err := h.UpdateStatus(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "VALIDATION_ERROR" {
		t.Errorf("want VALIDATION_ERROR, got %s", code)
	}
}

func TestUpdateStatus_NotFound_BookingNotFound(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	h := handler.NewTrackingHandler(&MockTrackingService{
		UpdateStatusFunc: func(_ context.Context, _ uuid.UUID, _ model.BookingStatus, _ uuid.UUID) error {
			return service.ErrBookingNotFound
		},
	})

	req, rec := makeRequest(http.MethodPatch, "/", `{"status":"loaded"}`)
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	setUserID(c, operatorID.String())

	if err := h.UpdateStatus(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", rec.Code)
	}
}

func TestUpdateStatus_BadRequest_InvalidTransition(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	h := handler.NewTrackingHandler(&MockTrackingService{
		UpdateStatusFunc: func(_ context.Context, _ uuid.UUID, _ model.BookingStatus, _ uuid.UUID) error {
			return service.ErrInvalidStatusTransition
		},
	})

	req, rec := makeRequest(http.MethodPatch, "/", `{"status":"accepted"}`)
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	setUserID(c, operatorID.String())

	if err := h.UpdateStatus(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "VALIDATION_ERROR" {
		t.Errorf("want VALIDATION_ERROR, got %s", code)
	}
}

func TestUpdateStatus_Success(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	bookingID := uuid.New()
	var capturedStatus model.BookingStatus

	h := handler.NewTrackingHandler(&MockTrackingService{
		UpdateStatusFunc: func(_ context.Context, id uuid.UUID, status model.BookingStatus, opID uuid.UUID) error {
			if id != bookingID {
				t.Errorf("unexpected bookingID: %v", id)
			}
			if opID != operatorID {
				t.Errorf("unexpected operatorID: %v", opID)
			}
			capturedStatus = status
			return nil
		},
	})

	req, rec := makeRequest(http.MethodPatch, "/", `{"status":"loaded"}`)
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(bookingID.String())
	setUserID(c, operatorID.String())

	if err := h.UpdateStatus(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
	if capturedStatus != model.BookingStatusLoaded {
		t.Errorf("want loaded, got %v", capturedStatus)
	}
}

// ---- GetByTrackingNumber: 空文字 ----

func TestGetByTrackingNumber_BadRequest_EmptyTrackingNumber(t *testing.T) {
	e := newEcho()
	h := handler.NewTrackingHandler(&MockTrackingService{})

	req, rec := makeRequest(http.MethodGet, "/", "")
	c := e.NewContext(req, rec)
	c.SetParamNames("tracking_number")
	c.SetParamValues("")

	if err := h.GetByTrackingNumber(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "BAD_REQUEST" {
		t.Errorf("want BAD_REQUEST, got %s", code)
	}
}
