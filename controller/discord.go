package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/system_setting"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type DiscordResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

type DiscordUser struct {
	UID  string `json:"id"`
	ID   string `json:"username"`
	Name string `json:"global_name"`
}

// DiscordAuthResult contains both user info and access token for verification
type DiscordAuthResult struct {
	User        *DiscordUser
	AccessToken string
}

// buildDiscordRedirectURI resolves the redirect_uri used for Discord token exchange.
// Priority:
// 1) redirect_uri query param provided by the frontend callback (ensures exact match)
// 2) configured ServerAddress
// 3) request host/protocol fallback
func buildDiscordRedirectURI(c *gin.Context) string {
	if uri := c.Query("redirect_uri"); uri != "" {
		return uri
	}

	base := strings.TrimSuffix(system_setting.ServerAddress, "/")
	if base != "" {
		return fmt.Sprintf("%s/oauth/discord", base)
	}

	proto := "http"
	forwardedProto := c.GetHeader("X-Forwarded-Proto")
	if forwardedProto == "" {
		forwardedProto = c.GetHeader("X-Forwarded-Protocol")
	}
	if strings.EqualFold(forwardedProto, "https") || c.Request.TLS != nil {
		proto = "https"
	}

	host := c.Request.Host
	if host == "" {
		return "/oauth/discord"
	}

	return fmt.Sprintf("%s://%s/oauth/discord", proto, host)
}

