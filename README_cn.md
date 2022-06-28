# sqlw - ra***w*** and simple ***w***rapped std sql

[![Doc][1]][2] [![Go Report Card][3]][4] [![MIT licensed][5]][6]

[1]: https://godoc.org/github.com/lesismal/sqlw?status.svg
[2]: https://godoc.org/github.com/lesismal/sqlw
[3]: https://goreportcard.com/badge/github.com/lesismal/sqlw
[4]: https://goreportcard.com/report/github.com/lesismal/sqlw
[5]: https://img.shields.io/badge/license-MIT-blue.svg
[6]: LICENSE


## 安装

```sh
go get github.com/lesismal/sqlw
```

## 使用

### 定义结构

**Noted**: 
1. 这里示例的结构体标签`db`，用于映射结构体与sql的表字段。
2. 如果您想使用一些三方工具自动生成结构体或sql表，但是三方工具有自定义的结构体标签，您可以在 `sqlw.Open` 时指定结构体标签，或者用 `db.SetTag()` 方法来修改该标签。

```golang
type Model struct {
	Id int64  `db:"id"`
	I  int64  `db:"i"`
	S  string `db:"s"`
}
```

### 创建sqlw.DB实例

```golang
import (
    _ "github.com/go-sql-driver/mysql"
    "github.com/lesismal/sqlw"
)

// "db" 是您的结构体用于与sql表字段映射的标签, 如果您使用三方工具自动生成结构体或sql表，您可以根据该工具生成的实际标签作为参数
db, err := sqlw.Open("mysql", SqlConnStr, "db")
if err != nil {
    // handle err
}
```

### 创建/使用事务

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


### 创建/使用Stmt/预编译

```golang
stmt, err := db.Prepare(`your sql`)
if err != nil {
    // handle err
}

// curd logic using stmt
```

### 插入一条记录

```golang
model := Model{
    I: 1,
    S: "str_1",
}

result, err := db.Insert("insert into sqlw_test.sqlw_test", &model)
// result, err := db.Insert("insert into sqlw_test.sqlw_test(i,s)", &model) // 插入结构体指定字段
if err != nil {
    log.Panic(err)
}
log.Println("sql:", result.Sql())
```

### 插入多条记录

```golang
var models []*Model
for i:=0; i<3; i++{
    models = append(models, &Model{
        I: i,
        S: fmt.Sprintf("str_%v", i),
    })
}

result, err := db.Insert("insert into sqlw_test.sqlw_test", models)
// result, err := db.Insert("insert into sqlw_test.sqlw_test(i,s)", models) // 插入结构体指定字段
if err != nil {
    log.Panic(err)
}
log.Println("sql:", result.Sql())
```

### 删除记录

```golang
deleteId := 1
result, err := db.Delete("delete from sqlw_test.sqlw_test where id=?", deleteId)
if err != nil {
    log.Panic(err)
}
log.Println("sql:", result.Sql())
```

### 更新记录

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

### 查询单条记录

```golang
var model Model
selectId := 1
result, err := db.Select(&model, "select * from sqlw_test.sqlw_test where id=?", selectId)
// result, err := db.Select(&model, "select (i,s) from sqlw_test.sqlw_test where id=?", selectId) // 查询结构体指定字段
if err != nil {
    log.Panic(err)
}
log.Println("model:", model)
log.Println("sql:", result.Sql())
```

### 查询多条记录

```golang
var models []*Model // type []Model is also fine
result, err = db.Select(&models, "select * from sqlw_test.sqlw_test")
// result, err = db.Select(&models, "select (i,s) from sqlw_test.sqlw_test") // 查询结构体指定字段
if err != nil {
    log.Panic(err)
}
for i, v := range models {
    log.Printf("models[%v]: %v", i, v)
}
log.Println("sql:", result.Sql())
```

### 获取执行的sql语句及参数

> `sqlw.DB/Tx/Stmt` 的所有 `Query/QueryRow/Exec/Insert/Delete/Update/Select` 相关方法都会返回 `(sqlw.Result, error)`，
> 其中的 `sqlw.Result` 是非 nil 的，您可以通过 `sqlw.Result.Sql()` 获取实际执行的sql语句及参数并辅助日志或调试。

例如：
```golang
result, err := db.Insert(`insert into t(a,b) values(?,?)`, 1, 2)
if err != nil {
    // handle err
}
fmt.Println("sql:", result.Sql)
```

输出：
```sh
sql: insert into t(a,b) values(?,?), [1, 2]
```

### 更多示例
请参考：[sqlw_examples](https://github.com/lesismal/sqlw_examples)
