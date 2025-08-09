package etl

import (
	"fmt"
	"sync"
	"time"

	"github.com/jdetok/golib/errd"
)

type InsertStmnt struct {
	Tbl     string
	PrimKey string // define like "key" or "key1, key2"
	Cols    []string
	Vals    []any
	Rows    [][]any
	Chunks  [][][]any
}

type Table struct {
	Name    string
	PrimKey string
	PlTm    string
}

type LgTbls struct {
	lgs  []string
	tbls []Table
}

func MakeInsert(tbl, primKey string, cols []string, rows [][]any) InsertStmnt {
	var ins = InsertStmnt{
		Tbl:     tbl,
		PrimKey: primKey,
		Cols:    cols,
		Rows:    rows,
	}
	ins.FlattenVals()
	ins.ChunkVals()
	return ins
}

// flatten [][]any to []any with all values
func (ins *InsertStmnt) FlattenVals() {
	for _, r := range ins.Rows {
		ins.Vals = append(ins.Vals, r...)
	}
}

// flatten & return the values of a chunk of rows
func ValsFromSet(set [][]any) []any {
	var valsFlat []any
	for _, r := range set {
		valsFlat = append(valsFlat, r...)
	}
	return valsFlat
}

/*
populates ins.Chunks [][][]any with chunks
postgres sql.Exec() only allows 65,535 individual values to be inserted at once
ChunkVals populates ins.Chunks ([][][]any) with as many chunks ([][]any) with
as many full []any as necessary to keep the total number of values under 65,535.
* have found that setting the max vals in a chunk at 20,000 makes the individual
**execs much quicker
*/
func (ins *InsertStmnt) ChunkVals() {
	const PG_MAX int = 2000 // MUST BE < 65,535
	var totRows int = len(ins.Rows)
	var valsPer int = len(ins.Rows[0])
	var maxRows int = PG_MAX / valsPer
	var totVals int = len(ins.Vals)

	// number of chunks needed
	//subtracting by 1 enables ceiling integer division
	var numChunks int = (totVals + PG_MAX - 1) / PG_MAX

	// make slice of slice with 2 ends for start/end position in rows
	chunkPos := make([][2]int, 0, numChunks)
	for i := range numChunks {
		start := i * maxRows
		end := min((start + maxRows), totRows) // last row if < (start + tot)
		chunkPos = append(chunkPos, [2]int{start, end})
	}

	// append [][]any w/ start & end pos data from ins.Rows
	for _, c := range chunkPos {
		var valChunk [][]any = ins.Rows[c[0]:c[1]]
		ins.Chunks = append(ins.Chunks, valChunk)
	}
}

// loop through the chunks & attempt to insert all rows from each one
func (ins *InsertStmnt) InsertFast(cnf *Conf) error {
	e := errd.InitErr()
	var wg sync.WaitGroup
	var mu sync.Mutex
	errCh := make(chan error, len(ins.Chunks))

	for i, c := range ins.Chunks {
		wg.Add(1)
		go func(i int, c [][]any) {
			defer wg.Done()
			st := time.Now()
			cnf.L.WriteLog(
				fmt.Sprintf(
					"starting chunk %d/%d - %v", i+1, len(ins.Chunks), st))
			res, err := cnf.DB.Exec(ins.BuildStmnt(c), ValsFromSet(c)...)
			if err != nil {
				e.Msg = fmt.Sprintf("error inserting chunk %d/%d", i+1, len(ins.Chunks))
				errCh <- e.BuildErr(err)
				return
			}
			ra, _ := res.RowsAffected()
			mu.Lock()
			cnf.RowCnt += ra // add rows affected to total
			cnf.L.WriteLog(
				fmt.Sprint(
					fmt.Sprintf("chunk %d/%d complete | rowsets: %d | vals: %d\n",
						i+1, len(ins.Chunks), len(c), len(ValsFromSet(c))),
					fmt.Sprintln("- ", time.Now()),
					fmt.Sprintln("- ", time.Since(st)),
					fmt.Sprintf("-- %d new rows inserted into %s\n", ra, ins.Tbl),
					fmt.Sprintln("-- total rows affected: ", cnf.RowCnt),
				),
			)
			mu.Unlock()
			time.Sleep(1 * time.Second)

		}(i, c)

	}

	wg.Wait()
	close(errCh)
	if len(errCh) > 0 {
		err := <-errCh
		e.Msg = "one or more chunks failed to insert"
		return e.BuildErr(err)
	}

	return nil
}

// loop through the chunks & attempt to insert all rows from each one
func (ins *InsertStmnt) Insert(cnf *Conf) error {
	e := errd.InitErr()
	for i, c := range ins.Chunks {
		res, err := cnf.DB.Exec(ins.BuildStmnt(c), ValsFromSet(c)...)
		if err != nil {
			e.Msg = fmt.Sprintf("error inserting chunk %d/%d", i+1, len(ins.Chunks))
			return e.BuildErr(err)
		}
		ra, _ := res.RowsAffected()
		cnf.RowCnt += ra // add rows affected to total
		cnf.L.WriteLog(
			fmt.Sprint(
				fmt.Sprintf(
					"chunk %d/%d: rowsets: %d | vals: %d\n---- %d new rows inserted into %s",
					i+1, len(ins.Chunks), len(c), len(ValsFromSet(c)), ra, ins.Tbl),
				"\n---- total rows affected: ", cnf.RowCnt,
			))
	}
	return nil
}

// construct the SQL statement to execute
func (ins *InsertStmnt) BuildStmnt(chunk [][]any) string {
	stmnt := fmt.Sprintf("insert into %s (", ins.Tbl)
	ins.addCols(&stmnt)
	ins.addChunkParams(&stmnt, chunk)
	return fmt.Sprintf("%s on conflict (%s) do nothing", stmnt, ins.PrimKey)
}

// use ins.Cols to add list of columns to sql statement
func (ins *InsertStmnt) addCols(stmnt *string) {
	for i, c := range ins.Cols {
		*stmnt += c
		if i < (len(ins.Cols) - 1) {
			*stmnt += ", "
		}
	}
	*stmnt += ")"
}

/*
creates list of placeholder params like ($1, $2, $3...)
postgres sql.Exec() function only accepts 65,535 total values per call
the chunk funcs break the vals into as many chunks of less than 65,535 as needed
*/
func (ins *InsertStmnt) addChunkParams(stmnt *string, chunk [][]any) {
	*stmnt += " values "
	for i, r := range chunk {
		*stmnt += "("
		for j := range r {
			// postgres uses 1-type placeholders, i*rows + idx of current val
			*stmnt += fmt.Sprintf("$%d", i*len(r)+(j+1))
			if j < (len(r) - 1) {
				*stmnt += ", "
			}
		}
		*stmnt += ")"
		if i < (len(chunk) - 1) {
			*stmnt += ", "
		}
	}
}
