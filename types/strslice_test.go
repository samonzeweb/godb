package types

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStrSlice(t *testing.T) {
	Convey("Given StrSlice", t, func() {
		Convey("valid value", func() {
			v := []string{"a", "b", "c"}
			strSlice := StrSlice(v)
			So(len(strSlice), ShouldEqual, len(v))
			vx, err := strSlice.Value()
			So(err, ShouldEqual, nil)
			So(string(vx.([]byte)), ShouldEqual, `["a","b","c"]`)
		})

		Convey("nil value", func() {
			var strSlice StrSlice
			err := strSlice.Scan(nil)
			So(err, ShouldEqual, nil)
			So(len(strSlice), ShouldEqual, 0)
		})

		Convey("invalid value", func() {
			var strSlice StrSlice
			err := strSlice.Scan("1")
			So(err, ShouldNotEqual, nil)
			So(len(strSlice), ShouldEqual, 0)
		})

		Convey("parse null", func() {
			var strSlice StrSlice
			err := strSlice.Scan([]byte("null"))
			So(err, ShouldEqual, nil)
			So(strSlice, ShouldEqual, nil)
			So(len(strSlice), ShouldEqual, 0)
		})

		Convey("parse from JS", func() {
			var strSlice StrSlice
			err := strSlice.Scan([]byte(`["Ankara","Tokyo","Paris"]`))
			So(err, ShouldEqual, nil)
			So(len(strSlice), ShouldEqual, 3)
		})
	})
}
