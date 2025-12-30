package etl

import (
	"testing"

	"github.com/jdetok/bball-etl-cli/pkg/cnf"
)

func TestGamelogsByDate(t *testing.T) {
	c := &cnf.Conf{}
	df := "10/25/2025"
	dt := "11/10/2025"
	if err := GamelogsByDate(c, df, dt); err != nil {
		t.Fatal(err)
	}
}
