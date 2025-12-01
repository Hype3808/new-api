package model

import (
	"errors"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// CheckinLog represents a user check-in record
type CheckinLog struct {
	Id          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId      int       `json:"user_id" gorm:"not null;uniqueIndex:idx_user_date"`
	RewardQuota int       `json:"reward_quota" gorm:"not null"`
	QuotaBefore int       `json:"quota_before" gorm:"not null"`
	QuotaAfter  int       `json:"quota_after" gorm:"not null"`
	CheckinDate string    `json:"checkin_date" gorm:"type:varchar(10);not null;uniqueIndex:idx_user_date"` // YYYY-MM-DD format
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the table name for CheckinLog
func (CheckinLog) TableName() string {
	return "checkin_logs"
}

// Insert creates a new check-in log record
func (log *CheckinLog) Insert() error {
	if log.UserId == 0 {
		return errors.New("user_id is required")
	}
	if log.CheckinDate == "" {
		return errors.New("checkin_date is required")
	}
	return DB.Create(log).Error
}

// GetCheckinLogByUserIdAndDate retrieves a check-in log for a specific user and date
func GetCheckinLogByUserIdAndDate(userId int, date string) (*CheckinLog, error) {
	if userId == 0 {
		return nil, errors.New("user_id is required")
	}
	if date == "" {
		return nil, errors.New("date is required")
	}
	var log CheckinLog
	err := DB.Where("user_id = ? AND checkin_date = ?", userId, date).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// GetLastCheckinByUserId retrieves the most recent check-in log for a user
func GetLastCheckinByUserId(userId int) (*CheckinLog, error) {
	if userId == 0 {
		return nil, errors.New("user_id is required")
	}
	var log CheckinLog
	err := DB.Where("user_id = ?", userId).Order("created_at DESC").First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// GetCheckinLogsByUserId retrieves all check-in logs for a user with pagination
func GetCheckinLogsByUserId(userId int, pageInfo *common.PageInfo) ([]*CheckinLog, int64, error) {
	if userId == 0 {
		return nil, 0, errors.New("user_id is required")
	}
	var logs []*CheckinLog
	var total int64

	err := DB.Model(&CheckinLog{}).Where("user_id = ?", userId).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = DB.Where("user_id = ?", userId).
		Order("created_at DESC").
		Limit(pageInfo.GetPageSize()).
		Offset(pageInfo.GetStartIdx()).
		Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// HasCheckedInToday checks if a user has already checked in today
// Uses the provided date string in YYYY-MM-DD format
func HasCheckedInToday(userId int, todayDate string) (bool, error) {
	if userId == 0 {
		return false, errors.New("user_id is required")
	}
	var count int64
	err := DB.Model(&CheckinLog{}).Where("user_id = ? AND checkin_date = ?", userId, todayDate).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
