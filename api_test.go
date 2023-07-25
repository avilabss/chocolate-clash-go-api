package chocolateclashgoapi

import (
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetMember(t *testing.T) {
	api, err := Init(FWALeague)

	if err != nil {
		log.Panic(err)
	}

	memberTag := "#2UYCC2J9L"
	member, err := api.GetMember(memberTag, 10, 10, true)

	require.NoError(t, err)
	require.Equal(t, member.Tag, memberTag)

	file, _ := json.MarshalIndent(member, "", " ")
	_ = os.WriteFile("test-results/getMember.json", file, 0644)
}
