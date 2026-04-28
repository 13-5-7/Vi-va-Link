package handler_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/bus-logistics/backend/handler"
	"github.com/bus-logistics/backend/model"
	"github.com/bus-logistics/backend/repository"
	"github.com/bus-logistics/backend/service"
	"github.com/google/uuid"
)

func newSchedule(id, operatorID uuid.UUID) model.Schedule {
	return model.Schedule{
		ID:            id,
		OperatorID:    operatorID,
		OriginName:    "Tokyo Station",
		DestName:      "Osaka Station",
		DepartAt:      time.Now().Add(24 * time.Hour),
		ArriveAt:      time.Now().Add(30 * time.Hour),
		MaxWeightKg:   100,
		MaxSizeCm:     200,
		AvailWeightKg: 100,
		Status:        model.ScheduleStatusOpen,
		CreatedAt:     time.Now(),
	}
}

// ---- List ----

func TestScheduleList_Unauthorized_MissingUserID(t *testing.T) {
	e := newEcho()
	h := handler.NewScheduleHandler(&MockScheduleService{})

	req, rec := makeRequest(http.MethodGet, "/", "")
	c := e.NewContext(req, rec)

	if err := h.List(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rec.Code)
	}
}

func TestScheduleList_Unauthorized_InvalidUserID(t *testing.T) {
	e := newEcho()
	h := handler.NewScheduleHandler(&MockScheduleService{})

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

func TestScheduleList_Success_Empty(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	h := handler.NewScheduleHandler(&MockScheduleService{
		ListByOperatorFunc: func(_ context.Context, id uuid.UUID) ([]model.Schedule, error) {
			if id != operatorID {
				t.Errorf("unexpected operatorID: %v", id)
			}
			return []model.Schedule{}, nil
		},
	})

	req, rec := makeRequest(http.MethodGet, "/", "")
	c := e.NewContext(req, rec)
	setUserID(c, operatorID.String())

	if err := h.List(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
	body := decodeBody(t, rec)
	schedules, ok := body["schedules"].([]any)
	if !ok {
		t.Fatal("response must have 'schedules' array")
	}
	if len(schedules) != 0 {
		t.Errorf("want 0 schedules, got %d", len(schedules))
	}
}

func TestScheduleList_Success_WithSchedules(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	scheduleID := uuid.New()
	h := handler.NewScheduleHandler(&MockScheduleService{
		ListByOperatorFunc: func(_ context.Context, _ uuid.UUID) ([]model.Schedule, error) {
			return []model.Schedule{newSchedule(scheduleID, operatorID)}, nil
		},
	})

	req, rec := makeRequest(http.MethodGet, "/", "")
	c := e.NewContext(req, rec)
	setUserID(c, operatorID.String())

	if err := h.List(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
	body := decodeBody(t, rec)
	schedules := body["schedules"].([]any)
	if len(schedules) != 1 {
		t.Errorf("want 1 schedule, got %d", len(schedules))
	}
}

// ---- Create ----

func TestScheduleCreate_Unauthorized_MissingUserID(t *testing.T) {
	e := newEcho()
	h := handler.NewScheduleHandler(&MockScheduleService{})

	req, rec := makeRequest(http.MethodPost, "/", `{"origin_lat":35.0,"origin_lng":139.0}`)
	c := e.NewContext(req, rec)

	if err := h.Create(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rec.Code)
	}
}

func TestScheduleCreate_ValidationError_OriginRequired(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	h := handler.NewScheduleHandler(&MockScheduleService{
		CreateFunc: func(_ context.Context, _ service.CreateScheduleRequest) (*model.Schedule, error) {
			return nil, service.ErrOriginRequired
		},
	})

	req, rec := makeRequest(http.MethodPost, "/", `{"origin_lat":0,"origin_lng":0,"dest_lat":34.0,"dest_lng":135.0,"depart_at":"2099-01-01T00:00:00Z","arrive_at":"2099-01-01T06:00:00Z","max_weight_kg":100}`)
	c := e.NewContext(req, rec)
	setUserID(c, operatorID.String())

	if err := h.Create(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "VALIDATION_ERROR" {
		t.Errorf("want VALIDATION_ERROR, got %s", code)
	}
}

func TestScheduleCreate_ValidationError_DestRequired(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	h := handler.NewScheduleHandler(&MockScheduleService{
		CreateFunc: func(_ context.Context, _ service.CreateScheduleRequest) (*model.Schedule, error) {
			return nil, service.ErrDestRequired
		},
	})

	req, rec := makeRequest(http.MethodPost, "/", `{"origin_lat":35.0,"origin_lng":139.0,"dest_lat":0,"dest_lng":0,"depart_at":"2099-01-01T00:00:00Z","arrive_at":"2099-01-01T06:00:00Z","max_weight_kg":100}`)
	c := e.NewContext(req, rec)
	setUserID(c, operatorID.String())

	if err := h.Create(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "VALIDATION_ERROR" {
		t.Errorf("want VALIDATION_ERROR, got %s", code)
	}
}

func TestScheduleCreate_ValidationError_DepartAtPast(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	h := handler.NewScheduleHandler(&MockScheduleService{
		CreateFunc: func(_ context.Context, _ service.CreateScheduleRequest) (*model.Schedule, error) {
			return nil, service.ErrDepartAtPast
		},
	})

	req, rec := makeRequest(http.MethodPost, "/", `{"origin_lat":35.0,"origin_lng":139.0,"dest_lat":34.0,"dest_lng":135.0,"depart_at":"2000-01-01T00:00:00Z","arrive_at":"2000-01-01T06:00:00Z","max_weight_kg":100}`)
	c := e.NewContext(req, rec)
	setUserID(c, operatorID.String())

	if err := h.Create(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "VALIDATION_ERROR" {
		t.Errorf("want VALIDATION_ERROR, got %s", code)
	}
}

func TestScheduleCreate_Success(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	scheduleID := uuid.New()
	h := handler.NewScheduleHandler(&MockScheduleService{
		CreateFunc: func(_ context.Context, req service.CreateScheduleRequest) (*model.Schedule, error) {
			s := newSchedule(scheduleID, req.OperatorID)
			s.OriginName = req.OriginName
			s.DestName = req.DestName
			return &s, nil
		},
	})

	body := `{"origin_lat":35.68,"origin_lng":139.76,"origin_name":"Tokyo","dest_lat":34.69,"dest_lng":135.50,"dest_name":"Osaka","depart_at":"2099-06-01T10:00:00Z","arrive_at":"2099-06-01T16:00:00Z","max_weight_kg":100,"max_size_cm":200}`
	req, rec := makeRequest(http.MethodPost, "/", body)
	c := e.NewContext(req, rec)
	setUserID(c, operatorID.String())

	if err := h.Create(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("want 201, got %d\nbody: %s", rec.Code, rec.Body.String())
	}
	respBody := decodeBody(t, rec)
	if respBody["origin_name"] != "Tokyo" {
		t.Errorf("want origin_name=Tokyo, got %v", respBody["origin_name"])
	}
}

// ---- Search ----

func TestScheduleSearch_Success_Empty(t *testing.T) {
	e := newEcho()
	h := handler.NewScheduleHandler(&MockScheduleService{
		SearchFunc: func(_ context.Context, _ repository.ScheduleFilter) ([]model.Schedule, error) {
			return []model.Schedule{}, nil
		},
	})

	req, rec := makeRequest(http.MethodGet, "/?origin_lat_min=34.0&origin_lat_max=36.0", "")
	c := e.NewContext(req, rec)

	if err := h.Search(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
	body := decodeBody(t, rec)
	schedules, ok := body["schedules"].([]any)
	if !ok {
		t.Fatal("response must have 'schedules' array")
	}
	if len(schedules) != 0 {
		t.Errorf("want 0 schedules, got %d", len(schedules))
	}
}

func TestScheduleSearch_BadRequest_InvalidParam(t *testing.T) {
	e := newEcho()
	h := handler.NewScheduleHandler(&MockScheduleService{})

	req, rec := makeRequest(http.MethodGet, "/?origin_lat_min=not-a-float", "")
	c := e.NewContext(req, rec)

	if err := h.Search(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
}

func TestScheduleSearch_BadRequest_InvalidTimeParam(t *testing.T) {
	e := newEcho()
	h := handler.NewScheduleHandler(&MockScheduleService{})

	req, rec := makeRequest(http.MethodGet, "/?depart_at_from=not-a-time", "")
	c := e.NewContext(req, rec)

	if err := h.Search(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
}

// ---- GetByID ----

func TestScheduleGetByID_BadRequest_InvalidID(t *testing.T) {
	e := newEcho()
	h := handler.NewScheduleHandler(&MockScheduleService{})

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

func TestScheduleGetByID_NotFound(t *testing.T) {
	e := newEcho()
	h := handler.NewScheduleHandler(&MockScheduleService{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
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

func TestScheduleGetByID_Success(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	scheduleID := uuid.New()
	h := handler.NewScheduleHandler(&MockScheduleService{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*model.Schedule, error) {
			s := newSchedule(id, operatorID)
			return &s, nil
		},
	})

	req, rec := makeRequest(http.MethodGet, "/", "")
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(scheduleID.String())

	if err := h.GetByID(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
	body := decodeBody(t, rec)
	if body["origin_name"] != "Tokyo Station" {
		t.Errorf("want origin_name=Tokyo Station, got %v", body["origin_name"])
	}
}

// ---- UpdateStatus ----

func TestScheduleUpdateStatus_BadRequest_InvalidID(t *testing.T) {
	e := newEcho()
	h := handler.NewScheduleHandler(&MockScheduleService{})

	req, rec := makeRequest(http.MethodPatch, "/", `{"status":"full"}`)
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

func TestScheduleUpdateStatus_Unauthorized_MissingUserID(t *testing.T) {
	e := newEcho()
	h := handler.NewScheduleHandler(&MockScheduleService{})

	req, rec := makeRequest(http.MethodPatch, "/", `{"status":"full"}`)
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	if err := h.UpdateStatus(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rec.Code)
	}
}

func TestScheduleUpdateStatus_BadRequest_InvalidStatus(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	h := handler.NewScheduleHandler(&MockScheduleService{})

	req, rec := makeRequest(http.MethodPatch, "/", `{"status":"invalid_status"}`)
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

func TestScheduleUpdateStatus_NotFound(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	h := handler.NewScheduleHandler(&MockScheduleService{
		UpdateScheduleStatusFunc: func(_ context.Context, _ uuid.UUID, _ model.ScheduleStatus, _ uuid.UUID) error {
			return service.ErrScheduleNotFound
		},
	})

	req, rec := makeRequest(http.MethodPatch, "/", `{"status":"full"}`)
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

func TestScheduleUpdateStatus_BadRequest_InvalidTransition(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	h := handler.NewScheduleHandler(&MockScheduleService{
		UpdateScheduleStatusFunc: func(_ context.Context, _ uuid.UUID, _ model.ScheduleStatus, _ uuid.UUID) error {
			return service.ErrInvalidScheduleTransition
		},
	})

	req, rec := makeRequest(http.MethodPatch, "/", `{"status":"full"}`)
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

func TestScheduleUpdateStatus_Success(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	scheduleID := uuid.New()
	var capturedStatus model.ScheduleStatus

	h := handler.NewScheduleHandler(&MockScheduleService{
		UpdateScheduleStatusFunc: func(_ context.Context, id uuid.UUID, status model.ScheduleStatus, opID uuid.UUID) error {
			if id != scheduleID {
				t.Errorf("unexpected scheduleID: %v", id)
			}
			if opID != operatorID {
				t.Errorf("unexpected operatorID: %v", opID)
			}
			capturedStatus = status
			return nil
		},
	})

	req, rec := makeRequest(http.MethodPatch, "/", `{"status":"departed"}`)
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(scheduleID.String())
	setUserID(c, operatorID.String())

	if err := h.UpdateStatus(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
	if capturedStatus != model.ScheduleStatusDeparted {
		t.Errorf("want departed, got %v", capturedStatus)
	}
}

// ---- Delete ----

func TestScheduleDelete_BadRequest_InvalidID(t *testing.T) {
	e := newEcho()
	h := handler.NewScheduleHandler(&MockScheduleService{})

	req, rec := makeRequest(http.MethodDelete, "/", "")
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	if err := h.Delete(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
}

func TestScheduleDelete_Unauthorized_MissingUserID(t *testing.T) {
	e := newEcho()
	h := handler.NewScheduleHandler(&MockScheduleService{})

	req, rec := makeRequest(http.MethodDelete, "/", "")
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	if err := h.Delete(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rec.Code)
	}
}

func TestScheduleDelete_NotFound(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	h := handler.NewScheduleHandler(&MockScheduleService{
		DeleteFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
			return service.ErrScheduleNotFound
		},
	})

	req, rec := makeRequest(http.MethodDelete, "/", "")
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	setUserID(c, operatorID.String())

	if err := h.Delete(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", rec.Code)
	}
}

func TestScheduleDelete_Conflict_HasBookings(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	h := handler.NewScheduleHandler(&MockScheduleService{
		DeleteFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
			return service.ErrScheduleHasBookings
		},
	})

	req, rec := makeRequest(http.MethodDelete, "/", "")
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	setUserID(c, operatorID.String())

	if err := h.Delete(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("want 409, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "HAS_BOOKINGS" {
		t.Errorf("want HAS_BOOKINGS, got %s", code)
	}
}

func TestScheduleDelete_Conflict_InvalidStatus(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	h := handler.NewScheduleHandler(&MockScheduleService{
		DeleteFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
			return service.ErrInvalidScheduleTransition
		},
	})

	req, rec := makeRequest(http.MethodDelete, "/", "")
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	setUserID(c, operatorID.String())

	if err := h.Delete(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("want 409, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "INVALID_STATUS" {
		t.Errorf("want INVALID_STATUS, got %s", code)
	}
}

func TestScheduleDelete_Success(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	scheduleID := uuid.New()
	var capturedScheduleID, capturedOperatorID uuid.UUID

	h := handler.NewScheduleHandler(&MockScheduleService{
		DeleteFunc: func(_ context.Context, sID uuid.UUID, oID uuid.UUID) error {
			capturedScheduleID = sID
			capturedOperatorID = oID
			return nil
		},
	})

	req, rec := makeRequest(http.MethodDelete, "/", "")
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(scheduleID.String())
	setUserID(c, operatorID.String())

	if err := h.Delete(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d\nbody: %s", rec.Code, rec.Body.String())
	}
	if capturedScheduleID != scheduleID {
		t.Errorf("want scheduleID %v, got %v", scheduleID, capturedScheduleID)
	}
	if capturedOperatorID != operatorID {
		t.Errorf("want operatorID %v, got %v", operatorID, capturedOperatorID)
	}
}

// ---- Create: max_weight_kg / max_size_cm バリデーション ----

func TestScheduleCreate_ValidationError_ZeroMaxWeight(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	h := handler.NewScheduleHandler(&MockScheduleService{
		CreateFunc: func(_ context.Context, _ service.CreateScheduleRequest) (*model.Schedule, error) {
			return nil, service.ErrInvalidMaxWeight
		},
	})

	body := `{"origin_lat":35.68,"origin_lng":139.76,"dest_lat":34.69,"dest_lng":135.50,"depart_at":"2099-01-01T00:00:00Z","arrive_at":"2099-01-01T06:00:00Z","max_weight_kg":0,"max_size_cm":140}`
	req, rec := makeRequest(http.MethodPost, "/", body)
	c := e.NewContext(req, rec)
	setUserID(c, operatorID.String())

	if err := h.Create(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "VALIDATION_ERROR" {
		t.Errorf("want VALIDATION_ERROR, got %s", code)
	}
}

func TestScheduleCreate_ValidationError_ZeroMaxSize(t *testing.T) {
	e := newEcho()
	operatorID := uuid.New()
	h := handler.NewScheduleHandler(&MockScheduleService{
		CreateFunc: func(_ context.Context, _ service.CreateScheduleRequest) (*model.Schedule, error) {
			return nil, service.ErrInvalidMaxSize
		},
	})

	body := `{"origin_lat":35.68,"origin_lng":139.76,"dest_lat":34.69,"dest_lng":135.50,"depart_at":"2099-01-01T00:00:00Z","arrive_at":"2099-01-01T06:00:00Z","max_weight_kg":100,"max_size_cm":0}`
	req, rec := makeRequest(http.MethodPost, "/", body)
	c := e.NewContext(req, rec)
	setUserID(c, operatorID.String())

	if err := h.Create(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	if code := errCode(t, decodeBody(t, rec)); code != "VALIDATION_ERROR" {
		t.Errorf("want VALIDATION_ERROR, got %s", code)
	}
}
