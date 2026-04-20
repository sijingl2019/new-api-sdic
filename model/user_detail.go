package model

import (
	"os"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
)

type UserDetail struct {
	Username          string    `gorm:"type:varchar(255);not null" json:"username"`
	Name              string    `gorm:"type:varchar(255)" json:"name"`
	Mobile            string    `gorm:"type:varchar(32)" json:"mobile"`
	CompanyName       string    `gorm:"type:varchar(255)" json:"companyName"`
	FirstDeptName     string    `gorm:"type:varchar(255)" json:"firstDeptName"`
	SecondDeptName    string    `gorm:"column:second_dept_name;type:varchar(255)" json:"twoDeptName"`
	ThirdDeptName     string    `gorm:"column:third_dept_name;type:varchar(255)" json:"threeDeptName"`
	FourthDeptName    string    `gorm:"column:fourth_dept_name;type:varchar(255)" json:"fourDeptName"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"-"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"-"`
	JobCategory       string    `gorm:"type:varchar(255)" json:"jobCategory"`
	JobCategorySecond string    `gorm:"type:varchar(255)" json:"jobCategorySecond"`
	Status            string    `gorm:"type:varchar(50)" json:"status"`
	IsExclude         int       `gorm:"default:0" json:"isExclude"`
	ExcludeReason     string    `gorm:"type:text" json:"excludeReason"`
}

func (UserDetail) TableName() string {
	return "users_detail"
}

func InitUserDetail() error {
	var count int64
	err := DB.Model(&UserDetail{}).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // Already initialized
	}

	// Read json file
	file, err := os.Open("/users_dept.json")
	if err != nil {
		if os.IsNotExist(err) {
			common.SysLog("users_dept.json not found, skipping users_detail initialization")
			return nil
		}
		return err
	}
	defer file.Close()

	var result struct {
		Data []UserDetail `json:"data"`
	}

	// Rule 1: Use common.DecodeJson instead of io.ReadAll + Unmarshal
	err = common.DecodeJson(file, &result)
	if err != nil {
		return err
	}

	if len(result.Data) > 0 {
		// Use batch insert for performance
		chunkSize := 1000
		for i := 0; i < len(result.Data); i += chunkSize {
			end := i + chunkSize
			if end > len(result.Data) {
				end = len(result.Data)
			}
			err = DB.CreateInBatches(result.Data[i:end], chunkSize).Error
			if err != nil {
				common.SysError("failed to insert users_detail: " + err.Error())
				return err
			}
		}
		common.SysLog("users_detail initialized with " + strconv.Itoa(len(result.Data)) + " records")
	}

	return nil
}

func GetUserDetailByPhone(phone string) (*UserDetail, error) {
	var userDetail UserDetail
	err := DB.Where("mobile = ?", phone).First(&userDetail).Error
	return &userDetail, err
}
