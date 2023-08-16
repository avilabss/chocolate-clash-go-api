package chocolateclashgoapi

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type API struct {
	BaseUrl        string
	CollyCollector *colly.Collector
}

// Initilize choclateclash api by selecting an appropriate league
func Init(league string) (*API, error) {
	if league != FWALeague && league != OtherLeague {
		return nil, fmt.Errorf("%s: %s", ErrUnknownLeague, league)
	}

	api := API{}

	api.BaseUrl = fmt.Sprintf("https://%s.chocolateclash.com/cc_n", league)

	// Setup colly
	collyCollector := colly.NewCollector(
		// colly.AllowedDomains("cc.chocolateclash.com", "fwa.chocolateclash.com"),
		colly.AllowURLRevisit(),
		colly.Async(true),
	)

	collyCollector.Limit(&colly.LimitRule{
		// DomainGlob:  "*.chocolateclash.com/*",
		Delay:       1 * time.Second,
		RandomDelay: 1 * time.Second,
	})

	api.CollyCollector = collyCollector

	return &api, nil
}

func (api *API) FixWarPid(tag string, actions_count int, war_attacks_count int) error {
	var returnError error = nil
	c := api.CollyCollector.Clone()

	c.OnRequest(func(r *colly.Request) {
		if strings.Contains(r.URL.String(), "fixwarpidissue") {
			return
		}

		log.Printf("Fixing war pid issue: %s\n", tag)
	})

	c.OnError(func(_ *colly.Response, err error) {
		returnError = err
	})

	c.OnHTML("td:nth-child(2)", func(e *colly.HTMLElement) {
		e.ForEach("a:nth-child(3)", func(_ int, a *colly.HTMLElement) {
			warPidUrl := a.Attr("href")
			warPidUrl = fmt.Sprintf("%s/%s", api.BaseUrl, warPidUrl)

			if strings.Contains(warPidUrl, "fixwarpidissue") {
				log.Printf("Visiting fix url: %s", warPidUrl)
				c.Visit(warPidUrl)
			}
		})
	})

	url := fmt.Sprintf("%s/member.php?tag=%s&rlim=%v&slim=%v", api.BaseUrl, tag, actions_count, war_attacks_count)

	c.Visit(url)
	c.Wait()

	return returnError
}

