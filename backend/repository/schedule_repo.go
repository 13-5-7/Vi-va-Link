package repository

import (
	"context"
	"errors"
	"fmt"
	"time"
	"log"

	"github.com/bus-logistics/backend/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScheduleFilter struct {
	OriginLatMin *float64
	OriginLatMax *float64
	OriginLngMin *float64
	OriginLngMax *float64
	DestLatMin   *float64
	DestLatMax   *float64
	DestLngMin   *float64
	DestLngMax   *float64
	DepartAtFrom *time.Time
	DepartAtTo   *time.Time
}

type ScheduleRepository struct {
	pool *pgxpool.Pool
}

func NewScheduleRepository(pool *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{pool: pool}
}

func scanSchedule(row pgx.Row) (*model.Schedule, error) {
	log.Println("----repository scanSchedule called-----")

	var s model.Schedule
	// データベースからスケジュールの情報を読み取る
	err := row.Scan(
		&s.ID, &s.OperatorID,
		&s.OriginLat, &s.OriginLng, &s.OriginName,
		&s.DestLat, &s.DestLng, &s.DestName,
		&s.DepartAt, &s.ArriveAt,
		&s.MaxWeightKg, &s.MaxSizeCm, &s.AvailWeightKg,
		&s.Status, &s.RouteGeoJSON, &s.CreatedAt,
	)
	// エラーが発生した場合はエラーを返す
	if err != nil {
		return nil, err
	}
	return &s, nil
}

const scheduleColumns = `id, operator_id,
	origin_lat, origin_lng, origin_name,
	dest_lat, dest_lng, dest_name,
	depart_at, arrive_at,
	max_weight_kg, max_size_cm, avail_weight_kg,
	status, route_geojson, created_at`

func (r *ScheduleRepository) Create(ctx context.Context, s *model.Schedule) (*model.Schedule, error) {
	log.Println("----repository Create called-----")

	// データベースにスケジュールを挿入し、挿入されたスケジュールを返す
	row := r.pool.QueryRow(ctx,
		`INSERT INTO schedules (
			operator_id,
			origin_lat, origin_lng, origin_name,
			dest_lat, dest_lng, dest_name,
			depart_at, arrive_at,
			max_weight_kg, max_size_cm, avail_weight_kg,
			status, route_geojson
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING `+scheduleColumns,
		s.OperatorID,
		s.OriginLat, s.OriginLng, s.OriginName,
		s.DestLat, s.DestLng, s.DestName,
		s.DepartAt, s.ArriveAt,
		s.MaxWeightKg, s.MaxSizeCm, s.AvailWeightKg,
		s.Status, s.RouteGeoJSON,
	)
	return scanSchedule(row)
}

func (r *ScheduleRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.ScheduleStatus) error {
	log.Println("----repository UpdateStatus called-----")

	// データベースのスケジュールのstatusを更新する
	_, err := r.pool.Exec(ctx,
		`UPDATE schedules SET status = $1 WHERE id = $2`,
		status, id,
	)
	return err
}

// AddAvailWeight は avail_weight_kg を指定量だけ加算する（キャンセル時の回復用）
func (r *ScheduleRepository) AddAvailWeight(ctx context.Context, id uuid.UUID, delta float64) error {
	log.Println("----repository AddAvailWeight called-----")

	// データベースのスケジュールのavail_weight_kgをdeltaだけ加算する
	_, err := r.pool.Exec(ctx,
		`UPDATE schedules SET avail_weight_kg = avail_weight_kg + $1 WHERE id = $2`,
		delta, id,
	)
	return err
}

func (r *ScheduleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	log.Println("----repository Delete called-----")

	// データベースからスケジュールを削除する
	_, err := r.pool.Exec(ctx, `DELETE FROM schedules WHERE id = $1`, id)
	return err
}

