package repository_test

import (
	"testing"
	"time"

	"github.com/bus-logistics/backend/repository"
)

func ptr[T any](v T) *T { return &v }

// ---- ScheduleFilter フィールド設定 ----

func TestScheduleFilter_ZeroValue(t *testing.T) {
	var f repository.ScheduleFilter
	if f.OriginLatMin != nil {
		t.Error("OriginLatMin should be nil by default")
	}
	if f.DepartAtFrom != nil {
		t.Error("DepartAtFrom should be nil by default")
	}
}

func TestScheduleFilter_OriginBounds(t *testing.T) {
	f := repository.ScheduleFilter{
		OriginLatMin: ptr(34.0),
		OriginLatMax: ptr(36.0),
		OriginLngMin: ptr(138.0),
		OriginLngMax: ptr(140.0),
	}
	if *f.OriginLatMin != 34.0 {
		t.Errorf("want 34.0, got %v", *f.OriginLatMin)
	}
	if *f.OriginLatMax != 36.0 {
		t.Errorf("want 36.0, got %v", *f.OriginLatMax)
	}
	if *f.OriginLngMin != 138.0 {
		t.Errorf("want 138.0, got %v", *f.OriginLngMin)
	}
	if *f.OriginLngMax != 140.0 {
		t.Errorf("want 140.0, got %v", *f.OriginLngMax)
	}
}

func TestScheduleFilter_DestBounds(t *testing.T) {
	f := repository.ScheduleFilter{
		DestLatMin: ptr(33.0),
		DestLatMax: ptr(35.0),
		DestLngMin: ptr(134.0),
		DestLngMax: ptr(136.0),
	}
	if *f.DestLatMin != 33.0 {
		t.Errorf("want 33.0, got %v", *f.DestLatMin)
	}
	if *f.DestLngMax != 136.0 {
		t.Errorf("want 136.0, got %v", *f.DestLngMax)
	}
}

func TestScheduleFilter_DepartAtRange(t *testing.T) {
	from := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)
	f := repository.ScheduleFilter{
		DepartAtFrom: &from,
		DepartAtTo:   &to,
	}
	if !f.DepartAtFrom.Equal(from) {
		t.Errorf("want %v, got %v", from, *f.DepartAtFrom)
	}
	if !f.DepartAtTo.Equal(to) {
		t.Errorf("want %v, got %v", to, *f.DepartAtTo)
	}
}

func TestScheduleFilter_DepartAtFrom_BeforeTo(t *testing.T) {
	from := time.Date(2099, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2099, 6, 30, 0, 0, 0, 0, time.UTC)
	f := repository.ScheduleFilter{
		DepartAtFrom: &from,
		DepartAtTo:   &to,
	}
	if !f.DepartAtFrom.Before(*f.DepartAtTo) {
		t.Error("DepartAtFrom should be before DepartAtTo")
	}
}

func TestScheduleFilter_PartialFilter_OnlyOriginLat(t *testing.T) {
	f := repository.ScheduleFilter{
		OriginLatMin: ptr(35.0),
	}
	if f.OriginLatMin == nil {
		t.Error("OriginLatMin should be set")
	}
	// 他のフィールドは nil のまま
	if f.OriginLatMax != nil {
		t.Error("OriginLatMax should remain nil")
	}
	if f.DestLatMin != nil {
		t.Error("DestLatMin should remain nil")
	}
}
