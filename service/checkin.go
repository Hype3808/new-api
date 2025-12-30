package service

import (
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/checkin_setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
)

// formatQuotaWithCurrency converts quota (tokens) to display amount with currency symbol
func formatQuotaWithCurrency(quota int) string {
	quotaPerUnit := common.QuotaPerUnit
	usdAmount := float64(quota) / quotaPerUnit

	displayType := operation_setting.GetQuotaDisplayType()
	if displayType == operation_setting.QuotaDisplayTypeTokens {
		return fmt.Sprintf("%d", quota)
	}

	symbol := operation_setting.GetCurrencySymbol()
	
	// Get exchange rate for CNY or Custom currency
	var displayAmount float64
	if displayType == operation_setting.QuotaDisplayTypeCNY {
		displayAmount = usdAmount * operation_setting.USDExchangeRate
	} else if displayType == operation_setting.QuotaDisplayTypeCustom {
		generalSetting := operation_setting.GetGeneralSetting()
		displayAmount = usdAmount * generalSetting.CustomCurrencyExchangeRate
	} else {
		// USD
		displayAmount = usdAmount
	}

	return fmt.Sprintf("%s%.2f", symbol, displayAmount)
}

const (
	// CheckinCacheKeyFmt is the Redis key format for checkin status
	// Format: checkin:{userId}:{date}
	CheckinCacheKeyFmt = "checkin:%d:%s"
	// CheckinCacheDuration is the TTL for checkin cache (25 hours to cover timezone differences)
	CheckinCacheDuration = 25 * time.Hour
)

// GetTodayDateString returns today's date string in YYYY-MM-DD format (UTC+8)
func GetTodayDateString() string {
	// Use UTC+8 timezone for date calculation
	loc := time.FixedZone("UTC+8", 8*60*60)
	return time.Now().In(loc).Format("2006-01-02")
}

// GetNextCheckinTime returns the timestamp of next available checkin time (00:00 UTC+8)
func GetNextCheckinTime() int64 {
	loc := time.FixedZone("UTC+8", 8*60*60)
	now := time.Now().In(loc)
	// Get tomorrow's 00:00
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
	return tomorrow.Unix()
}

// getCheckinCacheKey generates the Redis cache key for checkin status
func getCheckinCacheKey(userId int, date string) string {
	return fmt.Sprintf(CheckinCacheKeyFmt, userId, date)
}

// HasCheckedInToday checks if a user has already checked in today
// Always checks database first for accuracy, then updates cache
func HasCheckedInToday(userId int) (bool, error) {
	todayDate := GetTodayDateString()

	// Always check database for accuracy (cache may be stale)
	hasCheckedIn, err := model.HasCheckedInToday(userId, todayDate)
	if err != nil {
		return false, err
	}

	// Update cache based on database result
	if common.RedisEnabled {
		cacheKey := getCheckinCacheKey(userId, todayDate)
		if hasCheckedIn {
			_ = common.RedisSet(cacheKey, "1", CheckinCacheDuration)
		} else {
			// Clear any stale cache
			_ = common.RedisDel(cacheKey)
		}
	}

	return hasCheckedIn, nil
}

// CanCheckin checks if a user can perform checkin
// Returns (canCheckin, reason, error)
func CanCheckin(userId int) (bool, string, error) {
	setting := checkin_setting.GetCheckinSetting()

	// Check if feature is enabled
	if !setting.Enabled {
		return false, "签到功能未启用", nil
	}

	// Check if already checked in today
	hasCheckedIn, err := HasCheckedInToday(userId)
	if err != nil {
		return false, "", err
	}
	if hasCheckedIn {
		return false, "今日已签到，请明天再来", nil
	}

	// Check quota threshold
	userQuota, err := model.GetUserQuota(userId, false)
	if err != nil {
		return false, "", err
	}
	if userQuota >= setting.QuotaThreshold {
		return false, fmt.Sprintf("当前额度超过签到阈值（%s），无法签到", formatQuotaWithCurrency(setting.QuotaThreshold)), nil
	}

	return true, "", nil
}


