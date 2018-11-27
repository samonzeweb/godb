package types

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStrSlice(t *testing.T) {
	Convey("Given StrSlice", t, func() {
		Convey("valid value", func() {
			v := []string{"a", "b", "c"}
			nullStrSlice := StrSlice(v)
			So(len(nullStrSlice), ShouldEqual, len(v))
			vx, err := nullStrSlice.Value()
			So(err, ShouldEqual, nil)
			So(string(vx.([]byte)), ShouldEqual, `["a","b","c"]`)

		})

		Convey("nil value", func() {
			var nullStrSlice StrSlice
			err := nullStrSlice.Scan(nil)
			So(err, ShouldEqual, nil)
			So(len(nullStrSlice), ShouldEqual, 0)
		})

		Convey("invalid value", func() {
			var nullStrSlice StrSlice
			err := nullStrSlice.Scan("1")
			So(err, ShouldNotEqual, nil)
			So(len(nullStrSlice), ShouldEqual, 0)
		})

		Convey("parse null", func() {
			var nullStrSlice StrSlice
			err := nullStrSlice.Scan([]byte("null"))
			So(err, ShouldEqual, nil)
			So(nullStrSlice, ShouldEqual, nil)
			So(len(nullStrSlice), ShouldEqual, 0)
		})

		Convey("parse from JS", func() {
			var nullStrSlice StrSlice
			err := nullStrSlice.Scan([]byte(`["Ankara","Tokyo","Paris"]`))
			So(err, ShouldEqual, nil)
			So(len(nullStrSlice), ShouldEqual, 3)
		})
	})
}
