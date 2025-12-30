package get

import (
	"encoding/json"
	"fmt"
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
