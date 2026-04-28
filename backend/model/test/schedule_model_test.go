package model_test

import (
	"testing"

	"github.com/bus-logistics/backend/model"
)

// ---- ScheduleStatus 定数値 ----

func TestScheduleStatus_Constants(t *testing.T) {
	tests := []struct {
		status model.ScheduleStatus
		want   string
	}{
		{model.ScheduleStatusOpen, "open"},
		{model.ScheduleStatusFull, "full"},
		{model.ScheduleStatusDeparted, "departed"},
		{model.ScheduleStatusArrived, "arrived"},
	}
	for _, tt := range tests {
		if string(tt.status) != tt.want {
			t.Errorf("want %q, got %q", tt.want, tt.status)
		}
	}
}

// ---- Schedule 構造体フィールド ----

func TestSchedule_ZeroValue(t *testing.T) {
	var s model.Schedule
	if s.MaxWeightKg != 0 {
		t.Errorf("want MaxWeightKg=0, got %v", s.MaxWeightKg)
	}
	if s.AvailWeightKg != 0 {
		t.Errorf("want AvailWeightKg=0, got %v", s.AvailWeightKg)
	}
	if s.Bookings != nil {
		t.Errorf("want Bookings=nil, got %v", s.Bookings)
	}
}

func TestSchedule_BookingsField(t *testing.T) {
	s := model.Schedule{
		Bookings: []model.Booking{
			{TrackingNumber: "TRK-001"},
			{TrackingNumber: "TRK-002"},
		},
	}
	if len(s.Bookings) != 2 {
		t.Errorf("want 2 bookings, got %d", len(s.Bookings))
	}
	if s.Bookings[0].TrackingNumber != "TRK-001" {
		t.Errorf("want TRK-001, got %s", s.Bookings[0].TrackingNumber)
	}
}
