package gorillahttp

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/prusya/eve-discord-service/pkg/system"
)

func TestNew(t *testing.T) {
	sys := &system.System{
		Config: &system.Config{
			WebServerAddress: ":8080",
		},
	}
	httpservice := New(sys)
	require.Equal(t, sys.HTTP, httpservice)
	require.Equal(t, ":8080", httpservice.server.Addr)
}

func TestStartStop(t *testing.T) {
	sys := &system.System{
		Config: &system.Config{
			WebServerAddress: ":8080",
		},
	}
	httpservice := New(sys)

	httpservice.Start()
	resp, err := http.Get("http://localhost:8080/api/healthcheck")
	require.Nil(t, err)
	require.Equal(t, 200, resp.StatusCode)
	resp.Body.Close()

	httpservice.Stop()
	resp, err = http.Get("http://localhost:8080/api/healthcheck")
	require.NotNil(t, err)
}
