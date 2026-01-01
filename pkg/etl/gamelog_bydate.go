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

type LgSeasons map[string][]string // i.e. ["00"][]string{"2024-25", "2025-26"}

type LgGameTypeMap map[string]map[string]bool // i.e. ["00"]["12/25/2025"]false (nba game on the date, NOT a playoff game)

func NewLgSeasonsMap(dateFrom, dateTo string) (LgSeasons, error) {
	datefmt := "01/06/2006"

	ls := LgSeasons{nba: []string{}, wnba: []string{}}

	// convert string dates to time
	df, err := time.Parse(datefmt, dateFrom)
	if err != nil {
		return nil, err
	}
	dt, err := time.Parse(datefmt, dateTo)
	if err != nil {
		return nil, err
	}

	y1 := int(df.Year())
	m1 := int(df.Month())
	y2 := int(dt.Year())
	m2 := int(dt.Month())

	switch {
	case m1 < 5:
		ls[nba] = append(ls[nba], fmt.Sprintf("%d-%s", y1-1, strconv.Itoa(y1)[2:]))
		ls[wnba] = append(ls[wnba], strconv.Itoa(y1-1))
	case m1 > 5 && m1 < 10:
		ls[nba] = append(ls[nba], fmt.Sprintf("%d-%s", y1-1, strconv.Itoa(y1)[2:]))
		ls[wnba] = append(ls[wnba], strconv.Itoa(y1))
	case m1 > 10:
		ls[nba] = append(ls[nba], fmt.Sprintf("%d-%s", y1, strconv.Itoa(y1 + 1)[2:]))
		ls[wnba] = append(ls[wnba], strconv.Itoa(y1))
	}

	for i := y1; i < y2; i++ {
		ls[nba] = append(ls[nba], fmt.Sprintf("%d-%s", i, strconv.Itoa(i + 1)[2:]))
		ls[wnba] = append(ls[wnba], strconv.Itoa(i+1))
	}

	switch {
	case m2 < 5:

	case m2 > 5 && m2 < 10:
		ls[wnba] = append(ls[wnba], strconv.Itoa(y2))
	case m2 > 10:
		ls[nba] = append(ls[nba], fmt.Sprintf("%d-%s", y2, strconv.Itoa(y2 + 1)[2:]))
		ls[wnba] = append(ls[wnba], strconv.Itoa(y2))
	}

	return ls, nil
}

func GetManyLgSchedules(dateFrom, dateTo string) (LgGameTypeMap, error) {
	lgSzns, err := NewLgSeasonsMap(dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	fmt.Println(lgSzns)

	lgSchedMap := LgGameTypeMap{}

	for lg, szns := range lgSzns {
		for i, szn := range szns {
			// req goes here
			fmt.Printf("req %d/%d: lg {%s} szn {%s}\n", i+1, len(szns)*len(lgSzns), lg, szn)
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

			resp := &RespSched{}

			if err := json.Unmarshal(body, &resp); err != nil {
				return nil, fmt.Errorf("error unmarshaling json body: %v", err)
			}

			if lgSchedMap[lg] == nil {
				lgSchedMap[lg] = map[string]bool{}
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
				gdate.Date = dt.Format("01/02/2006")
				lgSchedMap[lg][gdate.Date] = gdate.IsPlayoff

			}
		}
	}

	return lgSchedMap, nil
}
