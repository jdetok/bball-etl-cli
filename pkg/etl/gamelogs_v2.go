package etl

import "github.com/jdetok/bball-etl-cli/pkg/pgdb"

type ResultSet struct {
	Name    string   `json:"name"`
	Headers []string `json:"headers"`
	RowSet  [][]any  `json:"rowSet"`
}

type RespGamelogs struct {
	Resource   string      `json:"resource"`
	Parameters any         `json:"parameters"`
	ResultSets []ResultSet `json:"resultSets"`
}

type DBTargets map[string]pgdb.Table

func GamelogParams() *pgdb.LgTbls {
	return &pgdb.LgTbls{
		Lgs: []string{"00", "10"},
		Tbls: []pgdb.Table{
			{
				Name:    "intake.gm_team",
				PrimKey: "game_id, team_id",
				PlTm:    "T",
			},
			{
				Name:    "intake.gm_player",
				PrimKey: "game_id, player_id",
				PlTm:    "P",
			},
		},
		Into: map[string]pgdb.Table{ // can iterate through this
			"T": {
				Name:    "intake.gm_team",
				PrimKey: "game_id, team_id",
				PlTm:    "T",
			},
			"P": {
				Name:    "intake.gm_player",
				PrimKey: "game_id, player_id",
				PlTm:    "P",
			},
		},
	}
}

func GamelogDBTargets() DBTargets {
	return DBTargets{
		"T": {
			Name:    "intake.gm_team",
			PrimKey: "game_id, team_id",
		},
		"P": {
			Name:    "intake.gm_player",
			PrimKey: "game_id, player_id",
		},
	}
}
