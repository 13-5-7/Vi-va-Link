package handler

import (
	"errors"
	"net/http"
	"log"

	"github.com/bus-logistics/backend/service"
	"github.com/bus-logistics/backend/utils"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type BookingHandler struct {
	bookingService BookingServiceInterface
}

func NewBookingHandler(bookingService BookingServiceInterface) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

type createBookingRequest struct {
	ScheduleID     uuid.UUID `json:"schedule_id"`
	WeightKg       float64   `json:"weight_kg"`
	SizeCm         float64   `json:"size_cm"`
	ContentDesc    string    `json:"content_desc"`
	RecipientName  string    `json:"recipient_name"`
	RecipientPhone string    `json:"recipient_phone"`
	RecipientAddr  string    `json:"recipient_addr"`
}

// List returns the shipper's own bookings
func (h *BookingHandler) List(c echo.Context) error {
	log.Println("----handler List called-----")

	// コンテキストに存在する認証情報を取得
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "missing user_id")
	}
	shipperID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user_id")
	}

	// サービス層を呼び出してスケジュール一覧を取得
	bookings, err := h.bookingService.ListByShipper(c.Request().Context(), shipperID)
	if err != nil {
		return utils.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}

	// ドメインモデルをそのまま返さず、APIの仕様に合わせたマップに変換
	result := make([]map[string]any, 0, len(bookings))
	for _, b := range bookings {
		result = append(result, map[string]any{
			"id":               b.ID,
			"schedule_id":      b.ScheduleID,
			"shipper_id":       b.ShipperID,
			"tracking_number":  b.TrackingNumber,
			"weight_kg":        b.WeightKg,
			"size_cm":          b.SizeCm,
			"content_desc":     b.ContentDesc,
			"recipient_name":   b.RecipientName,
			"recipient_phone":  b.RecipientPhone,
			"recipient_addr":   b.RecipientAddr,
			"status":           b.Status,
			"status_updated_at":b.StatusUpdatedAt,
			"created_at":       b.CreatedAt,
		})
	}

	// 変換したスケジュールをJSONで返す
	return c.JSON(http.StatusOK, map[string]any{"bookings": result})
}

// Create 新規予約を作成し、結果をJSON形式で返す
// POST /api/vi/bookings
func (h *BookingHandler) Create(c echo.Context) error {
	log.Println("----handler/booking.go Create called-----")

	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "missing user_id")
	}
	shipperID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user_id")
	}

	var req createBookingRequest
	if err := c.Bind(&req); err != nil {
		return utils.NewAppError(http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	booking, err := h.bookingService.Create(c.Request().Context(), service.CreateBookingRequest{
		ScheduleID:     req.ScheduleID,
		ShipperID:      shipperID,
		WeightKg:       req.WeightKg,
		SizeCm:         req.SizeCm,
		ContentDesc:    req.ContentDesc,
		RecipientName:  req.RecipientName,
		RecipientPhone: req.RecipientPhone,
		RecipientAddr:  req.RecipientAddr,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCapacityExceeded):
			return utils.NewAppError(http.StatusConflict, "CAPACITY_EXCEEDED", err.Error())
		case errors.Is(err, service.ErrSizeExceeded):
			return utils.NewAppError(http.StatusConflict, "SIZE_EXCEEDED", err.Error())
		case errors.Is(err, service.ErrWeightLimitExceeded):
			return utils.NewAppError(http.StatusBadRequest, "WEIGHT_LIMIT_EXCEEDED", "1個あたりの重量は10kg以下にしてください")
		case errors.Is(err, service.ErrSizeLimitExceeded):
			return utils.NewAppError(http.StatusBadRequest, "SIZE_LIMIT_EXCEEDED", "3辺合計は140cm以下にしてください")
		case errors.Is(err, service.ErrScheduleNotFound):
			return utils.NewAppError(http.StatusNotFound, "NOT_FOUND", err.Error())
		default:
			return utils.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"id":               booking.ID,
		"schedule_id":      booking.ScheduleID,
		"shipper_id":       booking.ShipperID,
		"tracking_number":  booking.TrackingNumber,
		"weight_kg":        booking.WeightKg,
		"size_cm":          booking.SizeCm,
		"content_desc":     booking.ContentDesc,
		"recipient_name":   booking.RecipientName,
		"recipient_phone":  booking.RecipientPhone,
		"recipient_addr":   booking.RecipientAddr,
		"status":           booking.Status,
		"status_updated_at": booking.StatusUpdatedAt,
		"created_at":       booking.CreatedAt,
	})
}

// GetByID returns a single booking by ID
func (h *BookingHandler) GetByID(c echo.Context) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// Cancel は予約をキャンセルする（ステータスが accepted の場合のみ可能）
func (h *BookingHandler) Cancel(c echo.Context) error {
	log.Println("----handler Cancel called-----")

	idStr := c.Param("id")
	bookingID, err := uuid.Parse(idStr)
	if err != nil {
		// パース失敗 = 不正なリクエスト
		return utils.NewAppError(http.StatusBadRequest, "BAD_REQUEST", "invalid booking id")
	}

	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "missing user_id")
	}
	shipperID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user_id")
	}

	if err := h.bookingService.Cancel(c.Request().Context(), bookingID, shipperID); err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			return utils.NewAppError(http.StatusNotFound, "NOT_FOUND", "booking not found")
		case errors.Is(err, service.ErrForbidden):
			return utils.NewAppError(http.StatusForbidden, "FORBIDDEN", "you are not allowed to cancel this booking")
		// ステータス不整合によるエラーは 409 Conflict を返す
		case errors.Is(err, service.ErrCannotCancel):
			return utils.NewAppError(http.StatusConflict, "CANNOT_CANCEL", "booking cannot be cancelled in current status")
		default:
			return utils.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "cancelled"})
}
