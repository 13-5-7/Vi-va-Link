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
	"github.com/labstack/echo/v4"
)

func setUserID(c echo.Context, id string) {
	c.Set("user_id", id)
}

// ---- List ----

func TestBookingList_Unauthorized_MissingUserID(t *testing.T) {
	e := newEcho()
	h := handler.NewBookingHandler(&MockBookingService{})

	req, rec := makeRequest(http.MethodGet, "/", "")
	c := e.NewContext(req, rec)
	// user_id をセットしない

	if err := h.List(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rec.Code)
	}
}

func TestBookingList_Unauthorized_InvalidUserID(t *testing.T) {
	e := newEcho()
	h := handler.NewBookingHandler(&MockBookingService{})

	req, rec := makeRequest(http.MethodGet, "/", "")
	c := e.NewContext(req, rec)
	setUserID(c, "not-a-uuid")

	if err := h.List(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rec.Code)
	}
}

func TestBookingList_Success_Empty(t *testing.T) {
	e := newEcho()
	shipperID := uuid.New()
	h := handler.NewBookingHandler(&MockBookingService{
		ListByShipperFunc: func(_ context.Context, id uuid.UUID) ([]model.Booking, error) {
			if id != shipperID {
				t.Errorf("unexpected shipperID: %v", id)
			}
			return []model.Booking{}, nil
		},
	})

	req, rec := makeRequest(http.MethodGet, "/", "")
	c := e.NewContext(req, rec)
	setUserID(c, shipperID.String())

	if err := h.List(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
	body := decodeBody(t, rec)
	bookings, ok := body["bookings"].([]any)
	if !ok {
		t.Fatal("response must have 'bookings' array")
	}
	if len(bookings) != 0 {
		t.Errorf("want 0 bookings, got %d", len(bookings))
	}
}

func TestBookingList_Success_WithBookings(t *testing.T) {
	e := newEcho()
	shipperID := uuid.New()
	scheduleID := uuid.New()
	bookingID := uuid.New()

	h := handler.NewBookingHandler(&MockBookingService{
		ListByShipperFunc: func(_ context.Context, _ uuid.UUID) ([]model.Booking, error) {
			return []model.Booking{
				{
					ID:             bookingID,
					ScheduleID:     scheduleID,
					ShipperID:      shipperID,
					TrackingNumber: "TRK-TEST01",
					WeightKg:       5.0,
					Status:         model.BookingStatusAccepted,
					CreatedAt:      time.Now(),
				},
			}, nil
		},
	})

	req, rec := makeRequest(http.MethodGet, "/", "")
	c := e.NewContext(req, rec)
	setUserID(c, shipperID.String())

	if err := h.List(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
	body := decodeBody(t, rec)
	bookings := body["bookings"].([]any)
	if len(bookings) != 1 {
		t.Errorf("want 1 booking, got %d", len(bookings))
	}
}

// ---- Create ----

func TestBookingCreate_Unauthorized_MissingUserID(t *testing.T) {
	e := newEcho()
	h := handler.NewBookingHandler(&MockBookingService{})

	req, rec := makeRequest(http.MethodPost, "/", `{"schedule_id":"00000000-0000-0000-0000-000000000001","weight_kg":1}`)
	c := e.NewContext(req, rec)

	if err := h.Create(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rec.Code)
	}
}

func TestBookingCreate_Conflict_CapacityExceeded(t *testing.T) {
	e := newEcho()
	shipperID := uuid.New()
	h := handler.NewBookingHandler(&MockBookingService{
		CreateFunc: func(_ context.Context, _ service.CreateBookingRequest) (*model.Booking, error) {
			return nil, service.ErrCapacityExceeded
		},
	})

	req, rec := makeRequest(http.MethodPost, "/",
		`{"schedule_id":"00000000-0000-0000-0000-000000000001","weight_kg":999}`)
	c := e.NewContext(req, rec)
	setUserID(c, shipperID.String())

	if err := h.Create(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("want 409, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "CAPACITY_EXCEEDED" {
		t.Errorf("want CAPACITY_EXCEEDED, got %s", code)
	}
}

func TestBookingCreate_Conflict_SizeExceeded(t *testing.T) {
	e := newEcho()
	shipperID := uuid.New()
	h := handler.NewBookingHandler(&MockBookingService{
		CreateFunc: func(_ context.Context, _ service.CreateBookingRequest) (*model.Booking, error) {
			return nil, service.ErrSizeExceeded
		},
	})

	req, rec := makeRequest(http.MethodPost, "/",
		`{"schedule_id":"00000000-0000-0000-0000-000000000001","weight_kg":1,"size_cm":999}`)
	c := e.NewContext(req, rec)
	setUserID(c, shipperID.String())

	if err := h.Create(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("want 409, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "SIZE_EXCEEDED" {
		t.Errorf("want SIZE_EXCEEDED, got %s", code)
	}
}

func TestBookingCreate_NotFound_ScheduleNotFound(t *testing.T) {
	e := newEcho()
	shipperID := uuid.New()
	h := handler.NewBookingHandler(&MockBookingService{
		CreateFunc: func(_ context.Context, _ service.CreateBookingRequest) (*model.Booking, error) {
			return nil, service.ErrScheduleNotFound
		},
	})

	req, rec := makeRequest(http.MethodPost, "/",
		`{"schedule_id":"00000000-0000-0000-0000-000000000001","weight_kg":1}`)
	c := e.NewContext(req, rec)
	setUserID(c, shipperID.String())

	if err := h.Create(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", rec.Code)
	}
}

func TestBookingCreate_Success(t *testing.T) {
	e := newEcho()
	shipperID := uuid.New()
	scheduleID := uuid.New()
	bookingID := uuid.New()

	h := handler.NewBookingHandler(&MockBookingService{
		CreateFunc: func(_ context.Context, req service.CreateBookingRequest) (*model.Booking, error) {
			return &model.Booking{
				ID:             bookingID,
				ScheduleID:     scheduleID,
				ShipperID:      shipperID,
				TrackingNumber: "TRK-NEW001",
				WeightKg:       req.WeightKg,
				Status:         model.BookingStatusAccepted,
				CreatedAt:      time.Now(),
			}, nil
		},
	})

	body := `{"schedule_id":"` + scheduleID.String() + `","weight_kg":3.5,"size_cm":50,"content_desc":"test","recipient_name":"Taro","recipient_phone":"090","recipient_addr":"Tokyo"}`
	req, rec := makeRequest(http.MethodPost, "/", body)
	c := e.NewContext(req, rec)
	setUserID(c, shipperID.String())

	if err := h.Create(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("want 201, got %d\nbody: %s", rec.Code, rec.Body.String())
	}
	respBody := decodeBody(t, rec)
	if respBody["tracking_number"] != "TRK-NEW001" {
		t.Errorf("want TRK-NEW001, got %v", respBody["tracking_number"])
	}
}

// ---- GetByID ----

func TestBookingGetByID_BadRequest_InvalidID(t *testing.T) {
	e := newEcho()
	h := handler.NewBookingHandler(&MockBookingService{})

	req, rec := makeRequest(http.MethodGet, "/", "")
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	if err := h.GetByID(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
}

func TestBookingGetByID_NotFound(t *testing.T) {
	e := newEcho()
	h := handler.NewBookingHandler(&MockBookingService{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*model.Booking, error) {
			return nil, nil
		},
	})

	req, rec := makeRequest(http.MethodGet, "/", "")
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	if err := h.GetByID(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", rec.Code)
	}
}

func TestBookingGetByID_Success(t *testing.T) {
	e := newEcho()
	bookingID := uuid.New()
	h := handler.NewBookingHandler(&MockBookingService{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*model.Booking, error) {
			return &model.Booking{
				ID:             id,
				TrackingNumber: "TRK-FOUND",
				Status:         model.BookingStatusLoaded,
				CreatedAt:      time.Now(),
			}, nil
		},
	})

	req, rec := makeRequest(http.MethodGet, "/", "")
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(bookingID.String())

	if err := h.GetByID(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
	body := decodeBody(t, rec)
	if body["tracking_number"] != "TRK-FOUND" {
		t.Errorf("want TRK-FOUND, got %v", body["tracking_number"])
	}
}

// ---- Create: システム制限エラー ----

func TestBookingCreate_BadRequest_WeightLimitExceeded(t *testing.T) {
	e := newEcho()
	shipperID := uuid.New()
	h := handler.NewBookingHandler(&MockBookingService{
		CreateFunc: func(_ context.Context, _ service.CreateBookingRequest) (*model.Booking, error) {
			return nil, service.ErrWeightLimitExceeded
		},
	})

	req, rec := makeRequest(http.MethodPost, "/",
		`{"schedule_id":"00000000-0000-0000-0000-000000000001","weight_kg":11}`)
	c := e.NewContext(req, rec)
	setUserID(c, shipperID.String())

	if err := h.Create(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "WEIGHT_LIMIT_EXCEEDED" {
		t.Errorf("want WEIGHT_LIMIT_EXCEEDED, got %s", code)
	}
}

func TestBookingCreate_BadRequest_SizeLimitExceeded(t *testing.T) {
	e := newEcho()
	shipperID := uuid.New()
	h := handler.NewBookingHandler(&MockBookingService{
		CreateFunc: func(_ context.Context, _ service.CreateBookingRequest) (*model.Booking, error) {
			return nil, service.ErrSizeLimitExceeded
		},
	})

	req, rec := makeRequest(http.MethodPost, "/",
		`{"schedule_id":"00000000-0000-0000-0000-000000000001","weight_kg":1,"size_cm":141}`)
	c := e.NewContext(req, rec)
	setUserID(c, shipperID.String())

	if err := h.Create(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "SIZE_LIMIT_EXCEEDED" {
		t.Errorf("want SIZE_LIMIT_EXCEEDED, got %s", code)
	}
}

// ---- Cancel ----

func TestBookingCancel_BadRequest_InvalidID(t *testing.T) {
	e := newEcho()
	h := handler.NewBookingHandler(&MockBookingService{})

	req, rec := makeRequest(http.MethodDelete, "/", "")
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	if err := h.Cancel(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
}

func TestBookingCancel_Unauthorized_MissingUserID(t *testing.T) {
	e := newEcho()
	h := handler.NewBookingHandler(&MockBookingService{})

	req, rec := makeRequest(http.MethodDelete, "/", "")
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	if err := h.Cancel(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rec.Code)
	}
}

func TestBookingCancel_NotFound(t *testing.T) {
	e := newEcho()
	shipperID := uuid.New()
	h := handler.NewBookingHandler(&MockBookingService{
		CancelFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
			return service.ErrBookingNotFound
		},
	})

	req, rec := makeRequest(http.MethodDelete, "/", "")
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	setUserID(c, shipperID.String())

	if err := h.Cancel(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "NOT_FOUND" {
		t.Errorf("want NOT_FOUND, got %s", code)
	}
}

func TestBookingCancel_Forbidden(t *testing.T) {
	e := newEcho()
	shipperID := uuid.New()
	h := handler.NewBookingHandler(&MockBookingService{
		CancelFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
			return service.ErrForbidden
		},
	})

	req, rec := makeRequest(http.MethodDelete, "/", "")
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	setUserID(c, shipperID.String())

	if err := h.Cancel(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("want 403, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "FORBIDDEN" {
		t.Errorf("want FORBIDDEN, got %s", code)
	}
}

func TestBookingCancel_Conflict_CannotCancel(t *testing.T) {
	e := newEcho()
	shipperID := uuid.New()
	h := handler.NewBookingHandler(&MockBookingService{
		CancelFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
			return service.ErrCannotCancel
		},
	})

	req, rec := makeRequest(http.MethodDelete, "/", "")
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	setUserID(c, shipperID.String())

	if err := h.Cancel(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("want 409, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "CANNOT_CANCEL" {
		t.Errorf("want CANNOT_CANCEL, got %s", code)
	}
}

func TestBookingCancel_Success(t *testing.T) {
	e := newEcho()
	shipperID := uuid.New()
	bookingID := uuid.New()
	var capturedBookingID, capturedShipperID uuid.UUID

	h := handler.NewBookingHandler(&MockBookingService{
		CancelFunc: func(_ context.Context, bID uuid.UUID, sID uuid.UUID) error {
			capturedBookingID = bID
			capturedShipperID = sID
			return nil
		},
	})

	req, rec := makeRequest(http.MethodDelete, "/", "")
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(bookingID.String())
	setUserID(c, shipperID.String())

	if err := h.Cancel(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d\nbody: %s", rec.Code, rec.Body.String())
	}
	if capturedBookingID != bookingID {
		t.Errorf("want bookingID %v, got %v", bookingID, capturedBookingID)
	}
	if capturedShipperID != shipperID {
		t.Errorf("want shipperID %v, got %v", shipperID, capturedShipperID)
	}
}

// ---- Create: 0・負値バリデーション ----

func TestBookingCreate_BadRequest_ZeroWeight(t *testing.T) {
	e := newEcho()
	shipperID := uuid.New()
	h := handler.NewBookingHandler(&MockBookingService{
		CreateFunc: func(_ context.Context, _ service.CreateBookingRequest) (*model.Booking, error) {
			return nil, service.ErrWeightLimitExceeded
		},
	})

	req, rec := makeRequest(http.MethodPost, "/",
		`{"schedule_id":"00000000-0000-0000-0000-000000000001","weight_kg":0,"size_cm":50}`)
	c := e.NewContext(req, rec)
	setUserID(c, shipperID.String())

	if err := h.Create(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "WEIGHT_LIMIT_EXCEEDED" {
		t.Errorf("want WEIGHT_LIMIT_EXCEEDED, got %s", code)
	}
}

func TestBookingCreate_BadRequest_ZeroSize(t *testing.T) {
	e := newEcho()
	shipperID := uuid.New()
	h := handler.NewBookingHandler(&MockBookingService{
		CreateFunc: func(_ context.Context, _ service.CreateBookingRequest) (*model.Booking, error) {
			return nil, service.ErrSizeLimitExceeded
		},
	})

	req, rec := makeRequest(http.MethodPost, "/",
		`{"schedule_id":"00000000-0000-0000-0000-000000000001","weight_kg":5,"size_cm":0}`)
	c := e.NewContext(req, rec)
	setUserID(c, shipperID.String())

	if err := h.Create(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "SIZE_LIMIT_EXCEEDED" {
		t.Errorf("want SIZE_LIMIT_EXCEEDED, got %s", code)
	}
}