// Get member details
func (api *API) GetMember(tag string, actions_count int, war_attacks_count int, fix_errors bool) (*Member, error) {
	if strings.HasPrefix(tag, "#") {
		tag = url.QueryEscape(tag)
	}

	var returnError error = nil
	var member Member = Member{}

	if fix_errors {
		err := api.FixWarPid(tag, actions_count, war_attacks_count)

		if err != nil {
			return nil, fmt.Errorf("%s: %s", ErrFailedToFixWarPid, err)
		}
	}

	c := api.CollyCollector.Clone()

	c.OnRequest(func(r *colly.Request) {
		log.Printf("Getting member: %s\n", tag)
	})

	c.OnError(func(_ *colly.Response, err error) {
		returnError = err
	})

	// Top metadata
	c.OnHTML("#top", func(e *colly.HTMLElement) {
		e.ForEach("a", func(_ int, a *colly.HTMLElement) {
			href := a.Attr("href")

			if strings.HasPrefix(href, "clashofclans://") {
				member.InGameUrl = href
			} else if strings.HasPrefix(href, "clan.php?tag=") {
				member.Clan.Url = fmt.Sprintf("%s/%s", api.BaseUrl, href)
			}
		})

		text := e.Text

		// Compile regexp for elements
		regexpTag, _ := regexp.Compile(`for\s#[\w\d]+`)
		regexpName, _ := regexp.Compile(`Name:\s[\w\d.-]+`)
		regexpSync, _ := regexp.Compile(`Synchronized:\s\w+`)
		regexpDonations, _ := regexp.Compile(`Donates:\s\d+`)
		regexpDonationsReceived, _ := regexp.Compile(`Receives:\s\d+`)
		regexpTownHallLevel, _ := regexp.Compile(`Town\sHall:\s\d+`)
		regexpRole, _ := regexp.Compile(`Rank:\s[\w+-]+`)
		regexpClanMetaData, _ := regexp.Compile(`Clan:\s.+Donates:`)

		// Find string
		tagStr := regexpTag.FindString(text)
		nameStr := regexpName.FindString(text)
		syncStr := regexpSync.FindString(text)
		donationsStr := regexpDonations.FindString(text)
		donationsReceivedStr := regexpDonationsReceived.FindString(text)
		townHallLevelStr := regexpTownHallLevel.FindString(text)
		roleStr := regexpRole.FindString(text)
		clanMetaData := regexpClanMetaData.FindString(text)

		var tag string
		var name string
		var sync bool
		var donations int
		var donationsReceived int
		var townHallLevel int
		var role string
		var clanTag string
		var clanName string
		var clanLeague string

		tag = strings.Replace(tagStr, "for ", "", 1)
		name = strings.Replace(nameStr, "Name: ", "", 1)
		syncStr = strings.Replace(syncStr, "Synchronized: ", "", 1)

		if strings.ToLower(syncStr) == "yes" {
			sync = true
		} else {
			sync = false
		}

		donationsStr = strings.Replace(donationsStr, "Donates: ", "", 1)
		donations, _ = strconv.Atoi(donationsStr)

		donationsReceivedStr = strings.Replace(donationsReceivedStr, "Receives: ", "", 1)
		donationsReceived, _ = strconv.Atoi(donationsReceivedStr)

		townHallLevelStr = strings.Replace(townHallLevelStr, "Town Hall: ", "", 1)
		townHallLevel, _ = strconv.Atoi(townHallLevelStr)

		role = strings.Replace(roleStr, "Rank: ", "", 1)

		clanMetaData = strings.Replace(clanMetaData, "Clan: ", "", 1)
		clanMetaData = strings.Replace(clanMetaData, "Donates:", "", 1)
		clanMetaDataParts := strings.Split(clanMetaData, "(")

		clanName = strings.TrimSpace(clanMetaDataParts[0])

		clanTag = strings.TrimSpace(clanMetaDataParts[1])
		clanTag = strings.Replace(clanTag, ")", "", 1)
		clanTag = fmt.Sprintf("#%s", clanTag)

		clanLeague = strings.TrimSpace(clanMetaDataParts[2])
		clanLeague = strings.Replace(clanLeague, ")", "", 1)

		// Assign values
		member.Tag = tag
		member.Name = name
		member.Synchronized = sync
		member.Donations = donations
		member.DonationsReceived = donationsReceived
		member.TownHallLevel = townHallLevel
		member.Role = role
		member.Clan.Tag = clanTag
		member.Clan.Name = clanName
		member.Clan.League = clanLeague
	})

	// Tables
	c.OnHTML("table > tbody", func(e *colly.HTMLElement) {
		if strings.Contains(e.ChildText("tr:nth-child(1) > td:nth-child(2)"), "Action") {
			// Action table
			e.ForEach("tr:not(:first-child):not(:last-child)", func(_ int, a *colly.HTMLElement) {
				timestamp := a.ChildText("td:nth-child(1)")
				action := a.ChildText("td:nth-child(2)")
				clanUrl := a.ChildAttr("td:nth-child(3) > a", "href")
				clanName := a.ChildText("td:nth-child(3) > a")
				clanLeague := a.ChildText("td:nth-child(3) > span > span")

				clanUrlParts := strings.Split(clanUrl, "=")
				clanTag := clanUrlParts[len(clanUrlParts)-1]

				member.Actions = append(member.Actions, Action{
					Timestamp: timestamp,
					Action:    action,
					Clan: Clan{
						Tag:    fmt.Sprintf("#%s", clanTag),
						Name:   clanName,
						League: clanLeague,
						Url:    fmt.Sprintf("%s/%s", api.BaseUrl, clanUrl),
					},
				})
			})

		} else if strings.Contains(e.ChildText("tr:nth-child(1) > td:nth-child(2)"), "Information") {
			// Information table
			e.ForEach("tr:not(:first-child):not(:last-child)", func(_ int, a *colly.HTMLElement) {
				timestamp := a.ChildText("td:nth-child(1)")
				information := a.ChildText("td:nth-child(2)")
				fixWarPidUrl := a.ChildAttr("td:nth-child(2) > a:nth-child(3)", "href")
				memberOnClanUrl := a.ChildAttr("td:nth-child(2) > a:nth-child(4)", "href")
				opponentClanUrl := a.ChildAttr("td:nth-child(2) > a:nth-child(6)", "href")

				var fixWarPid bool
				var warPidUrl *string
				var memberOnClan *Clan
				var opponentClan *Clan
				var logColor *string

				if fixWarPidUrl != "" {
					logColor = nil
					fixWarPid = true
					fixWarPidUrl = fmt.Sprintf("%s/%s", api.BaseUrl, fixWarPidUrl)
					warPidUrl = &fixWarPidUrl
				} else {
					color := a.ChildAttr("td:nth-child(2) > span[style]", "style")
					color = strings.Replace(color, "color:", "", 1)
					color = strings.Replace(color, ";", "", 1)

					fixWarPid = false
					warPidUrl = nil
					logColor = &color
				}

				if memberOnClanUrl != "" {
					memberOnClanUrl = fmt.Sprintf("%s/%s", api.BaseUrl, memberOnClanUrl)

					memberOnClanUrlParts := strings.Split(memberOnClanUrl, "=")
					memberOnClanTag := memberOnClanUrlParts[len(memberOnClanUrlParts)-1]

					memberOnClanName := a.ChildText("td:nth-child(2) > a:nth-child(4)")
					memberOnClanLeague := a.ChildText("td:nth-child(2) > span:nth-child(5)")
					memberOnClanLeague = strings.Replace(memberOnClanLeague, "(", "", 1)
					memberOnClanLeague = strings.Replace(memberOnClanLeague, ")", "", 1)

					memberOnClan = &Clan{
						Tag:    memberOnClanTag,
						Name:   memberOnClanName,
						League: memberOnClanLeague,
						Url:    memberOnClanUrl,
					}
				} else {
					memberOnClan = nil
				}

				if opponentClanUrl != "" {
					opponentClanUrl = fmt.Sprintf("%s/%s", api.BaseUrl, opponentClanUrl)

					opponentClanUrlParts := strings.Split(opponentClanUrl, "=")
					opponentClanTag := opponentClanUrlParts[len(opponentClanUrlParts)-1]

					opponentClanName := a.ChildText("td:nth-child(2) > a:nth-child(6)")
					opponentClanLeague := a.ChildText("td:nth-child(2) > span:nth-child(7)")
					opponentClanLeague = strings.Replace(opponentClanLeague, "(", "", 1)
					opponentClanLeague = strings.Replace(opponentClanLeague, ")", "", 1)

					opponentClan = &Clan{
						Tag:    opponentClanTag,
						Name:   opponentClanName,
						League: opponentClanLeague,
						Url:    opponentClanUrl,
					}
				} else {
					opponentClan = nil
					logColor = nil
				}

				member.Attacks = append(member.Attacks, Attack{
					Timestamp:    timestamp,
					Information:  information,
					Color:        logColor,
					MemberOnClan: memberOnClan,
					OpponentClan: opponentClan,
					FixWarPid:    fixWarPid,
					FixWarPidUrl: warPidUrl,
				})
			})

		} else if strings.Contains(e.ChildText("tr:nth-child(1) > td:nth-child(2)"), "Note") {
			// Note table
			e.ForEach("tr:not(:first-child)", func(_ int, a *colly.HTMLElement) {
				timestamp := a.ChildText("td:nth-child(1)")
				note := a.ChildText("td:nth-child(2)")
				author := a.ChildText("td:nth-child(3)")

				member.Notes = append(member.Notes, Note{
					Timestamp: timestamp,
					Note:      note,
					Author:    author,
				})
			})
		}

	})

	url := fmt.Sprintf("%s/member.php?tag=%s&rlim=%v&slim=%v", api.BaseUrl, tag, actions_count, war_attacks_count)

	c.Visit(url)
	c.Wait()

	if returnError != nil {
		return nil, returnError
	}

	return &member, nil
}
