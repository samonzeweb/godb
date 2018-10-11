package mssql

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/samonzeweb/godb/adapters"
	"github.com/samonzeweb/godb/dberror"
	"github.com/samonzeweb/godb/dbreflect"

	_ "github.com/denisenkom/go-mssqldb"
)

// init registers types of mssql package corresponding to fields values
func init() {
	dbreflect.RegisterScannableStruct(Rowversion{})
}

type MSSQL struct{}

var Adapter = MSSQL{}

func (MSSQL) DriverName() string {
	return "sqlserver"
}

func (MSSQL) Quote(identifier string) string {
	return "[" + identifier + "]"
}

func (MSSQL) ReplacePlaceholders(originalPlaceholder string, sql string) string {
	sqlBuffer := bytes.NewBuffer(make([]byte, 0, len(sql)))
	count := 1
	for {
		pp := strings.Index(sql, originalPlaceholder)
		if pp == -1 {
			break
		}
		sqlBuffer.WriteString(sql[:pp])
		sqlBuffer.WriteString("@p")
		sqlBuffer.WriteString(strconv.Itoa(count))
		count++
		sql = sql[pp+1:]
	}
	sqlBuffer.WriteString(sql)
	return sqlBuffer.String()
}

func (m MSSQL) ReturningBuild(columns []string) string {
	suffixBuffer := bytes.NewBuffer(make([]byte, 0, 16*len(columns)+1))
	suffixBuffer.WriteString("OUTPUT ")
	for i, column := range columns {
		if i > 0 {
			suffixBuffer.WriteString(", ")
		}
		suffixBuffer.WriteString(column)
	}
	return suffixBuffer.String()
}

func (m MSSQL) FormatForNewValues(columns []string) []string {
	formatedColumns := make([]string, 0, len(columns))
	for _, column := range columns {
		formatedColumns = append(formatedColumns, "INSERTED."+m.Quote(column))
	}
	return formatedColumns
}

func (m MSSQL) GetReturningPosition() adapters.ReturningPosition {
	return adapters.ReturningSQLServer
}

func (MSSQL) BuildLimit(limit int) *adapters.SQLPart {
	sqlPart := adapters.SQLPart{}
	sqlPart.Sql = "FETCH NEXT ? ROWS ONLY"
	sqlPart.Arguments = make([]interface{}, 0, 1)
	sqlPart.Arguments = append(sqlPart.Arguments, limit)
	return &sqlPart
}

func (MSSQL) BuildOffset(offset int) *adapters.SQLPart {
	sqlPart := adapters.SQLPart{}
	sqlPart.Sql = "OFFSET ? ROWS"
	sqlPart.Arguments = make([]interface{}, 0, 1)
	sqlPart.Arguments = append(sqlPart.Arguments, offset)
	return &sqlPart
}

func (MSSQL) IsOffsetFirst() bool {
	return true
}

type ErrorWithNumber interface {
	SQLErrorNumber() int32
}

func (MSSQL) ParseError(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(ErrorWithNumber); ok {
		switch e.SQLErrorNumber() {
		case 2601:
			return dberror.UniqueConstraint{Message: err.Error(), Field: "", Err: err}
		case 2627:
			return dberror.UniqueConstraint{Message: err.Error(), Field: "", Err: err}
		case 547:
			return dberror.CheckConstraint{Message: err.Error(), Field: dberror.ExtractStr(err.Error(), "column '", "'"), Err: err}
		}
	}

	return err
}
