package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/repository"
)

type SpendingResponse struct {
	CurrentCostCents int64     `json:"current_cost_cents"`
	BudgetCents      int64     `json:"budget_cents"`
	SpendingPercent  int       `json:"spending_percent"`
	RemainingCents   int64     `json:"remaining_cents"`
	Month            time.Time `json:"month"`
	AlertStatus      string    `json:"alert_status"` // "normal", "warning", "critical"
	LastAlertSentAt  *time.Time `json:"last_alert_sent_at,omitempty"`
}

type SpendingHistoryResponse struct {
	Items  []MonthlyCostItem `json:"items"`
	Total  int64             `json:"total"`
	Limit  int32             `json:"limit"`
	Offset int32             `json:"offset"`
}

type MonthlyCostItem struct {
	Month                   time.Time `json:"month"`
	InfrastructureCostCents int64     `json:"infrastructure_cost_cents"`
	ClaudeTokens            int64     `json:"claude_tokens"`
	TotalCostCents          int64     `json:"total_cost_cents"`
	BudgetCents             int64     `json:"budget_cents"`
	SpendingPercent         int       `json:"spending_percent"`
}

type AlertHistoryResponse struct {
	Items  []AlertItem `json:"items"`
	Total  int64       `json:"total"`
	Limit  int32       `json:"limit"`
	Offset int32       `json:"offset"`
}

type AlertItem struct {
	ID               int64     `json:"id"`
	AlertLevel       string    `json:"alert_level"` // "warning" or "critical"
	ThresholdPercent int32     `json:"threshold_percent"`
	SentAt           time.Time `json:"sent_at"`
}

// GetSpendingStatus returns current spending status for owner's business
func (h *Handler) GetSpendingStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value("userID").(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get owner's business
	business, err := h.Queries.GetBusinessByOwner(ctx, ownerID)
	if err != nil {
		slog.Error("failed to get business", "error", err)
		http.Error(w, "business not found", http.StatusNotFound)
		return
	}

	// Get current month's spending
	currentMonth := domain.GetCurrentMonth()
	monthlyCost, err := h.Queries.GetOrCreateMonthlyCost(ctx, repository.GetOrCreateMonthlyCostParams{
		BusinessID: business.ID,
		Month:      currentMonth,
	})
	if err != nil {
		// Try to create if doesn't exist
		monthlyCost, err = h.Queries.CreateMonthlyCost(ctx, repository.CreateMonthlyCostParams{
			BusinessID:         business.ID,
			Month:              currentMonth,
			MonthlyBudgetCents: h.Config.MonthlyBudgetCents,
		})
		if err != nil {
			slog.Error("failed to create monthly cost", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	// Calculate spending percentage
	spendingPercent := domain.CalculateSpendingPercentage(
		monthlyCost.TotalCostCents,
		monthlyCost.MonthlyBudgetCents,
	)

	// Get alert status
	warn, critical := domain.ShouldSendAlert(
		spendingPercent,
		h.Config.WarnThresholdPct,
		h.Config.CriticalThresholdPct,
	)
	alertStatus := "normal"
	if critical {
		alertStatus = "critical"
	} else if warn {
		alertStatus = "warning"
	}

	// Get last alert sent time
	var lastAlertSentAt *time.Time
	if warn {
		recentAlert, err := h.Queries.GetRecentSpendingAlert(ctx, repository.GetRecentSpendingAlertParams{
			BusinessID:    business.ID,
			MonthlyCostID: monthlyCost.ID,
			AlertLevel:    "critical",
		})
		if err == nil {
			lastAlertSentAt = &recentAlert.SentAt
		}
	}

	remaining := monthlyCost.MonthlyBudgetCents - monthlyCost.TotalCostCents
	if remaining < 0 {
		remaining = 0
	}

	response := SpendingResponse{
		CurrentCostCents: monthlyCost.TotalCostCents,
		BudgetCents:      monthlyCost.MonthlyBudgetCents,
		SpendingPercent:  spendingPercent,
		RemainingCents:   remaining,
		Month:            monthlyCost.Month,
		AlertStatus:      alertStatus,
		LastAlertSentAt:  lastAlertSentAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSpendingHistory returns paginated monthly spending history
func (h *Handler) GetSpendingHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value("userID").(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get owner's business
	business, err := h.Queries.GetBusinessByOwner(ctx, ownerID)
	if err != nil {
		slog.Error("failed to get business", "error", err)
		http.Error(w, "business not found", http.StatusNotFound)
		return
	}

	limit := int32(10)
	offset := int32(0)

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 32); err == nil {
			limit = int32(parsed)
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.ParseInt(o, 10, 32); err == nil {
			offset = int32(parsed)
		}
	}

	costs, err := h.Queries.ListMonthlyCosts(ctx, repository.ListMonthlyCostsParams{
		BusinessID: business.ID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		slog.Error("failed to list costs", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	items := make([]MonthlyCostItem, len(costs))
	for i, cost := range costs {
		spendingPercent := domain.CalculateSpendingPercentage(
			cost.TotalCostCents,
			cost.MonthlyBudgetCents,
		)
		items[i] = MonthlyCostItem{
			Month:                   cost.Month,
			InfrastructureCostCents: cost.InfrastructureCostCents,
			ClaudeTokens:            cost.ClaudeApiTokens,
			TotalCostCents:          cost.TotalCostCents,
			BudgetCents:             cost.MonthlyBudgetCents,
			SpendingPercent:         spendingPercent,
		}
	}

	response := SpendingHistoryResponse{
		Items:  items,
		Total:  int64(len(items)),
		Limit:  limit,
		Offset: offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAlertHistory returns paginated alert history
func (h *Handler) GetAlertHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value("userID").(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get owner's business
	business, err := h.Queries.GetBusinessByOwner(ctx, ownerID)
	if err != nil {
		slog.Error("failed to get business", "error", err)
		http.Error(w, "business not found", http.StatusNotFound)
		return
	}

	limit := int32(10)
	offset := int32(0)

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 32); err == nil {
			limit = int32(parsed)
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.ParseInt(o, 10, 32); err == nil {
			offset = int32(parsed)
		}
	}

	alerts, err := h.Queries.ListSpendingAlerts(ctx, repository.ListSpendingAlertsParams{
		BusinessID: business.ID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		slog.Error("failed to list alerts", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	items := make([]AlertItem, len(alerts))
	for i, alert := range alerts {
		items[i] = AlertItem{
			ID:               alert.ID,
			AlertLevel:       alert.AlertLevel,
			ThresholdPercent: alert.ThresholdPercent,
			SentAt:           alert.SentAt,
		}
	}

	response := AlertHistoryResponse{
		Items:  items,
		Total:  int64(len(items)),
		Limit:  limit,
		Offset: offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
