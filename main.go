package main

/*
#cgo CFLAGS: -I/usr/include
#cgo LDFLAGS: -lsqlite3

#include <stdio.h>
#include <string.h>
#include <sqlite3.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"strings"
	"unsafe"
)

type DBConnection = C.sqlite3

type QueryResults struct {
	columns []string
	values  [][]any
}

func (r *QueryResults) String() string {
	if len(r.values) == 0 {
		return "Empty Result Set"
	}

	columnMaxLength := make([]int, len(r.columns))
	for i, col := range r.columns {
		columnMaxLength[i] = len(col)
	}

	for _, row := range r.values {
		for i, val := range row {
			length := len(fmt.Sprintf("%v", val))
			if length > columnMaxLength[i] {
				columnMaxLength[i] = length
			}
		}
	}

	output := fmt.Sprintf("QueryResults: Size = %d\n", len(r.values))
	var header string
	for i, col := range r.columns {
		padLength := columnMaxLength[i] - len(col)
		header += "| " + col + strings.Repeat(" ", padLength+1)
	}
	header += " |\n"

	output += "┌" + strings.Repeat("-", len(header)-3) + "┐\n"
	output += header
	output += "|" + strings.Repeat("-", len(header)-3) + "|\n"

	for _, row := range r.values {
		var line string
		for i, val := range row {
			padLength := columnMaxLength[i] - len(fmt.Sprintf("%v", val))
			line += "| " + fmt.Sprintf("%v%s ", val, strings.Repeat(" ", padLength))
		}
		output += line + " |" + "\n"
	}
	output += "└" + strings.Repeat("-", len(header)-3) + "┘"

	return output
}

type DBError struct {
	msg string
}

func (e *DBError) Error() string {
	return e.msg
}

func Connect(dbName string) (*DBConnection, error) {
	var db *DBConnection
	dbName_c := C.CString(dbName)
	defer C.free(unsafe.Pointer(dbName_c))
	if C.sqlite3_open(dbName_c, &db) != C.SQLITE_OK {
		fmt.Println("no good")
		return nil, &DBError{msg: fmt.Sprintf("Failed to open database: %s", dbName)}
	}
	return db, nil
}

func (db *DBConnection) Execute(sql string, params ...any) (*QueryResults, error) {
	var stmt *C.sqlite3_stmt
	sql_c := C.CString(sql)
	defer C.free(unsafe.Pointer(sql_c))

	if C.sqlite3_prepare(db, sql_c, -1, &stmt, nil) != C.SQLITE_OK {
		return nil, &DBError{msg: "Failed to prepare query"}
	}
	defer C.sqlite3_finalize(stmt)

	parameterCount := C.sqlite3_bind_parameter_count(stmt)
	for i := range parameterCount {
		val := params[i]
		switch v := val.(type) {
		case int:
			if C.sqlite3_bind_int(stmt, i+1, C.int(v)) != C.SQLITE_OK {
				return nil, &DBError{msg: "Failed to bind int parameter"}
			}

		case string:
			bindValue := C.CString(v)
			defer C.free(unsafe.Pointer(bindValue))
			if C.sqlite3_bind_text(stmt, i+1, bindValue, -1, C.SQLITE_STATIC) != C.SQLITE_OK {
				return nil, &DBError{msg: "Failed to bind string parameter"}
			}

		default:
			return nil, &DBError{msg: fmt.Sprintf("Type %T not yet implemented", v)}
		}
	}

	cols := C.sqlite3_column_count(stmt)
	headers := make([]string, cols)
	for i := range cols {
		headers[i] = C.GoString(C.sqlite3_column_name(stmt, i))
	}

	results := make([][]any, 0)
	for C.sqlite3_step(stmt) == C.SQLITE_ROW {
		entry := make([]any, 0)
		for i := range cols {
			datatype := C.sqlite3_column_type(stmt, i)
			switch datatype {
			case C.SQLITE_INTEGER:
				val := C.sqlite3_column_int(stmt, i)
				entry = append(entry, val)
			case C.SQLITE_TEXT:
				val := C.GoString((*C.char)(unsafe.Pointer(C.sqlite3_column_text(stmt, i))))
				entry = append(entry, val)
			case C.SQLITE_NULL:
				val := "null"
				entry = append(entry, val)
			}
		}
		results = append(results, entry)
	}

	queryResult := &QueryResults{
		columns: headers,
		values:  results,
	}

	return queryResult, nil
}

func (db *DBConnection) Close() error {
	if C.sqlite3_close(db) != C.SQLITE_OK {
		return &DBError{msg: "Failed to close db"}
	}
	return nil
}

func main() {
	db, _ := Connect("test.db")
	defer db.Close()
	results, _ := db.Execute("select * from users;")
	// results, _ := db.Execute("alter table users add column sogginess int;")
	// results, _ := db.Execute("insert into users (name) values (?)", "floyd")
	fmt.Println(results)
}