func getDiscordUserInfoByCode(code string, redirectURI string) (*DiscordAuthResult, error) {
	if code == "" {
		return nil, errors.New("无效的参数")
	}

	// Build OAuth scope based on guild verification settings
	// When guild verification is enabled, we need additional scopes
	values := url.Values{}
	values.Set("client_id", system_setting.GetDiscordSettings().ClientId)
	values.Set("client_secret", system_setting.GetDiscordSettings().ClientSecret)
	values.Set("code", code)
	values.Set("grant_type", "authorization_code")
	values.Set("redirect_uri", redirectURI)
	formData := values.Encode()
	req, err := http.NewRequest("POST", "https://discord.com/api/v10/oauth2/token", strings.NewReader(formData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		common.SysLog(err.Error())
		return nil, errors.New("无法连接至 Discord 服务器，请稍后重试！")
	}
	defer res.Body.Close()

	// Read response body for debugging
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// Log the response for debugging
	if res.StatusCode != http.StatusOK {
		common.SysError(fmt.Sprintf("Discord Token API returned status %d: %s", res.StatusCode, string(bodyBytes)))
		return nil, errors.New("Discord 获取 Token 失败，请检查设置！")
	}

	var discordResponse DiscordResponse
	err = json.Unmarshal(bodyBytes, &discordResponse)
	if err != nil {
		common.SysError(fmt.Sprintf("Discord Token API response parse error: %s, body: %s", err.Error(), string(bodyBytes)))
		return nil, err
	}

	if discordResponse.AccessToken == "" {
		common.SysError(fmt.Sprintf("Discord 获取 Token 失败，响应: %s", string(bodyBytes)))
		return nil, errors.New("Discord 获取 Token 失败，请检查设置！")
	}

	req, err = http.NewRequest("GET", "https://discord.com/api/v10/users/@me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+discordResponse.AccessToken)
	res2, err := client.Do(req)
	if err != nil {
		common.SysLog(err.Error())
		return nil, errors.New("无法连接至 Discord 服务器，请稍后重试！")
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		common.SysError("Discord 获取用户信息失败！请检查设置！")
		return nil, errors.New("Discord 获取用户信息失败！请检查设置！")
	}

	var discordUser DiscordUser
	err = json.NewDecoder(res2.Body).Decode(&discordUser)
	if err != nil {
		return nil, err
	}
	if discordUser.UID == "" || discordUser.ID == "" {
		common.SysError("Discord 获取用户信息为空！请检查设置！")
		return nil, errors.New("Discord 获取用户信息为空！请检查设置！")
	}
	return &DiscordAuthResult{
		User:        &discordUser,
		AccessToken: discordResponse.AccessToken,
	}, nil
}

// GetDiscordOAuthURL returns the Discord OAuth URL with appropriate scopes
// When guild verification is enabled, includes guilds and guilds.members.read scopes
func GetDiscordOAuthURL() string {
	settings := system_setting.GetDiscordSettings()
	baseScopes := "identify"

	// Add guild-related scopes if guild verification is enabled
	if settings.GuildVerifyEnabled {
		baseScopes = "identify guilds guilds.members.read"
	}

	return fmt.Sprintf(
		"https://discord.com/api/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s",
		settings.ClientId,
		url.QueryEscape(fmt.Sprintf("%s/oauth/discord", system_setting.ServerAddress)),
		url.QueryEscape(baseScopes),
	)
}

func DiscordOAuth(c *gin.Context) {
	session := sessions.Default(c)
	state := c.Query("state")
	if state == "" || session.Get("oauth_state") == nil || state != session.Get("oauth_state").(string) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "state is empty or not same",
		})
		return
	}

	// Clear the oauth_state immediately after validation to prevent reuse
	session.Delete("oauth_state")
	session.Save()

	username := session.Get("username")
	if username != nil {
		DiscordBind(c)
		return
	}
	if !system_setting.GetDiscordSettings().Enabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 Discord 登录以及注册",
		})
		return
	}
	code := c.Query("code")
	redirectURI := buildDiscordRedirectURI(c)
	authResult, err := getDiscordUserInfoByCode(code, redirectURI)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	discordUser := authResult.User
	accessToken := authResult.AccessToken

	user := model.User{
		DiscordId: discordUser.UID,
	}
	if model.IsDiscordIdAlreadyTaken(user.DiscordId) {
		// Existing user login flow
		err := user.FillUserByDiscordId()
		if err != nil {
			common.SysError(fmt.Sprintf("Discord OAuth: FillUserByDiscordId failed for UID %s: %s", discordUser.UID, err.Error()))
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "用户信息获取失败，请重试",
			})
			return
		}

		// Double check that user was actually found
		if user.Id == 0 {
			common.SysError(fmt.Sprintf("Discord OAuth: user.Id is 0 after FillUserByDiscordId for UID %s", discordUser.UID))
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "用户信息获取失败，请重试",
			})
			return
		}

		// Verify Discord guild/role requirements for existing user login
		verified, verifyMsg, verifyErr := service.VerifyDiscordUserForLogin(accessToken)
		if verifyErr != nil {
			common.SysError(fmt.Sprintf("Discord verification error for user %d: %s", user.Id, verifyErr.Error()))
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": verifyMsg,
			})
			return
		}
		if !verified {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": verifyMsg,
			})
			return
		}
	} else {
		// New user registration flow
		if common.RegisterEnabled {
			// Verify Discord guild/role requirements for new user registration
			verified, verifyMsg, verifyErr := service.VerifyDiscordUser(accessToken)
			if verifyErr != nil {
				common.SysError(fmt.Sprintf("Discord verification error for new user: %s", verifyErr.Error()))
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": verifyMsg,
				})
				return
			}
			if !verified {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": verifyMsg,
				})
				return
			}

			if discordUser.ID != "" {
				user.Username = discordUser.ID
			} else {
				user.Username = "discord_" + strconv.Itoa(model.GetMaxUserId()+1)
			}
			if discordUser.Name != "" {
				user.DisplayName = discordUser.Name
			} else {
				user.DisplayName = "Discord User"
			}
			err := user.Insert(0)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
				})
				return
			}
		} else {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "管理员关闭了新用户注册",
			})
			return
		}
	}

	if user.Status != common.UserStatusEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "用户已被封禁",
			"success": false,
		})
		return
	}
	setupLogin(&user, c)
}

func DiscordBind(c *gin.Context) {
	if !system_setting.GetDiscordSettings().Enabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 Discord 登录以及注册",
		})
		return
	}
	code := c.Query("code")
	redirectURI := buildDiscordRedirectURI(c)
	authResult, err := getDiscordUserInfoByCode(code, redirectURI)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	discordUser := authResult.User
	user := model.User{
		DiscordId: discordUser.UID,
	}
	if model.IsDiscordIdAlreadyTaken(user.DiscordId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该 Discord 账户已被绑定",
		})
		return
	}
	session := sessions.Default(c)
	id := session.Get("id")
	user.Id = id.(int)
	err = user.FillUserById()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	user.DiscordId = discordUser.UID
	err = user.Update(false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "bind",
	})
}
