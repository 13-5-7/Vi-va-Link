package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
	"log"

	"github.com/bus-logistics/backend/model"
	"github.com/bus-logistics/backend/repository"
	"github.com/bus-logistics/backend/service"
	"github.com/bus-logistics/backend/utils"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ScheduleHandler struct {
	scheduleService ScheduleServiceInterface
}

func NewScheduleHandler(scheduleService ScheduleServiceInterface) *ScheduleHandler {
	log.Println("-----NewScheduleHandler called-----")
	if scheduleService == nil {
		log.Fatal("scheduleService is required for ScheduleHandler")
	}
	return &ScheduleHandler{scheduleService: scheduleService}
}

type createScheduleRequest struct {
	OriginLat    float64         `json:"origin_lat"`
	OriginLng    float64         `json:"origin_lng"`
	OriginName   string          `json:"origin_name"`
	DestLat      float64         `json:"dest_lat"`
	DestLng      float64         `json:"dest_lng"`
	DestName     string          `json:"dest_name"`
	DepartAt     time.Time       `json:"depart_at"`
	ArriveAt     time.Time       `json:"arrive_at"`
	MaxWeightKg  float64         `json:"max_weight_kg"`
	MaxSizeCm    float64         `json:"max_size_cm"`
	RouteGeoJSON json.RawMessage `json:"route_geojson"`
}

// List returns the operator's own schedules
func (h *ScheduleHandler) List(c echo.Context) error {
	log.Println("----handler List called-----")

	// コンテキストに存在する認証情報を取得
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "missing user_id")
	}
	operatorID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user_id")
	}

	// サービス層を呼び出してスケジュールを取得
	schedules, err := h.scheduleService.ListByOperator(c.Request().Context(), operatorID)
	if err != nil {
		return utils.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}

	// スケジュールをAPIレスポンス用の形式に変換
	result := make([]map[string]any, 0, len(schedules))
	for _, s := range schedules {
		result = append(result, scheduleToMap(s, true))
	}

	// 変換したスケジュールをJSONで返す
	return c.JSON(http.StatusOK, map[string]any{"schedules": result})
}

