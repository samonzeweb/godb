package postgresql

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/samonzeweb/godb/adapters"
	"github.com/samonzeweb/godb/dberror"
)

type PostgreSQL struct{}

var Adapter = PostgreSQL{}

func (PostgreSQL) DriverName() string {
	return "postgres"
}

func (PostgreSQL) Quote(identifier string) string {
	return "\"" + identifier + "\""
}

func (PostgreSQL) ReplacePlaceholders(originalPlaceholder string, sql string) string {
	sqlBuffer := bytes.NewBuffer(make([]byte, 0, len(sql)))
	count := 1
	for {
		pp := strings.Index(sql, originalPlaceholder)
		if pp == -1 {
			break
		}
		sqlBuffer.WriteString(sql[:pp])
		sqlBuffer.WriteString("$")
		sqlBuffer.WriteString(strconv.Itoa(count))
		count++
		sql = sql[pp+1:]
	}
	sqlBuffer.WriteString(sql)
	return sqlBuffer.String()
}

func (p PostgreSQL) ReturningBuild(columns []string) string {
	suffixBuffer := bytes.NewBuffer(make([]byte, 0, 16*len(columns)+1))
	suffixBuffer.WriteString("RETURNING ")
	for i, column := range columns {
		if i > 0 {
			suffixBuffer.WriteString(", ")
		}
		suffixBuffer.WriteString(column)
	}
	return suffixBuffer.String()
}

func (p PostgreSQL) FormatForNewValues(columns []string) []string {
	formattedColumns := make([]string, 0, len(columns))
	for _, column := range columns {
		formattedColumns = append(formattedColumns, p.Quote(column))
	}
	return formattedColumns
}

func (p PostgreSQL) GetReturningPosition() adapters.ReturningPosition {
	return adapters.ReturningPostgreSQL
}

func (p PostgreSQL) ParseError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "duplicate key value violates unique constraint") || strings.Contains(errMsg, "(SQLSTATE 23505)"):
		return dberror.UniqueConstraint{Message: errMsg, Field: dberror.ExtractStr(errMsg, "constraint \"", "\""), Err: err}
	case strings.Contains(errMsg, "violates foreign key constraint") || strings.Contains(errMsg, "(SQLSTATE 23503)"):
		return dberror.ForeignKeyConstraint{Message: errMsg, Field: dberror.ExtractStr(errMsg, "constraint \"", "\""), Err: err}
	case strings.Contains(errMsg, "violates check constraint") || strings.Contains(errMsg, "(SQLSTATE 23514)"):
		return dberror.CheckConstraint{Message: errMsg, Field: dberror.ExtractStr(errMsg, "constraint \"", "\""), Err: err}
	}

	return err
}
