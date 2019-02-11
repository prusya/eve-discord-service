package pgdiscordstore

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	db := &sqlx.DB{}
	store := New(db)
	require.Equal(t, db, store.db)
}
