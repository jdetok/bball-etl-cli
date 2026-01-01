package etl

import (
	"fmt"
	"testing"
)

func TestNewLgSeasonsMap(t *testing.T) {
	ls, err := NewLgSeasonsMap("09/09/2022", "11/02/2026")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(ls)
}
