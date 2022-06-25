package sqlw

import (
	"database/sql"
)

type sqlResult struct {
	sql.Result
	query string
}

type Result = *sqlResult

func (r *sqlResult) Sql() string {
	return r.query
}

func newResult(r sql.Result, query string) Result {
	return &sqlResult{r, query}
}
