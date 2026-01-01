package etl

import (
	"fmt"
	"strconv"
	"time"
)

/* CUSTOM MODE BY DATE
- user passes -mode custom, and at least one of -dateFrom or -dateTo
- first need to validate in 01/02/2006 notation
- func to accept a start and end date, return laegues mapped to years
- call schedule endpoint
	- need to get schedule for each league for each season within the date range
	- create large map with all the dates for each season
*/

type LgSeasons map[string][]string // i.e. ["00"][]string{"2024-25", "2025-26"}

func NewLgSeasonsMap(dateFrom, dateTo string) (LgSeasons, error) {
	datefmt := "01/06/2006"

	nba := "00"
	wnba := "10"

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
