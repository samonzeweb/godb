package types

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFloat64Slice(t *testing.T) {
	Convey("Given Float64Slice", t, func() {
		Convey("valid value", func() {
			v := []float64{1.1, 2.2, 3.3}
			float64Slice := Float64Slice(v)
			So(len(float64Slice), ShouldEqual, len(v))
			vx, err := float64Slice.Value()
			So(err, ShouldEqual, nil)
			So(string(vx.([]byte)), ShouldEqual, `[1.1,2.2,3.3]`)
		})

		Convey("nil value", func() {
			var float64Slice Float64Slice
			err := float64Slice.Scan(nil)
			So(err, ShouldEqual, nil)
			So(len(float64Slice), ShouldEqual, 0)
		})

		Convey("invalid value", func() {
			var float64Slice Float64Slice
			err := float64Slice.Scan("a")
			So(err, ShouldNotEqual, nil)
			So(len(float64Slice), ShouldEqual, 0)
		})

		Convey("parse null", func() {
			var float64Slice Float64Slice
			err := float64Slice.Scan([]byte("null"))
			So(err, ShouldEqual, nil)
			So(float64Slice, ShouldEqual, nil)
			So(len(float64Slice), ShouldEqual, 0)
		})

		Convey("parse from JS", func() {
			var float64Slice Float64Slice
			err := float64Slice.Scan([]byte(`[1.1,2.2,3.3]`))
			So(err, ShouldEqual, nil)
			So(len(float64Slice), ShouldEqual, 3)
		})
	})
}
