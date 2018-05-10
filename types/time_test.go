package types

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"time"
)

func TestNullTime(t *testing.T) {
	Convey("Given NullTime", t, func() {
		Convey("valid value", func() {
			now := time.Now()
			nullTime := NullTimeFrom(now)
			So(nullTime.Valid, ShouldEqual, true)
			So(nullTime.Time, ShouldEqual, now)
			v, err := nullTime.Value()
			So(err, ShouldEqual, nil)
			So(v, ShouldEqual, now)

		})

		Convey("nil value", func() {
			var nullTime NullTime
			err := nullTime.Scan(nil)
			So(err, ShouldNotEqual, nil)
			So(nullTime.Valid, ShouldEqual, false)
			So(nullTime.Time.Second(), ShouldEqual, 0)
		})

		Convey("invalid value", func() {
			var nullTime NullTime
			err := nullTime.Scan(int64(1))
			So(err, ShouldNotEqual, nil)
			So(nullTime.Valid, ShouldEqual, false)
			So(nullTime.Time.Second(), ShouldEqual, 0)
		})
	})
}
