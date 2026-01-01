package etl

import (
	"fmt"
	"time"

	"github.com/jdetok/bball-etl-cli/pkg/cnf"
)

// Conf struct, only have to pass this to access logger, db, row count, etc

func RunNightlyETL(c *cnf.Conf, df, dt string) error {
	if err := CrntPlayersETL(c); err != nil {
		return fmt.Errorf("error with current players ETL: %v", err)
	}

	if err := GLogDailyETL(c, df, dt); err != nil {
		return fmt.Errorf("error with nightly game log ETL: %v", err)
	}

	c.Lg.Infof("\nfinished with nightly ETL | total rows affected: %d", c.RowCnt)
	return nil
}

func RunSeasonETL(c *cnf.Conf, startY, endY string) error {
	szns, err := SznBSlice(startY, endY)
	if err != nil {
		return fmt.Errorf("error making seasons string: %v", err)
	}

	for _, s := range szns {
		sra := c.RowCnt // capture row count at start of each season
		stT := time.Now()

		// players etl for season
		if err := SznPlayersETL(c, "1", s); err != nil {
			return fmt.Errorf("error getting players for %s: %v", s, err)
		}

		// get team and player game logcs for the season
		err = GLogSeasonETL(c, s)
		if err != nil {
			return fmt.Errorf("error inserting data for %s: %v", s, err)
		} // log finished with season etl
		c.Lg.Infof(`finished with %s season ETL after %v
++ total rows before: %d | total rows after: %d,
++ rows affected from %s fetch: %d
++ total rows affected: %d
`, s, time.Since(stT), sra, c.RowCnt, s, c.RowCnt-sra, c.RowCnt)
	} // log finished with ETL
	c.Lg.Infof("finished %d seasons between %s and %s | total rows affected: %d",
		len(szns), startY, endY, c.RowCnt)
	return nil
}
