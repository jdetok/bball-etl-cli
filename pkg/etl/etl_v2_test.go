package etl

import (
	"testing"

	"github.com/jdetok/bball-etl-cli/pkg/cnf"
)

func TestGamelogsByDate(t *testing.T) {
	c := &cnf.Conf{}
	if err := DailyGamelogs(c); err != nil {
		t.Fatal(err)
	}
}
