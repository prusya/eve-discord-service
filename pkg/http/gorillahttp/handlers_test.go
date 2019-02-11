package gorillahttp

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeserializeEveChar(t *testing.T) {
	referenceEC := eveChar{
		EveCharID:     1,
		EveCorpID:     2,
		EveAlliID:     3,
		EveCharName:   "char name",
		EveCorpName:   "corp name",
		EveAlliName:   "alli name",
		EveCorpTicker: "corp ticker",
		EveAlliTicker: "alli ticker",
	}
	j, err := json.Marshal(&referenceEC)
	require.Nil(t, err)
	data := base64.StdEncoding.EncodeToString(j)

	ec := deserializeEveChar(data)
	require.Equal(t, referenceEC.EveCharID, ec.EveCharID)
	require.Equal(t, referenceEC.EveCharName, ec.EveCharName)
}
