# HIDEN GO SQLITE DRIVER
Hiden is a SQLite database driver written in Go.

It currently does not satisfy Go's database/sql interface, but probs will eventually.


## API
Keeping it super simple, Hiden only has 3 points of contact:
* `Connect`
* `Close`
* `Execute`

### `Connect`
Used to connect to a SQLite database. On success returns a pointer to the database object and nil error. On failure returns nil database pointer and a DBError.
Usage:
```go
db, err := Connect("mysqlite.db")
```

### `Close`
Used to close connection to the SQLite database. Returns an error on failure.
Usage:
```go
db, err := Connect("mysqlite.db")
defer db.Close()
```

### `Execute`
Used to execute raw SQL. Returns a QueryResults object and nil error, on success. On failure, returns nil result and DBError.
Usage:
```go
db, err := Connect("mysqlite.db")
defer db.Close()

results, err := db.Execute("select * from users;")
```

The QueryResults object returned from `Execute` has default formatted string output.
For example
```go
fmt.Println(results)
```
will product something like:
```QueryResults: Size = 2
┌------------┐
| id | name  |
|------------|
| 1  | Alice |
| 2  | Bob   |
└------------┘
```
