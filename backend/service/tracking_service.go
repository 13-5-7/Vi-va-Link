package service

import (
	"context"
	"errors"
	"log"

	"github.com/bus-logistics/backend/model"
	"github.com/bus-logistics/backend/repository"
	"github.com/bus-logistics/backend/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrBookingNotFound         = errors.New("booking not found")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

type TrackingInfo struct {
	Booking  *model.Booking
	Schedule *model.Schedule
}

type TrackingService struct {
	pool         *pgxpool.Pool
	bookingRepo  *repository.BookingRepository
	scheduleRepo *repository.ScheduleRepository
	trackingRepo *repository.TrackingRepository
}

func NewTrackingService(
	pool *pgxpool.Pool,
	bookingRepo *repository.BookingRepository,
	scheduleRepo *repository.ScheduleRepository,
	trackingRepo *repository.TrackingRepository,
) *TrackingService {
	return &TrackingService{
		pool:         pool,
		bookingRepo:  bookingRepo,
		scheduleRepo: scheduleRepo,
		trackingRepo: trackingRepo,
	}
}

// GetByTrackingNumber は tracking_number で予約情報とスケジュール情報を返す
func (s *TrackingService) GetByTrackingNumber(ctx context.Context, trackingNumber string) (*TrackingInfo, error) {
	log.Println("----service GetByTrackingNumber called-----")

	booking, err := s.bookingRepo.FindByTrackingNumber(ctx, trackingNumber)
	if err != nil {
		return nil, err
	}
	if utils.IsEmpty(booking) {
		return nil, ErrBookingNotFound
	}

	schedule, err := s.scheduleRepo.FindByID(ctx, booking.ScheduleID)
	if err != nil {
		return nil, err
	}

	return &TrackingInfo{
		Booking:  booking,
		Schedule: schedule,
	}, nil
}

// UpdateStatus はステータスを更新する（前方向遷移のみ許可）
func (s *TrackingService) UpdateStatus(ctx context.Context, bookingID uuid.UUID, newStatus model.BookingStatus, operatorID uuid.UUID) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}
