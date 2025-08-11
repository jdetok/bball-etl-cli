# general get request architecture 07/24/2025
# most current 7/24 9pm:
this takes a premade GetReq var, requests the json, marshals it to Resp type
```go
func main() {
	resp, err := RequestResp(leagueGameLog)
	if err != nil {
		log.Fatalf("error getting response: %e", err)
	}
	ProcessResp(resp)
}
```
## GetReq type
```go
type GetReq struct {
	Host     string
	Endpoint string
	Params   []Pair
	Headers  []Pair
}
// EXAMPLE
var commonPlayerInfo = GetReq{
	Host:     HOST,
    Headers:  HDRS,
	Endpoint: "/stats/commonplayerinfo",
	Params:   []Pair{{"LeagueID", "10"}, {"PlayerID", "2544"}},
}
```
## Resp type
```go
type Resp struct {
	Resource   string      `json:"resource"`
	Parameters any         `json:"parameters"`
	ResultSets []ResultSet `json:"resultSets"`
}

type ResultSet struct {
	Name    string   `json:"name"`
	Headers []string `json:"headers"`
	RowSet  [][]any  `json:"rowSet"`
}

```

## entrypoint: Get function
the Get function accepts a host, end[point], params, & headers 
(both slices of key-val pairs) and returns a response body, HTTP status code, &
error  
example that prints the response body: 
```go
body, _, err := commonPlayerInfo.GetRespBody()
	if err != nil {
		log.Fatal(err)
	}
fmt.Println(string(body))
```
### global host & headers
```go
const HOST string = "stats.nba.com"
var HDRS = []Pair{
	{"Accept", "application/json"},
	{"Connection", "keep-alive"},
	{"Referer", "https://www.nba.com"},
	{"Origin", "https://www.nba.com"},
	{"User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36"},
}
```
## Pair type
the Pair struct is used both for headers & URL params
```go
type Pair struct {
	Key string
	Val string
}
```
- both addHdrs and addParams accept a slice of Pair
    - addHdrs also acccepts a `http.Request`, loops through the slice & adds a 
    header to the request for each Pair
    - addParams also accepts the base url, loops through & adds the key value 
    parameter pairs to the url string
