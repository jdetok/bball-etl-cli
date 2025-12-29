package etl

import (
	"encoding/json"
	"fmt"
)

// calls scheduleleaguev2 endpoint for schedule start/end dates

type RespSched struct {
	Dates GameDates `json:"leagueSchedule"`
}

// main json object in response body after endpoint/params
type GameDates struct {
	GmDates []GameDate `json:"gameDates"`
}

type GameDate struct {
	Date string `json:"gameDate"`
}

func SchedReq(league, season string) GetReq {
	var gr = GetReq{
		Host:     HOST,
		Headers:  HDRS,
		Endpoint: "/stats/scheduleleaguev2",
		Params: []Pair{
			{"LeagueID", league},
			{"Season", season},
		},
	}
	return gr
}

func RequestSchedule(gr GetReq) error {
	fmt.Printf("requesting data from %s...\n", gr.Endpoint)
	body, err := gr.BodyFromReq()
	if err != nil {
		return fmt.Errorf("error getting schedule response: %v", err)
	}

	var resp RespSched
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("error unmarshaling schedule response: %v", err)
	}

	fmt.Println(resp.Dates)
	return nil
}