// Create registers a new schedule for the operator
func (h *ScheduleHandler) Create(c echo.Context) error {
	log.Println("----handler Create called-----")

	// コンテキストに存在する認証情報を取得
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "missing user_id")
	}
	operatorID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user_id")
	}

	var req createScheduleRequest
	// リクエストボディのバインド
	if err := c.Bind(&req); err != nil {
		return utils.NewAppError(http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}
	// バリデーションはサービス層で行う
	schedule, err := h.scheduleService.Create(c.Request().Context(), service.CreateScheduleRequest{
		OperatorID:   operatorID,
		OriginLat:    req.OriginLat,
		OriginLng:    req.OriginLng,
		OriginName:   req.OriginName,
		DestLat:      req.DestLat,
		DestLng:      req.DestLng,
		DestName:     req.DestName,
		DepartAt:     req.DepartAt,
		ArriveAt:     req.ArriveAt,
		MaxWeightKg:  req.MaxWeightKg,
		MaxSizeCm:    req.MaxSizeCm,
		RouteGeoJSON: req.RouteGeoJSON,
	})
	// エラーハンドリング
	// サービス層からのエラーに応じて適切なHTTPステータスコードとエラーメッセージを返す
	if err != nil {
		switch {
		// バリデーションエラーの場合は400 Bad Requestを返す
		case errors.Is(err, service.ErrOriginRequired),
			 errors.Is(err, service.ErrDestRequired),
			 errors.Is(err, service.ErrDepartAtPast),
			 errors.Is(err, service.ErrInvalidMaxWeight),
			 errors.Is(err, service.ErrInvalidMaxSize):
			return utils.NewAppError(http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		// 認証エラーの場合は401 Unauthorizedを返す
		default:
			return utils.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
	}
	return c.JSON(http.StatusCreated, scheduleToMap(*schedule, false))
}

// Search searches schedules by location and time range (for shippers)
func (h *ScheduleHandler) Search(c echo.Context) error {
	log.Println("----handler Search called-----")

	filter := repository.ScheduleFilter{}

	// クエリパラメータを取得し、数値(float64)に変換する
	parseFloat := func(key string) (*float64, error) {
		v := c.QueryParam(key)
		if utils.IsEmpty(v){
			return nil, nil
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, err
		}
		return &f, nil
	}

	// クエリパラメータを取得し、RFC3339形式の日時(time.Time)に変換する
	parseTime := func(key string) (*time.Time, error) {
		v := c.QueryParam(key)
		if utils.IsEmpty(v) {
			return nil, nil
		}
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, err
		}
		return &t, nil
	}

	// クエリパラメータを解析し、filter構造体の各フィールド(float64)に格納する
	floatFields := []struct {
		key string
		dst **float64
	}{
		{"origin_lat_min", &filter.OriginLatMin},
		{"origin_lat_max", &filter.OriginLatMax},
		{"origin_lng_min", &filter.OriginLngMin},
		{"origin_lng_max", &filter.OriginLngMax},
		{"dest_lat_min", &filter.DestLatMin},
		{"dest_lat_max", &filter.DestLatMax},
		{"dest_lng_min", &filter.DestLngMin},
		{"dest_lng_max", &filter.DestLngMax},
	}
	for _, f := range floatFields {
		val, err := parseFloat(f.key)
		if err != nil {
			return utils.NewAppError(http.StatusBadRequest, "BAD_REQUEST", "invalid parameter: "+f.key)
		}
		*f.dst = val
	}

	// クエリパラメータを解析し、filter構造体の各フィールド(time.Time)に格納する
	timeFields := []struct {
		key string
		dst **time.Time
	}{
		{"depart_at_from", &filter.DepartAtFrom},
		{"depart_at_to", &filter.DepartAtTo},
	}
	for _, f := range timeFields {
		val, err := parseTime(f.key)
		if err != nil {
			return utils.NewAppError(http.StatusBadRequest, "BAD_REQUEST", "invalid parameter: "+f.key+" (RFC3339 required)")
		}
		*f.dst = val
	}

	// 
	schedules, err := h.scheduleService.Search(c.Request().Context(), filter)
	if err != nil {
		return utils.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}

	result := make([]map[string]any, 0, len(schedules))
	for _, s := range schedules {
		result = append(result, scheduleToMap(s, true))
	}

	return c.JSON(http.StatusOK, map[string]any{"schedules": result})
}


// UpdateStatus updates the schedule status and cascades to bookings
func (h *ScheduleHandler) UpdateStatus(c echo.Context) error {
	log.Println("----handler UpdateStatus called-----")

	// パスパラメータからスケジュールIDを取得
	idStr := c.Param("id")
	scheduleID, err := uuid.Parse(idStr)
	if err != nil {
		return utils.NewAppError(http.StatusBadRequest, "BAD_REQUEST", "invalid schedule id")
	}

	// コンテキストに存在する認証情報を取得
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "missing user_id")
	}
	operatorID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user_id")
	}

	// リクエストボディから新しいステータスを取得
	var req struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		return utils.NewAppError(http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	// バリデーション: statusは'full', 'departed', 'arrived'のいずれかでなければならない
	newStatus := model.ScheduleStatus(req.Status)
	if newStatus != model.ScheduleStatusFull && newStatus != model.ScheduleStatusDeparted && newStatus != model.ScheduleStatusArrived {
		return utils.NewAppError(http.StatusBadRequest, "VALIDATION_ERROR", "status must be 'full', 'departed' or 'arrived'")
	}

	if err := h.scheduleService.UpdateScheduleStatus(c.Request().Context(), scheduleID, newStatus, operatorID); err != nil {
		switch {
		// スケジュールが見つからない場合は404 Not Foundを返す
		case errors.Is(err, service.ErrInvalidScheduleTransition):
			return utils.NewAppError(http.StatusBadRequest, "VALIDATION_ERROR", err.Error())

		// スケジュールが見つからない場合は404 Not Foundを返す
		case errors.Is(err, service.ErrScheduleNotFound):
			return utils.NewAppError(http.StatusNotFound, "NOT_FOUND", "schedule not found")
		// 認証エラーの場合は401 Unauthorizedを返す
		default:
			return utils.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
	}

	// 成功した場合は新しいステータスを返す
	return c.JSON(http.StatusOK, map[string]any{"status": req.Status})
}

// GetByID はパスパラメータのIDに該当するスケジュールを1件取得し、JSONで返却
func (h *ScheduleHandler) GetByID(c echo.Context) error {
	log.Println("----handler GetByID called-----")

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		// パース失敗 = 不正なリクエスト
		return utils.NewAppError(http.StatusBadRequest, "BAD_REQUEST", "invalid schedule id")
	}
	
	schedule, err := h.scheduleService.GetByID(c.Request().Context(), id)
	if err != nil {
		// 内部エラーはログ等に詳細は残すが、クライアントには汎用的なメッセージを返す
		return utils.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
	if schedule == nil {
		return utils.NewAppError(http.StatusNotFound, "NOT_FOUND", "schedule not found")
	}

	return c.JSON(http.StatusOK, scheduleToMap(*schedule, true))
}

// Delete removes a schedule (only if no bookings and not departed)
func (h *ScheduleHandler) Delete(c echo.Context) error {
	log.Println("----handler Delete called-----")

	// パスパラメータからスケジュールIDを取得
	idStr := c.Param("id")
	scheduleID, err := uuid.Parse(idStr)
	if err != nil {
		return utils.NewAppError(http.StatusBadRequest, "BAD_REQUEST", "invalid schedule id")
	}

	// コンテキストに存在する認証情報を取得
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "missing user_id")
	}
	operatorID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user_id")
	}

	// サービス層を呼び出してスケジュールを削除
	if err := h.scheduleService.Delete(c.Request().Context(), scheduleID, operatorID); err != nil {
		switch {
		// スケジュールが見つからない場合は404 Not Foundを返す
		case errors.Is(err, service.ErrScheduleNotFound):
			return utils.NewAppError(http.StatusNotFound, "NOT_FOUND", "schedule not found")
		// スケジュールに予約が存在する場合は409 Conflictを返す
		case errors.Is(err, service.ErrScheduleHasBookings):
			return utils.NewAppError(http.StatusConflict, "CONFLICT", "予約が存在するスケジュールは削除できません")
		// スケジュールがすでに出発済みの場合は409 Conflictを返す
		case errors.Is(err, service.ErrInvalidScheduleTransition):
			return utils.NewAppError(http.StatusConflict, "CONFLICT", "出発済みのスケジュールは削除できません")
		default:
			// その他のエラーは500 Internal Server Errorを返す
			return utils.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

// Cancel cancels a schedule (only open/full, cascades to accepted bookings)
func (h *ScheduleHandler) Cancel(c echo.Context) error {
	log.Println("----handler Cancel called-----")

	// パスパラメータからスケジュールIDを取得
	idStr := c.Param("id")
	scheduleID, err := uuid.Parse(idStr)
	if err != nil {
		return utils.NewAppError(http.StatusBadRequest, "BAD_REQUEST", "invalid schedule id")
	}

	// コンテキストに存在する認証情報を取得
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "missing user_id")
	}
	operatorID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user_id")
	}

	// サービス層を呼び出してスケジュールをキャンセル
	if err := h.scheduleService.Cancel(c.Request().Context(), scheduleID, operatorID); err != nil {
		switch {
		// スケジュールが見つからない場合は404 Not Foundを返す
		case errors.Is(err, service.ErrScheduleNotFound):
			return utils.NewAppError(http.StatusNotFound, "NOT_FOUND", "schedule not found")
		// スケジュールがすでにキャンセル済みの場合は409 Conflictを返す
		case errors.Is(err, service.ErrScheduleAlreadyCancelled):
			return utils.NewAppError(http.StatusConflict, "CONFLICT", "スケジュールはすでにキャンセル済みです")
		// スケジュールに予約が存在する場合は409 Conflictを返す
		case errors.Is(err, service.ErrInvalidScheduleTransition):
			return utils.NewAppError(http.StatusConflict, "CONFLICT", "出発済み・到着済みのスケジュールはキャンセルできません")
		default:
			// その他のエラーは500 Internal Server Errorを返す
			return utils.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "cancelled"})
}
