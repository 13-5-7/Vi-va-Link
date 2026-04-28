package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"
	"log"

	"github.com/bus-logistics/backend/model"
	"github.com/bus-logistics/backend/repository"
	"github.com/bus-logistics/backend/utils"
	"github.com/google/uuid"
)

var (
	ErrOriginRequired = errors.New("origin location is required")
	ErrDestRequired   = errors.New("destination location is required")
	ErrDepartAtPast   = errors.New("depart_at must be in the future")
)

type CreateScheduleRequest struct {
	OperatorID   uuid.UUID
	OriginLat    float64
	OriginLng    float64
	OriginName   string
	DestLat      float64
	DestLng      float64
	DestName     string
	DepartAt     time.Time
	ArriveAt     time.Time
	MaxWeightKg  float64
	MaxSizeCm    float64
	RouteGeoJSON json.RawMessage
}

type ScheduleService struct {
	repo        *repository.ScheduleRepository
	bookingRepo *repository.BookingRepository
}

func NewScheduleService(repo *repository.ScheduleRepository, bookingRepo *repository.BookingRepository) *ScheduleService {
	return &ScheduleService{repo: repo, bookingRepo: bookingRepo}
}

var ErrInvalidMaxWeight = errors.New("max_weight_kg must be greater than 0")
var ErrInvalidMaxSize   = errors.New("max_size_cm must be greater than 0")

func (s *ScheduleService) Create(ctx context.Context, req CreateScheduleRequest) (*model.Schedule, error) {
	log.Println("----service Create called-----")

	// バリデーション
	// origin/dest の緯度経度の両方が空の場合はエラー
	if utils.IsEmpty(req.OriginLat) && utils.IsEmpty(req.OriginLng) {
		return nil, ErrOriginRequired
	}
	if utils.IsEmpty(req.DestLat) && utils.IsEmpty(req.DestLng) {
		return nil, ErrDestRequired
	}
	// depart_at が過去の場合はエラー
	if !req.DepartAt.After(time.Now()) {
		return nil, ErrDepartAtPast
	}
	// max_weight_kg が0以下の場合はエラー
	if req.MaxWeightKg <= 0 {
		return nil, ErrInvalidMaxWeight
	}
	// max_size_cm が0以下の場合はエラー
	if req.MaxSizeCm <= 0 {
		return nil, ErrInvalidMaxSize
	}
	// route_geojson はバリデーションは行わない（OSRMのレスポンスをそのまま保存する想定）

	// スケジュール作成
	schedule := &model.Schedule {
		OperatorID:    req.OperatorID,
		OriginLat:     req.OriginLat,
		OriginLng:     req.OriginLng,
		OriginName:    req.OriginName,
		DestLat:       req.DestLat,
		DestLng:       req.DestLng,
		DestName:      req.DestName,
		DepartAt:      req.DepartAt,
		ArriveAt:      req.ArriveAt,
		MaxWeightKg:   req.MaxWeightKg,
		MaxSizeCm:     req.MaxSizeCm,
		AvailWeightKg: req.MaxWeightKg,
		Status:        model.ScheduleStatusOpen,
		RouteGeoJSON:  req.RouteGeoJSON,
	}

	return s.repo.Create(ctx, schedule)
}

func (s *ScheduleService) ListByOperator(ctx context.Context, operatorID uuid.UUID) ([]model.Schedule, error) {
	return s.repo.ListByOperator(ctx, operatorID)
}

func (s *ScheduleService) Search(ctx context.Context, filter repository.ScheduleFilter) ([]model.Schedule, error) {
	return s.repo.Search(ctx, filter)
}

func (s *ScheduleService) GetByID(ctx context.Context, id uuid.UUID) (*model.Schedule, error) {
	return s.repo.FindByID(ctx, id)
}

var ErrInvalidScheduleTransition = errors.New("invalid schedule status transition")
var ErrScheduleHasBookings = errors.New("schedule has bookings and cannot be deleted")

// scheduleStatusOrder はスケジュールステータスの順序
var scheduleStatusOrder = map[model.ScheduleStatus]int{
	model.ScheduleStatusOpen:     0,
	model.ScheduleStatusFull:     1,
	model.ScheduleStatusDeparted: 2,
	model.ScheduleStatusArrived:  3,
}

