package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

type TokenSetting struct {
	RequireGroupSelection bool `json:"require_group_selection"` // 是否要求在创建令牌时必须选择分组
}

// 默认配置
var tokenSetting = TokenSetting{
	RequireGroupSelection: false, // 默认关闭，不强制要求选择分组
}

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("token_setting", &tokenSetting)
}

func GetTokenSetting() *TokenSetting {
	return &tokenSetting
}

// IsGroupSelectionRequired 返回是否要求选择令牌分组
func IsGroupSelectionRequired() bool {
	return tokenSetting.RequireGroupSelection
}
