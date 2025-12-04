package etl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

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

// send http get request, return bytes from response body
func RespFromClient(l logd.Logger, req *http.Request) ([]byte, error) {
	e := errd.InitErr()

	const retries = 3
	baseDelay := (5 * time.Second)

	for attempt := 0; attempt <= retries; attempt++ {
		curDel := baseDelay << attempt
		clnReq := req.Clone(req.Context())
		res, err := http.DefaultClient.Do(clnReq)
		if err != nil {
			// RETRY REQUEST IF < RETRIES
			if attempt < retries {
				e.Msg = fmt.Sprintf(
					"HTTP client error (attempt %d/%d): %v - retrying after %v seconds",
					attempt+1, retries+1, err, curDel)
				l.WriteLog(e.Msg)
				time.Sleep(curDel) // exponential backoff
				continue
			}
			e.Msg = "*500 - HTTP client error occurred, no response received"
			l.WriteLog(fmt.Sprintf("%s: %v", e.Msg, err))
			return nil, e.BuildErr(err)
		}

		// capture status code
		status := res.StatusCode

		// SUCCESS CODE RECEIVED - RETURN BODY
		if status >= 200 && status < 300 {
			body, err := func() ([]byte, error) {
				// defer response body close
				defer res.Body.Close()
				return io.ReadAll(res.Body)
			}()
			if err != nil {
				e.Msg = fmt.Sprint(status, "- error reading response body")
				l.WriteLog(e.Msg)
				return nil, e.BuildErr(err)
			}
			return body, nil
		}

		// retryable statuses: 429 (rate limit) and 5xx (server errors)
		if status == 429 || (status >= 500 && status < 600) {
			_ = res.Body.Close() // ensure body is closed before retrying
			if attempt < retries {
				e.Msg = fmt.Sprintf("%d - retryable response (attempt %d/%d)",
					status, attempt+1, retries+1)
				l.WriteLog(e.Msg)
				time.Sleep(baseDelay << attempt)
				continue
			}
			// final attempt failed
			e.Msg = fmt.Sprintf("%d - final retry failed", status)
			l.WriteLog(e.Msg)
			return nil, e.BuildErr(fmt.Errorf("http status %d", status))
		}

		// non-retryable client error (4xx, except 429)
		_ = res.Body.Close() // ensure body is closed before returning
		e.Msg = fmt.Sprint(status, "- non-retryable HTTP error occurred")
		l.WriteLog(e.Msg)
		return nil, e.BuildErr(fmt.Errorf("http status %d", status))
	}

	// should not reach here
	e.Msg = "unexpected error in RespFromClient retry loop"
	l.WriteLog(e.Msg)
	return nil, e.NewErr()
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
