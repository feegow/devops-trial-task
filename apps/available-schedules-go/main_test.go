package main

import (
	"testing"
	"time"
)

func TestAlignToHalfHour(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "already aligned to 00",
			input:    time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
			expected: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			name:     "already aligned to 30",
			input:    time.Date(2025, 1, 1, 10, 30, 0, 0, time.UTC),
			expected: time.Date(2025, 1, 1, 10, 30, 0, 0, time.UTC),
		},
		{
			name:     "round up from 15 to 30",
			input:    time.Date(2025, 1, 1, 10, 15, 0, 0, time.UTC),
			expected: time.Date(2025, 1, 1, 10, 30, 0, 0, time.UTC),
		},
		{
			name:     "round up from 45 to 00",
			input:    time.Date(2025, 1, 1, 10, 45, 0, 0, time.UTC),
			expected: time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := alignToHalfHour(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestResolveProfessional(t *testing.T) {
	// ID existente
	p := resolveProfessional(2684)
	if p.ID != 2684 {
		t.Errorf("expected ID 2684, got %d", p.ID)
	}

	// ID inexistente retorna primeiro
	p = resolveProfessional(9999)
	if p.ID != professionals[0].ID {
		t.Errorf("expected default professional, got %d", p.ID)
	}
}

func TestResolveUnit(t *testing.T) {
	// ID existente
	u := resolveUnit(901)
	if u.ID != 901 {
		t.Errorf("expected ID 901, got %d", u.ID)
	}

	// ID inexistente retorna primeiro
	u = resolveUnit(9999)
	if u.ID != units[0].ID {
		t.Errorf("expected default unit, got %d", u.ID)
	}
}
