package dgodiscordservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"

	"github.com/prusya/eve-discord-service/pkg/discord"
	"github.com/prusya/eve-discord-service/pkg/system"
)

const (
	serviceName = "dgodiscordservice"
)

// Service implements discord.Service interface backed by discordgo lib.
type Service struct {
	store    discord.Store
	system   *system.System
	dgo      *discordgo.Session
	stopChan chan struct{}
}

// userData defines a response from users validation server.
type userData struct {
	EveCharID     int32
	EveCharName   string
	EveCorpTicker string
	EveAlliTicker string
	Valid         bool
}

// New creates a new Service and prepares it to Start.
func New(sys *system.System, store discord.Store) *Service {
	dgo, err := discordgo.New("Bot " + sys.Config.DiscordBotToken)
	system.HandleError(err)

	s := Service{
		store:    store,
		system:   sys,
		dgo:      dgo,
		stopChan: make(chan struct{}, 1),
	}

	sys.Discord = &s

	return &s
}

// Start starts the Service.
func (s *Service) Start() {
	err := s.dgo.Open()
	system.HandleError(err, serviceName+" dgo.Open")

	validateUsersT := time.NewTicker(20 * time.Minute)
	refreshTokensT := time.NewTicker(24 * time.Hour)
	go func() {
		go s.RefreshTokens()
		select {
		case <-validateUsersT.C:
			go s.ValidateUsers()
		case <-refreshTokensT.C:
			go s.RefreshTokens()
		case <-s.stopChan:
			validateUsersT.Stop()
			refreshTokensT.Stop()
			break
		}
	}()
}

// Stop stops the Service.
func (s *Service) Stop() {
	s.stopChan <- struct{}{}
	s.dgo.Close()
}

// GetStore returns discord.Store.
func (s *Service) GetStore() discord.Store {
	return s.store
}

// GuildAddUser adds user to a guild.
func (s *Service) GuildAddUser(u *discord.User) {
	tickers := strings.TrimSpace(
		fmt.Sprintf("%s %s", u.EveAlliTicker, u.EveCorpTicker))
	nickname := fmt.Sprintf("[%s] %s", tickers, u.EveCharName)
	err := s.dgo.GuildMemberAdd(u.DiscordAccessToken,
		s.system.Config.DiscordGuildID, u.DiscordID, nickname,
		s.system.Config.DiscordGuildRoles, false, false)
	system.HandleError(err, serviceName+".GuildAddUser", u)
}

func recoverPanic() {
	recover()
}

// guildKickUser deletes a user from guild.
func (s *Service) guildKickUser(did string) {
	err := s.dgo.GuildMemberDelete(s.system.Config.DiscordGuildID, did)
	system.HandleError(err, serviceName+".guildKickUser", "discordID="+did)
}

// processInvalidUser deletes user from guild and marks store record as inactive.
func (s *Service) processInvalidUser(u *discord.User) {
	defer recoverPanic()

	s.guildKickUser(u.DiscordID)
	s.store.SetUserInactiveByDiscordID(u.DiscordID)
}

// processUpdatedUser changes user's nickname in guild and updates store record.
func (s *Service) processUpdatedUser(u userData, user *discord.User) {
	defer recoverPanic()

	// Update nickname in guild.
	tickers := strings.TrimSpace(
		fmt.Sprintf("%s %s", u.EveAlliTicker, u.EveCorpTicker))
	nickname := fmt.Sprintf("[%s] %s", tickers, u.EveCharName)
	err := s.dgo.GuildMemberNickname(s.system.Config.DiscordGuildID,
		user.DiscordID, nickname)
	system.HandleError(err, serviceName+".ValidateUsers GuildMemberNickname",
		u, user)

	// Update store record.
	user.EveCorpTicker = u.EveCorpTicker
	user.EveAlliTicker = u.EveAlliTicker
	s.store.UpdateUser(user)
}

// ValidateUsers keeps user records up to date, sets proper nicknames in guild,
// deletes users from guild if they don't have access to discord service.
func (s *Service) ValidateUsers() {
	defer recoverPanic()

	// We need to check only active users.
	ids := s.store.ActiveUsersCharIDs()

	// Send ids to the validation server.
	payload, _ := json.Marshal(ids)
	resp, err := http.Post(s.system.Config.UsersValidationEndpoint,
		"application/json", bytes.NewBuffer(payload))
	system.HandleError(err, serviceName+".ValidateUsers http.Post", ids)
	if resp.StatusCode != 200 {
		err = errors.New("non 200 response from validation server")
		system.HandleError(err, serviceName+".ValidateUsers", resp.StatusCode)
	}

	usersData := []userData{}
	d := json.NewDecoder(resp.Body)
	d.UseNumber()
	d.Decode(&usersData)
	resp.Body.Close()

	// Process response from the validation server.
	users := s.store.Users()
	for _, u := range usersData {
		for _, user := range users {
			if user.EveCharID == u.EveCharID {
				// Invalid users should only be kicked from the guild.
				if !u.Valid {
					s.processInvalidUser(user)
					continue
				}

				// Rename user if corp or alli has changed.
				if user.EveCorpTicker != u.EveCorpTicker ||
					user.EveAlliTicker != u.EveAlliTicker {
					s.processUpdatedUser(u, user)
				}
			}
		}
	}
}

// refreshToken obtains and stores a fresh access token.
// Kicks user from guild if token was revoked and
// `DiscordKickUserWithRevokedToken` option is set to true.
func (s *Service) refreshToken(u *discord.User, data *url.Values) {
	defer recoverPanic()

	// Request refresh.
	resp, err := http.PostForm("https://discordapp.com/api/v6/oauth2/token", *data)
	system.HandleError(err, serviceName+".refreshToken PostForm", data)

	// Decode response.
	var tokenData map[string]interface{}
	d := json.NewDecoder(resp.Body)
	d.UseNumber()
	err = d.Decode(&tokenData)
	system.HandleError(err, serviceName+".refreshToken Decode", data)
	defer resp.Body.Close()

	if _, ok := tokenData["error"]; ok {
		// Check if access was revoked.
		if tokenData["error"].(string) == "invalid_grant" {
			if s.system.Config.DiscordKickUserWithRevokedToken {
				s.processInvalidUser(u)
			} else {
				s.store.SetUserInactiveByDiscordID(u.DiscordID)
			}
			return
		}
		system.HandleError(errors.New("refresh token error"),
			serviceName+".refreshToken Decode", data, u, tokenData)
	}

	// Absence of `access_token` means either user revoked access or the
	// token was granted for another app or some other error occured.
	if _, ok := tokenData["access_token"]; !ok {
		system.HandleError(errors.New("access_token not found"),
			serviceName+".refreshToken Decode", data, u, tokenData)
	}

	// Store new token.
	u.DiscordAccessToken = tokenData["access_token"].(string)
	s.store.UpdateUser(u)
}

// RefreshTokens gets and stores fresh access tokens.
func (s *Service) RefreshTokens() {
	users := s.store.UsersWithValidToken()
	v := url.Values{}
	v.Set("client_id", s.system.Config.DiscordClientID)
	v.Set("client_secret", s.system.Config.DiscordClientSecret)
	v.Set("grant_type", "refresh_token")
	v.Set("redirect_uri", s.system.Config.DiscordCallbackURL)

	for _, user := range users {
		v.Set("refresh_token", user.DiscordRefreshToken)
		v.Set("scope", user.DiscordTokenScope)
		s.refreshToken(user, &v)
	}
}
