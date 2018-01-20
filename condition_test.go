package godb

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestQ(t *testing.T) {
	Convey("Q build", t, func() {
		Convey("A simple condition with one placeholder and one argument", func() {
			sql := "id = ?"
			q := Q(sql, 123)
			So(q.sql, ShouldEqual, sql)
		})

		Convey("A simple condition with multiple placeholders arguments", func() {
			sql := "id = ? AND is_deleted = ?"
			q := Q(sql, 123, 0)
			So(q.sql, ShouldEqual, sql)
		})

		Convey("A condition expanding slice argument", func() {
			q := Q("id IN (?)", []int{123, 456})
			So(q.sql, ShouldEqual, "id IN (?,?)")
		})

		Convey("A condition with slice and non slice arguments", func() {
			q := Q("id IN (?) AND is_deleted = ?", []int{123, 456}, 0)
			So(q.sql, ShouldEqual, "id IN (?,?) AND is_deleted = ?")
		})
	})

	Convey("Q set error field ...", t, func() {
		Convey("Error is set if arguments and placeholders count does not math", func() {
			sql := "id = ?"
			q := Q(sql, 123, 456)
			So(q.error, ShouldNotBeNil)
		})

		Convey("Error is set if a given slice is empty", func() {
			q := Q("id IN (?)", []int{})
			So(q.error, ShouldNotBeNil)
		})

		Convey("Error is set if a given slice is nil", func() {
			q := Q("id IN (?)", nil)
			So(q.error, ShouldNotBeNil)
		})
	})
}

func TestAND(t *testing.T) {
	Convey("Given multiple conditions", t, func() {
		c1 := Q("id = ?", 123)
		c2 := Q("is_deleted = ?", 0)

		Convey("And conjunction joins conditions with 'AND'", func() {
			So(And(c1, c2).sql, ShouldEqual, "id = ? AND is_deleted = ?")
		})

		Convey("And conjunction joins arguments", func() {
			So(len(And(c1, c2).args), ShouldEqual, 2)
		})
	})

	Convey("Given one condition", t, func() {
		sql := "id = ?"
		c := Q(sql, 123)

		Convey("And conjunction return a similar condition", func() {
			So(And(c).sql, ShouldEqual, sql)
			So(len(And(c).args), ShouldEqual, 1)
		})
	})

	Convey("Given conditions with at least one error", t, func() {
		c := &Condition{error: fmt.Errorf("A mysterious error has occurred")}

		Convey("And return a condition with an error", func() {
			So(And(c).error, ShouldNotBeNil)
		})
	})
}

func TestOR(t *testing.T) {
	Convey("Given multiple conditions", t, func() {
		c1 := Q("id = ?", 123)
		c2 := Q("id = ?", 456)

		Convey("Or conjunction joins conditions with 'OR'", func() {
			So(Or(c1, c2).sql, ShouldContainSubstring, "id = ? OR id = ?")
		})

		Convey("Or conjunction joins arguments", func() {
			So(len(Or(c1, c2).args), ShouldEqual, 2)
		})
	})

	Convey("Given one condition", t, func() {
		sql := "id = ?"
		c := Q(sql, 123)

		Convey("Or conjunction return a similar condition", func() {
			So(Or(c).sql, ShouldEqual, sql)
			So(len(Or(c).args), ShouldEqual, 1)
		})
	})

	Convey("Given conditions with at least one error", t, func() {
		c := &Condition{error: fmt.Errorf("A mysterious error has occurred")}

		Convey("Or return a condition with an error", func() {
			So(Or(c).error, ShouldNotBeNil)
		})
	})
}

func TestNOT(t *testing.T) {
	Convey("Given a condition", t, func() {
		c := Q("id = ?", 123)

		Convey("Not surround the condition with 'NOT (' and ')'", func() {
			So(Not(c).sql, ShouldEqual, "NOT (id = ?)")
		})

		Convey("Not return a condition with the original arguments", func() {
			So(len(Not(c).args), ShouldEqual, len(c.args))
		})
	})

	Convey("Given a condition with an error", t, func() {
		c := &Condition{error: fmt.Errorf("A mysterious error has occurred")}

		Convey("Not return a condition with an error", func() {
			So(Not(c).error, ShouldNotBeNil)
		})
	})
}

func TestAllTogether(t *testing.T) {
	Convey("Given a 'complex' condition", t, func() {
		q := And(Q("id IN (?)", []int{1, 2, 3, 4, 5}),
			Not(Q("is_deleted = ?", 1)),
			Or(Q("category_id = ?", 123),
				Q("tag IN (?)", []string{"foo", "bar", "baz"}),
			),
		)

		Convey("The SQL string is correct", func() {
			sql := "id IN (?,?,?,?,?) AND NOT (is_deleted = ?) AND (category_id = ? OR tag IN (?,?,?))"
			So(q.sql, ShouldEqual, sql)
		})

		Convey("The aruments count is correct", func() {
			So(len(q.args), ShouldEqual, 10)
		})
	})
}
