package gorillahttp

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"

	"github.com/prusya/eve-discord-service/pkg/discord"
	"github.com/prusya/eve-discord-service/pkg/system"
)

type eveChar struct {
	EveCharID     int32
	EveCorpID     int32
	EveAlliID     int32
	EveCharName   string
	EveCorpName   string
	EveAlliName   string
	EveCorpTicker string
	EveAlliTicker string
	Valid         bool
}

type discordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	AvatarID      string `json:"avatar"`
	Discriminator string `json:"discriminator"`
	Locale        string `json:"locale"`
	MFAEnabled    bool   `json:"mfa_enabled"`
	Verified      bool   `json:"verified"`
	Flags         int    `json:"flags"`
}

// respondWithJSON receives a payload of any type, converts it into json
// and writes resulting json to a response writer.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// respondWithError receives a message string, converts it into json
// and writes resulting json to a response writer.
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondOK writes json `status: ok` to w.
func respondOK(w http.ResponseWriter) {
	respondWithJSON(w, 200, map[string]string{"status": "ok"})
}

// respond401 responds with http Unauthorized 401.
func respond401(w http.ResponseWriter) {
	respondWithError(w, 401, http.StatusText(401))
}

// respond403 responds with http Forbidden 403.
func respond403(w http.ResponseWriter) {
	respondWithError(w, 403, http.StatusText(403))
}

// NotFoundH responds with http Not Found 404.
func NotFoundH(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, 404, http.StatusText(404))
}

// MethodNotAllowedH responds with http Method Not Allowed 405.
func MethodNotAllowedH(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, 405, http.StatusText(405))
}

// recoverPanic recovers and responds with http 500 in case of a panic.
func recoverPanic(w http.ResponseWriter) {
	if r := recover(); r != nil {
		respondWithError(w, 500, fmt.Sprintf("%s", r))
	}
}

// HealthCheckH calls respondOK.
func HealthCheckH(w http.ResponseWriter, r *http.Request) {
	respondOK(w)
}

// DiscordAuthH generates a cookie with a unique state and returns a login url
// for discord sso.
// https://discordapp.com/developers/docs/topics/oauth2
func (s *Service) DiscordAuthH(w http.ResponseWriter, r *http.Request) {
	// Create a new state.
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	// Get a session and store the state.
	// Normally, it should create a new empty session.
	session, _ := s.session.Get(r, "DiscordAuth")
	session.Values["state"] = state
	session.Save(r, w)

	// Generate a login url and redirect.
	url := s.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOnline,
		oauth2.SetAuthURLParam("scope", strings.Join(s.system.Config.DiscordAuthScopes, " ")))
	respondWithJSON(w, 200, url)
}

// DiscordAuthCallbackH handles the callback request from discord sso
// and completes the auth process.
// https://discordapp.com/developers/docs/topics/oauth2
func (s *Service) DiscordAuthCallbackH(w http.ResponseWriter, r *http.Request) {
	defer recoverPanic(w)

	// Get values from the callback url.
	code := r.FormValue("code")
	urlState := r.FormValue("state")

	// Get the state from the session and invalidate the session.
	session, _ := s.session.Get(r, "DiscordAuth")
	cookieState := session.Values["state"]
	session.Options.MaxAge = -1
	session.Values = make(map[interface{}]interface{})
	session.Save(r, w)

	// States from the cookie and the url must match.
	if cookieState != urlState {
		system.HandleError(errors.New("cookie state != url state"),
			fmt.Sprintf("cookieState=%s urlState=%s", cookieState, urlState))
	}

	// Exchange the code for a oauth2 token.
	oauth2token, err := s.oauth2Config.Exchange(
		context.WithValue(oauth2.NoContext, oauth2.HTTPClient, oauth2.HTTPClient), code)
	system.HandleError(err, serviceName+".DiscordAuthCallbackH oauth2Config.Exchange")

	// Get discord's user and eve's character information.
	du := fetchDiscordUser(oauth2token.AccessToken)
	cookie, err := r.Cookie("char")
	system.HandleError(err, serviceName+".DiscordAuthCallbackH get eve char cookie")
	eu := deserializeEveChar(cookie.Value)

	// Create or update user record in the discord service's store.
	user := discord.User{
		EveCharID:           eu.EveCharID,
		EveCharName:         eu.EveCharName,
		EveCorpTicker:       eu.EveCorpTicker,
		EveAlliTicker:       eu.EveAlliTicker,
		DiscordID:           du.ID,
		DiscordAccessToken:  oauth2token.AccessToken,
		DiscordRefreshToken: oauth2token.RefreshToken,
		DiscordTokenScope:   oauth2token.Extra("scope").(string),
		DiscordTokenIsValid: true,
		Active:              true,
	}
	if s.system.Discord.GetStore().DiscordIDExists(user.DiscordID) {
		s.system.Discord.GetStore().UpdateUser(&user)
	} else {
		s.system.Discord.GetStore().CreateUser(&user)
	}

	// Finally, add user to the guild.
	s.system.Discord.GuildAddUser(&user)

	respondOK(w)
}

// fetchDiscordUser retrieves user's information for provided token.
func fetchDiscordUser(accessToken string) *discordUser {
	// Request user's information.
	req, err := http.NewRequest("GET", "https://discordapp.com/api/users/@me", nil)
	system.HandleError(err)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	system.HandleError(err)
	defer resp.Body.Close()

	// Decode the result.
	u := discordUser{}
	d := json.NewDecoder(resp.Body)
	d.UseNumber()
	err = d.Decode(&u)
	system.HandleError(err)

	return &u
}

// deserializeEveChar converts base64 encoded json with eve char data into struct.
func deserializeEveChar(data string) *eveChar {
	// Decode base64 into json.
	j, err := base64.StdEncoding.DecodeString(data)
	system.HandleError(err, serviceName+".deserializeUser DecodeString", "data="+data)

	// Decode json into struct.
	var ec eveChar
	d := json.NewDecoder(bytes.NewReader(j))
	d.UseNumber()
	err = d.Decode(&ec)
	system.HandleError(err, serviceName+".deserializeEveChar Decode", "data="+data)

	return &ec
}
