package tablenamer

import (
	"testing"

	"sync"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSnakeCaseCacheInit(t *testing.T) {
	Convey("Snake case global cache", t, func() {
		So(sCache, ShouldNotBeNil)
		So(sCache, ShouldHaveSameTypeAs, sync.Map{})
	})
}

func TestTableNamer(t *testing.T) {
	Convey("Given English", t, func() {
		Convey("ToSnakeCase", func() {
			So(ToSnakeCase("AuthorBook"), ShouldEqual, "author_book")
			So(ToSnakeCase("Author"), ShouldEqual, "author")
			So(ToSnakeCase("author"), ShouldEqual, "author")
			So(ToSnakeCase("AuthorID"), ShouldEqual, "author_id")
			So(ToSnakeCase("AuthorIDBook"), ShouldEqual, "author_id_book")
		})
		snp := SnakePlural()
		Convey("SnakePlural()", func() {
			So(snp("AuthorBook", false), ShouldEqual, "author_books")
			So(snp("AuthorBook", true), ShouldEqual, "AuthorBook")
			So(snp("Author", false), ShouldEqual, "authors")
			So(snp("author", false), ShouldEqual, "authors")
			So(snp("AuthorID", false), ShouldEqual, "author_ids")
			So(snp("AuthorIDBook", false), ShouldEqual, "author_id_books")
		})
		sn := Snake()
		Convey("Snake()", func() {
			So(sn("AuthorBook", false), ShouldEqual, "author_book")
			So(sn("AuthorBook", true), ShouldEqual, "AuthorBook")
			So(sn("Author", false), ShouldEqual, "author")
			So(sn("author", false), ShouldEqual, "author")
			So(sn("AuthorID", false), ShouldEqual, "author_id")
			So(sn("AuthorIDBook", false), ShouldEqual, "author_id_book")
		})

		s := Same()
		Convey("Same()", func() {
			So(s("AuthorBook", false), ShouldEqual, "AuthorBook")
			So(s("AuthorBook", true), ShouldEqual, "AuthorBook")
			So(s("Author", false), ShouldEqual, "Author")
			So(s("author", false), ShouldEqual, "author")
			So(s("AuthorID", false), ShouldEqual, "AuthorID")
			So(s("AuthorIDBook", false), ShouldEqual, "AuthorIDBook")
		})

		p := Plural()
		Convey("Plural()", func() {
			So(p("AuthorBook", false), ShouldEqual, "AuthorBooks")
			So(p("AuthorBook", true), ShouldEqual, "AuthorBook")
			So(p("Author", false), ShouldEqual, "Authors")
			So(p("author", false), ShouldEqual, "authors")
			So(p("AuthorID", false), ShouldEqual, "AuthorIDs")
			So(p("AuthorIDBook", false), ShouldEqual, "AuthorIDBooks")
		})

		Convey("Plural() caching", func() {
			So(p("AuthorBook", false), ShouldEqual, "AuthorBooks")
			So(p("AuthorBook", true), ShouldEqual, "AuthorBook")
			So(p("Author", false), ShouldEqual, "Authors")
			So(p("author", false), ShouldEqual, "authors")
			So(p("AuthorID", false), ShouldEqual, "AuthorIDs")
			So(p("AuthorIDBook", false), ShouldEqual, "AuthorIDBooks")
		})
	})
}
