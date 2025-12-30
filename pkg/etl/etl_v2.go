package etl

import (
	"fmt"
	"net/http"

	"github.com/jdetok/bball-etl-cli/pkg/cnf"
	"github.com/jdetok/bball-etl-cli/pkg/get"
)

func GamelogsByDate(c *cnf.Conf, df, dt string) error {
	dbTarget := GamelogParams()
	fmt.Println(dbTarget.Into)

	sch, err := GetCurrentLgSchedules()
	if err != nil {
		return fmt.Errorf("error getting schedules: %v", err)
	}

	fmt.Println("dates in nba season:", len(sch.LgMap["00"]))
	fmt.Println("dates in wnba season:", len(sch.LgMap["10"]))

	rm := NewGamelogsReq("00", "2025-26", "Regular+Season", "P", "10/25/2025", "11/10/2025")
	fmt.Println(rm.URL)

	req, err := http.NewRequest(http.MethodGet, rm.URL, nil)
	if err != nil {
		return fmt.Errorf("error calling %s: %v", rm.URL, err)
	}
	rm.AddHeaders(req)
	body, err := get.RespFromClient(req)
	if err != nil {
		return fmt.Errorf("error getting response: %v", err)
	}
	fmt.Println(string(body))
	return nil
}

func NewGamelogsReq(lg, szn, sType, plTm, df, dt string) *get.ReqestMeta {
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
	rm.URL = rm.MakeQueryStr()
	return rm
}
