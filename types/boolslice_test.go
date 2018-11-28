package types

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBoolSlice(t *testing.T) {
	Convey("Given BoolSlice", t, func() {
		Convey("valid value", func() {
			v := []bool{true, false, true}
			nullBoolSlice := BoolSlice(v)
			So(len(nullBoolSlice), ShouldEqual, len(v))
			vx, err := nullBoolSlice.Value()
			So(err, ShouldEqual, nil)
			So(string(vx.([]byte)), ShouldEqual, `[true,false,true]`)

		})

		Convey("nil value", func() {
			var nullBoolSlice BoolSlice
			err := nullBoolSlice.Scan(nil)
			So(err, ShouldEqual, nil)
			So(len(nullBoolSlice), ShouldEqual, 0)
		})

		Convey("invalid value", func() {
			var nullBoolSlice BoolSlice
			err := nullBoolSlice.Scan("a")
			So(err, ShouldNotEqual, nil)
			So(len(nullBoolSlice), ShouldEqual, 0)
		})

		Convey("parse null", func() {
			var nullBoolSlice BoolSlice
			err := nullBoolSlice.Scan([]byte("null"))
			So(err, ShouldEqual, nil)
			So(nullBoolSlice, ShouldEqual, nil)
			So(len(nullBoolSlice), ShouldEqual, 0)
		})

		Convey("parse from JS", func() {
			var nullBoolSlice BoolSlice
			err := nullBoolSlice.Scan([]byte(`[true,false,true]`))
			So(err, ShouldEqual, nil)
			So(len(nullBoolSlice), ShouldEqual, 3)
		})
	})
}