func (r *ScheduleRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Schedule, error) {
	log.Println("----repository FindByID called-----")

	// データベースからスケジュールを1件取得する
	query := `SELECT id, operator_id, origin_lat, origin_lng, origin_name,
			  dest_lat, dest_lng, dest_name, depart_at, arrive_at,
			  max_weight_kg, max_size_cm, avail_weight_kg, status,
			  route_geojson, created_at FROM schedules WHERE id = $1`

	// クエリを実行してスケジュールを取得する
	row := r.pool.QueryRow(ctx, query, id)
	s, err := scanSchedule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// 取得したスケジュールに紐づく予約を全て取得するクエリ
	bookingQuery := `
		SELECT id, schedule_id, shipper_id, tracking_number, weight_kg,
			   size_cm, content_desc, recipient_name, status, created_at
		FROM bookings
		WHERE schedule_id = $1
		ORDER BY created_at DESC`
	
	rows, err := r.pool.Query(ctx, bookingQuery, s.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bookings: %w", err)
	}
	defer rows.Close()


	s.Bookings = []model.Booking{}
	// クエリの結果をスケジュールの予約のスライスに変換してスケジュールにセットする
	for rows.Next() {
		var b model.Booking
		// データベースからスケジュールの予約の情報を読み取る
		err := rows.Scan(
			&b.ID, &b.ScheduleID, &b.ShipperID, &b.TrackingNumber, &b.WeightKg,
			&b.SizeCm, &b.ContentDesc, &b.RecipientName, &b.Status, &b.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		// 取得したスケジュールの予約をスケジュールの予約のスライスに追加する
		s.Bookings = append(s.Bookings, b)
	}

	return s, nil
}

func (r *ScheduleRepository) ListByOperator(ctx context.Context, operatorID uuid.UUID) ([]model.Schedule, error) {
	log.Println("----repository ListByOperator called-----")

	// データベースから指定されたオペレーターIDのスケジュールを全て取得する
	rows, err := r.pool.Query(ctx,
		`SELECT `+scheduleColumns+` FROM schedules WHERE operator_id = $1 ORDER BY depart_at DESC`,
		operatorID,
	)
	// エラーが発生した場合はエラーを返す
	if err != nil {
		return nil, err
	}
	// クエリの結果をスケジュールのスライスに変換して返す
	defer rows.Close()
	return collectSchedules(ctx, r, rows)
}

func (r *ScheduleRepository) Search(ctx context.Context, filter ScheduleFilter) ([]model.Schedule, error) {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

func collectSchedules(ctx context.Context, r *ScheduleRepository, rows pgx.Rows) ([]model.Schedule, error) {
	log.Println("----repository collectSchedules called-----")

	var schedules []model.Schedule
	for rows.Next() {
		var s model.Schedule
		err := rows.Scan(
			&s.ID, &s.OperatorID,
			&s.OriginLat, &s.OriginLng, &s.OriginName,
			&s.DestLat, &s.DestLng, &s.DestName,
			&s.DepartAt, &s.ArriveAt,
			&s.MaxWeightKg, &s.MaxSizeCm, &s.AvailWeightKg,
			&s.Status, &s.RouteGeoJSON, &s.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// 各スケジュールの予約を取得
		s.Bookings = []model.Booking{}
		bRows, err := r.pool.Query(ctx,
			`SELECT id, schedule_id, shipper_id, tracking_number, weight_kg, status, recipient_name
			 FROM bookings WHERE schedule_id = $1 ORDER BY created_at DESC`, s.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch bookings for schedule %s: %w", s.ID, err)
		}
		for bRows.Next() {
			var b model.Booking
			if scanErr := bRows.Scan(&b.ID, &b.ScheduleID, &b.ShipperID, &b.TrackingNumber, &b.WeightKg, &b.Status, &b.RecipientName); scanErr != nil {
				bRows.Close()
				return nil, scanErr
			}
			s.Bookings = append(s.Bookings, b)
		}
		bRows.Close()
		if err := bRows.Err(); err != nil {
			return nil, err
		}

		schedules = append(schedules, s)
	}
	return schedules, rows.Err()
}
