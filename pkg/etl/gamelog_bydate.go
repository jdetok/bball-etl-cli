package etl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jdetok/bball-etl-cli/pkg/get"
)

/*
	CUSTOM MODE BY DATE

- user passes -mode custom, and at least one of -dateFrom or -dateTo
- first need to validate in 01/02/2006 notation
- func to accept a start and end date, return laegues mapped to years
- call schedule endpoint
  - need to get schedule for each league for each season within the date range
  - create large map with all the dates for each season
*/
var nba string = "00"
var wnba string = "10"
var dateLayout string = "01/02/2006"

type LgSeasons map[string][]string // i.e. ["00"][]string{"2024-25", "2025-26"}

type LgSznDates map[string]map[string]map[string]bool // i.e. ["00"]["2025-26"]["12/25/2025"]false (nba game on the date, NOT a playoff game)

func NewLgSeasonsMap(dateFrom, dateTo string) (LgSeasons, error) {
	ls := LgSeasons{nba: []string{}, wnba: []string{}}

	// convert string dates to time
	df, err := time.Parse(dateLayout, dateFrom)
	if err != nil {
		return nil, err
	}
	dt, err := time.Parse(dateLayout, dateTo)
	if err != nil {
		return nil, err
	}

	y1 := int(df.Year())
	m1 := int(df.Month())
	y2 := int(dt.Year())
	m2 := int(dt.Month())

	// first seasons
	switch {
	case m1 <= 5:
		ls[nba] = appendIfNew(ls[nba], fmt.Sprintf("%d-%s", y1-1, strconv.Itoa(y1)[2:]))
		ls[wnba] = appendIfNew(ls[wnba], strconv.Itoa(y1-1))
	case m1 > 5 && m1 < 10:
		ls[nba] = appendIfNew(ls[nba], fmt.Sprintf("%d-%s", y1-1, strconv.Itoa(y1)[2:]))
		ls[wnba] = appendIfNew(ls[wnba], strconv.Itoa(y1))
	case m1 >= 10:
		ls[nba] = appendIfNew(ls[nba], fmt.Sprintf("%d-%s", y1, strconv.Itoa(y1 + 1)[2:]))
		ls[wnba] = appendIfNew(ls[wnba], strconv.Itoa(y1))
	}

	// all seasons betweeen first and last
	for i := y1; i < y2; i++ {
		nbaSzn := fmt.Sprintf("%d-%s", i, strconv.Itoa(i + 1)[2:])
		wnbaSzn := strconv.Itoa(i + 1)
		if ls[nba][len(ls[nba])-1] != nbaSzn {
			ls[nba] = appendIfNew(ls[nba], nbaSzn)
		}
		if ls[wnba][len(ls[wnba])-1] != wnbaSzn {
			ls[wnba] = appendIfNew(ls[wnba], wnbaSzn)
		}
	}

	// last seasons
	switch {
	case m2 < 10:
		ls[nba] = appendIfNew(ls[nba], fmt.Sprintf("%d-%s", y2-1, strconv.Itoa(y2)[2:]))
		ls[wnba] = appendIfNew(ls[wnba], strconv.Itoa(y2))
	case m2 >= 10:
		ls[nba] = appendIfNew(ls[nba], fmt.Sprintf("%d-%s", y2, strconv.Itoa(y2 + 1)[2:]))
		ls[wnba] = appendIfNew(ls[wnba], strconv.Itoa(y2))
	}
	return ls, nil
}

func appendIfNew(s []string, v string) []string {
	if len(s) == 0 || s[len(s)-1] != v {
		return append(s, v)
	}
	return s
}

func GetManyLgSchedules(dateFrom, dateTo string) (LgSznDates, error) {
	lgSzns, err := NewLgSeasonsMap(dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	// convert string dates to time
	d1, err := time.Parse(dateLayout, dateFrom)
	if err != nil {
		return nil, err
	}
	d2, err := time.Parse(dateLayout, dateTo)
	if err != nil {
		return nil, err
	}

	fmt.Println(lgSzns)

	lgSchedMap := LgSznDates{}

	for lg, szns := range lgSzns {
		if lgSchedMap[lg] == nil {
			lgSchedMap[lg] = map[string]map[string]bool{}
		}
		for i, szn := range szns {
			if lgSchedMap[lg][szn] == nil {
				lgSchedMap[lg][szn] = map[string]bool{}
			}
			// req goes here
			fmt.Printf("req %d/%d for lg {%s} | szn {%s}...", i+1, len(szns), lg, szn)
			rm := get.NewRequest(
				get.HOST, "/stats/scheduleleaguev2", get.HDRMAP,
				map[string]string{
					"LeagueID": lg,
					"Season":   szn,
				},
			)
			rm.URL = rm.MakeUQueryStr()
			req, err := http.NewRequest(http.MethodGet, rm.URL, nil)

			rm.AddHeaders(req)

			body, err := get.RespFromClient(req)
			if err != nil {
				return nil, fmt.Errorf("error getting response: %v", err)
			}
			fmt.Println(len(body), "bytes received from", rm.Endpt)
			resp := &RespSched{}

			if err := json.Unmarshal(body, &resp); err != nil {
				return nil, fmt.Errorf("error unmarshaling json body: %v", err)
			}

			for ig := range resp.Schedule.Dates {
				gdate := &resp.Schedule.Dates[ig]
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

				if dt.Before(d1) || dt.After(d2) {
					// fmt.Println(lg, ":", dt, "|", d1, "|", d2)
					continue
				}

				gdate.Date = dt.Format("01/02/2006")
				lgSchedMap[lg][szn][gdate.Date] = gdate.IsPlayoff

			}
		}
	}

	return lgSchedMap, nil
}