// UpdateScheduleStatus はスケジュールのステータスを更新する（荷物への連動なし）
func (s *ScheduleService) UpdateScheduleStatus(ctx context.Context, scheduleID uuid.UUID, newStatus model.ScheduleStatus, operatorID uuid.UUID) error {
	log.Println("----service UpdateScheduleStatus called-----")

	// スケジュールが存在するか、オペレーターが所有しているかをチェック
	schedule, err := s.repo.FindByID(ctx, scheduleID)
	if err != nil {
		return err
	}
	if schedule == nil {
		return ErrScheduleNotFound
	}

	// スケジュールがオペレーターのものでない場合は ErrScheduleNotFound を返す（セキュリティ上、存在しないのと同じ扱いにする）
	currentOrder, ok1 := scheduleStatusOrder[schedule.Status]
	nextOrder, ok2 := scheduleStatusOrder[newStatus]
	if !ok1 || !ok2 || nextOrder <= currentOrder {
		return ErrInvalidScheduleTransition
	}

	// ステータスを更新する
	return s.repo.UpdateStatus(ctx, scheduleID, newStatus)
}

// Delete はスケジュールを削除する（予約がある場合・出発済みは不可）
// cancelled 予約のみの場合は削除可能
func (s *ScheduleService) Delete(ctx context.Context, scheduleID uuid.UUID, operatorID uuid.UUID) error {
	log.Println("----service Delete called-----")

	// スケジュールが存在するか、オペレーターが所有しているか、出発済みでないかをチェック
	schedule, err := s.repo.FindByID(ctx, scheduleID)
	// エラーハンドリング
	if err != nil {
		return err
	}
	// スケジュールが見つからない場合は ErrScheduleNotFound を返す
	if schedule == nil {
		return ErrScheduleNotFound
	}
	// スケジュールがオペレーターのものでない場合は ErrScheduleNotFound を返す（セキュリティ上、存在しないのと同じ扱いにする）
	if schedule.OperatorID != operatorID {
		return ErrScheduleNotFound
	}
	// スケジュールが出発済みの場合は ErrInvalidScheduleTransition を返す
	if schedule.Status == model.ScheduleStatusDeparted || schedule.Status == model.ScheduleStatusArrived {
		return ErrInvalidScheduleTransition
	}

	// 予約のステータスが cancelled 以外のものがある場合は削除不可
	activeBookings := 0
	for _, b := range schedule.Bookings {
		if b.Status != model.BookingStatusCancelled {
			activeBookings++
		}
	}
	// アクティブな予約がある場合は ErrScheduleHasBookings を返す
	if activeBookings > 0 {
		return ErrScheduleHasBookings
	}
	return s.repo.Delete(ctx, scheduleID)
}

var ErrScheduleAlreadyCancelled = errors.New("schedule is already cancelled")

// Cancel はスケジュールをキャンセルする
// open/full のスケジュールのみキャンセル可能
// accepted 状態の予約は全て cancelled に変更し、avail_weight_kg を回復する
func (s *ScheduleService) Cancel(ctx context.Context, scheduleID uuid.UUID, operatorID uuid.UUID) error {
	log.Println("----service Cancel called-----")

	schedule, err := s.repo.FindByID(ctx, scheduleID)
	// エラーハンドリング
	if err != nil {
		return err
	}
	// スケジュールが見つからない場合は ErrScheduleNotFound を返す
	if schedule == nil {
		return ErrScheduleNotFound
	}
	// スケジュールがオペレーターのものでない場合は ErrScheduleNotFound を返す
	if schedule.OperatorID != operatorID {
		return ErrScheduleNotFound
	}
	// スケジュールがすでにキャンセル済みの場合は ErrScheduleAlreadyCancelled を返す
	if schedule.Status == model.ScheduleStatusCancelled {
		return ErrScheduleAlreadyCancelled
	}
	// スケジュールが出発済み・到着済みの場合は ErrInvalidScheduleTransition を返す
	if schedule.Status == model.ScheduleStatusDeparted || schedule.Status == model.ScheduleStatusArrived {
		return ErrInvalidScheduleTransition
	}

	for _, b := range schedule.Bookings {
		// accepted 状態の予約は全て cancelled に変更し、avail_weight_kg を回復する
		if b.Status == model.BookingStatusAccepted {
			// トランザクション内で予約のステータスを cancelled に更新し、スケジュールの avail_weight_kg を回復する
			if err := s.bookingRepo.UpdateStatusDirect(ctx, b.ID, model.BookingStatusCancelled); err != nil {
				return err
			}
			// スケジュールの avail_weight_kg を回復する
			if err := s.repo.AddAvailWeight(ctx, scheduleID, b.WeightKg); err != nil {
				return err
			}
		}
	}

	// スケジュールのステータスを cancelled に更新する
	return s.repo.UpdateStatus(ctx, scheduleID, model.ScheduleStatusCancelled)
}

