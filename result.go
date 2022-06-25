package sqlw

import (
	"database/sql"
	"fmt"
)

type sqlResult struct {
	sql.Result
	query string
	args  []interface{}
}

type Result = *sqlResult

func (r *sqlResult) Sql() string {
	return fmt.Sprintf(`"%s", %v`, r.query, r.args)
}

func newResult(r sql.Result, query string, args []interface{}) Result {
	return &sqlResult{r, query, args}
}
