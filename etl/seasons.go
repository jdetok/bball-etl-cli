package etl

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jdetok/golib/errd"
	"github.com/jdetok/golib/logd"
)

type SeasonLeague struct {
	Szn  string
	WSzn string
}

// pass a time (usually time.Now()), return string with yesterday's date
func Yesterday(dt time.Time) string {
	return dt.Add(-24 * time.Hour).Format("01/02/2006")
}

func SznBSlice(l logd.Logger, start, end string) ([]string, error) {
	e := errd.InitErr()
	startYr, errS := strconv.Atoi(start)
	endYr, errE := strconv.Atoi(end)
	numY := endYr - startYr

	if errS != nil || errE != nil {
		e.Msg = "error converting start or end year to int"
		l.WriteLog(e.Msg)
		return nil, e.NewErr()
	}

	var szns []string
	for y := numY; y > 0; y-- {
		szns = append(szns,
			fmt.Sprintf(
				"%d-%s", startYr+y, strconv.Itoa(startYr + (y + 1))[2:]),
		)
	}
	szns = append(szns, fmt.Sprintf("%d-%s", startYr, strconv.Itoa(startYr + 1)[2:]))
	return szns, nil
}

func SznSlice(l logd.Logger, start, end string) ([]string, error) {
	e := errd.InitErr()
	startYr, errS := strconv.Atoi(start)
	endYr, errE := strconv.Atoi(end)
	numY := endYr - startYr

	if errS != nil || errE != nil {
		e.Msg = "error converting start or end year to int"
		l.WriteLog(e.Msg)
		return nil, e.NewErr()
	}

	var szns []string
	for y := range numY {
		szns = append(szns,
			fmt.Sprintf(
				"%d-%s", startYr+y, strconv.Itoa(startYr + (y + 1))[2:]),
		)
	}
	szns = append(szns, fmt.Sprintf("%d-%s", endYr, strconv.Itoa(endYr + 1)[2:]))
	return szns, nil
}

/*
returns slice of season strings for date (generally pass time.Now())
calling in 2025 will return 2024-25 and 2025-26 and so on
*/
func CurrentSzns(dt time.Time) []string {
	var cyyy string = dt.Format("2006")
	var cy string = dt.AddDate(1, 0, 0).Format("06")

	var pyyy string = dt.AddDate(-1, 0, 0).Format("2006")
	var py string = dt.Format("06")

	return []string{
		fmt.Sprint(pyyy, "-", py),
		fmt.Sprint(cyyy, "-", cy),
	}
}

func GetSeasons() SeasonLeague {
	var sl SeasonLeague
	var crnt []string = CurrentSzns(time.Now())

	m, err := strconv.Atoi(time.Now().Format("1"))
	if err != nil {
		fmt.Println(err)
	}

	// beginning of year through april
	sl.Szn = crnt[0]
	sl.WSzn = crnt[0]

	// may through september
	if m > 5 && m < 10 {
		sl.WSzn = crnt[1]
	}

	// october through end of year
	if m > 10 {
		sl.Szn = crnt[1]
		sl.WSzn = crnt[1]
	}

	// fmt.Printf("NBA Season: %s | WNBA Season: %s\n", sl.Szn, sl.WSzn)
	return sl
}
