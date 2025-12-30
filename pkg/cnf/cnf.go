package cnf

import (
	"database/sql"

	"github.com/jdetok/bball-etl-cli/pkg/logd"
)

type Conf struct {
	Lg     *logd.Logd
	DB     *sql.DB
	RowCnt int64 // row counter
	Errs   []string
}
