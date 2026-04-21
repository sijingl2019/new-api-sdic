package model

import "strings"

type DepartmentLog struct {
	CompanyName    string `json:"company_name,omitempty" gorm:"column:company_name"`
	FirstDeptName  string `json:"first_dept_name,omitempty" gorm:"column:first_dept_name"`
	SecondDeptName string `json:"second_dept_name,omitempty" gorm:"column:second_dept_name"`
	ThirdDeptName  string `json:"third_dept_name,omitempty" gorm:"column:third_dept_name"`
	ModelName      string `json:"model_name,omitempty" gorm:"column:model_name"`
	PromptTokens   int    `json:"prompt_tokens" gorm:"column:prompt_tokens"`
	CompleteTokens int    `json:"complete_tokens" gorm:"column:complete_tokens"`
	TotalTokens    int    `json:"total_tokens" gorm:"column:total_tokens"`
	EmployeeCount  int    `json:"employee_count" gorm:"column:employee_count"`
	UseCount       int    `json:"use_count" gorm:"column:use_count"`
	NotUseCount    int    `json:"not_use_count" gorm:"column:not_use_count"`
}

type PersonalStatLog struct {
	Name           string `json:"name" gorm:"column:name"`
	FirstDeptName  string `json:"first_dept_name" gorm:"column:first_dept_name"`
	SecondDeptName string `json:"second_dept_name" gorm:"column:second_dept_name"`
	ThirdDeptName  string `json:"third_dept_name" gorm:"column:third_dept_name"`
	PromptTokens   int    `json:"prompt_tokens" gorm:"column:prompt_tokens"`
	CompleteTokens int    `json:"complete_tokens" gorm:"column:complete_tokens"`
	TotalTokens    int    `json:"total_tokens" gorm:"column:total_tokens"`
}

// validDimensions defines the allowed dimension keys and their corresponding SQL column expressions.
var validDimensions = map[string]string{
	"company_name":    "ud.company_name",
	"first_dept_name": "ud.first_dept_name",
	"second_dept_name": "ud.second_dept_name",
	"third_dept_name": "ud.third_dept_name",
	"model_name":      "l.model_name",
}

// parseDimensions parses a comma-separated dimension string and returns
// the list of valid dimension keys in a fixed order and whether model_name is included.
func parseDimensions(dimensions string) (keys []string, hasModelName bool) {
	order := []string{"company_name", "first_dept_name", "second_dept_name", "third_dept_name", "model_name"}
	if dimensions == "" {
		// default: all department dimensions, no model
		return order[:4], false
	}
	parts := strings.Split(dimensions, ",")
	seen := make(map[string]bool)
	for _, p := range parts {
		seen[strings.TrimSpace(p)] = true
	}
	for _, k := range order {
		if seen[k] {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return order[:4], false
	}
	hasModelName = seen["model_name"]
	return keys, hasModelName
}

func GetDepartmentLogs(startTimestamp int64, endTimestamp int64, companyName, firstDeptName, secondDeptName, thirdDeptName string, startIdx int, num int, sort string, dimensions string) ([]*DepartmentLog, int64, error) {
	var logs []*DepartmentLog
	var total int64

	dimKeys, hasModelName := parseDimensions(dimensions)

	// Build group columns
	var groupColParts []string
	for _, k := range dimKeys {
		groupColParts = append(groupColParts, validDimensions[k])
	}
	groupCols := strings.Join(groupColParts, ", ")

	// Build select columns
	var selectParts []string
	for _, k := range dimKeys {
		selectParts = append(selectParts, validDimensions[k])
	}

	baseDb := DB.Table("users_detail ud").
		Joins("LEFT JOIN users u ON ud.username = u.username")

	joinLogCondition := "LEFT JOIN logs l ON u.id = l.user_id"
	var joinArgs []interface{}
	joinConditionParts := []string{"l.type = ?"}
	joinArgs = append(joinArgs, LogTypeConsume)

	if startTimestamp != 0 {
		joinConditionParts = append(joinConditionParts, "l.created_at >= ?")
		joinArgs = append(joinArgs, startTimestamp)
	}
	if endTimestamp != 0 {
		joinConditionParts = append(joinConditionParts, "l.created_at <= ?")
		joinArgs = append(joinArgs, endTimestamp)
	}
	for _, part := range joinConditionParts {
		joinLogCondition += " AND " + part
	}

	db := baseDb.Joins(joinLogCondition, joinArgs...)

	if companyName != "" {
		db = db.Where("ud.company_name = ?", companyName)
	}
	if firstDeptName != "" {
		db = db.Where("ud.first_dept_name = ?", firstDeptName)
	}
	if secondDeptName != "" {
		db = db.Where("ud.second_dept_name = ?", secondDeptName)
	}
	if thirdDeptName != "" {
		db = db.Where("ud.third_dept_name = ?", thirdDeptName)
	}

	// Count query
	countDb := DB.Table("users_detail ud")
	if hasModelName {
		// When model_name dimension is selected, count requires joining logs
		countJoinCondition := "LEFT JOIN users u ON ud.username = u.username"
		countDb = countDb.Joins(countJoinCondition)
		countLogJoin := "LEFT JOIN logs l ON u.id = l.user_id"
		var countJoinArgs []interface{}
		countLogJoinParts := []string{"l.type = ?"}
		countJoinArgs = append(countJoinArgs, LogTypeConsume)
		if startTimestamp != 0 {
			countLogJoinParts = append(countLogJoinParts, "l.created_at >= ?")
			countJoinArgs = append(countJoinArgs, startTimestamp)
		}
		if endTimestamp != 0 {
			countLogJoinParts = append(countLogJoinParts, "l.created_at <= ?")
			countJoinArgs = append(countJoinArgs, endTimestamp)
		}
		for _, part := range countLogJoinParts {
			countLogJoin += " AND " + part
		}
		countDb = countDb.Joins(countLogJoin, countJoinArgs...)
		// Filter out NULL model_name (users with no logs)
		countDb = countDb.Where("l.model_name IS NOT NULL AND l.model_name <> ''")
	}
	if companyName != "" {
		countDb = countDb.Where("ud.company_name = ?", companyName)
	}
	if firstDeptName != "" {
		countDb = countDb.Where("ud.first_dept_name = ?", firstDeptName)
	}
	if secondDeptName != "" {
		countDb = countDb.Where("ud.second_dept_name = ?", secondDeptName)
	}
	if thirdDeptName != "" {
		countDb = countDb.Where("ud.third_dept_name = ?", thirdDeptName)
	}
	countDb = countDb.Select(groupCols).Group(groupCols)

	err := DB.Table("(?) AS temp", countDb).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return logs, 0, nil
	}

	// Build aggregation select
	selectParts = append(selectParts,
		"COUNT(DISTINCT ud.username) as employee_count",
		"COUNT(DISTINCT l.user_id) as use_count",
		"COUNT(DISTINCT ud.username) - COUNT(DISTINCT l.user_id) as not_use_count",
		"COALESCE(SUM(l.prompt_tokens), 0) as prompt_tokens",
		"COALESCE(SUM(l.completion_tokens), 0) as complete_tokens",
		"COALESCE(SUM(l.prompt_tokens), 0) + COALESCE(SUM(l.completion_tokens), 0) as total_tokens",
	)

	db = db.Select(strings.Join(selectParts, ", ")).Group(groupCols)

	if sort != "" {
		db = db.Order(sort)
	} else {
		db = db.Order("prompt_tokens DESC")
	}

	if num > 0 {
		db = db.Limit(num).Offset(startIdx)
	}

	err = db.Find(&logs).Error
	return logs, total, err
}

