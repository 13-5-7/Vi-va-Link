package model_test

import (
	"testing"

	"github.com/bus-logistics/backend/model"
)

// ---- BookingStatus 定数値 ----

func TestBookingStatus_Constants(t *testing.T) {
	tests := []struct {
		status model.BookingStatus
		want   string
	}{
		{model.BookingStatusAccepted, "accepted"},
		{model.BookingStatusLoaded, "loaded"},
		{model.BookingStatusInTransit, "in_transit"},
		{model.BookingStatusDelivered, "delivered"},
	}
	for _, tt := range tests {
		if string(tt.status) != tt.want {
			t.Errorf("want %q, got %q", tt.want, tt.status)
		}
	}
}

// ---- StatusOrder ----

func TestStatusOrder_AllStatusesDefined(t *testing.T) {
	statuses := []model.BookingStatus{
		model.BookingStatusAccepted,
		model.BookingStatusLoaded,
		model.BookingStatusInTransit,
		model.BookingStatusDelivered,
	}
	for _, s := range statuses {
		if _, ok := model.StatusOrder[s]; !ok {
			t.Errorf("StatusOrder missing entry for %q", s)
		}
	}
}

func TestStatusOrder_Ascending(t *testing.T) {
	// accepted < loaded < in_transit < delivered の順序を確認
	ordered := []model.BookingStatus{
		model.BookingStatusAccepted,
		model.BookingStatusLoaded,
		model.BookingStatusInTransit,
		model.BookingStatusDelivered,
	}
	for i := 1; i < len(ordered); i++ {
		prev := model.StatusOrder[ordered[i-1]]
		curr := model.StatusOrder[ordered[i]]
		if prev >= curr {
			t.Errorf("StatusOrder: %q (%d) should be less than %q (%d)", ordered[i-1], prev, ordered[i], curr)
		}
	}
}

// ---- CanTransitionTo ----

func TestCanTransitionTo_ValidForwardTransitions(t *testing.T) {
	tests := []struct {
		from model.BookingStatus
		to   model.BookingStatus
	}{
		{model.BookingStatusAccepted, model.BookingStatusLoaded},
		{model.BookingStatusAccepted, model.BookingStatusInTransit},
		{model.BookingStatusAccepted, model.BookingStatusDelivered},
		{model.BookingStatusLoaded, model.BookingStatusInTransit},
		{model.BookingStatusLoaded, model.BookingStatusDelivered},
		{model.BookingStatusInTransit, model.BookingStatusDelivered},
	}
	for _, tt := range tests {
		if !tt.from.CanTransitionTo(tt.to) {
			t.Errorf("expected %q -> %q to be valid", tt.from, tt.to)
		}
	}
}

func TestCanTransitionTo_InvalidBackwardTransitions(t *testing.T) {
	tests := []struct {
		from model.BookingStatus
		to   model.BookingStatus
	}{
		{model.BookingStatusLoaded, model.BookingStatusAccepted},
		{model.BookingStatusInTransit, model.BookingStatusLoaded},
		{model.BookingStatusInTransit, model.BookingStatusAccepted},
		{model.BookingStatusDelivered, model.BookingStatusInTransit},
		{model.BookingStatusDelivered, model.BookingStatusLoaded},
		{model.BookingStatusDelivered, model.BookingStatusAccepted},
	}
	for _, tt := range tests {
		if tt.from.CanTransitionTo(tt.to) {
			t.Errorf("expected %q -> %q to be invalid (backward)", tt.from, tt.to)
		}
	}
}

func TestCanTransitionTo_SameStatus(t *testing.T) {
	statuses := []model.BookingStatus{
		model.BookingStatusAccepted,
		model.BookingStatusLoaded,
		model.BookingStatusInTransit,
		model.BookingStatusDelivered,
	}
	for _, s := range statuses {
		if s.CanTransitionTo(s) {
			t.Errorf("expected same-status transition %q -> %q to be invalid", s, s)
		}
	}
}

func TestCanTransitionTo_UnknownStatus(t *testing.T) {
	unknown := model.BookingStatus("unknown")
	if unknown.CanTransitionTo(model.BookingStatusLoaded) {
		t.Error("unknown -> loaded should be invalid")
	}
	if model.BookingStatusAccepted.CanTransitionTo(unknown) {
		t.Error("accepted -> unknown should be invalid")
	}
}
