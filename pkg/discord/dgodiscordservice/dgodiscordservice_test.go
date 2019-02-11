package dgodiscordservice

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/prusya/eve-discord-service/pkg/discord/pgdiscordstore"
	"github.com/prusya/eve-discord-service/pkg/system"
)

func TestDgodiscordservice(t *testing.T) {
	sys := &system.System{
		Config: &system.Config{},
	}
	db := &sqlx.DB{}
	store := pgdiscordstore.New(db)

	t.Run("TestNew", func(t *testing.T) {
		dgodiscordservice := New(sys, store)
		require.Equal(t, dgodiscordservice, sys.Discord)
	})

	t.Run("TestGetStore", func(t *testing.T) {
		dgodiscordservice := New(sys, store)
		require.Equal(t, store, dgodiscordservice.GetStore())
	})

	t.Run("TestRecoverPanic", func(t *testing.T) {
		defer recoverPanic()

		panic("panic")
		t.Fail()
	})
}
