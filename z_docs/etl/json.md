# json in the format below is marshaled into the Resp type: 
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
# json example
```json
{
    "resource": "endpoint example",
    "parameters": {
        "param": "parameter example"
    },
    "resultSets": [
        {
            "name": "endpoint example",
            "headers": [
                "col1",
                "col2",
                "col3"
            ],
            "rowSet": [
                [
                    "val1",
                    "val2",
                    "val3"
                ],
                [
                    "val1",
                    "val2",
                    "val3"
                ]
            ]
        }
    ]
}
```
