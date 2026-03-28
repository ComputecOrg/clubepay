package domain

import "time"

// CheckUsageLimit verifies if current usage count has reached the limit.
func CheckUsageLimit(limitType string, limitCount int, currentCount int64) error {
	if currentCount >= int64(limitCount) {
		return ErrUsageLimitReached
	}
	return nil
}

// DailyRange returns the start (inclusive) and end (exclusive) of the current day.
func DailyRange(now time.Time) (start, end time.Time) {
	start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end = start.AddDate(0, 0, 1)
	return
}

// MonthlyRange returns the start (inclusive) and end (exclusive) of the current month.
func MonthlyRange(now time.Time) (start, end time.Time) {
	start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end = start.AddDate(0, 1, 0)
	return
}
