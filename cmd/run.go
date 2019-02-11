package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"

	"github.com/prusya/eve-discord-service/pkg/discord/dgodiscordservice"
	"github.com/prusya/eve-discord-service/pkg/discord/pgdiscordstore"
	"github.com/prusya/eve-discord-service/pkg/http/gorillahttp"
	"github.com/prusya/eve-discord-service/pkg/system"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "starts the service",
	Long: `usage: eve-discord-service run
It will start the main functionality of the service.
Make sure to run "eve-discord-service init" and fill the config file before "run".`,
	Run: func(cmd *cobra.Command, args []string) {
		// Moved from root cmd cos we need to read config on demand, not always.
		initConfig()

		// Setup logger.
		date := time.Now().Format("2006-01-02_15-04-05")
		f, err := os.OpenFile("logs/log_"+date+".txt", os.O_WRONLY|os.O_CREATE, 0644)
		system.HandleError(err)
		defer f.Close()
		log.SetOutput(f)
		log.SetFlags(log.LstdFlags | log.Lshortfile)

		// Prepare gracefull shutdown.
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGKILL)

		// Create shared System for services.
		sys := system.New(sigChan)

		// Connect to db.
		db, err := sqlx.Connect("postgres", sys.Config.PgConnString)
		system.HandleError(err)
		defer db.Close()

		// Create http service.
		httpService := gorillahttp.New(sys)
		defer httpService.Stop()

		// Create discord service.
		discordStore := pgdiscordstore.New(db)
		discordStore.Init()
		discordService := dgodiscordservice.New(sys, discordStore)
		defer discordService.Stop()

		// Start services.
		httpService.Start()
		discordService.Start()

		// Handle graceful shutdown.
		// Actual shutdown is performed in deferred Stop calls.
		<-sigChan
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
