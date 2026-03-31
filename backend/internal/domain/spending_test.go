package domain

import (
	"testing"
	"time"
)

func TestCalculateSpendingPercentage(t *testing.T) {
	tests := []struct {
		name             string
		currentCostCents int64
		budgetCents      int64
		expected         int
	}{
		{
			name:             "zero cost",
			currentCostCents: 0,
			budgetCents:      500000,
			expected:         0,
		},
		{
			name:             "50 percent",
			currentCostCents: 250000,
			budgetCents:      500000,
			expected:         50,
		},
		{
			name:             "80 percent",
			currentCostCents: 400000,
			budgetCents:      500000,
			expected:         80,
		},
		{
			name:             "95 percent",
			currentCostCents: 475000,
			budgetCents:      500000,
			expected:         95,
		},
		{
			name:             "100 percent",
			currentCostCents: 500000,
			budgetCents:      500000,
			expected:         100,
		},
		{
			name:             "over 100 percent",
			currentCostCents: 550000,
			budgetCents:      500000,
			expected:         110,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateSpendingPercentage(tt.currentCostCents, tt.budgetCents)
			if result != tt.expected {
				t.Errorf("got %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestShouldSendAlert(t *testing.T) {
	tests := []struct {
		name                string
		currentPercentage   int
		warnThreshold       int
		criticalThreshold   int
		expectedWarn        bool
		expectedCritical    bool
	}{
		{
			name:              "below warn threshold",
			currentPercentage: 70,
			warnThreshold:     80,
			criticalThreshold: 95,
			expectedWarn:      false,
			expectedCritical:  false,
		},
		{
			name:              "at warn threshold",
			currentPercentage: 80,
			warnThreshold:     80,
			criticalThreshold: 95,
			expectedWarn:      true,
			expectedCritical:  false,
		},
		{
			name:              "between warn and critical",
			currentPercentage: 90,
			warnThreshold:     80,
			criticalThreshold: 95,
			expectedWarn:      true,
			expectedCritical:  false,
		},
		{
			name:              "at critical threshold",
			currentPercentage: 95,
			warnThreshold:     80,
			criticalThreshold: 95,
			expectedWarn:      true,
			expectedCritical:  true,
		},
		{
			name:              "above critical",
			currentPercentage: 100,
			warnThreshold:     80,
			criticalThreshold: 95,
			expectedWarn:      true,
			expectedCritical:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warn, critical := ShouldSendAlert(tt.currentPercentage, tt.warnThreshold, tt.criticalThreshold)
			if warn != tt.expectedWarn || critical != tt.expectedCritical {
				t.Errorf("got warn=%v critical=%v, want warn=%v critical=%v",
					warn, critical, tt.expectedWarn, tt.expectedCritical)
			}
		})
	}
}

func TestGetCurrentMonth(t *testing.T) {
	now := time.Now()
	expected := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	result := GetCurrentMonth()
	if result != expected {
		t.Errorf("got %v, want %v", result, expected)
	}
}

func TestGetMonthStart(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "beginning of month",
			input:    time.Date(2026, 3, 1, 15, 30, 0, 0, time.UTC),
			expected: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "middle of month",
			input:    time.Date(2026, 3, 15, 15, 30, 0, 0, time.UTC),
			expected: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "end of month",
			input:    time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC),
			expected: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMonthStart(tt.input)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}
