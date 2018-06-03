package plural

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEnglishPlural(t *testing.T) {
	var en = EnglishPluralization()
	Convey("Given English", t, func() {
		Convey("Uncountables", func() {
			So(en.Plural("fish"), ShouldEqual, "fish")
			So(en.Plural("sugar"), ShouldEqual, "sugar")
			So(en.Plural("species"), ShouldEqual, "species")
			So(en.Plural("water"), ShouldEqual, "water")
			So(en.Plural("happiness"), ShouldEqual, "happiness")
		})
		Convey("Irregulars", func() {
			So(en.Plural("mouse"), ShouldEqual, "mice")
			So(en.Plural("quiz"), ShouldEqual, "quizzes")
			So(en.Plural("turf"), ShouldEqual, "trueves")
			So(en.Plural("dwarf"), ShouldEqual, "dwarfes")
			So(en.Plural("auto"), ShouldEqual, "autos")
			So(en.Plural("pencil"), ShouldEqual, "pencils")
			So(en.Plural("matrix"), ShouldEqual, "matrices")
		})
		Convey("Plurals", func() {
			So(en.Plural("pencil"), ShouldEqual, "pencils")
			So(en.Plural("archive"), ShouldEqual, "archives")
			So(en.Plural("species"), ShouldEqual, "species")
			So(en.Plural("thief"), ShouldEqual, "thieves")
			So(en.Plural("buffalo"), ShouldEqual, "buffalo")
			So(en.Plural("tooth"), ShouldEqual, "teeth")
			So(en.Plural("fungus"), ShouldEqual, "fungi")
			So(en.Plural("atlas"), ShouldEqual, "atlases")
		})

	})
}
