package godb

import (
	"database/sql"
	"testing"

	"github.com/samonzeweb/godb/adapters/sqlite"
	. "github.com/smartystreets/goconvey/convey"
)

type structWithColumns struct {
	Foo int    `db:"foo"`
	Bar string `db:"bar"`
}

func TestNewSelectStatement(t *testing.T) {
	Convey("Create a select statement", t, func() {

		Convey("Without columns", func() {
			db := &DB{}
			q := db.SelectFrom("dummies")
			Convey("The table name is not empty", func() {
				So(len(q.fromTables), ShouldEqual, 1)
				So(q.fromTables[0], ShouldEqual, "dummies")
			})
		})
	})
}

func TestSelectColumns(t *testing.T) {
	Convey("Given a select statement", t, func() {
		db := &DB{}
		q := db.SelectFrom("dummies")

		Convey("Columns add columns after the existing list", func() {
			q.Columns("foo", "bar", "baz")
			So(len(q.columns), ShouldEqual, 3)
			So(q.columns[0], ShouldEqual, "foo")
			So(q.columns[1], ShouldEqual, "bar")
		})

		Convey("You can't use Columns after ColumnsFromStruct", func() {
			_, _, err := q.Columns("baz").
				ColumnsFromStruct(&structWithColumns{}).
				ToSQL()
			So(err, ShouldNotBeNil)
		})

	})
}

func TestSelectColumnsFromStruct(t *testing.T) {
	Convey("Given a select statement", t, func() {
		db := &DB{adapter: sqlite.Adapter}
		q := db.SelectFrom("dummies")

		Convey("ColumnsFromStruct add columns", func() {
			q.ColumnsFromStruct(&structWithColumns{})
			So(len(q.columns), ShouldEqual, 2)
			So(q.columns[0], ShouldEqual, "\"foo\"")
			So(q.columns[1], ShouldEqual, "\"bar\"")
		})

		Convey("You can't use ColumnsFromStruct after Columns", func() {
			_, _, err := q.ColumnsFromStruct(&structWithColumns{}).
				Columns("baz").
				ToSQL()
			So(err, ShouldNotBeNil)
		})
	})
}

func TestSelectFrom(t *testing.T) {
	Convey("Given a select statement", t, func() {
		db := &DB{}
		q := db.SelectFrom("dummies").
			Columns("foo", "bar", "baz")

		Convey("Calling From append a table name to the list", func() {
			q.From("others")
			So(len(q.fromTables), ShouldEqual, 2)
			So(q.fromTables[0], ShouldEqual, "dummies")
			So(q.fromTables[1], ShouldEqual, "others")
		})
	})
}

func TestSelectLeftJoin(t *testing.T) {
	Convey("Given a select query", t, func() {
		db := &DB{}
		q := db.SelectFrom("dummies").
			Columns("foo", "bar", "baz")

		Convey("Calling LeftJoin will add the given string to the joins list", func() {
			q.LeftJoin("others", "othersalias", Q("othersalias.id = dummies.other_id"))
			So(len(q.joins), ShouldEqual, 1)
		})
	})
}

func TestSelectInnerJoin(t *testing.T) {
	Convey("Given a select query", t, func() {
		db := &DB{}
		q := db.SelectFrom("dummies").
			Columns("foo", "bar", "baz")

		Convey("Calling LeftJoin will add the given string to the joins list", func() {
			q.InnerJoin("others", "othersalias", Q("othersalias.id = dummies.other_id"))
			So(len(q.joins), ShouldEqual, 1)
		})
	})
}

func TestSelectWhere(t *testing.T) {
	Convey("Given a select query", t, func() {
		db := &DB{}
		q := db.SelectFrom("dummies").
			Columns("foo", "bar", "baz")

		Convey("Call Where will add a new condition", func() {
			sql := "id = ?"
			q.Where(sql, 123)
			So(len(q.where), ShouldEqual, 1)
			So(q.where[0].sql, ShouldEqual, sql)
		})
	})
}

func TestSelectWhereQ(t *testing.T) {
	Convey("Given a select query", t, func() {
		db := &DB{}
		q := db.SelectFrom("dummies").
			Columns("foo", "bar", "baz")

		Convey("Call WhereQ will add the given condition", func() {
			qc := Q("id = ?", 123)
			q.WhereQ(qc)
			So(len(q.where), ShouldEqual, 1)
			So(q.where[0], ShouldEqual, qc)
		})
	})
}

