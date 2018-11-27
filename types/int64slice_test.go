package types

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInt64Slice(t *testing.T) {
	Convey("Given Int64Slice", t, func() {
		Convey("valid value", func() {
			v := []int64{1, 2, 3}
			nullInt64Slice := Int64Slice(v)
			So(len(nullInt64Slice), ShouldEqual, len(v))
			vx, err := nullInt64Slice.Value()
			So(err, ShouldEqual, nil)
			So(string(vx.([]byte)), ShouldEqual, `[1,2,3]`)

		})

		Convey("nil value", func() {
			var nullInt64Slice Int64Slice
			err := nullInt64Slice.Scan(nil)
			So(err, ShouldEqual, nil)
			So(len(nullInt64Slice), ShouldEqual, 0)
		})

		Convey("invalid value", func() {
			var nullInt64Slice Int64Slice
			err := nullInt64Slice.Scan("a")
			So(err, ShouldNotEqual, nil)
			So(len(nullInt64Slice), ShouldEqual, 0)
		})

		Convey("parse null", func() {
			var nullInt64Slice Int64Slice
			err := nullInt64Slice.Scan([]byte("null"))
			So(err, ShouldEqual, nil)
			So(nullInt64Slice, ShouldEqual, nil)
			So(len(nullInt64Slice), ShouldEqual, 0)
		})

		Convey("parse from JS", func() {
			var nullInt64Slice Int64Slice
			err := nullInt64Slice.Scan([]byte(`[1,2,3]`))
			So(err, ShouldEqual, nil)
			So(len(nullInt64Slice), ShouldEqual, 3)
		})
	})
}
