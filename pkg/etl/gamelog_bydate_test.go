package etl

import (
	"testing"
)

func TestNewLgSeasonsMap(t *testing.T) {
	_, err := NewLgSeasonsMap("09/09/2022", "01/02/2026")
	if err != nil {
		t.Fatal(err)
	}
}
