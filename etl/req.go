package etl

import (
	"fmt"
	"net/http"

	"github.com/jdetok/golib/errd"
	"github.com/jdetok/golib/logd"
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

/*
make new request with url returned from MakeFullURL
add gr.Headers to req with addHdrs
use RespFromClient to do the http req, return the resp body []byte
*/
func (gr *GetReq) BodyFromReq(l logd.Logger) ([]byte, error) {
	e := errd.InitErr()
	req, err := http.NewRequest(http.MethodGet, gr.MakeFulLURL(), nil)
	if err != nil {
		e.Msg = fmt.Sprintf("error calling %s", gr.MakeFulLURL())
		l.WriteLog(e.Msg)
		return nil, e.BuildErr(err)

	}
	gr.addHdrs(req)
	body, err := RespFromClient(l, req)
	if err != nil {
		return nil, err
	}
	return body, nil
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
