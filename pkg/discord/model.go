package discord

// User defines a model for a database and represents a discord user.
type User struct {
	ID int `storm:"id,unique,increment" db:"id"`

	// EveCharID is not unique cos same character can connect using
	// different discord accounts.
	// Distinction is only made by DiscordID.
	EveCharID     int32  `db:"eve_char_id"`
	EveCharName   string `db:"eve_char_name"`
	EveCorpTicker string `db:"eve_corp_ticker"`
	EveAlliTicker string `db:"eve_alli_ticker"`

	DiscordID           string `storm:"unique" db:"discord_id"`
	DiscordAccessToken  string `db:"discord_access_token"`
	DiscordRefreshToken string `db:"discord_refresh_token"`
	DiscordTokenScope   string `db:"discord_token_scope"`
	DiscordTokenIsValid bool   `db:"discord_token_is_valid"`

	Active bool `db:"active"`
}

// Store defines an interface of how to interact with user model on db level.
type Store interface {
	Init()
	Drop()
	CreateUser(u *User)
	Users() []*User
	ActiveUsersCharIDs() []int32
	UsersWithValidToken() []*User
	UpdateUser(u *User)
	SetUserInactiveByDiscordID(did string)
	DiscordIDExists(did string) bool
}

// Service defines an interface of how to ineract with discord service.
type Service interface {
	Start()
	Stop()
	GetStore() Store
	ValidateUsers()
	RefreshTokens()
	GuildAddUser(u *User)
}
