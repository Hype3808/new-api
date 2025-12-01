package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/setting/system_setting"
)

const (
	// Discord API base URL
	DiscordAPIBaseURL = "https://discord.com/api/v10"
	// Discord API timeout
	DiscordAPITimeout = 10 * time.Second
)

// DiscordVerificationResult represents the result of Discord verification
type DiscordVerificationResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetUserGuilds retrieves the list of guilds the user is a member of
// Calls Discord API endpoint: GET /users/@me/guilds
// Requires OAuth scope: guilds
func GetUserGuilds(accessToken string) ([]dto.DiscordGuild, error) {
	if accessToken == "" {
		return nil, errors.New("access token is required")
	}

	url := fmt.Sprintf("%s/users/@me/guilds", DiscordAPIBaseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout: DiscordAPITimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		common.SysLog(fmt.Sprintf("Discord API error (GetUserGuilds): %s", err.Error()))
		return nil, errors.New("无法连接至 Discord 服务器，请稍后重试")
	}
	defer resp.Body.Close()

	// Read response body for debugging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		common.SysError(fmt.Sprintf("Discord API returned status %d for GetUserGuilds", resp.StatusCode))
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, errors.New("Discord 授权已过期或缺少 guilds 权限，请重新授权")
		}
		return nil, errors.New("Discord 获取服务器列表失败")
	}

	var guilds []dto.DiscordGuild
	if err := json.Unmarshal(bodyBytes, &guilds); err != nil {
		common.SysError(fmt.Sprintf("Discord GetUserGuilds parse error: %s", err.Error()))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return guilds, nil
}

// GetUserGuildMember retrieves the user's member information in a specific guild
// Calls Discord API endpoint: GET /users/@me/guilds/{guild.id}/member
// Requires OAuth scope: guilds.members.read
func GetUserGuildMember(accessToken string, guildId string) (*dto.DiscordGuildMember, error) {
	if accessToken == "" {
		return nil, errors.New("access token is required")
	}
	if guildId == "" {
		return nil, errors.New("guild ID is required")
	}

	url := fmt.Sprintf("%s/users/@me/guilds/%s/member", DiscordAPIBaseURL, guildId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout: DiscordAPITimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		common.SysLog(fmt.Sprintf("Discord API error (GetUserGuildMember): %s", err.Error()))
		return nil, errors.New("无法连接至 Discord 服务器，请稍后重试")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// User is not a member of this guild
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		common.SysLog(fmt.Sprintf("Discord API returned status %d for GetUserGuildMember", resp.StatusCode))
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, errors.New("Discord 授权已过期，请重新登录")
		}
		return nil, errors.New("Discord 获取成员信息失败")
	}

	var member dto.DiscordGuildMember
	if err := json.NewDecoder(resp.Body).Decode(&member); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &member, nil
}


// VerifyGuildMembership verifies if a user is a member of the specified guild
// Returns true if the user is a member, false otherwise
func VerifyGuildMembership(accessToken string, guildId string) (bool, error) {
	if guildId == "" {
		// No guild requirement, verification passes
		return true, nil
	}

	guilds, err := GetUserGuilds(accessToken)
	if err != nil {
		return false, err
	}

	for _, guild := range guilds {
		if guild.Id == guildId {
			return true, nil
		}
	}
	return false, nil
}

// VerifyUserRoles verifies if a user has at least one of the required roles in the specified guild
// Returns true if the user has at least one required role, false otherwise
// If requiredRoleIds is empty, verification passes
func VerifyUserRoles(accessToken string, guildId string, requiredRoleIds []string) (bool, error) {
	if len(requiredRoleIds) == 0 {
		// No role requirement, verification passes
		return true, nil
	}

	if guildId == "" {
		return false, errors.New("guild ID is required for role verification")
	}

	member, err := GetUserGuildMember(accessToken, guildId)
	if err != nil {
		return false, err
	}

	if member == nil {
		// User is not a member of the guild
		return false, nil
	}

	// Check if user has at least one of the required roles
	for _, userRole := range member.Roles {
		for _, requiredRole := range requiredRoleIds {
			if userRole == requiredRole {
				return true, nil
			}
		}
	}

	return false, nil
}


// VerifyDiscordUser performs the complete Discord verification flow
// Checks guild membership and role requirements based on system settings
// Returns (success, errorMessage, error)
// - success: true if verification passed or verification is disabled
// - errorMessage: human-readable message explaining why verification failed
// - error: system error if API call failed
func VerifyDiscordUser(accessToken string) (bool, string, error) {
	settings := system_setting.GetDiscordSettings()

	// If guild verification is disabled, skip all checks
	if !settings.GuildVerifyEnabled {
		return true, "", nil
	}

	// Check guild membership
	if settings.RequiredGuildId != "" {
		isMember, err := VerifyGuildMembership(accessToken, settings.RequiredGuildId)
		if err != nil {
			common.SysLog(fmt.Sprintf("Discord guild verification error: %s", err.Error()))
			return false, "Discord 验证失败，请稍后重试", err
		}
		if !isMember {
			return false, "您需要加入指定的 Discord 服务器才能注册", nil
		}

		// Check role requirements (only if guild membership is verified)
		requiredRoles := settings.GetRequiredRoleIdsList()
		if len(requiredRoles) > 0 {
			hasRole, err := VerifyUserRoles(accessToken, settings.RequiredGuildId, requiredRoles)
			if err != nil {
				common.SysLog(fmt.Sprintf("Discord role verification error: %s", err.Error()))
				return false, "Discord 验证失败，请稍后重试", err
			}
			if !hasRole {
				return false, "您需要拥有指定的身份组才能注册", nil
			}
		}
	}

	return true, "", nil
}

// VerifyDiscordUserForLogin performs Discord verification for existing user login
// Similar to VerifyDiscordUser but with login-specific error messages
func VerifyDiscordUserForLogin(accessToken string) (bool, string, error) {
	settings := system_setting.GetDiscordSettings()

	// If guild verification is disabled, skip all checks
	if !settings.GuildVerifyEnabled {
		return true, "", nil
	}

	// Check guild membership
	if settings.RequiredGuildId != "" {
		isMember, err := VerifyGuildMembership(accessToken, settings.RequiredGuildId)
		if err != nil {
			common.SysLog(fmt.Sprintf("Discord guild verification error (login): %s", err.Error()))
			return false, "Discord 验证失败，请稍后重试", err
		}
		if !isMember {
			return false, "您已不在指定的 Discord 服务器中，无法登录", nil
		}

		// Check role requirements (only if guild membership is verified)
		requiredRoles := settings.GetRequiredRoleIdsList()
		if len(requiredRoles) > 0 {
			hasRole, err := VerifyUserRoles(accessToken, settings.RequiredGuildId, requiredRoles)
			if err != nil {
				common.SysLog(fmt.Sprintf("Discord role verification error (login): %s", err.Error()))
				return false, "Discord 验证失败，请稍后重试", err
			}
			if !hasRole {
				return false, "您已不拥有指定的身份组，无法登录", nil
			}
		}
	}

	return true, "", nil
}
