package mssql

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestReturningBuild(t *testing.T) {
	Convey("Given list of columns", t, func() {
		columns := []string{"id", "other_stuff"}
		Convey("ReturningBuild build a RETURNING clause with the given columns", func() {
			returningClause := Adapter.ReturningBuild(columns)
			So(returningClause, ShouldEqual, "OUTPUT id, other_stuff")
		})
	})
}

func TestFormatForNewValues(t *testing.T) {
	Convey("Given list of columns", t, func() {
		columns := []string{"id", "other_stuff"}
		Convey("FormatForNewValues returns a list of columns, each quoted prefixed with 'INSERTED'", func() {
			formatedColumns := Adapter.FormatForNewValues(columns)
			So(len(formatedColumns), ShouldEqual, len(columns))
			for i, column := range columns {
				So(formatedColumns[i], ShouldEqual, "INSERTED.["+column+"]")
			}
		})
	})
}
