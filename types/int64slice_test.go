package types

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInt64Slice(t *testing.T) {
	Convey("Given Int64Slice", t, func() {
		Convey("valid value", func() {
			v := []int64{1, 2, 3}
			int64Slice := Int64Slice(v)
			So(len(int64Slice), ShouldEqual, len(v))
			vx, err := int64Slice.Value()
			So(err, ShouldEqual, nil)
			So(string(vx.([]byte)), ShouldEqual, `[1,2,3]`)
		})

		Convey("nil value", func() {
			var int64Slice Int64Slice
			err := int64Slice.Scan(nil)
			So(err, ShouldEqual, nil)
			So(len(int64Slice), ShouldEqual, 0)
		})

		Convey("invalid value", func() {
			var int64Slice Int64Slice
			err := int64Slice.Scan("a")
			So(err, ShouldNotEqual, nil)
			So(len(int64Slice), ShouldEqual, 0)
		})

		Convey("parse null", func() {
			var int64Slice Int64Slice
			err := int64Slice.Scan([]byte("null"))
			So(err, ShouldEqual, nil)
			So(int64Slice, ShouldEqual, nil)
			So(len(int64Slice), ShouldEqual, 0)
		})

		Convey("parse from JS", func() {
			var int64Slice Int64Slice
			err := int64Slice.Scan([]byte(`[1,2,3]`))
			So(err, ShouldEqual, nil)
			So(len(int64Slice), ShouldEqual, 3)
		})
	})
}
