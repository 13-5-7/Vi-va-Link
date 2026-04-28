package model

import (
	"time"

	"github.com/google/uuid"
)

type BookingStatus string

const (
	BookingStatusAccepted  BookingStatus = "accepted"
	BookingStatusLoaded    BookingStatus = "loaded"
	BookingStatusInTransit BookingStatus = "in_transit"
	BookingStatusDelivered BookingStatus = "delivered"
	BookingStatusCancelled BookingStatus = "cancelled"
)

// StatusOrder はステータスの順序を定義（前方向遷移のみ許可）
var StatusOrder = map[BookingStatus]int{
	BookingStatusAccepted:  0,
	BookingStatusLoaded:    1,
	BookingStatusInTransit: 2,
	BookingStatusDelivered: 3,
}

// CanTransitionTo は現在のステータスから次のステータスへの遷移が有効かチェック
func (s BookingStatus) CanTransitionTo(next BookingStatus) bool {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

type Booking struct {
	ID              uuid.UUID     `db:"id"               json:"id"`
	ScheduleID      uuid.UUID     `db:"schedule_id"      json:"schedule_id"`
	ShipperID       uuid.UUID     `db:"shipper_id"       json:"shipper_id"`
	TrackingNumber  string        `db:"tracking_number"  json:"tracking_number"`
	WeightKg        float64       `db:"weight_kg"        json:"weight_kg"`
	SizeCm          float64       `db:"size_cm"          json:"size_cm"`
	ContentDesc     string        `db:"content_desc"     json:"content_desc"`
	RecipientName   string        `db:"recipient_name"   json:"recipient_name"`
	RecipientPhone  string        `db:"recipient_phone"  json:"recipient_phone"`
	RecipientAddr   string        `db:"recipient_addr"   json:"recipient_addr"`
	Status          BookingStatus `db:"status"           json:"status"`
	StatusUpdatedAt time.Time     `db:"status_updated_at" json:"status_updated_at"`
	CreatedAt       time.Time     `db:"created_at"       json:"created_at"`
}
