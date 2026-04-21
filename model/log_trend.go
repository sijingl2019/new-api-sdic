package model

import (
	"time"

	"github.com/QuantumNous/new-api/common"
)

type DailyStat struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type TokenStat struct {
	Date   string `json:"date"`
	Tokens int64  `json:"tokens"`
}

type TrendData struct {
	DailyRegistrations []DailyStat `json:"daily_registrations"`
	DailyCalls         []DailyStat `json:"daily_calls"`
	DailyTokens        []TokenStat `json:"daily_tokens"`
}

func GetLogsTrend(startTimestamp int64, endTimestamp int64) (*TrendData, error) {
	trend := &TrendData{
		DailyRegistrations: make([]DailyStat, 0),
		DailyCalls:         make([]DailyStat, 0),
		DailyTokens:        make([]TokenStat, 0),
	}

	// Default time range: start of this week to now
	if startTimestamp == 0 {
		startTimestamp = getStartOfWeek()
	}
	if endTimestamp == 0 {
		endTimestamp = time.Now().Unix()
	}

	// Get daily user registrations
	registrations, err := getDailyRegistrations(startTimestamp, endTimestamp)
	if err != nil {
		return nil, err
	}
	trend.DailyRegistrations = registrations

	// Get daily API call counts
	calls, err := getDailyCalls(startTimestamp, endTimestamp)
	if err != nil {
		return nil, err
	}
	trend.DailyCalls = calls

	// Get daily token consumption
	tokens, err := getDailyTokens(startTimestamp, endTimestamp)
	if err != nil {
		return nil, err
	}
	trend.DailyTokens = tokens

	return trend, nil
}

func getStartOfWeek() int64 {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	startOfWeek := now.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
	return startOfWeek.Unix()
}

func getDateFormatExpr(table string, timestampCol string) string {
	if common.UsingMySQL {
		return "date_format(from_unixtime(" + timestampCol + "), '%Y-%m-%d')"
	} else if common.UsingPostgreSQL {
		return "to_char(to_timestamp(" + timestampCol + "), 'YYYY-MM-DD')"
	} else {
		// SQLite
		return "date(" + timestampCol + ", 'unixepoch')"
	}
}

func getDailyRegistrations(startTimestamp int64, endTimestamp int64) ([]DailyStat, error) {
	var results []DailyStat

	dateExpr := getDateFormatExpr("users", "created_at")

	err := DB.Model(&User{}).
		Select(dateExpr+" as date, count(*) as count").
		Where("created_at >= ? AND created_at <= ?", startTimestamp, endTimestamp).
		Group(dateExpr).
		Order("date asc").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func getDailyCalls(startTimestamp int64, endTimestamp int64) ([]DailyStat, error) {
	var results []DailyStat

	dateExpr := getDateFormatExpr("logs", "created_at")

	err := LOG_DB.Model(&Log{}).
		Select(dateExpr+" as date, count(*) as count").
		Where("created_at >= ? AND created_at <= ?", startTimestamp, endTimestamp).
		Group(dateExpr).
		Order("date asc").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func getDailyTokens(startTimestamp int64, endTimestamp int64) ([]TokenStat, error) {
	var results []TokenStat

	dateExpr := getDateFormatExpr("logs", "created_at")

	err := LOG_DB.Model(&Log{}).
		Select(dateExpr+" as date, (sum(prompt_tokens) + sum(completion_tokens)) as tokens").
		Where("created_at >= ? AND created_at <= ?", startTimestamp, endTimestamp).
		Group(dateExpr).
		Order("date asc").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}