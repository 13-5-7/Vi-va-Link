package service

import (
	"context"
	"errors"
	//"fmt"

	"github.com/bus-logistics/backend/model"
	"github.com/bus-logistics/backend/repository"
	"github.com/google/uuid"
	//"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrCapacityExceeded    = errors.New("weight capacity exceeded")
	ErrSizeExceeded        = errors.New("size limit exceeded")
	ErrScheduleNotFound    = errors.New("schedule not found")
	ErrWeightLimitExceeded = errors.New("weight exceeds system limit of 10kg per item")
	ErrSizeLimitExceeded   = errors.New("size exceeds system limit of 140cm (3-side total)")
	ErrCannotCancel        = errors.New("booking cannot be cancelled in current status")
	ErrForbidden           = errors.New("forbidden")
)

// システム制限値
const (
	MaxWeightKgPerItem = 10.0  // 1個あたり最大重量 (kg)
	MaxSizeCmPerItem   = 140.0 // 1個あたり最大サイズ・3辺合計 (cm)
)

type CreateBookingRequest struct {
	ScheduleID     uuid.UUID
	ShipperID      uuid.UUID
	WeightKg       float64
	SizeCm         float64
	ContentDesc    string
	RecipientName  string
	RecipientPhone string
	RecipientAddr  string
}

type BookingService struct {
	pool        *pgxpool.Pool
	bookingRepo *repository.BookingRepository
}

func NewBookingService(pool *pgxpool.Pool, bookingRepo *repository.BookingRepository) *BookingService {
	return &BookingService{pool: pool, bookingRepo: bookingRepo}
}

func (s *BookingService) Create(ctx context.Context, req CreateBookingRequest) (*model.Booking, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

func (s *BookingService) ListByShipper(ctx context.Context, shipperID uuid.UUID) ([]model.Booking, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

func (s *BookingService) GetByID(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// Cancel はステータスが accepted の予約をキャンセルし、スケジュールの残余重量を回復する
func (s *BookingService) Cancel(ctx context.Context, bookingID uuid.UUID, shipperID uuid.UUID) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}
