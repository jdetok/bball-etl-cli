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

// unmarshal []byte body into Resp struct
func UnmarshalInto(body []byte) (Resp, error) {
	var resp Resp
	if err := json.Unmarshal(body, &resp); err != nil {
		return resp, fmt.Errorf("error unmarshaling: %e", err)
	}
	return resp, nil
}
