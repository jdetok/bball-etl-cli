package etl

import (
	"fmt"

	"github.com/jdetok/bball-etl-cli/pkg/cnf"
	"github.com/jdetok/bball-etl-cli/pkg/get"
	"github.com/jdetok/bball-etl-cli/pkg/pgdb"
)

func PlayerReq(onlyCurrent, league, season string) get.GetReq {
	var gr = get.GetReq{
		Host:     get.HOST,
		Headers:  get.HDRS,
		Endpoint: "/stats/commonallplayers",
		Params: []get.Pair{
			{"IsOnlyCurrentSeason", onlyCurrent},
			{"LeagueID", league},
			{"Season", season},
		},
	}
	return gr
}

func PlayersParams() pgdb.LgTbls {
	var lt pgdb.LgTbls
	lt.Lgs = []string{"00", "10"}
	lt.Tbls = []pgdb.Table{
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
func SznPlayersETL(c *cnf.Conf, onlyCurrent, season string) error {
	pp := PlayersParams()
	c.Lg.Infof("attempting players ETL for %s nba/wnba seasons", season)

	for i := range pp.Lgs {
		var lg string
		switch pp.Lgs[i] {
		case "00":
			lg = "nba"
		case "10":
			lg = "wnba"
		}

		c.Lg.Infof("attempting to insert %s %s players", season, lg)
		// r := PlayerReq(onlyCurrent, p[0], p[1])
		r := PlayerReq(onlyCurrent, pp.Lgs[i], season)
		resp, err := get.RequestResp(r)
		if err != nil {
			return fmt.Errorf("error getting response for %s: lg: %s szn: %s: %v", r.Endpoint, lg, season, err)
		}

		// get cols/rows from resp, return early when no rows in response
		var cols []string = resp.ResultSets[0].Headers
		var rows [][]any = resp.ResultSets[0].RowSet
		// ProcessResp(resp)
		fmt.Println("Cols Length:", len(cols), "Rows Length:", len(rows))

		if len(rows) == 0 {
			c.Lg.Infof("response returned 0 rows, exiting")
			return nil
		}
		c.Lg.Infof("response returned %d fields & %d rows", len(cols), len(rows))

		// prepare the sql statement & chunks of values
		ins := pgdb.MakeInsert(
			pp.Tbls[i].Name,
			pp.Tbls[i].PrimKey,
			cols,
			rows,
		) // attempt to insert rows from response
		ins.InsertFast(c)

		c.Lg.Infof("%s %s players ETL complete", season, lg)
	}
	c.Lg.Infof("players ETL complete for %s", season)
	return nil
}

func CrntPlayersETL(c *cnf.Conf) error {
	sl := GetSeasons()
	var szns = []string{sl.Szn, sl.WSzn}
	pp := PlayersParams()

	c.Lg.Infof("attempting current players ETL for %s nba season and %s wnba season", sl.Szn, sl.WSzn)

	for i := range pp.Lgs {
		var lg string
		switch pp.Lgs[i] {
		case "00":
			lg = "nba"
		case "10":
			lg = "wnba"
		}

		c.Lg.Infof("attempting to insert current %s players", lg)
		// r := PlayerReq(onlyCurrent, p[0], p[1])
		r := PlayerReq("1", pp.Lgs[i], szns[i])
		resp, err := get.RequestResp(r)
		if err != nil {
			return fmt.Errorf("error getting response for %s: %v", r.Endpoint, err)
		}

		// get cols/rows from resp, return early when no rows in response
		var cols []string = resp.ResultSets[0].Headers
		var rows [][]any = resp.ResultSets[0].RowSet
		// ProcessResp(resp)
		fmt.Println("Cols Length:", len(cols), "Rows Length:", len(rows))

		if len(rows) == 0 {
			c.Lg.Infof("response returned 0 rows, exiting")
			return nil
		}
		c.Lg.Infof("response returned %d fields & %d rows", len(cols), len(rows))

		// prepare the sql statement & chunks of values
		ins := pgdb.MakeInsert(
			pp.Tbls[i].Name,
			pp.Tbls[i].PrimKey,
			cols,
			rows,
		) // attempt to insert rows from response
		ins.InsertFast(c)

		c.Lg.Infof("current %s players ETL complete", lg)
	}
	c.Lg.Infof("current players ETL complete for all leagues")
	return nil
}
