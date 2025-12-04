package etl

import (
	"fmt"
	"sync"

	"github.com/jdetok/golib/errd"
)

type ConcurETL struct {
	Cnf *Conf
	// send to tmout if HTTP request returns 429
	// to make this work, before launching a goroutine that does a http request,
	// check if the timeout chan is empty
	tmout chan struct{}
	dberr chan string
	hterr chan string
	gnerr chan string
}

// glim is amount of gorountines
func MakeConcurETL(cnf *Conf, glim int) *ConcurETL {
	cc := ConcurETL{
		Cnf:   cnf,
		tmout: make(chan struct{}),
		dberr: make(chan string),
		hterr: make(chan string),
		gnerr: make(chan string),
	}
	return &cc
}

// safe concurrent build mode
func (cc *ConcurETL) CCBuildModeETL(st, en string) (string, error) {

	return fmt.Sprintf("\n---- db build etl complete | total rows affected: %d",
		cc.Cnf.RowCnt), nil
}

// CONCURRENT ETL
func CCSeasonsETL(cnf *Conf, st, en string) (string, error) {
	var rwg sync.WaitGroup // read waitgroup
	var pwg sync.WaitGroup // produce (write) waitgroup

	// return formatted slice of seasons from startY, endY | "2025-26" format
	szns, err := SznBSlice(cnf.L, st, en)
	if err != nil {
		return fmt.Sprintf(
			"error making slice of season string from %s to %s", st, en), nil
	}

	pchan := make(chan string, len(szns))
	rwg.Add(1)
	go func(ch <-chan string) {
		defer rwg.Done()
		for s := range ch {
			fmt.Println(s)
		}
	}(pchan)

	// create a chan
	for _, s := range szns {
		pwg.Add(1)
		go func(cnf *Conf, s string) {
			defer pwg.Done()
			e := errd.InitErr()
			lt := GLogParams()

			err := GetManyGLogs(cnf, lt.lgs, lt.tbls, s)
			if err != nil {
				e.Msg = fmt.Sprintf("error running ETL for %s", s)
				pchan <- fmt.Errorf("%s\n%v", e.Msg, err).Error()
			}
			pchan <- fmt.Sprintf("successfully got game logs for %s", s)
		}(cnf, s)
	}

	pwg.Wait()
	close(pchan)

	rwg.Wait()
	return "completed succesfully", nil
}
