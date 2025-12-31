package etl

import (
	"testing"
	"time"

	"github.com/jdetok/bball-etl-cli/pkg/cnf"
)

func TestGamelogsByDate(t *testing.T) {
	c := &cnf.Conf{}
	if err := DailyGamelogs(c, Yesterday(time.Now())); err != nil {
		t.Fatal(err)
	}
}