func TestSelectGroupBy(t *testing.T) {
	Convey("Given a select query", t, func() {
		db := &DB{}
		q := db.SelectFrom("dummies").
			Columns("foo", "count(*)")

		Convey("Calling GroupBy will add the given string to the groupBy list", func() {
			groupBy := "foo"
			q.GroupBy(groupBy)
			So(len(q.groupBy), ShouldEqual, 1)
			So(q.groupBy[0], ShouldEqual, groupBy)
		})
	})
}

func TestSelectHaving(t *testing.T) {
	Convey("Given a select query", t, func() {
		db := &DB{}
		q := db.SelectFrom("dummies").
			Columns("foo", "count(*)")

		Convey("Call Having will add a new condition", func() {
			sql := "count(*) > 1"
			q.Having(sql)
			So(len(q.having), ShouldEqual, 1)
			So(q.having[0].sql, ShouldEqual, sql)
		})
	})
}

func TestSelectHavingQ(t *testing.T) {
	Convey("Given a select query", t, func() {
		db := &DB{}
		q := db.SelectFrom("dummies").
			Columns("foo", "count(*)")

		Convey("Call WhereQ will add the given condition", func() {
			qc := Q("count(*) > 1")
			q.HavingQ(qc)
			So(len(q.having), ShouldEqual, 1)
			So(q.having[0], ShouldEqual, qc)
		})
	})
}

func TestSelectOrderBy(t *testing.T) {
	Convey("Given a select query", t, func() {
		db := &DB{}
		q := db.SelectFrom("dummies").
			Columns("foo", "bar", "baz")

		Convey("Calling OrderBy will add the given string to the orderBy list", func() {
			orderBy := "foo"
			q.OrderBy(orderBy)
			So(len(q.orderBy), ShouldEqual, 1)
			So(q.orderBy[0], ShouldEqual, orderBy)
		})
	})
}

func TestSelectOffet(t *testing.T) {
	Convey("Given a select query", t, func() {
		db := &DB{}
		q := db.SelectFrom("dummies").
			Columns("foo", "bar", "baz")

		Convey("Calling Offset will set the offset value", func() {
			q.Offset(123)
			So(q.offset, ShouldNotBeNil)
			So(*q.offset, ShouldEqual, 123)
		})
	})
}

func TestSelectLimit(t *testing.T) {
	Convey("Given a select query", t, func() {
		db := &DB{}
		q := db.SelectFrom("dummies").
			Columns("foo", "bar", "baz")

		Convey("Calling Limit will set the offset value", func() {
			q.Limit(123)
			So(q.limit, ShouldNotBeNil)
			So(*q.limit, ShouldEqual, 123)
		})
	})
}

func TestSelectSuffix(t *testing.T) {
	Convey("Given a select query", t, func() {
		db := &DB{}
		q := db.SelectFrom("dummies").
			Columns("foo", "bar", "baz")

		Convey("Calling Suffix will add the given string to the suffixes list", func() {
			suffix := "FOR UPDATE"
			q.Suffix(suffix)
			So(len(q.suffixes), ShouldEqual, 1)
			So(q.suffixes[0], ShouldEqual, suffix)
		})
	})
}

