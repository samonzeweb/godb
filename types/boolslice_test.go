package types

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBoolSlice(t *testing.T) {
	Convey("Given BoolSlice", t, func() {
		Convey("valid value", func() {
			v := []bool{true, false, true}
			boolSlice := BoolSlice(v)
			So(len(boolSlice), ShouldEqual, len(v))
			vx, err := boolSlice.Value()
			So(err, ShouldEqual, nil)
			So(string(vx.([]byte)), ShouldEqual, `[true,false,true]`)
		})

		Convey("nil value", func() {
			var boolSlice BoolSlice
			err := boolSlice.Scan(nil)
			So(err, ShouldEqual, nil)
			So(len(boolSlice), ShouldEqual, 0)
		})

		Convey("invalid value", func() {
			var boolSlice BoolSlice
			err := boolSlice.Scan("a")
			So(err, ShouldNotEqual, nil)
			So(len(boolSlice), ShouldEqual, 0)
		})

		Convey("parse null", func() {
			var boolSlice BoolSlice
			err := boolSlice.Scan([]byte("null"))
			So(err, ShouldEqual, nil)
			So(boolSlice, ShouldEqual, nil)
			So(len(boolSlice), ShouldEqual, 0)
		})

		Convey("parse from JS", func() {
			var boolSlice BoolSlice
			err := boolSlice.Scan([]byte(`[true,false,true]`))
			So(err, ShouldEqual, nil)
			So(len(boolSlice), ShouldEqual, 3)
		})
	})
}
