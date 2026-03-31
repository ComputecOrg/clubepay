package domain

import "time"

// CalculateSpendingPercentage returns the percentage of budget spent (0-100+)
func CalculateSpendingPercentage(currentCostCents, budgetCents int64) int {
	if budgetCents == 0 {
		return 0
	}
	return int((currentCostCents * 100) / budgetCents)
}

// ShouldSendAlert determines if warning or critical alerts should be sent
func ShouldSendAlert(currentPercentage, warnThreshold, criticalThreshold int) (warn, critical bool) {
	if currentPercentage >= criticalThreshold {
		return true, true
	}
	if currentPercentage >= warnThreshold {
		return true, false
	}
	return false, false
}

// GetCurrentMonth returns the first day of the current month
func GetCurrentMonth() time.Time {
	return GetMonthStart(time.Now())
}

// GetMonthStart returns the first day of the month for a given date
func GetMonthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}