func GetAllDepartmentLogsForExport(startTimestamp int64, endTimestamp int64, companyName, firstDeptName, secondDeptName, thirdDeptName string, dimensions string) ([]*DepartmentLog, error) {
	logs, _, err := GetDepartmentLogs(startTimestamp, endTimestamp, companyName, firstDeptName, secondDeptName, thirdDeptName, 0, 0, "prompt_tokens DESC", dimensions)
	return logs, err
}

func GetPersonalStatLogsForExport(startTimestamp int64, endTimestamp int64, companyName, firstDeptName, secondDeptName, thirdDeptName string) ([]*PersonalStatLog, error) {
	var logs []*PersonalStatLog

	baseDb := DB.Table("users_detail ud").
		Joins("LEFT JOIN users u ON ud.username = u.username")

	joinLogCondition := "LEFT JOIN logs l ON u.id = l.user_id"
	var joinArgs []interface{}
	joinConditionParts := []string{"l.type = ?"}
	joinArgs = append(joinArgs, LogTypeConsume)

	if startTimestamp != 0 {
		joinConditionParts = append(joinConditionParts, "l.created_at >= ?")
		joinArgs = append(joinArgs, startTimestamp)
	}
	if endTimestamp != 0 {
		joinConditionParts = append(joinConditionParts, "l.created_at <= ?")
		joinArgs = append(joinArgs, endTimestamp)
	}
	for _, part := range joinConditionParts {
		joinLogCondition += " AND " + part
	}

	db := baseDb.Joins(joinLogCondition, joinArgs...)

	if companyName != "" {
		db = db.Where("ud.company_name = ?", companyName)
	}
	if firstDeptName != "" {
		db = db.Where("ud.first_dept_name = ?", firstDeptName)
	}
	if secondDeptName != "" {
		db = db.Where("ud.second_dept_name = ?", secondDeptName)
	}
	if thirdDeptName != "" {
		db = db.Where("ud.third_dept_name = ?", thirdDeptName)
	}

	groupCols := "ud.name, ud.first_dept_name, ud.second_dept_name, ud.third_dept_name"

	db = db.Select(`
		ud.name, 
		ud.first_dept_name, 
		ud.second_dept_name, 
		ud.third_dept_name,
		COALESCE(SUM(l.prompt_tokens), 0) as prompt_tokens,
		COALESCE(SUM(l.completion_tokens), 0) as complete_tokens,
		COALESCE(SUM(l.prompt_tokens), 0) + COALESCE(SUM(l.completion_tokens), 0) as total_tokens
	`).Group(groupCols).Order("total_tokens DESC")

	err := db.Find(&logs).Error
	return logs, err
}
