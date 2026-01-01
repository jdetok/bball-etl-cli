package etl

import (
	"fmt"
	"testing"
	"time"
)

func TestGetSeasonsFromDate(t *testing.T) {
	// yday := Yesterday(time.Now())

	yday := time.Now().Add(-24 * time.Hour)

	tst24Date, _ := time.Parse("01/02/2006", "12/25/2024")

	tst24WDate, _ := time.Parse("01/02/2006", "08/18/2024")

	sl1, err := GetSeasonsFromDate(yday)
	if err != nil {
		t.Error(err)
	}

	sl2, err := GetSeasonsFromDate(tst24Date)
	if err != nil {
		t.Error(err)
	}

	sl3, err := GetSeasonsFromDate(tst24WDate)
	if err != nil {
		t.Error(err)
	}

	for _, sl := range []*SeasonLeague{sl1, sl2, sl3} {
		fmt.Printf("nba season: %s | wnba season: %s\n", sl.Szn, sl.WSzn)
	}

}
