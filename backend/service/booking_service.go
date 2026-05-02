package service

import (
	"context"
	"errors"
	"log"
	//"fmt"

	"github.com/bus-logistics/backend/model"
	"github.com/bus-logistics/backend/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	return s.bookingRepo.ListByShipper(ctx, shipperID)
}

func (s *BookingService) GetByID(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// Cancel はステータスが accepted の予約をキャンセルし、スケジュールの残余重量を回復する
// ステータスが 'accepted' の場合のみ実行可能
func (s *BookingService) Cancel(ctx context.Context, bookingID uuid.UUID, shipperID uuid.UUID) error {
	log.Println("----service Cancel called-----")

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	// エラー発生時はロールバック、正常時は Commit 後に呼ばれても影響なし
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	var dbShipperID uuid.UUID
	var status model.BookingStatus
	var weightKg float64
	var scheduleID uuid.UUID

	// 更新対象をロックして取得 (Race Condition 防止)
	err = tx.QueryRow(ctx,
		`SELECT shipper_id, status, weight_kg, schedule_id FROM bookings WHERE id = $1 FOR UPDATE`,
		bookingID,
	).Scan(&dbShipperID, &status, &weightKg, &scheduleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = ErrBookingNotFound
		}
		return err
	}

	// 所有権とステータスのチェック
	if dbShipperID != shipperID {
		err = ErrForbidden
		return err
	}

	if status != model.BookingStatusAccepted {
		err = ErrCannotCancel
		return err
	}

	// 予約ステータスの更新
	_, err = tx.Exec(ctx,
		`UPDATE bookings SET status = 'cancelled', status_updated_at = NOW() WHERE id = $1`,
		bookingID,
	)
	if err != nil {
		return err
	}

	// スケジュールの在庫（可能重量）を戻す
	_, err = tx.Exec(ctx,
		`UPDATE schedules SET avail_weight_kg = avail_weight_kg + $1 WHERE id = $2`,
		weightKg, scheduleID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
