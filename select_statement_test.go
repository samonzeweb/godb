package godb

import "testing"

func TestSimpleSelectStatement(t *testing.T) {
	selectStmt := newSelectStatement(&DB{}, "dummies").Columns("foo").Distinct()

	sql, _, err := selectStmt.ToSQL()
	checkToSQL(t, "SELECT DISTINCT foo FROM dummies", sql, err)
}

func TestSelectStatementWhere(t *testing.T) {
	selectStmt := newSelectStatement(&DB{}, "dummies").
		Columns("foo").
		Where("id = ?", 123)

	sql, args, err := selectStmt.ToSQL()
	checkToSQL(t, "SELECT foo FROM dummies WHERE id = ?", sql, err)

	if len(args) != 1 || args[0] != 123 {
		t.Fatal("Bad arguments list :", args)
	}
}

func TestSelectStatementWhereQ(t *testing.T) {
	selectStmt := newSelectStatement(&DB{}, "dummies").
		Columns("foo").
		WhereQ(Q("id = ?", 1))

	sql, _, err := selectStmt.ToSQL()
	checkToSQL(t, "SELECT foo FROM dummies WHERE id = ?", sql, err)
}

func TestSelectStatementGroupByHaving(t *testing.T) {
	selectStmt := newSelectStatement(&DB{}, "dummies").
		Columns("foo", "count(*)").
		GroupBy("foo").
		Having("count(*) > 1")

	sql, _, err := selectStmt.ToSQL()
	checkToSQL(t, "SELECT foo, count(*) FROM dummies GROUP BY foo HAVING count(*) > 1", sql, err)
}

func TestSelectStatementOrderOffsetLimit(t *testing.T) {
	selectStmt := newSelectStatement(&DB{}, "dummies").
		Columns("foo", "bar").
		OrderBy("foo").
		Offset(2).
		Limit(1)

	sql, _, err := selectStmt.ToSQL()
	checkToSQL(t, "SELECT foo, bar FROM dummies ORDER BY foo OFFSET ? LIMIT ?", sql, err)
}

func TestSelectStatementSuffix(t *testing.T) {
	selectStmt := newSelectStatement(&DB{}, "dummies").
		Columns("foo").
		Suffix("FOR UPDATE")

	sql, _, err := selectStmt.ToSQL()
	checkToSQL(t, "SELECT foo FROM dummies FOR UPDATE", sql, err)
}

func TestLeftJoin(t *testing.T) {
	selectStmt := newSelectStatement(&DB{}, "dummies").
		Columns("id", "foo", "other1.bar").
		LeftJoin("other", "other1", Q("dummies.id = other1.dummy_id"))

	sql, _, err := selectStmt.ToSQL()
	checkToSQL(t, "SELECT id, foo, other1.bar FROM dummies LEFT JOIN other AS other1 ON dummies.id = other1.dummy_id", sql, err)
}
