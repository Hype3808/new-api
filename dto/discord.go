package dto

// DiscordGuild represents a Discord guild/server
// Used when calling Discord API endpoint /users/@me/guilds
type DiscordGuild struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Owner       bool   `json:"owner"`
	Permissions string `json:"permissions"`
}

// DiscordGuildMember represents a user's membership in a Discord guild
// Used when calling Discord API endpoint /users/@me/guilds/{guild.id}/member
type DiscordGuildMember struct {
	User     *DiscordUser `json:"user"`
	Nick     string       `json:"nick"`
	Roles    []string     `json:"roles"`
	JoinedAt string       `json:"joined_at"`
}

// DiscordUser represents basic Discord user information
type DiscordUser struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"global_name"`
}

// DiscordRole represents a Discord role/identity group
type DiscordRole struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
