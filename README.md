eve-discord-service
---

[![CircleCI](https://circleci.com/gh/prusya/eve-discord-service.svg?style=shield)](https://circleci.com/gh/prusya/eve-discord-service) [![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/prusya/eve-discord-service) [![Maintainability](https://api.codeclimate.com/v1/badges/7f2f285d036b7196263c/maintainability)](https://codeclimate.com/github/prusya/eve-discord-service/maintainability) [![codecov](https://codecov.io/gh/prusya/eve-discord-service/branch/master/graph/badge.svg)](https://codecov.io/gh/prusya/eve-discord-service) [![license](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](https://raw.githubusercontent.com/prusya/eve-discord-service/master/LICENSE)

Service to add/del discord users to/from guilds.

## features

* Discord sso auth
* Automatic addition/deletion of users to/from discord guild
* Automatic rename of users based on their eve char's name and corp/alli ticker

## requirements
`go 1.11+`

`postgresql 9.2+`

## install

```bash
go install github.com/prusya/eve-discord-service
```

or

Get the latest binary from releases

## build from source

```bash
go get -u github.com/prusya/eve-discord-service

# inside `github.com/prusya/eve-discord-service` directory
go build
```

## usage

```bash
# navigate to directory where you want to store config and log files
# you must have create/write permissions

eve-discord-service init

# fill in config.json file

eve-discord-service run
```

## config file

example config is in `example.config.json` in repo's root
```
accept http requests on this address
"WebServerAddress": "127.0.0.1:8083"

32 characters long string used
(should be filled automatically during `eve-discord-service init`)
"SessionStoreKey": "ransomString"

create discord application and get credentials here
  https://discordapp.com/developers/applications/
"DiscordClientID": "clientID"
"DiscordClientSecret": "clientSecret"
"DiscordBotToken": "token"
"DiscordCallbackURL": "https://hostaddress/auth/discord/callback"

discord's sso endpoints
(should be filled automatically during `eve-discord-service init`)
"DiscordAuthURL": "https://discordapp.com/api/oauth2/authorize"
"DiscordTokenURL": "https://discordapp.com/api/oauth2/token"

request these scopes from user during login
(at least `identify` and `guilds.join` are required to use this service)
"DiscordAuthScopes": ["identify", "guilds.join"]
  
add user to this guild upon successful auth
"DiscordGuildID": "1234567890"

assign this roles to user upon successful auth
"DiscordGuildRoles": ["1234567890"]

whether to kick user from guild or not if access token was revoked
"DiscordKickUserWithRevokedToken": true,

send requests to validate users to this endpoint(address where `eve-auth-gateway-service` runs)
"UsersValidationEndpoint": "http://127.0.0.1:8081/api/validation/discord"

connection string to connect to postgresql database
set `sslmode=disable` if secure connection is not configured
"PgConnString": "postgres://username:password@hostaddress/dbname?sslmode=verify-full"
```

## notes

This service is a part of bundle of other services and is supposed to be run behind and contacted only by [https://github.com/prusya/eve-auth-gateway-service](https://github.com/prusya/eve-auth-gateway-service)

This service does not provide mechanisms to restrict connections to it and it's up to end user to limit who can connect to it. Suggested solution is to use a firewall to allow connections only from `eve-auth-gateway-service` host