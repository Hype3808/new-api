package dto

// CheckinResult represents the result of a check-in operation
type CheckinResult struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	RewardQuota int    `json:"reward_quota,omitempty"`
	NewQuota    int    `json:"new_quota,omitempty"`
}

// CheckinStatus represents the current check-in status for a user
type CheckinStatus struct {
	Enabled                 bool   `json:"enabled"`                    // 签到功能是否启用
	CanCheckin              bool   `json:"can_checkin"`                // 是否可以签到
	HasCheckedIn            bool   `json:"has_checked_in"`             // 今日是否已签到
	LastCheckinTime         int64  `json:"last_checkin_time"`          // 上次签到时间戳
	NextCheckinTime         int64  `json:"next_checkin_time"`          // 下次可签到时间戳
	RewardQuota             int    `json:"reward_quota"`               // 签到奖励额度
	QuotaThreshold          int    `json:"quota_threshold"`            // 签到阈值
	QuotaThresholdInclusive bool   `json:"quota_threshold_inclusive"`  // 签到阈值是否包含等于
	CurrentQuota            int    `json:"current_quota"`              // 当前额度
	Reason                  string `json:"reason,omitempty"`           // 不能签到的原因
}
