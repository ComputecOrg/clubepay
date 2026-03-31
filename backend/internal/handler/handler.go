package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/clubepay/backend/internal/config"
	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/email"
	"github.com/clubepay/backend/internal/psp"
	"github.com/clubepay/backend/internal/repository"
	"github.com/clubepay/backend/internal/service"
)

type Handler struct {
	Auth          *service.AuthService
	Business      *service.BusinessService
	Plans         *service.PlanService
	Subscriptions *service.SubscriptionService
	Usage         *service.UsageService
	Referrals     *service.ReferralService
	Config        *config.Config
	// Keep these for webhook, cron, public, and search handlers that need direct access
	Queries *repository.Queries
	PSP     psp.PSP
	Email   email.Sender
}

func New(q *repository.Queries, cfg *config.Config, p psp.PSP, e email.Sender) *Handler {
	return &Handler{
		Auth:          service.NewAuthService(q, cfg, e),
		Business:      service.NewBusinessService(q),
		Plans:         service.NewPlanService(q),
		Subscriptions: service.NewSubscriptionService(q, p, e),
		Usage:         service.NewUsageService(q),
		Referrals:     service.NewReferralService(q),
		Config:        cfg,
		Queries:       q,
		PSP:           p,
		Email:         e,
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func readJSON(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func pgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func pgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func handleServiceError(w http.ResponseWriter, err error) {
	var svcErr *domain.ServiceError
	if errors.As(err, &svcErr) {
		writeError(w, svcErr.Code, svcErr.Message)
		return
	}
	slog.Error("unhandled service error", "error", err)
	writeError(w, http.StatusInternalServerError, "erro interno")
}

// SendSpendingAlerts checks all businesses and sends alerts if thresholds are reached
func (h *Handler) SendSpendingAlerts(ctx context.Context) error {
	// Get all businesses
	businesses, err := h.Queries.ListAllBusinesses(ctx)
	if err != nil {
		return err
	}

	currentMonth := domain.GetCurrentMonth()

	for _, business := range businesses {
		// Get or create monthly cost
		monthlyCost, err := h.Queries.GetOrCreateMonthlyCost(ctx, repository.GetOrCreateMonthlyCostParams{
			BusinessID: business.ID,
			Month:      pgTypeDate(currentMonth),
		})
		if err != nil {
			// Try to create if not found
			monthlyCost, err = h.Queries.CreateMonthlyCost(ctx, repository.CreateMonthlyCostParams{
				BusinessID:         business.ID,
				Month:              pgTypeDate(currentMonth),
				MonthlyBudgetCents: h.Config.MonthlyBudgetCents,
			})
			if err != nil {
				slog.Error("failed to get or create monthly cost", "businessID", business.ID, "error", err)
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
				h.sendCriticalAlert(ctx, &business, &monthlyCost, spendingPercent)
			}
		} else if warn {
			// Check if warning already sent
			existingAlert, _ := h.Queries.GetRecentSpendingAlert(ctx, repository.GetRecentSpendingAlertParams{
				BusinessID:    business.ID,
				MonthlyCostID: monthlyCost.ID,
				AlertLevel:    "warning",
			})

			if existingAlert.ID == 0 {
				h.sendWarningAlert(ctx, &business, &monthlyCost, spendingPercent)
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
		slog.Error("failed to create alert", "error", err)
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
		slog.Error("failed to send alert email", "error", err)
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
		slog.Error("failed to create alert", "error", err)
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
