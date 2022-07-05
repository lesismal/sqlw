// Copyright 2022 lesismal. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

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

func (r *sqlResult) LastInsertId() (int64, error) {
	if r.Result != nil {
		return r.Result.LastInsertId()
	}
	return 0, nil
}

func (r *sqlResult) RowsAffected() (int64, error) {
	if r.Result != nil {
		return r.Result.RowsAffected()
	}
	return 0, nil
}

func (r *sqlResult) Sql() string {
	return fmt.Sprintf(`"%s", %v`, r.query, r.args)
}

func newResult(r sql.Result, query string, args []interface{}) Result {
	return &sqlResult{r, query, args}
}