func TestSelectToSQL(t *testing.T) {
	Convey("Given a select query with columns and table", t, func() {
		db := &DB{}
		q := db.SelectFrom("dummies").
			Columns("foo", "bar", "baz")

		Convey("ToSQL create a SQL request", func() {
			sql, _, err := q.ToSQL()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual, "SELECT foo, bar, baz FROM dummies")
		})

		Convey("Calling Distinct will add the distinct clause to SQL", func() {
			q.Distinct()
			sql, _, _ := q.ToSQL()
			So(sql, ShouldStartWith, "SELECT DISTINCT")
		})

		Convey("Calling Join will add the specified join clause to SQL", func() {
			q.LeftJoin("others", "othersalias", Q("othersalias.id = dummies.other_id"))
			sql, _, _ := q.ToSQL()
			So(sql, ShouldEndWith, "LEFT JOIN others AS othersalias ON othersalias.id = dummies.other_id")
		})

		Convey("Calling Where multiple times", func() {
			q.Where("id = ?", 123).Where("is_deleted = ?", 0)
			sql, args, _ := q.ToSQL()

			Convey("will add all the specified conditions clause to SQL using the 'AND' conjunction", func() {
				So(sql, ShouldEndWith, "WHERE id = ? AND is_deleted = ?")
			})

			Convey("will add given arguments in the correct order", func() {
				So(len(args), ShouldEqual, 2)
				So(args[0].(int), ShouldEqual, 123)
				So(args[1].(int), ShouldEqual, 0)
			})
		})

		Convey("Calling GroupBy will add the specified group by clause to SQL", func() {
			q.GroupBy("foo").GroupBy("bar")
			sql, _, _ := q.ToSQL()
			So(sql, ShouldEndWith, "GROUP BY foo, bar")
		})

		Convey("Calling Having multiple times", func() {
			q.GroupBy("foo")
			q.Having("count(*) > ?", 0).Having("count(*) < ?", 10)
			sql, args, _ := q.ToSQL()

			Convey("will add all the specified conditions clause to SQL using the 'AND' conjunction", func() {
				So(sql, ShouldEndWith, "count(*) > ? AND count(*) < ?")
			})

			Convey("will add given arguments in the correct order", func() {
				So(len(args), ShouldEqual, 2)
				So(args[0].(int), ShouldEqual, 0)
				So(args[1].(int), ShouldEqual, 10)
			})
		})

		Convey("Calling OrderBy will add the specified order by clause to SQL", func() {
			q.OrderBy("foo").OrderBy("bar")
			sql, _, _ := q.ToSQL()
			So(sql, ShouldEndWith, "ORDER BY foo, bar")
		})

		Convey("Calling Offset will add the offset clause to SQL", func() {
			q.Offset(10)
			sql, args, _ := q.ToSQL()
			So(sql, ShouldEndWith, "OFFSET ?")
			So(args[0].(int), ShouldEqual, 10)
		})

		Convey("Calling Limit will add the limit clause to SQL", func() {
			q.Limit(10)
			sql, args, _ := q.ToSQL()
			So(sql, ShouldEndWith, "LIMIT ?")
			So(args[0].(int), ShouldEqual, 10)
		})

		Convey("Calling Suffix will add the given clause to SQL", func() {
			q.Suffix("FOR UPDATE")
			sql, _, _ := q.ToSQL()
			So(sql, ShouldEndWith, "FOR UPDATE")
		})

	})
}

func TestSelectToSQLErrors(t *testing.T) {
	Convey("Columns are mandatory", t, func() {
		db := &DB{}
		q := db.SelectFrom("dummies")
		_, _, err := q.ToSQL()
		So(err, ShouldNotBeNil)
	})

	Convey("Calling Having without GroupBy will returns an error", t, func() {
		db := &DB{}
		q := db.SelectFrom("dummies").
			Columns("foo", "count(*)")
		q.Having("count(*) > 1")
		_, _, err := q.ToSQL()
		So(err, ShouldNotBeNil)
	})
}

