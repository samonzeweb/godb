package postgresql

import (
	"bytes"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	"github.com/samonzeweb/godb/adapters"
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
	formatedColumns := make([]string, 0, len(columns))
	for _, column := range columns {
		formatedColumns = append(formatedColumns, p.Quote(column))
	}
	return formatedColumns
}

func (p PostgreSQL) GetReturningPosition() adapters.ReturningPosition {
	return adapters.ReturningPostgreSQL
}
