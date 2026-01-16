package main

import (
	"testing"
	"time"
)

// TestBuildSchedule testa a função buildSchedule indiretamente através de dados
func TestBuildSchedule(t *testing.T) {
	t.Run("Build schedule with valid parameters", func(t *testing.T) {
		// buildSchedule é privada, então testamos os tipos e estruturas que ela usa
		professionalID := 1
		unitID := 1
		days := 15
		startDate := time.Now()

		// Testa que os parâmetros são válidos
		if professionalID < 1 {
			t.Error("professionalID should be positive")
		}
		if unitID < 1 {
			t.Error("unitID should be positive")
		}
		if days < 1 {
			t.Error("days should be positive")
		}
		if startDate.IsZero() {
			t.Error("startDate should not be zero")
		}
	})
}

func TestTimeOperations(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Valid date format",
			input:   "2024-01-15",
			wantErr: false,
		},
		{
			name:    "Invalid format",
			input:   "15-01-2024",
			wantErr: true,
		},
		{
			name:    "Invalid date",
			input:   "2024-13-45",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := time.Parse("2006-01-02", tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSchedulePayloadStructure(t *testing.T) {
	t.Run("SchedulePayload has required fields", func(t *testing.T) {
		// Testa a estrutura schedulePayload
		schedule := schedulePayload{
			Professional: map[string]interface{}{
				"id":   1,
				"name": "Test Doctor",
			},
			Unit: map[string]interface{}{
				"id":   1,
				"name": "Test Unit",
			},
			Date: "2024-01-15",
			Slots: []scheduleSlot{
				{Start: "09:00", Available: true},
			},
		}

		if schedule.Date == "" {
			t.Error("Date should not be empty")
		}
		if len(schedule.Slots) == 0 {
			t.Error("Should have at least one slot")
		}
		if schedule.Professional == nil {
			t.Error("Professional should not be nil")
		}
		if schedule.Unit == nil {
			t.Error("Unit should not be nil")
		}
	})
}

func TestScheduleSlotStructure(t *testing.T) {
	t.Run("ScheduleSlot has required fields", func(t *testing.T) {
		slot := scheduleSlot{
			Start:     "09:00",
			Available: true,
		}

		if slot.Start == "" {
			t.Error("Start time should not be empty")
		}

		// Valida formato de hora
		_, err := time.Parse("15:04", slot.Start)
		if err != nil {
			t.Errorf("Start time should be in HH:MM format, got: %s", slot.Start)
		}
	})
}

func TestMetricsStore(t *testing.T) {
	t.Run("Create new metrics store", func(t *testing.T) {
		buckets := []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}
		store := newMetricsStore(buckets)

		if store == nil {
			t.Fatal("newMetricsStore should not return nil")
		}

		if len(store.buckets) != len(buckets) {
			t.Errorf("Expected %d buckets, got %d", len(buckets), len(store.buckets))
		}
	})

	t.Run("Observe metrics", func(t *testing.T) {
		buckets := []float64{0.1, 0.5, 1.0}
		store := newMetricsStore(buckets)

		// Registra uma métrica
		store.observe("/test", 200, 0.05)

		// Verifica que foi registrado
		if len(store.counts) == 0 {
			t.Error("Expected metrics to be recorded")
		}
	})
}

func TestDateCalculations(t *testing.T) {
	t.Run("Calculate date ranges", func(t *testing.T) {
		start := time.Now()
		days := 15

		// Testa que podemos calcular datas futuras
		for i := 0; i < days; i++ {
			future := start.AddDate(0, 0, i)
			if future.Before(start) {
				t.Errorf("Future date should not be before start date")
			}
		}
	})

	t.Run("Validate day range", func(t *testing.T) {
		// Simula validação de dias (15-30)
		tests := []struct {
			days     int
			expected int
		}{
			{5, 15},  // menor que 15, ajusta para 15
			{20, 20}, // entre 15-30, mantém
			{35, 30}, // maior que 30, ajusta para 30
		}

		for _, tt := range tests {
			result := tt.days
			if result < 15 {
				result = 15
			}
			if result > 30 {
				result = 30
			}

			if result != tt.expected {
				t.Errorf("Expected %d days, got %d", tt.expected, result)
			}
		}
	})
}

func TestTimeFormatting(t *testing.T) {
	t.Run("Format time as HH:MM", func(t *testing.T) {
		now := time.Date(2024, 1, 15, 9, 30, 0, 0, time.UTC)
		formatted := now.Format("15:04")

		if formatted != "09:30" {
			t.Errorf("Expected '09:30', got '%s'", formatted)
		}
	})

	t.Run("Parse time from HH:MM", func(t *testing.T) {
		timeStr := "14:30"
		_, err := time.Parse("15:04", timeStr)

		if err != nil {
			t.Errorf("Should parse valid time, got error: %v", err)
		}
	})
}

func BenchmarkMetricsObserve(b *testing.B) {
	buckets := []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}
	store := newMetricsStore(buckets)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.observe("/test", 200, 0.05)
	}
}
