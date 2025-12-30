package etl

import "github.com/jdetok/bball-etl-cli/pkg/pgdb"

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
