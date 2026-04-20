package controller

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
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

	logs, total, err := model.GetDepartmentLogs(startTimestamp, endTimestamp, companyName, firstDeptName, secondDeptName, thirdDeptName, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
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

	unitLogs, err := model.GetAllDepartmentLogsForExport(startTimestamp, endTimestamp, companyName, firstDeptName, secondDeptName, thirdDeptName)
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
	
	// Write headers for sheet 1
	headers1 := []string{"公司", "一级部门", "二级部门", "三级部门", "prompt_tokens", "complete_tokens", "员工人数", "员工使用人数", "员工未使用人数"}
	for i, header := range headers1 {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet1, colName+"1", header)
	}

	// Write data for sheet 1
	for rowIndex, log := range unitLogs {
		row := rowIndex + 2
		f.SetCellValue(sheet1, fmt.Sprintf("A%d", row), log.CompanyName)
		f.SetCellValue(sheet1, fmt.Sprintf("B%d", row), log.FirstDeptName)
		f.SetCellValue(sheet1, fmt.Sprintf("C%d", row), log.SecondDeptName)
		f.SetCellValue(sheet1, fmt.Sprintf("D%d", row), log.ThirdDeptName)
		f.SetCellValue(sheet1, fmt.Sprintf("E%d", row), log.PromptTokens)
		f.SetCellValue(sheet1, fmt.Sprintf("F%d", row), log.CompleteTokens)
		f.SetCellValue(sheet1, fmt.Sprintf("G%d", row), log.EmployeeCount)
		f.SetCellValue(sheet1, fmt.Sprintf("H%d", row), log.UseCount)
		f.SetCellValue(sheet1, fmt.Sprintf("I%d", row), log.NotUseCount)
	}

	sheet2 := "个人统计信息"
	f.NewSheet(sheet2)

	// Write headers for sheet 2
	headers2 := []string{"姓名", "一级部门", "二级部门", "三级部门", "prompt_tokens", "complete_tokens"}
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
