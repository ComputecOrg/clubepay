package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/repository"
)

// SendSpendingAlerts checks all businesses and sends alerts if thresholds are reached
func (h *Handler) SendSpendingAlerts(ctx context.Context) error {
	// Get all businesses
	businesses, err := h.Queries.ListAllBusinesses(ctx)
	if err != nil {
		return fmt.Errorf("failed to list businesses: %w", err)
	}

	currentMonth := domain.GetCurrentMonth()

	for _, business := range businesses {
		// Get or create monthly cost
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
				slog.Error("failed to get/create monthly cost", "businessID", business.ID, "error", err)
				continue
			}
		}

		// Calculate percentage
		spendingPercent := domain.CalculateSpendingPercentage(
			monthlyCost.TotalCostCents,
			monthlyCost.MonthlyBudgetCents,
		)

		// Check if alert should be sent
		warn, critical := domain.ShouldSendAlert(
			spendingPercent,
			h.Config.WarnThresholdPct,
			h.Config.CriticalThresholdPct,
		)

		if critical {
			// Check if critical alert already sent this month
			existingAlert, _ := h.Queries.GetRecentSpendingAlert(ctx, repository.GetRecentSpendingAlertParams{
				BusinessID:    business.ID,
				MonthlyCostID: monthlyCost.ID,
				AlertLevel:    "critical",
			})

			// Only send if no recent critical alert
			if existingAlert.ID == 0 {
				h.sendCriticalAlert(ctx, business, monthlyCost, spendingPercent)
			}
		} else if warn {
			// Check if warning already sent
			existingAlert, _ := h.Queries.GetRecentSpendingAlert(ctx, repository.GetRecentSpendingAlertParams{
				BusinessID:    business.ID,
				MonthlyCostID: monthlyCost.ID,
				AlertLevel:    "warning",
			})

			if existingAlert.ID == 0 {
				h.sendWarningAlert(ctx, business, monthlyCost, spendingPercent)
			}
		}
	}

	return nil
}

func (h *Handler) sendWarningAlert(ctx context.Context, business *repository.Business, cost *repository.MonthlyCost, percent int) {
	// Create alert record
	_, err := h.Queries.CreateSpendingAlert(ctx, repository.CreateSpendingAlertParams{
		BusinessID:       business.ID,
		MonthlyCostID:    cost.ID,
		AlertLevel:       "warning",
		ThresholdPercent: int32(h.Config.WarnThresholdPct),
	})
	if err != nil {
		slog.Error("failed to create warning alert", "error", err)
		return
	}

	// Send email
	subject := fmt.Sprintf("⚠️ ClubePay: Seu negócio atingiu %d%% do orçamento", percent)
	body := fmt.Sprintf(
		"Olá,\n\nSeu negócio %s atingiu %d%% do orçamento mensal (%d%% de limite).\n\n"+
			"Gastos do mês: R$%.2f\n"+
			"Orçamento: R$%.2f\n\nAcesse o painel para mais detalhes.",
		business.Name,
		percent,
		h.Config.WarnThresholdPct,
		float64(cost.TotalCostCents)/100,
		float64(cost.MonthlyBudgetCents)/100,
	)

	err = h.Email.Send(h.Config.SpendingAlertEmail, subject, body)
	if err != nil {
		slog.Error("failed to send warning alert email", "error", err)
	}
}

func (h *Handler) sendCriticalAlert(ctx context.Context, business *repository.Business, cost *repository.MonthlyCost, percent int) {
	// Create alert record
	_, err := h.Queries.CreateSpendingAlert(ctx, repository.CreateSpendingAlertParams{
		BusinessID:       business.ID,
		MonthlyCostID:    cost.ID,
		AlertLevel:       "critical",
		ThresholdPercent: int32(h.Config.CriticalThresholdPct),
	})
	if err != nil {
		slog.Error("failed to create critical alert", "error", err)
		return
	}

	// Send email with escalation
	subject := fmt.Sprintf("🚨 CRÍTICO: ClubePay - Seu negócio atingiu %d%% do orçamento", percent)
	body := fmt.Sprintf(
		"ALERTA CRÍTICO!\n\nSeu negócio %s atingiu %d%% do orçamento mensal.\n\n"+
			"Gastos do mês: R$%.2f\n"+
			"Orçamento: R$%.2f\n\nTome ação imediatamente.",
		business.Name,
		percent,
		float64(cost.TotalCostCents)/100,
		float64(cost.MonthlyBudgetCents)/100,
	)

	err = h.Email.Send(h.Config.SpendingAlertEmail, subject, body)
	if err != nil {
		slog.Error("failed to send critical alert email", "error", err)
	}
}
