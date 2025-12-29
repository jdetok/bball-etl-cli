package etl

import (
	"fmt"
)

func PlayerReq(onlyCurrent, league, season string) GetReq {
	var gr = GetReq{
		Host:     HOST,
		Headers:  HDRS,
		Endpoint: "/stats/commonallplayers",
		Params: []Pair{
			{"IsOnlyCurrentSeason", onlyCurrent},
			{"LeagueID", league},
			{"Season", season},
		},
	}
	return gr
}

func PlayersParams() LgTbls {
	var lt LgTbls
	lt.lgs = []string{"00", "10"}
	lt.tbls = []Table{
		{
			Name:    "intake.player",
			PrimKey: "person_id",
		},
		{
			Name:    "intake.wplayer",
			PrimKey: "person_id",
		},
	}
	return lt
}

// SAME AS CURRENT PLAYER ETL BUT FOR INDIVIDUAL SEASON
// WILL NEED A NEW GET SEASONS FUNCTION AS WELL
func SznPlayersETL(cnf *Conf, onlyCurrent, season string) error {
	pp := PlayersParams()
	cnf.Lg.Infof("attempting players ETL for %s nba/wnba seasons", season)

	for i := range pp.lgs {
		var lg string
		switch pp.lgs[i] {
		case "00":
			lg = "nba"
		case "10":
			lg = "wnba"
		}

		cnf.Lg.Infof(fmt.Sprintf("attempting to insert %s %s players", season, lg))
		// r := PlayerReq(onlyCurrent, p[0], p[1])
		r := PlayerReq(onlyCurrent, pp.lgs[i], season)
		resp, err := RequestResp(r)
		if err != nil {
			return fmt.Errorf("error getting response for %s: lg: %s szn: %s: %v", r.Endpoint, lg, season, err)
		}

		// get cols/rows from resp, return early when no rows in response
		var cols []string = resp.ResultSets[0].Headers
		var rows [][]any = resp.ResultSets[0].RowSet
		// ProcessResp(resp)
		fmt.Println("Cols Length:", len(cols), "Rows Length:", len(rows))

		if len(rows) == 0 {
			cnf.Lg.Infof("response returned 0 rows, exiting")
			return nil
		}
		cnf.Lg.Infof("response returned %d fields & %d rows", len(cols), len(rows))

		// prepare the sql statement & chunks of values
		ins := MakeInsert(
			pp.tbls[i].Name,
			pp.tbls[i].PrimKey,
			cols,
			rows,
		) // attempt to insert rows from response
		ins.InsertFast(cnf)

		cnf.Lg.Infof("%s %s players ETL complete", season, lg)
	}
	cnf.Lg.Infof("players ETL complete for ", season)
	return nil
}

func CrntPlayersETL(cnf *Conf) error {
	sl := GetSeasons()
	var szns = []string{sl.Szn, sl.WSzn}
	pp := PlayersParams()

	cnf.Lg.Infof("attempting current players ETL for %s nba season and %s wnba season", sl.Szn, sl.WSzn)

	for i := range pp.lgs {
		var lg string
		switch pp.lgs[i] {
		case "00":
			lg = "nba"
		case "10":
			lg = "wnba"
		}

		cnf.Lg.Infof("attempting to insert current %s players", lg)
		// r := PlayerReq(onlyCurrent, p[0], p[1])
		r := PlayerReq("1", pp.lgs[i], szns[i])
		resp, err := RequestResp(r)
		if err != nil {
			return fmt.Errorf("error getting response for %s: %v", r.Endpoint, err)
		}

		// get cols/rows from resp, return early when no rows in response
		var cols []string = resp.ResultSets[0].Headers
		var rows [][]any = resp.ResultSets[0].RowSet
		// ProcessResp(resp)
		fmt.Println("Cols Length:", len(cols), "Rows Length:", len(rows))

		if len(rows) == 0 {
			cnf.Lg.Infof("response returned 0 rows, exiting")
			return nil
		}
		cnf.Lg.Infof("response returned %d fields & %d rows", len(cols), len(rows))

		// prepare the sql statement & chunks of values
		ins := MakeInsert(
			pp.tbls[i].Name,
			pp.tbls[i].PrimKey,
			cols,
			rows,
		) // attempt to insert rows from response
		ins.InsertFast(cnf)

		cnf.Lg.Infof("current %s players ETL complete", lg)
	}
	cnf.Lg.Infof("current players ETL complete for all leagues")
	return nil
}
