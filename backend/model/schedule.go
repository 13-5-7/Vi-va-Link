package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ScheduleStatus string

const (
	ScheduleStatusOpen      ScheduleStatus = "open"
	ScheduleStatusFull      ScheduleStatus = "full"
	ScheduleStatusDeparted  ScheduleStatus = "departed"
	ScheduleStatusArrived   ScheduleStatus = "arrived"
	ScheduleStatusCancelled ScheduleStatus = "cancelled"
)

type Schedule struct {
	ID            uuid.UUID       `db:"id" json:"id"`
	OperatorID    uuid.UUID       `db:"operator_id" json:"operator_id"`
	OriginLat     float64         `db:"origin_lat" json:"origin_lat"`
	OriginLng     float64         `db:"origin_lng" json:"origin_lng"`
	OriginName    string          `db:"origin_name" json:"origin_name"`
	DestLat       float64         `db:"dest_lat" json:"dest_lat"`
	DestLng       float64         `db:"dest_lng" json:"dest_lng"`
	DestName      string          `db:"dest_name" json:"dest_name"`
	DepartAt      time.Time       `db:"depart_at" json:"depart_at"`
	ArriveAt      time.Time       `db:"arrive_at" json:"arrive_at"`
	MaxWeightKg   float64         `db:"max_weight_kg" json:"max_weight_kg"`
	MaxSizeCm     float64         `db:"max_size_cm" json:"max_size_cm"`
	AvailWeightKg float64         `db:"avail_weight_kg" json:"avail_weight_kg"`
	Status        ScheduleStatus  `db:"status" json:"status"`
	RouteGeoJSON  json.RawMessage `db:"route_geojson" json:"route_geojson"`
	CreatedAt     time.Time       `db:"created_at" json:"created_at"`
	Bookings      []Booking       `db:"-" json:"bookings"` // これでフロントエンドの res.data.bookings と一致します
}
