package checkin_setting

import "github.com/QuantumNous/new-api/setting/config"

// CheckinSetting 签到功能配置
type CheckinSetting struct {
	Enabled        bool `json:"enabled"`         // 是否启用签到功能
	RewardQuota    int  `json:"reward_quota"`    // 签到奖励额度
	QuotaThreshold int  `json:"quota_threshold"` // 签到额度阈值（用户额度必须低于此值才能签到）
}

// 默认配置
var defaultCheckinSetting = CheckinSetting{
	Enabled:        false,   // 默认关闭签到功能
	RewardQuota:    500000,  // 默认签到奖励 500000 额度
	QuotaThreshold: 1000000, // 默认签到阈值 1000000 额度
}

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("checkin", &defaultCheckinSetting)
}

// GetCheckinSetting 获取签到配置
func GetCheckinSetting() *CheckinSetting {
	return &defaultCheckinSetting
}