func TestSelectDo(t *testing.T) {
	Convey("Given a test database", t, func() {
		db := fixturesSetup(t)
		defer db.Close()

		Convey("Do execute the query and fills a given instance", func() {
			singleDummy := Dummy{}
			selectStmt := db.SelectFrom("dummies").
				Columns("id", "a_text", "another_text", "an_integer").
				Where("an_integer = ?", 13)

			err := selectStmt.Do(&singleDummy)
			So(err, ShouldBeNil)
			So(singleDummy.ID, ShouldBeGreaterThan, 0)
			So(singleDummy.AText, ShouldEqual, "Third")
			So(singleDummy.AnotherText, ShouldEqual, "Troisième")
			So(singleDummy.AnInteger, ShouldEqual, 13)
		})

		Convey("Do execute the query and fills a slice", func() {
			dummiesSlice := make([]Dummy, 0)
			selectStmt := db.SelectFrom("dummies").
				Columns("id", "a_text", "another_text", "an_integer").
				OrderBy("an_integer")

			err := selectStmt.Do(&dummiesSlice)
			So(err, ShouldBeNil)
			So(len(dummiesSlice), ShouldEqual, 3)
			So(dummiesSlice[0].ID, ShouldBeGreaterThan, 0)
			So(dummiesSlice[0].AText, ShouldEqual, "First")
			So(dummiesSlice[0].AnotherText, ShouldEqual, "Premier")
			So(dummiesSlice[0].AnInteger, ShouldEqual, 11)
			So(dummiesSlice[1].AnInteger, ShouldEqual, 12)
			So(dummiesSlice[2].AnInteger, ShouldEqual, 13)
		})

		Convey("Do execute the query and fills a slice of pointers", func() {
			dummiesSlice := make([]*Dummy, 0)
			selectStmt := db.SelectFrom("dummies").
				Columns("id", "a_text", "another_text", "an_integer").
				OrderBy("an_integer")

			err := selectStmt.Do(&dummiesSlice)
			So(err, ShouldBeNil)
			So(len(dummiesSlice), ShouldEqual, 3)
			So(dummiesSlice[0].ID, ShouldBeGreaterThan, 0)
			So(dummiesSlice[0].AText, ShouldEqual, "First")
			So(dummiesSlice[0].AnotherText, ShouldEqual, "Premier")
			So(dummiesSlice[0].AnInteger, ShouldEqual, 11)
			So(dummiesSlice[1].AnInteger, ShouldEqual, 12)
			So(dummiesSlice[2].AnInteger, ShouldEqual, 13)

			Convey("Do compute time consumed by SQL query", func() {
				So(db.ConsumedTime(), ShouldBeGreaterThan, 0)
			})
		})

		Convey("Do returns sql.sql.ErrNoRows if a single instance if requested but not found", func() {
			dummy := Dummy{}
			err := db.SelectFrom("dummies").Columns("id").Where("1 > 2").Do(&dummy)
			So(err, ShouldEqual, sql.ErrNoRows)
		})

		Convey("Do fills nullable fields", func() {
			dummiesSlice := make([]Dummy, 0)
			selectStmt := db.SelectFrom("dummies").
				Columns("id", "a_nullable_string").
				OrderBy("an_integer")

			err := selectStmt.Do(&dummiesSlice)
			So(err, ShouldBeNil)
			So(len(dummiesSlice), ShouldEqual, 3)
			So(dummiesSlice[0].ANullableString.Valid, ShouldBeTrue)
			So(dummiesSlice[0].ANullableString.String, ShouldEqual, "Not empty")
			So(dummiesSlice[1].ANullableString.Valid, ShouldBeTrue)
			So(dummiesSlice[1].ANullableString.String, ShouldEqual, "")
			So(dummiesSlice[2].ANullableString.Valid, ShouldBeFalse)
		})

		Convey("ColumnsFromStruct with auto added all columns for single record", func() {
			singleDummy := Dummy{}
			selectStmt := db.SelectFrom("dummies").
				Where("an_integer = ?", 13)

			err := selectStmt.Do(&singleDummy)
			So(err, ShouldBeNil)
			So(singleDummy.ID, ShouldBeGreaterThan, 0)

			So(err, ShouldBeNil)
			So(len(selectStmt.columns), ShouldEqual, 6)
			So(singleDummy.AnInteger, ShouldEqual, 13)
		})

		Convey("ColumnAlias with auto added all columns for single record", func() {
			type DummyAlias struct {
				ID              int            `db:"id,key,auto"`
				AText           string         `db:"new_name"`
				AnotherText     string         `db:"another_text"`
				AnInteger       int            `db:"an_integer"`
				ANullableString sql.NullString `db:"a_nullable_string"`
				Version         int            `db:"version,oplock"`
			}
			singleDummy := DummyAlias{}
			selectStmt := db.SelectFrom("dummies").
				Where("an_integer = ?", 13)
			selectStmt.ColumnAlias("a_text", "new_name")
			selectStmt.OrderBy("new_name")
			err := selectStmt.Do(&singleDummy)
			So(err, ShouldBeNil)
			So(singleDummy.ID, ShouldBeGreaterThan, 0)

			So(err, ShouldBeNil)
			So(len(selectStmt.columns), ShouldEqual, 6)
			So(singleDummy.AnInteger, ShouldEqual, 13)
		})

		Convey("ColumnsFromStruct with auto added all columns for slice", func() {
			dummiesSlice := make([]*Dummy, 0)
			selectStmt := db.SelectFrom("dummies").
				OrderBy("an_integer")

			err := selectStmt.Do(&dummiesSlice)
			So(err, ShouldBeNil)
			So(len(selectStmt.columns), ShouldEqual, 6)
			So(len(dummiesSlice), ShouldEqual, 3)
			So(dummiesSlice[0].AnInteger, ShouldEqual, 11)
			So(dummiesSlice[1].AnInteger, ShouldEqual, 12)
			So(dummiesSlice[2].AnInteger, ShouldEqual, 13)
		})
	})
}

