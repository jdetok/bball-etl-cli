package etl

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jdetok/golib/errd"
	"github.com/jdetok/golib/logd"
)

// Conf struct, only have to pass this to access logger, db, row count, etc
type Conf struct {
	L      logd.Logger
	DB     *sql.DB
	RowCnt int64 // row counter
	Errs   []string
}

func RunNightlyETL(cnf Conf) error {
	e := errd.InitErr()

	if err := CrntPlayersETL(cnf); err != nil {
		e.Msg = "error with current players ETL"
		cnf.L.WriteLog(e.Msg)
		return e.BuildErr(err)
	}

	if err := GLogDailyETL(&cnf); err != nil {
		e.Msg = "error with nightly game log ETL"
		cnf.L.WriteLog(e.Msg)
		return e.BuildErr(err)
	}

	cnf.L.WriteLog(fmt.Sprintf(
		"\n====  finished with nightly ETL | total rows affected: %d", cnf.RowCnt))
	return nil
}

// HIGH LEVEL SEASON ETL CALL - fetch & insert ALL data for each season from start - end
func RunSeasonETL(cnf Conf, startY, endY string) error {
	e := errd.InitErr()

	// return formatted slice of seasons from startY, endY | "2025-26" format
	szns, err := SznBSlice(cnf.L, startY, endY)
	if err != nil {
		e.Msg = "error making seasons string"
		cnf.L.WriteLog(e.Msg)
		log.Fatal(e.BuildErr(err))
	}

	// iterate through each season
	// convert to concurrent
	for _, s := range szns {
		sra := cnf.RowCnt // capture row count at start of each season
		stT := time.Now()

		// players etl for season
		if err := SznPlayersETL(cnf, "1", s); err != nil {
			e.Msg = fmt.Sprint("error getting players for ", s)
			return e.NewErr()
		}

		// get team and player game logs for the season
		err = GLogSeasonETL(&cnf, s)
		if err != nil {
			e.Msg = fmt.Sprint("error inserting data for ", s)
			cnf.L.WriteLog(e.Msg)
			fmt.Println(e.BuildErr(err))
		} // log finished with season etl
		cnf.L.WriteLog(fmt.Sprint(
			fmt.Sprintf("====  finished with %s season ETL after %v",
				s, time.Since(stT)),
			fmt.Sprintf(
				"\n== total rows before: %d | total rows after: %d",
				sra, cnf.RowCnt),
			fmt.Sprintf(
				"\n== rows affected from %s fetch: %d", s, cnf.RowCnt-sra),
			fmt.Sprintf(
				"\n== total rows affected: %d", cnf.RowCnt)))
	} // log finished with ETL
	cnf.L.WriteLog(fmt.Sprintf(
		"\n====  finished %d seasons between %s and %s | total rows affected: %d",
		len(szns), startY, endY, cnf.RowCnt,
	))

	return nil
}
