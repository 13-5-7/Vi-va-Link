package repository

import (
	"context"
	//"errors"
	"log"

	"github.com/bus-logistics/backend/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingRepository struct {
	pool *pgxpool.Pool
}

func NewBookingRepository(pool *pgxpool.Pool) *BookingRepository {
	return &BookingRepository{pool: pool}
}

const bookingColumns = `id, schedule_id, shipper_id, tracking_number,
	weight_kg, size_cm, content_desc,
	recipient_name, recipient_phone, recipient_addr,
	status, status_updated_at, created_at`

func scanBooking(row pgx.Row) (*model.Booking, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// Create はトランザクション内でbookingをINSERTする
func (r *BookingRepository) Create(ctx context.Context, tx pgx.Tx, booking *model.Booking) (*model.Booking, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// FindByID はIDでbookingを検索する
func (r *BookingRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// FindByTrackingNumber はトラッキング番号でbookingを検索する
func (r *BookingRepository) FindByTrackingNumber(ctx context.Context, trackingNumber string) (*model.Booking, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// ListByShipper は荷主IDに紐づくbooking一覧を返す
func (r *BookingRepository) ListByShipper(ctx context.Context, shipperID uuid.UUID) ([]model.Booking, error) {
	log.Println("----repository ListByShipper called-----")

	// shipperIDをキーに予約一覧（降順）を取得する
	rows, err := r.pool.Query(ctx,
		`SELECT `+bookingColumns+` FROM bookings WHERE shipper_id = $1 ORDER BY created_at DESC`,
		shipperID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 取得したデータをマッピング
	var bookings []model.Booking
	for rows.Next() {
		var b model.Booking
		err := rows.Scan(
			&b.ID, &b.ScheduleID, &b.ShipperID, &b.TrackingNumber,
			&b.WeightKg, &b.SizeCm, &b.ContentDesc,
			&b.RecipientName, &b.RecipientPhone, &b.RecipientAddr,
			&b.Status, &b.StatusUpdatedAt, &b.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, rows.Err()
}

// UpdateStatus はトランザクション内でbookingのステータスをUPDATEする
func (r *BookingRepository) UpdateStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status model.BookingStatus) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// ListBySchedule はスケジュールIDに紐づくbooking一覧を返す
func (r *BookingRepository) ListBySchedule(ctx context.Context, scheduleID uuid.UUID) ([]model.Booking, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// UpdateStatusDirect はトランザクションなしでbookingのステータスを直接UPDATEする（スケジュール連動用）
func (r *BookingRepository) UpdateStatusDirect(ctx context.Context, id uuid.UUID, status model.BookingStatus) error {
	log.Println("----repository UpdateStatusDirect called-----")

	// データベースのbookingのstatusを更新する
	_, err := r.pool.Exec(ctx,
		`UPDATE bookings SET status = $1, status_updated_at = NOW() WHERE id = $2`,
		status, id,
	)
	return err
}
