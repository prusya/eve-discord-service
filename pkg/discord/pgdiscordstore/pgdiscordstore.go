package pgdiscordstore

import (
	"github.com/jmoiron/sqlx"

	"github.com/prusya/eve-discord-service/pkg/discord"
	"github.com/prusya/eve-discord-service/pkg/system"
)

const (
	storeName            = "pgdiscordstore"
	createUserTableQuery = `
	CREATE TABLE IF NOT EXISTS "discord_user"
	(
		id                     SERIAL PRIMARY KEY,
		eve_char_id            INTEGER NOT NULL,
		eve_char_name          VARCHAR(50) NOT NULL,
		eve_corp_ticker        VARCHAR(50) NOT NULL,
		eve_alli_ticker        VARCHAR(50) NOT NULL,
		discord_id             VARCHAR(50) NOT NULL UNIQUE,
		discord_access_token   TEXT NOT NULL,
		discord_refresh_token  TEXT NOT NULL,
		discord_token_scope    TEXT NOT NULL,
		discord_token_is_valid BOOLEAN,
		active                 BOOLEAN
	)`
	createUserQuery = `
	INSERT INTO "discord_user"
	(eve_char_id, eve_char_name, eve_corp_ticker, eve_alli_ticker, 
		discord_id, discord_access_token, discord_refresh_token, 
		discord_token_scope, discord_token_is_valid, active)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	setUserInactiveByDiscordIDQuery = `
	UPDATE "discord_user"
	SET active = 'f'
	WHERE discord_id = $1`
	updateUserQuery = `
	UPDATE "discord_user"
	SET eve_char_id = $1,
		eve_char_name = $2,
		eve_corp_ticker = $3,
		eve_alli_ticker = $4,
		discord_id = $5,
		discord_access_token = $6,
		discord_refresh_token = $7,
		discord_token_scope = $8,
		discord_token_is_valid = $9,
		active = $10
	WHERE id = $11`
)

// Store implements discord.Store interface backed by postgresql and sqlx.
type Store struct {
	db *sqlx.DB
}

// New creates a new Store.
func New(db *sqlx.DB) *Store {
	s := Store{
		db: db,
	}

	return &s
}

// Init prepares db for usage.
func (s *Store) Init() {
	_, err := s.db.Exec(createUserTableQuery)
	system.HandleError(err, storeName+".Init")
}

// Drop placeholder.
func (s *Store) Drop() {}

// CreateUser stores a discord.User record.
func (s *Store) CreateUser(u *discord.User) {
	_, err := s.db.Exec(createUserQuery, u.EveCharID, u.EveCharName,
		u.EveCorpTicker, u.EveAlliTicker, u.DiscordID, u.DiscordAccessToken,
		u.DiscordRefreshToken, u.DiscordTokenScope, u.DiscordTokenIsValid,
		u.Active)
	system.HandleError(err, storeName+".CreateUser", u)
}

// Users returns all discord.User records.
func (s *Store) Users() []*discord.User {
	var users []*discord.User
	err := s.db.Select(&users, `SELECT * FROM "discord_user"`)
	system.HandleError(err, storeName+".Users")

	return users
}

// UsersWithValidToken returns user records with `discord_token_is_valid`
// set to true.
func (s *Store) UsersWithValidToken() []*discord.User {
	var users []*discord.User
	err := s.db.Select(&users, `SELECT * FROM "discord_user" where discord_token_is_valid`)
	system.HandleError(err, storeName+".UsersWithValidToken")

	return users
}

// UpdateUser updates a discord.User record.
func (s *Store) UpdateUser(u *discord.User) {
	_, err := s.db.Exec(updateUserQuery, u.EveCharID, u.EveCharName,
		u.EveCorpTicker, u.EveAlliTicker, u.DiscordID, u.DiscordAccessToken,
		u.DiscordRefreshToken, u.DiscordTokenScope, u.DiscordTokenIsValid,
		u.Active, u.ID)
	system.HandleError(err, storeName+".UpdateUser", u)
}

// SetUserInactiveByDiscordID sets `active` to false for provided did.
func (s *Store) SetUserInactiveByDiscordID(did string) {
	_, err := s.db.Exec(setUserInactiveByDiscordIDQuery, did)
	system.HandleError(err, storeName+".SetUserInactiveByDiscordID", "did="+did)
}

// ActiveUsersCharIDs returns EveCharIDs of users with `Active` set to true.
func (s *Store) ActiveUsersCharIDs() []int32 {
	var ids []int32
	err := s.db.Select(&ids, `SELECT eve_char_id FROM "discord_user" WHERE active`)
	system.HandleError(err, storeName+".ActiveUsersCharIDs")

	return ids
}

// DiscordIDExists checks if record with provided did exists.
func (s *Store) DiscordIDExists(did string) bool {
	var exists bool
	err := s.db.Get(&exists,
		`SELECT EXISTS(SELECT 1 FROM "discord_user" WHERE discord_id=$1)`, did)
	system.HandleError(err, storeName+".DiscordIDExists", "did="+did)

	return exists
}
