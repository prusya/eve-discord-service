package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/prusya/eve-discord-service/pkg/system"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "creates config.json and logs/",
	Long: `usage: eve-discord-service init
It will create a config template and logs directory.
You must create them in order or run the service.
Make sure to fill the config with proper values.`,
	Run: func(cmd *cobra.Command, args []string) {
		c := system.Config{
			WebServerAddress: "127.0.0.1:8082",

			SessionStoreKey: randomString(32),

			DiscordClientID:     "",
			DiscordClientSecret: "",
			DiscordCallbackURL:  "",
			DiscordBotToken:     "",
			DiscordAuthURL:      "https://discordapp.com/api/oauth2/authorize",
			DiscordTokenURL:     "https://discordapp.com/api/oauth2/token",
			DiscordAuthScopes:   []string{"identify", "guilds.join", "guilds"},
			DiscordGuildID:      "",
			DiscordGuildRoles:   []string{""},

			DiscordKickUserWithRevokedToken: true,

			UsersValidationEndpoint: "http://127.0.0.1:8081/api/validation/discord",

			PgConnString: "postgres://username:password@hostaddress/dbname?sslmode=verify-full",
		}

		// Create config.json file.
		cj, _ := json.MarshalIndent(c, "", "  ")
		err := ioutil.WriteFile("config.json", cj, 0644)
		if err != nil {
			panic(err)
		}
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		fmt.Println("Config file created at", cwd)

		// Create logs dir.
		err = os.Mkdir("logs", 0644)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	rand.Seed(time.Now().UnixNano())
}

func randomString(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^*()_+|-=")

	b := make([]rune, n)

	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}
