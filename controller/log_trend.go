package controller

import (
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
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

func GetLogsTrend(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	trendData, err := model.GetLogsTrend(startTimestamp, endTimestamp)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, trendData)
}