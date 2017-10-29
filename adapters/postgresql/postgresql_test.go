package postgresql

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestReplacePlaceholders(t *testing.T) {
	Convey("Given a SQL string containing placeholders", t, func() {
		sql := "SELECT id, dummy FROM dummies WHERE id > ? AND id < ?"
		Convey("ReplacePlaceholders change all placeholders with $xx", func() {
			sqlWithNewPlaceholders := Adapter.ReplacePlaceholders("?", sql)
			So(sqlWithNewPlaceholders, ShouldEqual, "SELECT id, dummy FROM dummies WHERE id > $1 AND id < $2")
		})
	})
}

func TestReturningBuild(t *testing.T) {
	Convey("Given list of columns", t, func() {
		columns := []string{"id", "other_stuff"}
		Convey("ReturningBuild build a RETURNING clause with the given columns", func() {
			returningClause := Adapter.ReturningBuild(columns)
			So(returningClause, ShouldEqual, "RETURNING \"id\",\"other_stuff\"")
		})
	})
}

func TestFormatForNewValues(t *testing.T) {
	Convey("Given list of columns", t, func() {
		columns := []string{"id", "other_stuff"}
		Convey("FormatForNewValues returns a copy of the columns list (without change)", func() {
			formatedColumns := Adapter.FormatForNewValues(columns)
			So(len(formatedColumns), ShouldEqual, len(columns))
			for i, column := range columns {
				So(formatedColumns[i], ShouldEqual, column)
			}
		})
	})
}
