package get

const HOST string = "stats.nba.com"

var HDRS = []Pair{
	{"Accept", "application/json"},
	{"Connection", "keep-alive"},
	{"Referer", "https://www.nba.com"},
	{"Origin", "https://www.nba.com"},
	{"User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36"},
}

var HDRMAP = map[string]string{
	"Accept":     "application/json",
	"Connection": "keep-alive",
	"Referer":    "https://www.nba.com",
	"Origin":     "https://www.nba.com",
	"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36",
}

/*

https://stats.nba.com/stats/leaguegamelog?Counter=0&Sorter=DATE&PlayerOrTeam=P&Direction=DESC&LeagueID=00&Season=2025-26&SeasonType=Regular+Season&DateFrom=10/25/2025&DateTo=11/10/2025
https://stats.nba.com/stats/leaguegamelog?LeagueID=00&Season=&SeasonType=Regular+Season&Counter=0&Sorter=DATE&Direction=DESC&PlayerOrTeam=T&DateFrom=12/29/2025&DateTo=12/29/2025
*/
