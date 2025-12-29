package etl

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jdetok/bball-etl-cli/pkg/logd"
)

// Conf struct, only have to pass this to access logger, db, row count, etc
type Conf struct {
	Lg     *logd.Logd
	DB     *sql.DB
	RowCnt int64 // row counter
	Errs   []string
}

func RunNightlyETL(cnf *Conf) error {
	if err := CrntPlayersETL(cnf); err != nil {
		return fmt.Errorf("error with current players ETL: %v", err)
	}

	if err := GLogDailyETL(cnf); err != nil {
		return fmt.Errorf("error with nightly game log ETL: %v", err)
	}

	cnf.Lg.Infof("\nfinished with nightly ETL | total rows affected: %d", cnf.RowCnt)
	return nil
}

func RunSeasonETL(cnf *Conf, startY, endY string) error {
	szns, err := SznBSlice(startY, endY)
	if err != nil {
		return fmt.Errorf("error making seasons string")
	}

	for _, s := range szns {
		sra := cnf.RowCnt // capture row count at start of each season
		stT := time.Now()

		// players etl for season
		if err := SznPlayersETL(cnf, "1", s); err != nil {
			return fmt.Errorf("error getting players for ", s)
		}

		// get team and player game logs for the season
		err = GLogSeasonETL(cnf, s)
		if err != nil {
			return fmt.Errorf("error inserting data for %s: %v", s, err)
		} // log finished with season etl
		cnf.Lg.Infof(`finished with %s season ETL after %v
++ total rows before: %d | total rows after: %d,
++ rows affected from %s fetch: %d
++ total rows affected: %d
`, s, time.Since(stT), sra, cnf.RowCnt, s, cnf.RowCnt-sra, cnf.RowCnt)
	} // log finished with ETL
	cnf.Lg.Infof("finished %d seasons between %s and %s | total rows affected: %d",
		len(szns), startY, endY, cnf.RowCnt)
	return nil
}
