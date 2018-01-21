package godb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewSQLBuffer(t *testing.T) {
	Convey("Calling NewSQLBuffer creates the buffer using the given lengths", t, func() {
		b := NewSQLBuffer(10, 20)

		So(b.sql.Cap(), ShouldEqual, 10)
		So(cap(b.arguments), ShouldEqual, 20)
	})
}

func TestSQLBufferSQL(t *testing.T) {
	Convey("SQL returns the builded query string", t, func() {
		b := NewSQLBuffer(16, 4)
		sql := "I'M SQL"
		b.Write(sql)
		So(b.SQL(), ShouldEqual, sql)
	})
}

func TestSQLBufferArguments(t *testing.T) {
	Convey("Arguments returns the arguments from the builded query", t, func() {
		b := NewSQLBuffer(16, 4)
		b.Write("", 1, 2, 3)
		arguments := b.Arguments()
		So(len(arguments), ShouldEqual, 3)
		So(arguments[0], ShouldEqual, 1)
		So(arguments[1], ShouldEqual, 2)
		So(arguments[2], ShouldEqual, 3)
	})
}

func TestSQLBufferSQLLen(t *testing.T) {
	Convey("SQLLen returns the len of the sql buffer", t, func() {
		b := NewSQLBuffer(16, 4)
		b.Write("I'M SQL")
		So(b.SQLLen(), ShouldEqual, 7)
	})
}

func TestSQLBufferWrite(t *testing.T) {
	Convey("Write append string and arguments to the buffer", t, func() {
		b := NewSQLBuffer(16, 4)
		sql := "WHERE id = ?"
		b.Write(sql, 123)
		So(b.SQL(), ShouldEqual, sql)
		So(len(b.Arguments()), ShouldEqual, 1)
		So(b.Arguments()[0], ShouldEqual, 123)
	})
}

func TestSQLBufferWriteIfNotEmpty(t *testing.T) {
	Convey("WriteIfNotEmpty append string and arguments to the buffer only if the buffer isn't empty", t, func() {
		b := NewSQLBuffer(16, 4)
		b.WriteIfNotEmpty("something", 123)
		So(b.SQLLen(), ShouldEqual, 0)
		So(len(b.Arguments()), ShouldEqual, 0)

		b.Write("will not be empty")
		sql := "WHERE id = ?"
		b.WriteIfNotEmpty(sql, 123)
		So(b.SQL(), ShouldEndWith, sql)
		So(len(b.Arguments()), ShouldEqual, 1)
		So(b.Arguments()[0], ShouldEqual, 123)
	})
}

func TestSQLBufferWriteBytes(t *testing.T) {
	Convey("WriteBytes append bytes and arguments to the buffer", t, func() {
		b := NewSQLBuffer(16, 4)
		sql := "WHERE id = ?"
		sqlBytes := []byte(sql)
		b.WriteBytes(sqlBytes, 123)
		So(b.SQL(), ShouldEqual, sql)
		So(len(b.Arguments()), ShouldEqual, 1)
		So(b.Arguments()[0], ShouldEqual, 123)
	})
}

func TestSQLBufferWriteStrings(t *testing.T) {
	Convey("WriteStrings writes the givens strings, using a separator", t, func() {
		b := NewSQLBuffer(16, 4)
		b.WriteStrings(",", "A", "B", "C")
		So(b.SQL(), ShouldEqual, "A,B,C")
	})
}

func TestSQLBufferAppend(t *testing.T) {
	Convey("Append appends the given buffer to the current", t, func() {
		b := NewSQLBuffer(16, 4)
		b.Write("FIRST", 123)
		other := NewSQLBuffer(0, 0)
		other.Write("SECOND", 456)
		b.Append(other)
		So(b.SQL(), ShouldEqual, "FIRSTSECOND")
		So(len(b.Arguments()), ShouldEqual, 2)
	})
}

func TestSQLBufferWriteCondition(t *testing.T) {
	Convey("WriteCondition write the condition to the buffer", t, func() {
		b := NewSQLBuffer(16, 4)
		sql := "WHERE id = ?"
		q := Q(sql, 123)
		b.WriteCondition(q)
		So(b.SQL(), ShouldEqual, sql)
		So(len(b.Arguments()), ShouldEqual, 1)
		So(b.Arguments()[0], ShouldEqual, 123)
	})
}
