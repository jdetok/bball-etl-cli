package etl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jdetok/bball-etl-cli/pkg/cnf"
	"github.com/jdetok/bball-etl-cli/pkg/get"
	"github.com/jdetok/bball-etl-cli/pkg/pgdb"
)

type ParamsToReq struct {
	lgs      []string
	sTypes   []string
	lgSznMap map[string]string
}

func Tst() error {
	sl, err := GetSeasonsFromDate(time.Now())
	if err != nil {
		return err
	}
	// sl := GetSeasons()
	fmt.Println(sl.Szn)
	fmt.Println(sl.WSzn)
	return nil
}

func GamelogParamsToReq(day string) (*ParamsToReq, error) {
	lgs := []string{}
	sTypes := []string{}
	// get day as time

	d, err := time.Parse("01/02/2006", day)
	if err != nil {
		return nil, err
	}

	lgSched, err := GetLgSchedules(d)
	if err != nil {
		return nil, fmt.Errorf("error getting schedules: %v", err)
	}

	fmt.Println("schedules for", d)

	tmpLgs := []string{"00", "10"}
	for _, lg := range tmpLgs {
		if lgSched.LgMap[lg] != nil {
			if gmIsPlayoff, ok := lgSched.LgMap[lg][day]; ok {
				lgs = append(lgs, lg)
				var sType string
				if gmIsPlayoff {
					sType = "Playoffs"
				} else {
					sType = "Regular+Season"
				}
				sTypes = append(sTypes, sType)
			}
		}
	}
	return &ParamsToReq{lgs, sTypes, lgSched.LgSzn}, nil
}

// run nightly (soon after all games conclude, usually around 00:30)
// references league schedules endpoint to only run etl for season/lg active for date
func DailyGamelogs(c *cnf.Conf, date string) error {
	// yday := Yesterday(time.Now())

	p, err := GamelogParamsToReq(date)
	if err != nil {
		return fmt.Errorf("error getting request parameters: %v", err)
	}

	c.Lg.Infof("fetching for %s: league(s): %v | season(s): %v, season type(s): %v",
		date, p.lgs, p.lgSznMap, p.sTypes)

	dbTargets := GamelogDBTargets()
	for pt, tbl := range dbTargets {
		for _, lg := range p.lgs {
			for _, st := range p.sTypes {

				resp, err := GamelogsByDate(c, lg, p.lgSznMap[lg], st, pt, date, date)
				if err != nil {
					return fmt.Errorf("error getting gamelog\n| %s | %s | %s | %s | %s |\n%v",
						lg, p.lgSznMap[lg], st, pt, date, err)
				}
				var cols []string = resp.ResultSets[0].Headers
				var rows [][]any = resp.ResultSets[0].RowSet
				if len(rows) == 0 {
					c.Lg.Infof("response returned 0 rows, exiting")
					return nil
				}
				c.Lg.Infof("response returned %d fields & %d rows", len(cols), len(rows))
				ins := pgdb.MakeInsert(
					tbl.Name,
					tbl.PrimKey,
					cols,
					rows,
				)
				if err := ins.Insert(c); err != nil {
					return fmt.Errorf("error attempting to insert %d rows into %s: %v",
						len(rows), tbl.Name, err)
				}
			}
		}
	}
	c.Lg.Infof("DailyGamelogs for %v complete", date)
	return nil
}

func GamelogsByDate(c *cnf.Conf, lg, szn, sType, pt, df, dt string) (*RespGamelogs, error) {
	rm := NewReqMeta(lg, szn, sType, pt, df, dt)

	c.Lg.Infof("sending get request to:\n+ %s", rm.URL)

	req, err := http.NewRequest(http.MethodGet, rm.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("error calling %s: %v", rm.URL, err)
	}
	rm.AddHeaders(req)
	body, err := get.RespFromClient(req)
	if err != nil {
		return nil, fmt.Errorf("error getting response: %v", err)
	}

	c.Lg.Infof("%d bytes received from %s", len(body), rm.Endpt)

	resp := &RespGamelogs{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("error unmarshaling json body: %v", err)
	}
	return resp, nil
}

func NewReqMeta(lg, szn, sType, plTm, df, dt string) *get.ReqestMeta {
	order := []string{"LeagueID", "Season", "SeasonType", "Counter", "Sorter", "Direction",
		"PlayerOrTeam", "DateFrom", "DateTo"}

	rm := &get.ReqestMeta{
		Host:  get.HOST,
		Hdrs:  get.HDRMAP,
		Endpt: "/stats/leaguegamelog",
		Params: map[string]string{
			"LeagueID":     lg,
			"Season":       szn,
			"SeasonType":   sType,
			"Counter":      "0",
			"Sorter":       "DATE",
			"Direction":    "DESC",
			"PlayerOrTeam": plTm,
			"DateFrom":     df,
			"DateTo":       dt,
		},
	}
	// rm.URL = fmt.Sprintf("https://%s%s")
	rm.URL = rm.MakeQueryStr(order)
	return rm
}