func TestCount(t *testing.T) {
	Convey("Given a test database", t, func() {
		db := fixturesSetup(t)
		defer db.Close()

		Convey("Count returns the count of row mathing the request", func() {
			selectStmt := db.SelectFrom("dummies")
			count, err := selectStmt.Count()
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 3)

			selectStmt = db.SelectFrom("dummies").Where("an_integer = ?", 12)
			count, err = selectStmt.Count()
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 1)

			Convey("Do compute time consumed by SQL query", func() {
				So(db.ConsumedTime(), ShouldBeGreaterThan, 0)
			})
		})
	})
}

func TestScanx(t *testing.T) {
	Convey("Given a test database", t, func() {
		db := fixturesSetup(t)
		defer db.Close()

		Convey("Scanx runs and fills destination columns", func() {
			selectStmt := db.SelectFrom("dummies")
			num := 0
			err := selectStmt.Where("an_integer = ?", 12).Columns("an_integer").Scanx(&num)
			So(err, ShouldBeNil)
			So(num, ShouldEqual, 12)

			Convey("Do compute time consumed by SQL query", func() {
				So(db.ConsumedTime(), ShouldBeGreaterThan, 0)
			})
		})
	})
}

func TestSelectDoWithIterator(t *testing.T) {
	Convey("Given a test database", t, func() {
		db := fixturesSetup(t)
		defer db.Close()

		Convey("DoWithIterator executes the query and returns an iterator", func() {

			selectStmt := db.SelectFrom("dummies").
				Columns("id", "a_text", "another_text", "an_integer").
				OrderBy("an_integer")

			iter, err := selectStmt.DoWithIterator()
			So(err, ShouldBeNil)
			defer iter.Close()

			Convey("Fetch data with Scan", func() {
				count := 0
				for iter.Next() {
					count++
					singleDummy := Dummy{}
					err := iter.Scan(&singleDummy)
					So(err, ShouldBeNil)

					if count == 1 {
						So(singleDummy.ID, ShouldBeGreaterThan, 0)
						So(singleDummy.AText, ShouldEqual, "First")
						So(singleDummy.AnotherText, ShouldEqual, "Premier")
						So(singleDummy.AnInteger, ShouldEqual, 11)
					}

				}
				So(count, ShouldEqual, 3)
				So(iter.Err(), ShouldBeNil)
			})

			Convey("Fetch data with Scanx", func() {
				count := 0
				for iter.Next() {
					var ID, anInteger int
					var aText, anotherText string
					err := iter.Scanx(&ID, &aText, &anotherText, &anInteger)
					So(err, ShouldBeNil)

					if count == 1 {
						So(ID, ShouldBeGreaterThan, 0)
						So(aText, ShouldEqual, "First")
						So(anotherText, ShouldEqual, "Premier")
						So(anInteger, ShouldEqual, 11)
					}

				}
				So(iter.Err(), ShouldBeNil)
			})
		})
	})
}

func TestSelectDoUsingStructForColumns(t *testing.T) {
	Convey("Given a test database", t, func() {
		db := fixturesSetup(t)
		defer db.Close()

		Convey("Do execute the query with two tables and fills a given instance", func() {
			fromTwoTables := make([]FromTwoTables, 0)

			selectStmt := db.SelectFrom("dummies").
				ColumnsFromStruct(&FromTwoTables{}).
				LeftJoin("relatedtodummies", "relatedtodummies", Q("relatedtodummies.dummies_id = dummies.id")).
				OrderBy("dummies.id")

			err := selectStmt.Do(&fromTwoTables)
			So(err, ShouldBeNil)
			So(len(fromTwoTables), ShouldEqual, 3)
			So(fromTwoTables[0].Dummy.AText, ShouldEqual, "First")
			So(fromTwoTables[0].RelatedToDummy.AText, ShouldEqual, "REL_First")
		})

	})
}