// DoCheckin performs the checkin operation for a user
// Returns the checkin result with reward information
func DoCheckin(userId int) (*dto.CheckinResult, error) {
	setting := checkin_setting.GetCheckinSetting()

	// Check if feature is enabled
	if !setting.Enabled {
		return &dto.CheckinResult{
			Success: false,
			Message: "签到功能未启用",
		}, nil
	}

	todayDate := GetTodayDateString()

	// CRITICAL: Check database directly to prevent race conditions
	hasCheckedIn, err := model.HasCheckedInToday(userId, todayDate)
	if err != nil {
		return nil, err
	}
	if hasCheckedIn {
		return &dto.CheckinResult{
			Success: false,
			Message: "今日已签到，请明天再来",
		}, nil
	}

	// Get current quota from database (not cache)
	quotaBefore, err := model.GetUserQuota(userId, true)
	if err != nil {
		return nil, err
	}

	// Check quota threshold
	if quotaBefore >= setting.QuotaThreshold {
		return &dto.CheckinResult{
			Success: false,
			Message: fmt.Sprintf("当前额度超过签到阈值（%s），无法签到", formatQuotaWithCurrency(setting.QuotaThreshold)),
		}, nil
	}

	// Create checkin log FIRST to prevent duplicate checkins
	// This acts as a lock - if another request tries to insert, it will fail due to unique constraint
	rewardQuota := setting.RewardQuota
	checkinLog := &model.CheckinLog{
		UserId:      userId,
		RewardQuota: rewardQuota,
		QuotaBefore: quotaBefore,
		QuotaAfter:  quotaBefore + rewardQuota,
		CheckinDate: todayDate,
	}
	err = checkinLog.Insert()
	if err != nil {
		// If insert fails, it might be a duplicate - check again
		hasCheckedIn, _ := model.HasCheckedInToday(userId, todayDate)
		if hasCheckedIn {
			return &dto.CheckinResult{
				Success: false,
				Message: "今日已签到，请明天再来",
			}, nil
		}
		common.SysError(fmt.Sprintf("failed to insert checkin log for user %d: %s", userId, err.Error()))
		return nil, err
	}

	// Now increase user quota
	err = model.IncreaseUserQuota(userId, rewardQuota, true)
	if err != nil {
		common.SysError(fmt.Sprintf("failed to increase quota for user %d after checkin: %s", userId, err.Error()))
		// Don't return error since checkin log was created
	}

	quotaAfter := quotaBefore + rewardQuota

	// Update Redis cache
	if common.RedisEnabled {
		cacheKey := getCheckinCacheKey(userId, todayDate)
		_ = common.RedisSet(cacheKey, "1", CheckinCacheDuration)
	}

	return &dto.CheckinResult{
		Success:     true,
		Message:     "签到成功",
		RewardQuota: rewardQuota,
		NewQuota:    quotaAfter,
	}, nil
}

// GetCheckinStatus returns the current checkin status for a user
func GetCheckinStatus(userId int) (*dto.CheckinStatus, error) {
	setting := checkin_setting.GetCheckinSetting()

	status := &dto.CheckinStatus{
		Enabled:        setting.Enabled,
		RewardQuota:    setting.RewardQuota,
		QuotaThreshold: setting.QuotaThreshold,
	}

	// If feature is disabled, return early
	if !setting.Enabled {
		status.CanCheckin = false
		status.Reason = "签到功能未启用"
		return status, nil
	}

	// Get user quota
	userQuota, err := model.GetUserQuota(userId, false)
	if err != nil {
		return nil, err
	}
	status.CurrentQuota = userQuota

	// Check if already checked in today
	hasCheckedIn, err := HasCheckedInToday(userId)
	if err != nil {
		return nil, err
	}
	status.HasCheckedIn = hasCheckedIn

	// Get last checkin time
	lastCheckin, err := model.GetLastCheckinByUserId(userId)
	if err == nil && lastCheckin != nil {
		status.LastCheckinTime = lastCheckin.CreatedAt.Unix()
	}

	// Calculate next checkin time
	status.NextCheckinTime = GetNextCheckinTime()

	// Determine if can checkin
	if hasCheckedIn {
		status.CanCheckin = false
		status.Reason = "今日已签到，请明天再来"
	} else if userQuota >= setting.QuotaThreshold {
		status.CanCheckin = false
		status.Reason = fmt.Sprintf("当前额度超过签到阈值（%s），无法签到", formatQuotaWithCurrency(setting.QuotaThreshold))
	} else {
		status.CanCheckin = true
	}

	return status, nil
}

// GetCheckinStatusSimple returns a simplified checkin status check
// Used for quick eligibility checks without full status details
func GetCheckinStatusSimple(userId int) (enabled bool, canCheckin bool, reason string, err error) {
	setting := checkin_setting.GetCheckinSetting()

	if !setting.Enabled {
		return false, false, "签到功能未启用", nil
	}

	canCheckin, reason, err = CanCheckin(userId)
	if err != nil {
		return setting.Enabled, false, "", err
	}

	return setting.Enabled, canCheckin, reason, nil
}
