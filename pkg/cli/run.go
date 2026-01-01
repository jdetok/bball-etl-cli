package cli

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jdetok/bball-etl-cli/pkg/cnf"
	"github.com/jdetok/bball-etl-cli/pkg/etl"
	"github.com/jdetok/bball-etl-cli/pkg/maild"
)

type RunMode struct {
	StartTime    time.Time
	Cnf          *cnf.Conf
	Args         *CLIArgs
	EmailUserVar string
	EmailPassVar string
	EmailHostVar string
	EmailPortVar string
}

type modeFn func(*RunMode) error

// map args to -mode to an etl function
func BuildRunModes(r *RunMode) map[string]modeFn {
	return map[string]modeFn{
		"":       func(*RunMode) error { return daily(r) },
		"daily":  func(*RunMode) error { return daily(r) },
		"dly":    func(*RunMode) error { return daily(r) },
		"d":      func(*RunMode) error { return daily(r) },
		"email":  func(*RunMode) error { return email(r) },
		"build":  func(*RunMode) error { return build(r) },
		"custom": func(*RunMode) error { return custom(r) },
		"tst":    func(*RunMode) error { return tst(r) },
	}
}

func daily(r *RunMode) error {
	return etl.DailyGamelogs(r.Cnf, etl.Yesterday(time.Now()))
}

func email(r *RunMode) error {
	return emailLogFile(r.Cnf, *r.Args.Atch.Dest, r.EmailUserVar, r.EmailPassVar, r.EmailHostVar, r.EmailPortVar)
}

func build(r *RunMode) error {
	return etl.RunSeasonETL(r.Cnf, *r.Args.StartYr.Dest, *r.Args.EndYr.Dest)
}

func tst(_ *RunMode) error {
	return etl.Tst()
}

func custom(r *RunMode) error {
	szn := *r.Args.Szn.Dest
	lg := *r.Args.Lg.Dest
	df := *r.Args.DateFrom.Dest
	dt := *r.Args.DateTo.Dest

	if szn == "" && df == "" && dt == "" {
		return fmt.Errorf("a season or date must be specified in custom mode")
	}

	if df != "" || dt != "" {
		updateDfDt(&df, &dt)
		return etl.CustomGamelogsByDate(r.Cnf, df, dt)
		// return etl.DailyGamelogs(r.Cnf, df)
		// return etl.
		// return etl.RunNightlyETL(r.Cnf, df, dt)
	}

	switch lg {
	case "":
		return etl.GLogSeasonETL(r.Cnf, szn)
	case "nba", "wnba":
		return etl.LgSznGlogs(r.Cnf, lg, szn)
	default:
		return fmt.Errorf("invalid league: %s", lg)
	}
}

// if one of datefrom/dateto is empty, fill it with the other
func updateDfDt(df, dt *string) {
	if *dt == "" && *df != "" {
		dt = df
	} else if *df == "" && *dt != "" {
		df = dt
	}
}

// send an email with atch attached, also loads environment variables for gmail auth
func emailLogFile(c *cnf.Conf, atch, userN, passN, hostN, portN string) error {
	if atch == "" {
		return fmt.Errorf("must pass an attachment in email mode")
	}
	if _, err := os.Stat(atch); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("fatal: file to attach not found at %s: %v", atch, err)
		}
		return fmt.Errorf("fatal: error occured finding os.Stat(%s): %v", atch, err)
	}
	if err := maild.EmailLog(atch, userN, passN, hostN, portN); err != nil {
		return fmt.Errorf("error emailing log: %v", err)
	}
	c.Lg.Infof("email with %s attached successfully sent", atch)
	return nil
}
