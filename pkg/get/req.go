package get

import (
	"fmt"
	"io"
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

// each nba endpoint/pg table pair should implement
type Fetcher interface {
	Fetch() error
}

type ReqestMeta struct {
	Host   string
	Endpt  string
	URL    string
	Hdrs   map[string]string
	Params map[string]string
}

func NewRequest(host, endpoint string, hdrs, params map[string]string) *ReqestMeta {
	return &ReqestMeta{
		Host:   host,
		Endpt:  endpoint,
		Hdrs:   hdrs,
		Params: params,
	}
}

func (rm *ReqestMeta) Fetch() error {

	return nil
}

func (rm *ReqestMeta) AddHeaders(r *http.Request) {
	for k, v := range rm.Hdrs {
		r.Header.Add(k, v)
	}
}

func (rm *ReqestMeta) MakeQueryStr(order []string) string {
	var url string = fmt.Sprintf("https://%s%s?", rm.Host, rm.Endpt)

	for _, param := range order {
		url = url + fmt.Sprintf("%s=%s&", param, rm.Params[param])
	}

	return url[0 : len(url)-1]
}

func (rm *ReqestMeta) MakeUQueryStr() string {
	var url string = fmt.Sprintf("https://%s%s?", rm.Host, rm.Endpt)

	for k, v := range rm.Params {
		url = url + fmt.Sprintf("%s=%s&", k, v)
	}
	return url[0 : len(url)-1]
}

// make new request with url returned from MakeFullURL
// add gr.Headers to req with addHdrs
// use RespFromClient to do the http req, return the resp body []byte
func (gr *GetReq) BodyFromReq() ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, gr.MakeFulLURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("error calling %s: %v", gr.MakeFulLURL(), err)

	}
	gr.addHdrs(req)
	body, err := RespFromClient(req)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// use http client to perform http request
// get & return body as []byte
func RespFromClient(req *http.Request) ([]byte, error) {
	var errMsg string
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if res != nil {
			if res.StatusCode == 429 {
				errMsg = fmt.Sprint(res.StatusCode, "- timeout error")
			} else {
				errMsg = fmt.Sprint(res.StatusCode, "- HTTP client error occured")
			}
		}
		errMsg = "*500 - HTTP client error occured, no response received"
		return nil, fmt.Errorf("%s: %v", errMsg, err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%d error reading response body: %v", res.StatusCode, err)
	}
	return body, nil

}

// endptURL to concat endpoint to base url
// makeQryStr to loop through gr.Params & make query string
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

// pass a defined GetReq struct, unmarshals body & returns as Resp struct
func RequestResp(gr GetReq) (*Resp, error) {
	var resp Resp
	body, err := gr.BodyFromReq()
	if err != nil {
		return nil, err
	}
	resp, err = UnmarshalInto(body)
	if err != nil {
		return &resp, fmt.Errorf("error unmarshaling: %e", err)
	}
	return &resp, nil
}
