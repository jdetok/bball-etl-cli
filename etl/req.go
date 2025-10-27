package etl

import (
	"net/http"
)

// build GetReq types to request data from new endpoints
type GetReq struct {
	Host     string
	Endpoint string
	Params   []Pair
	Headers  []Pair
}

// basic key value type
type Pair struct {
	Key string
	Val string
}

/*
endptURL to concat endpoint to base url
makeQryStr to loop through gr.Params & make query string
*/
func (gr *GetReq) MakeFulLURL() string {
	bUrl := gr.endptURL()
	return gr.makeQryStr(bUrl)
}

// concat endpoint to host
func (gr *GetReq) endptURL() string {
	return "https://" + gr.Host + gr.Endpoint
}

// makes the query string from gr.Params
func (gr *GetReq) makeQryStr(bUrl string) string {
	var url string = bUrl + "?"
	for i, p := range gr.Params {
		url = url + (p.Key + "=" + p.Val)
		if i < len(gr.Params)-1 {
			url += "&"
		}
	}
	return url
}

// loop through gr.Headers & add each as a header to the request
func (gr *GetReq) addHdrs(r *http.Request) {
	for _, h := range gr.Headers {
		r.Header.Add(h.Key, h.Val)
	}
}
