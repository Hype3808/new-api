package system_setting

import (
	"strings"

	"github.com/QuantumNous/new-api/setting/config"
)

type DiscordSettings struct {
	Enabled            bool   `json:"enabled"`
	ClientId           string `json:"client_id"`
	ClientSecret       string `json:"client_secret"`
	GuildVerifyEnabled bool   `json:"guild_verify_enabled"` // 是否启用服务器验证
	RequiredGuildId    string `json:"required_guild_id"`    // 必须加入的服务器ID
	RequiredRoleIds    string `json:"required_role_ids"`    // 必须拥有的身份组ID列表(逗号分隔，满足其一即可)
}

// 默认配置
var defaultDiscordSettings = DiscordSettings{}

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("discord", &defaultDiscordSettings)
}

func GetDiscordSettings() *DiscordSettings {
	return &defaultDiscordSettings
}

// GetRequiredRoleIdsList 获取身份组ID列表
func (s *DiscordSettings) GetRequiredRoleIdsList() []string {
	if s.RequiredRoleIds == "" {
		return []string{}
	}
	// 按逗号分隔并去除空格
	parts := strings.Split(s.RequiredRoleIds, ",")
	var roles []string
	for _, role := range parts {
		trimmed := strings.TrimSpace(role)
		if trimmed != "" {
			roles = append(roles, trimmed)
		}
	}
	return roles
}
