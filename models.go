package chocolateclashgoapi

const (
	FWALeague   = "fwa"
	OtherLeague = "cc"
)

type Member struct {
	Tag               string   `json:"tag"`
	Name              string   `json:"name"`
	Synchronized      bool     `json:"synchronized"`
	InGameUrl         string   `json:"inGameUrl"`
	Donations         int      `json:"donations"`
	DonationsReceived int      `json:"donationsReceived"`
	TownHallLevel     int      `json:"townHallLevel"`
	Role              string   `json:"role"`
	Clan              Clan     `json:"clan"`
	Actions           []Action `json:"actions"`
	Attacks           []Attack `json:"attacks"`
	Notes             []Note   `json:"notes"`
}

type Clan struct {
	Tag    string `json:"tag"`
	Name   string `json:"name"`
	League string `json:"league"`
	Url    string `json:"url"`
}

type Action struct {
	Timestamp string `json:"timestamp"`
	Action    string `json:"action"`
	Clan      Clan   `json:"clan"`
}

type Attack struct {
	Timestamp    string  `json:"timestamp"`
	Information  string  `json:"information"`
	Color        *string `json:"color"`
	MemberOnClan *Clan   `json:"memberOnClan"`
	OpponentClan *Clan   `json:"opponentClan"`
	FixWarPid    bool    `json:"fixWarPid"`
	FixWarPidUrl *string `json:"fixWarPidUrl"`
}

type Note struct {
	Timestamp string `json:"timestamp"`
	Note      string `json:"note"`
	Author    string `json:"author"`
}
