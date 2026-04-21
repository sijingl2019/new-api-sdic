package controller

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func GetDepartmentLogs(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	companyName := c.Query("company_name")
	firstDeptName := c.Query("first_dept_name")
	secondDeptName := c.Query("second_dept_name")
	thirdDeptName := c.Query("third_dept_name")
	sort := c.Query("sort")
	dimensions := c.Query("dimensions")

	logs, total, err := model.GetDepartmentLogs(startTimestamp, endTimestamp, companyName, firstDeptName, secondDeptName, thirdDeptName, pageInfo.GetStartIdx(), pageInfo.GetPageSize(), sort, dimensions)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(logs)
	common.ApiSuccess(c, pageInfo)
}

func ExportDepartmentLogs(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	companyName := c.Query("company_name")
	firstDeptName := c.Query("first_dept_name")
	secondDeptName := c.Query("second_dept_name")
	thirdDeptName := c.Query("third_dept_name")
	dimensions := c.Query("dimensions")

	unitLogs, err := model.GetAllDepartmentLogsForExport(startTimestamp, endTimestamp, companyName, firstDeptName, secondDeptName, thirdDeptName, dimensions)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	personalLogs, err := model.GetPersonalStatLogsForExport(startTimestamp, endTimestamp, companyName, firstDeptName, secondDeptName, thirdDeptName)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet1 := "单位的统计信息"
	f.SetSheetName("Sheet1", sheet1)

	// Parse dimensions to determine which columns to include in export
	dimKeys := parseDimensionKeys(dimensions)
	hasModelName := false
	for _, k := range dimKeys {
		if k == "model_name" {
			hasModelName = true
			break
		}
	}

	// Build headers dynamically
	dimLabelMap := map[string]string{
		"company_name":     "公司",
		"first_dept_name":  "一级部门",
		"second_dept_name": "二级部门",
		"third_dept_name":  "三级部门",
		"model_name":       "模型名称",
	}

	var headers1 []string
	for _, k := range dimKeys {
		headers1 = append(headers1, dimLabelMap[k])
	}
	headers1 = append(headers1, "prompt_tokens", "complete_tokens", "total_tokens", "员工人数", "员工使用人数", "员工未使用人数")

	for i, header := range headers1 {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet1, colName+"1", header)
	}

	// Write data for sheet 1
	for rowIndex, log := range unitLogs {
		row := rowIndex + 2
		col := 1
		for _, k := range dimKeys {
			colName, _ := excelize.ColumnNumberToName(col)
			switch k {
			case "company_name":
				f.SetCellValue(sheet1, fmt.Sprintf("%s%d", colName, row), log.CompanyName)
			case "first_dept_name":
				f.SetCellValue(sheet1, fmt.Sprintf("%s%d", colName, row), log.FirstDeptName)
			case "second_dept_name":
				f.SetCellValue(sheet1, fmt.Sprintf("%s%d", colName, row), log.SecondDeptName)
			case "third_dept_name":
				f.SetCellValue(sheet1, fmt.Sprintf("%s%d", colName, row), log.ThirdDeptName)
			case "model_name":
				f.SetCellValue(sheet1, fmt.Sprintf("%s%d", colName, row), log.ModelName)
			}
			col++
		}
		colName, _ := excelize.ColumnNumberToName(col)
		f.SetCellValue(sheet1, fmt.Sprintf("%s%d", colName, row), log.PromptTokens)
		col++
		colName, _ = excelize.ColumnNumberToName(col)
		f.SetCellValue(sheet1, fmt.Sprintf("%s%d", colName, row), log.CompleteTokens)
		col++
		colName, _ = excelize.ColumnNumberToName(col)
		f.SetCellValue(sheet1, fmt.Sprintf("%s%d", colName, row), log.TotalTokens)
		col++
		colName, _ = excelize.ColumnNumberToName(col)
		f.SetCellValue(sheet1, fmt.Sprintf("%s%d", colName, row), log.EmployeeCount)
		col++
		colName, _ = excelize.ColumnNumberToName(col)
		f.SetCellValue(sheet1, fmt.Sprintf("%s%d", colName, row), log.UseCount)
		col++
		colName, _ = excelize.ColumnNumberToName(col)
		f.SetCellValue(sheet1, fmt.Sprintf("%s%d", colName, row), log.NotUseCount)
	}

	// Only add personal sheet if model_name is not a dimension (personal stats don't have model breakdown)
	if !hasModelName {
		sheet2 := "个人统计信息"
		f.NewSheet(sheet2)

		// Write headers for sheet 2
		headers2 := []string{"姓名", "一级部门", "二级部门", "三级部门", "prompt_tokens", "complete_tokens", "total_tokens"}
		for i, header := range headers2 {
			colName, _ := excelize.ColumnNumberToName(i + 1)
			f.SetCellValue(sheet2, colName+"1", header)
		}

		// Write data for sheet 2
		for rowIndex, log := range personalLogs {
			row := rowIndex + 2
			f.SetCellValue(sheet2, fmt.Sprintf("A%d", row), log.Name)
			f.SetCellValue(sheet2, fmt.Sprintf("B%d", row), log.FirstDeptName)
			f.SetCellValue(sheet2, fmt.Sprintf("C%d", row), log.SecondDeptName)
			f.SetCellValue(sheet2, fmt.Sprintf("D%d", row), log.ThirdDeptName)
			f.SetCellValue(sheet2, fmt.Sprintf("E%d", row), log.PromptTokens)
			f.SetCellValue(sheet2, fmt.Sprintf("F%d", row), log.CompleteTokens)
			f.SetCellValue(sheet2, fmt.Sprintf("G%d", row), log.TotalTokens)
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		common.ApiError(c, err)
		return
	}

	fileName := fmt.Sprintf("department_logs_%s.xlsx", time.Now().Format("20060102150405"))
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// parseDimensionKeys parses comma-separated dimension string for the controller.
func parseDimensionKeys(dimensions string) []string {
	validKeys := map[string]bool{
		"company_name":     true,
		"first_dept_name":  true,
		"second_dept_name": true,
		"third_dept_name":  true,
		"model_name":       true,
	}
	if dimensions == "" {
		return []string{"company_name", "first_dept_name", "second_dept_name", "third_dept_name"}
	}
	parts := strings.Split(dimensions, ",")
	var keys []string
	seen := make(map[string]bool)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if validKeys[p] && !seen[p] {
			keys = append(keys, p)
			seen[p] = true
		}
	}
	if len(keys) == 0 {
		return []string{"company_name", "first_dept_name", "second_dept_name", "third_dept_name"}
	}
	return keys
}
