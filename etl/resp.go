package etl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jdetok/golib/errd"
	"github.com/jdetok/golib/logd"
)

type Resp struct {
	Resource   string      `json:"resource"`
	Parameters any         `json:"parameters"`
	ResultSets []ResultSet `json:"resultSets"`
}

// main json object in response body after endpoint/params
type ResultSet struct {
	Name    string   `json:"name"`
	Headers []string `json:"headers"`
	RowSet  [][]any  `json:"rowSet"`
}

// pass a defined GetReq struct, unmarshals body & returns as Resp struct
func RequestResp(l logd.Logger, gr GetReq) (Resp, error) {
	e := errd.InitErr()
	var resp Resp
	body, err := gr.BodyFromReq(l)
	if err != nil {
		e.Msg = fmt.Sprintf("error getting response for %s", gr.Endpoint)
		l.WriteLog(e.Msg)
		return resp, e.BuildErr(err)
	}
	resp, err = UnmarshalInto(body)
	if err != nil {
		return resp, fmt.Errorf("error unmarshaling: %e", err)
	}
	return resp, nil
}

/*
use http client to perform http request
get & return body as []byte
*/
func RespFromClient(l logd.Logger, req *http.Request) ([]byte, error) {
	e := errd.InitErr()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if res != nil {
			if res.StatusCode == 429 {
				e.Msg = fmt.Sprint(res.StatusCode, "- timeout error")
			} else {
				e.Msg = fmt.Sprint(res.StatusCode, "- HTTP client error occured")
			}
			l.WriteLog(e.Msg)
			return nil, e.BuildErr(err)
		}
		e.Msg = "*500 - HTTP client error occured, no response received"
		return nil, e.NewErr()
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		e.Msg = fmt.Sprint(res.StatusCode, "- error reading response body")
		l.WriteLog(e.Msg)
		return nil, e.BuildErr(err)
	}
	return body, nil
}

/*
pass resp returned from RequestResp
placeholder print `header - val` to console
*/
func ProcessResp(resp Resp) {
	// fmt.Println(resp.ResultSets[0].RowSet[0]...)
	for _, r := range resp.ResultSets[0].RowSet {
		for i, x := range r {
			fmt.Printf("%v: %v\n", resp.ResultSets[0].Headers[i], x)
		}
		fmt.Println("*******")
	}
}

// unmarshal []byte body into Resp struct
func UnmarshalInto(body []byte) (Resp, error) {
	var resp Resp
	if err := json.Unmarshal(body, &resp); err != nil {
		return resp, fmt.Errorf("error unmarshaling: %e", err)
	}
	return resp, nil
}
