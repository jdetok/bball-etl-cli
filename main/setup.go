package main

import (
	"io"

	"github.com/jdetok/bball-etl-cli/pkg/conn"
	"github.com/jdetok/bball-etl-cli/pkg/logd"
)

func getLogWriter(logf string) (io.Writer, error) {
	var tmpLg io.Writer
	var err error
	switch logf {
	case "", "cli":
		tmpLg, err = logd.GetLogWriter(true, logf, "")
	default:
		tmpLg, err = logd.GetLogWriter(false, logf, "010206_150405")
	}
	if err != nil {
		return nil, err
	}
	return tmpLg, nil
}

func loadDBEnv(envf string) (*conn.DBEnv, error) {
	e := &conn.DBEnv{}
	var err error
	switch envf {
	case "", "skip": // don't load a .env file (env vars already exist)
		e, err = conn.Load(PG_HOSTN, PG_PORTN, PG_USERN, PG_PASSN, PG_DATAN)
	default:
		e, err = conn.LoadFromDotEnv(envf, PG_HOSTN, PG_PORTN, PG_USERN, PG_PASSN, PG_DATAN)
	}
	if err != nil {
		return nil, err
	}
	return e, nil
}
