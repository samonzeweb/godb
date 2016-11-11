package godb

import "testing"

func TestQ(t *testing.T) {
	sql := "id = ?"
	q := Q(sql, 123)
	if q.sql != sql {
		t.Error(q.sql, "!=", sql)
	}
	if len(q.args) != 1 {
		t.Error("Wrong placeholder count :", len(q.args))
	}

	sql = "id = ? AND is_deleted = ?"
	q = Q(sql, 123, 0)
	if q.sql != sql {
		t.Error(q.sql, "!=", sql)
	}
	if len(q.args) != 2 {
		t.Error("Wrong placeholder count :", len(q.args))
	}

	sql = "id IN (?)"
	q = Q(sql, []int{123, 456})
	if q.sql != "id IN (?,?)" {
		t.Error(q.sql, "!=", sql)
	}
	if len(q.args) != 2 {
		t.Error("Wrong placeholder count :", len(q.args))
	}

	t.Log("Q check arguments count")
	q = Q("id = ?", 123, 456)
	if q.Error == nil {
		t.Error("Q didn't produced an error when the argument count is incorrect")
	}
}

func TestAnd(t *testing.T) {
	c1 := Q("id = ?", 123)
	c2 := Q("is_deleted = ?", 0)
	q := And(c1, c2)
	if q.sql != "id = ? AND is_deleted = ?" {
		t.Error("And didn't build correct SQL :", q.sql)
	}
	if len(q.args) != 2 {
		t.Error("And didn't build correct arguments, count :", len(q.args))
	}

	q = And(c1)
	if q.sql != c1.sql {
		t.Error("And didn't build correct SQL :", q.sql)
	}
	if len(q.args) != 1 {
		t.Error("And didn't build correct arguments, count :", len(q.args))
	}

	q = And(Q("id = ?", 123, 456))
	if q.Error == nil {
		t.Error("And didn't propagated error")
	}
}

func TestOr(t *testing.T) {
	c1 := Q("id = ?", 123)
	c2 := Q("id = ?", 456)
	q := Or(c1, c2)
	if q.sql != "(id = ? OR id = ?)" {
		t.Error("Or didn't build correct SQL :", q.sql)
	}
	if len(q.args) != 2 {
		t.Error("Or didn't build correct arguments, count :", len(q.args))
	}

	q = Or(c1)
	if q.sql != c1.sql {
		t.Error("Or didn't build correct SQL :", q.sql)
	}
	if len(q.args) != 1 {
		t.Error("Or didn't build correct arguments, count :", len(q.args))
	}

	q = Or(Q("id = ?", 123, 456))
	if q.Error == nil {
		t.Error("Or didn't propagated error")
	}
}

func TestNot(t *testing.T) {
	q := Not(Q("is_deleted = ?", 0))
	if q.sql != "NOT (is_deleted = ?)" {
		t.Error("Not didn't build correct SQL :", q.sql)
	}
	if len(q.args) != 1 {
		t.Error("Not didn't build correct arguments, count :", len(q.args))
	}

	q = Not(Q("id = ?", 123, 456))
	if q.Error == nil {
		t.Error("Not didn't propagated error")
	}
}

func TestAllTogether(t *testing.T) {
	q := And(Q("id IN (?)", []int{1, 2, 3, 4, 5}),
		Not(Q("is_deleted = ?", 1)),
		Or(Q("category_id = ?", 123),
			Q("tag IN (?)", []string{"foo", "bar", "baz"}),
		),
	)
	if q.sql != "id IN (?,?,?,?,?) AND NOT (is_deleted = ?) AND (category_id = ? OR tag IN (?,?,?))" {
		t.Error("Incorrect SQL :", q.sql)
	}
	if len(q.args) != 10 {
		t.Error("Incorrect arguments, count :", len(q.args))
	}
}
