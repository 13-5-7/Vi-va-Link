package repository

import (
	"context"

	"github.com/bus-logistics/backend/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TrackingRepository struct {
	pool *pgxpool.Pool
}

func NewTrackingRepository(pool *pgxpool.Pool) *TrackingRepository {
	return &TrackingRepository{pool: pool}
}

// InsertStatusLog はステータス変更ログを booking_status_logs に挿入する
func (r *TrackingRepository) InsertStatusLog(ctx context.Context, tx pgx.Tx, bookingID uuid.UUID, oldStatus, newStatus model.BookingStatus, changedBy uuid.UUID) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}
