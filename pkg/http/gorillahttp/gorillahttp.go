package gorillahttp

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"

	"github.com/prusya/eve-discord-service/pkg/system"
)

const (
	serviceName = "gorillahttp"
)

// Service implements http.Service interface backed by gorilla toolkit.
type Service struct {
	system       *system.System
	router       *mux.Router
	server       *http.Server
	session      *sessions.FilesystemStore
	oauth2Config *oauth2.Config
}

// New creates a new Service and prepares it to Start.
func New(system *system.System) *Service {
	session := sessions.NewFilesystemStore("",
		[]byte(system.Config.SessionStoreKey))
	r := mux.NewRouter()

	s := Service{
		system:  system,
		router:  r,
		session: session,
		server: &http.Server{
			Addr:         system.Config.WebServerAddress,
			Handler:      r,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  20 * time.Second,
		},
	}
	s.Routes()

	s.oauth2Config = &oauth2.Config{
		ClientID:     s.system.Config.DiscordClientID,
		ClientSecret: s.system.Config.DiscordClientSecret,
		RedirectURL:  s.system.Config.DiscordCallbackURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  s.system.Config.DiscordAuthURL,
			TokenURL: s.system.Config.DiscordTokenURL,
		},
	}

	s.system.HTTP = &s

	return &s
}

// Start starts the Service.
func (s *Service) Start() {
	go func() {
		defer s.recoverPanic()
		err := s.server.ListenAndServe()
		if err != nil {
			if err != http.ErrServerClosed {
				system.HandleError(err, serviceName+".Start")
			}
		}
	}()
}

func (s *Service) recoverPanic() {
	if r := recover(); r != nil {
		s.system.SigChan <- os.Interrupt
	}
}

// Stop stops the Service.
func (s *Service) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.server.Shutdown(ctx)
}
