package etl

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/jdetok/bball-etl-cli/pkg/cnf"
	"github.com/jdetok/bball-etl-cli/pkg/get"
	"github.com/jdetok/bball-etl-cli/pkg/pgdb"
)

func GLogParams() pgdb.LgTbls {
	var lt pgdb.LgTbls
	lt.Lgs = []string{"00", "10"}
	lt.Tbls = []pgdb.Table{
		{
			Name:    "intake.gm_team",
			PrimKey: "game_id, team_id",
			PlTm:    "T",
		},
		{
			Name:    "intake.gm_player",
			PrimKey: "game_id, player_id",
			PlTm:    "P",
		},
	}
	return lt
}

func GameLogReqNew(league, season, sType, plTm, dateFrom, dateTo string) get.GetReq {
	var gr = get.GetReq{
		Host:     get.HOST,
		Headers:  get.HDRS,
		Endpoint: "/stats/leaguegamelog",
		Params: []get.Pair{
			{"LeagueID", league},
			{"Season", season},
			{"SeasonType", sType},
			{"Counter", "0"},
			{"Sorter", "DATE"},
			{"Direction", "DESC"},
			{"PlayerOrTeam", plTm},
			{"DateFrom", dateFrom},
			{"DateTo", dateTo},
		},
	}
	fmt.Println("requesting gamelogs:", gr.MakeFulLURL())
	return gr
}

// main api for creating gamelog requests
func GameLogETL(c *cnf.Conf, r get.GetReq, tbl, primKey string) error {
	// call endpoint in HTTP request, return Resp struct
	resp, err := get.RequestResp(r)
	if err != nil {
		return fmt.Errorf("error getting response for %s: %v", r.Endpoint, err)
	}

	// get cols/rows from resp, return early when no rows in response
	var cols []string = resp.ResultSets[0].Headers
	var rows [][]any = resp.ResultSets[0].RowSet
	if len(rows) == 0 {
		c.Lg.Infof("response returned 0 rows, exiting")
		return nil
	}
	c.Lg.Infof("response returned %d fields & %d rows", len(cols), len(rows))

	// prepare the sql statement & chunks of values
	ins := pgdb.MakeInsert(
		tbl,
		primKey,
		cols,
		rows,
	) // attempt to insert rows from response
	return ins.InsertFast(c)
}

// nightly game log fetch both PlayerTeam=P & T and NBA and WNBA
// using yeseterday's date as DateFrom/DateTo
func GLogDailyETL(c *cnf.Conf, df, dt string) error {
	lt := GLogParams()
	sl := GetSeasons()
	var szns = []string{sl.Szn, sl.WSzn}

	sch, err := GetCurrentLgSchedules()
	if err != nil {
		return fmt.Errorf("error getting schedules: %v", err)
	}

	sTypes := []string{"Playoffs", "Regular+Season"}
	// makes 4 calls to leaguegamelog endpoint
	for i, lg := range lt.Lgs { // outer loop, 2 calls per lg

		for _, t := range lt.Tbls {
			for _, s := range sTypes {
				if isPlOff, ok := sch.LgMap[lg][df]; ok {
					if (s == "Playoffs" && !isPlOff) || (s == "Regular+Season" && isPlOff) {
						continue
					}
					// create request
					r := GameLogReqNew(lg, szns[i], s, t.PlTm, df, dt)
					c.Lg.Infof("attempting to fetch %s: LG=%s, SZN=%s %s, PLTM=%s, DATE=%s",
						r.Endpoint, lg, szns[i], s, t.PlTm, df)
					// run etl
					err := GameLogETL(c, r, t.Name, t.PrimKey)
					if err != nil {
						return fmt.Errorf("error during daily game log ETL. LG=%s, SZN=%s, PLTM=%s, DATE=%s: %v",
							lg, szns[i], t.PlTm, df, err)
					}
				} else {
					c.Lg.Infof("SKIPPING LG=%s, SZN=%s %s, DATE=%s", lg, szns[i], s, df)
					continue
				}
			}
			// success, next call
			c.Lg.Infof("finished with LG=%s, SZN=%s, PLTM=%s, DATE=%s",
				lg, szns[i], t.PlTm, df)
		}
	}
	return nil
}

// run single season
func GetManyGLogs(c *cnf.Conf, lgs []string, tbls []pgdb.Table, szn string) error {
	for i := range lgs { // outer loop, 2 calls per lg
		sznY, err := strconv.Atoi(szn[:4])
		if err != nil {
			return fmt.Errorf("getting int from season %s", szn)
		} // no wnba pre 1997
		if lgs[i] == "10" && sznY < 1996 {
			c.Lg.Infof("skipping WNBA %s - first WNBA season was 1997-98", szn)
			continue
		} // loop through tables (PlTm, intake.gm_team, intake.gm_player)
		for _, t := range tbls {
			// get player/team reg and playoffs
			for _, s := range []string{"Regular+Season", "Playoffs"} {
				// create request
				r := GameLogReqNew(lgs[i], szn, s, t.PlTm, "", "")
				c.Lg.Infof("attempting to fetch %s: LG=%s, SZN=%s %s, PLTM=%s",
					r.Endpoint, lgs[i], szn, s, t.PlTm)

				// attempt to fetch & insert for current iteration
				// func returns run of insert
				err := GameLogETL(c, r, t.Name, t.PrimKey)
				if err != nil {
					return fmt.Errorf("error during daily game log ETL. LG=%s, SZN=%s %s, PLTM=%s",
						lgs[i], szn, s, t.PlTm)
				}
				// success, next call
				c.Lg.Infof("finished with LG=%s, SZN=%s %s, PLTM=%s", lgs[i], szn, s, t.PlTm)
			}
		}
	}
	return nil
}

// TODO: specific season/league ETL
func LgSznGlogs(c *cnf.Conf, lg, szn string) error {
	lt := GLogParams()
	var lg_id string
	switch lg {
	case "nba":
		lg_id = lt.Lgs[0] // "00"
	case "wnba":
		lg_id = lt.Lgs[1] // "10"
	}

	for _, t := range lt.Tbls {
		for _, s := range []string{"Regular+Season", "Playoffs"} {
			// create request
			r := GameLogReqNew(lg_id, szn, s, t.PlTm, "", "")
			c.Lg.Infof("attempting to fetch %s: LG=%s, SZN=%s %s, PLTM=%s", r.Endpoint, lg, szn, s, t.PlTm)

			// attempt to fetch & insert for current iteration
			// func returns run of insert
			err := GameLogETL(c, r, t.Name, t.PrimKey)
			if err != nil {
				return fmt.Errorf("error during daily game log ETL. LG=%s, SZN=%s %s, PLTM=%s",
					lg, szn, s, t.PlTm)
			}
			// success, next call
			c.Lg.Infof("finished with LG=%s, SZN=%s %s, PLTM=%s", lg, szn, s, t.PlTm)
		}
	}
	return nil
}

// should be able to use this for the custom mode without league specified
func GLogSeasonETL(c *cnf.Conf, szn string) error {
	lt := GLogParams()
	err := GetManyGLogs(c, lt.Lgs, lt.Tbls, szn)
	if err != nil {
		errMsg := fmt.Sprintf("error running ETL for %s: %v", szn, err)
		c.Errs = append(c.Errs, errMsg) // capture if an error occured
		return errors.New(errMsg)
	}
	return nil
}
