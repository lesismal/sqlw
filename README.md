# sqlw - ra***w*** and simple ***w***rapped std sql

[![Doc][1]][2] [![Go Report Card][3]][4] [![MIT licensed][5]][6]

[1]: https://godoc.org/github.com/lesismal/sqlw?status.svg
[2]: https://godoc.org/github.com/lesismal/sqlw
[3]: https://goreportcard.com/badge/github.com/lesismal/sqlw
[4]: https://goreportcard.com/report/github.com/lesismal/sqlw
[5]: https://img.shields.io/badge/license-MIT-blue.svg
[6]: LICENSE


## Install

```sh
go get github.com/lesismal/sqlw
```

## Usage

### Define Model

**Noted**: 
1. The `db` tags of the struct are used to map struct fields to sql table fields in our `insert/update/select` operations.
2. If you want to use some tools to auto-generate structs or sql tables but the tools use different struct tag names, you can set the tag when `sqlw.Open`, or modify it using `db.SetTag()`.

```golang
type Model struct {
	Id int64  `db:"id"`
	I  int64  `db:"i"`
	S  string `db:"s"`
}
```

### Open DB

```golang
import (
    _ "github.com/go-sql-driver/mysql"
    "github.com/lesismal/sqlw"
)

// "db" is Model struct tag name, if you want to use some tools to auto-generate structs or sql tables, 
// you can set the tag name according to your tools.
db, err := sqlw.Open("mysql", SqlConnStr, "db")
if err != nil {
    // handle err
}
```

### Transaction

```golang
tx, err := db.Begin()
if err != nil {
    // handle err
}
defer tx.Rollback()

// curd logic

err = tx.Commit()
if err != nil {
    // handle err
}
```


### Prepare/Stmt

```golang
stmt, err := db.Prepare(`your sql`)
if err != nil {
    // handle err
}

// curd logic using stmt
```

### Insert One Records

```golang
model := Model{
    I: 1,
    S: "str_1",
}

result, err := db.Insert("insert into sqlw_test.sqlw_test", &model)
// result, err := db.Insert("insert into sqlw_test.sqlw_test(i,s)", &model) // insert the specified fields
if err != nil {
    log.Panic(err)
}
log.Println("sql:", result.Sql())
```

### Insert Multi Records

```golang
var models []*Model
for i:=0; i<3; i++{
    models = append(models, &Model{
        I: i,
        S: fmt.Sprintf("str_%v", i),
    })
}

result, err := db.Insert("insert into sqlw_test.sqlw_test", models)
// result, err := db.Insert("insert into sqlw_test.sqlw_test(i,s)", models) // insert the specified fields
if err != nil {
    log.Panic(err)
}
log.Println("sql:", result.Sql())
```

### Delete

```golang
deleteId := 1
result, err := db.Delete("delete from sqlw_test.sqlw_test where id=?", deleteId)
if err != nil {
    log.Panic(err)
}
log.Println("sql:", result.Sql())
```

### Update

```golang
m := Model{
    I: 10,
    S: "str_10",
}

updateId := 1
result, err := db.Update("update sqlw_test.sqlw_test set i=?, s=? where id=?", &m, updateId)
if err != nil {
    log.Panic(err)
}
log.Println("sql:", result.Sql())
```

### Select One Record

```golang
var model Model
selectId := 1
result, err := db.Select(&model, "select * from sqlw_test.sqlw_test where id=?", selectId)
// result, err := db.Select(&model, "select (i,s) from sqlw_test.sqlw_test where id=?", selectId) // select the specified fields
if err != nil {
    log.Panic(err)
}
log.Println("model:", model)
log.Println("sql:", result.Sql())
```

### Select Multi Records

```golang
var models []*Model // type []Model is also fine
result, err = db.Select(&models, "select * from sqlw_test.sqlw_test")
// result, err = db.Select(&models, "select (i,s) from sqlw_test.sqlw_test") // select the specified fields
if err != nil {
    log.Panic(err)
}
for i, v := range models {
    log.Printf("models[%v]: %v", i, v)
}
log.Println("sql:", result.Sql())
```

### Get RawSql

> All `Query/QueryRow/Exec/Insert/Delete/Update/Select` related funcs of `sqlw.DB/Tx/Stmt` return 
> `(sqlw.Result, error)`.
> The `sqlw.Result` would always be a non-nil value to help users getting the raw sql, we can use 
> `sqlw.Result.Sql()` to print it out.

For example:
```golang
result, err := db.Insert(`insert into t(a,b) values(?,?)`, 1, 2)
if err != nil {
    // handle err
}
fmt.Println("sql:", result.Sql())
```

Output:
```sh
sql: insert into t(a,b) values(?,?), [1, 2]
```

### For More Examples
Please refer to: [sqlw_examples](https://github.com/lesismal/sqlw_examples)
