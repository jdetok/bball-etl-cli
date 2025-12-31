package etl

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jdetok/bball-etl-cli/pkg/get"
)

// calls scheduleleaguev2 endpoint for schedule start/end dates
const CURSZN_NBA = "https://cdn.nba.com/static/json/staticData/scheduleLeagueV2.json"
const CURSZN_WNBA = "https://cdn.wnba.com/static/json/staticData/scheduleLeagueV2.json"

type RespSched struct {
	Schedule LeagueSchedule `json:"leagueSchedule"`
}

type LeagueSchedule struct {
	Szn   string     `json:"seasonYear"`
	Lg    string     `json:"leagueId"`
	Dates []GameDate `json:"gameDates"`
}

type GameDate struct {
	GDate string `json:"gameDate"`
	Date  string // formatted
	Games []struct {
		Label string `json:"gameLabel"` // will be "" if regular season
	}
	IsPlayoff bool
}

type LeagueSchedules struct {
	N     LeagueSchedule
	W     LeagueSchedule
	LgMap map[string]map[string]bool
	LgSzn map[string]string
	NMap  map[string]bool
	WMap  map[string]bool
}

func GetCurrentLgSchedules() (*LeagueSchedules, error) {
	ls := &LeagueSchedules{
		LgMap: make(map[string]map[string]bool),
		LgSzn: make(map[string]string),
	}
	lgs := map[string]string{"nba": "00", "wnba": "10"}
	for lgName, lgId := range lgs {

		gr := get.GetReq{
			Host:     fmt.Sprintf("cdn.%s.com", lgName),
			Headers:  get.HDRS,
			Endpoint: "/static/json/staticData/scheduleLeagueV2.json",
		}
		resp, err := RequestSchedule(gr)
		if err != nil {
			return nil, err
		}
		ls.LgSzn[lgId] = resp.Schedule.Szn
		// fmt.Println("in cls:", ls.LgSzn)
		if ls.LgMap[resp.Schedule.Lg] == nil {
			ls.LgMap[resp.Schedule.Lg] = map[string]bool{}
		}
		for i := range resp.Schedule.Dates {
			gdate := &resp.Schedule.Dates[i]
			switch gdate.Games[0].Label {
			case "", "Preseason":
				gdate.IsPlayoff = false
			default: // NEED TO HARDEN THIS
				gdate.IsPlayoff = true
			}

			dt, err := time.Parse("01/02/2006 15:04:05", gdate.GDate)
			if err != nil {
				return nil, err
			}
			gdate.Date = dt.Format("01/02/2006")
			ls.LgMap[resp.Schedule.Lg][gdate.Date] = gdate.IsPlayoff
		}
		// switch lg {
		// case "nba":
		// 	ls.N = resp.Schedule
		// case "wnba":
		// 	ls.W = resp.Schedule
		// }
		// fmt.Println(resp)
	}
	// fmt.Println("bottom cls:", ls.LgSzn)
	return ls, nil
}

func SchedReq(league, season string) get.GetReq {
	var gr = get.GetReq{
		Host:     get.HOST,
		Headers:  get.HDRS,
		Endpoint: "/stats/scheduleleaguev2",
		Params: []get.Pair{
			{"LeagueID", league},
			{"Season", season},
		},
	}
	return gr
}

func RequestSchedule(gr get.GetReq) (*RespSched, error) {
	// fmt.Printf("requesting data from %s...\n", gr.Endpoint)
	body, err := gr.BodyFromReq()
	if err != nil {
		return nil, fmt.Errorf("error getting schedule response: %v", err)
	}

	var resp *RespSched
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("error unmarshaling schedule response: %v", err)
	}
	return resp, nil
}
