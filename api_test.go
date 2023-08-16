package chocolateclashgoapi

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetMember(t *testing.T) {
	api, err := Init(FWALeague)

	if err != nil {
		log.Panic(err)
	}

	memberTag := "#PRQL2VJR"
	member, err := api.GetMember(memberTag, 0, 20, true)

	require.NoError(t, err)
	require.Equal(t, member.Tag, memberTag)

	file, _ := json.MarshalIndent(member, "", " ")
	_ = os.WriteFile("test-results/getMember.json", file, 0644)
}

func TestMemberEligibility(t *testing.T) {
	api, err := Init(FWALeague)

	if err != nil {
		log.Panic(err)
	}

	memberTag := "#PRQL2VJR"
	member, _ := api.GetMember(memberTag, 0, 20, true)

	var eligibleAttacks []Attack

	for x := 0; x < len(member.Attacks); x++ {
		fullTimeStr := fmt.Sprintf("%s %s", member.Attacks[x].Timestamp, "00:00:00")
		fullTime, _ := time.Parse("2006-01-02 15:04:05", fullTimeStr)

		nowTime := time.Now().UTC()
		xDaysBefore := nowTime.AddDate(0, 0, -30)

		if fullTime.After(xDaysBefore) {
			eligibleAttacks = append(eligibleAttacks, member.Attacks[x])
		}
	}

	isEligible := true

	for x := 0; x < len(eligibleAttacks); x++ {
		color := *eligibleAttacks[x].Color

		if color == "purple" || color == "red" {
			isEligible = false
			break
		}
	}

	fmt.Printf("Member Eligible: %v\n", isEligible)
}
