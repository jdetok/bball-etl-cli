package etl

const HOST string = "stats.nba.com"

var HDRS = []Pair{
	{"Accept", "application/json"},
	{"Connection", "keep-alive"},
	{"Referer", "https://www.nba.com"},
	{"Origin", "https://www.nba.com"},
	{"User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36"},
}
