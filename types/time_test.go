package types

import (
	"testing"

	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNullTime(t *testing.T) {
	Convey("Given NullTime", t, func() {
		Convey("valid value", func() {
			now := time.Now()
			nullTime := ToNullTime(now)
			So(nullTime.Valid, ShouldEqual, true)
			So(nullTime.Time, ShouldEqual, now)
			v, err := nullTime.Value()
			So(err, ShouldEqual, nil)
			So(v, ShouldEqual, now)

		})

		Convey("nil value", func() {
			var nullTime NullTime
			err := nullTime.Scan(nil)
			So(err, ShouldEqual, nil)
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

		Convey("parse null", func() {
			var nullTime NullTime
			err := nullTime.UnmarshalJSON([]byte("null"))
			So(err, ShouldEqual, nil)
			So(nullTime.Valid, ShouldEqual, false)
			So(nullTime.Time.Second(), ShouldEqual, 0)
		})

		Convey("parse from JS", func() {
			var nullTime NullTime
			err := nullTime.UnmarshalJSON([]byte("2018-07-27T09:23:35.347Z"))
			So(err, ShouldEqual, nil)
			So(nullTime.Valid, ShouldEqual, true)
			So(nullTime.Time.Second(), ShouldNotEqual, 0)
		})

		Convey("serialize to JSON", func() {
			now := time.Now()
			nullTime := ToNullTime(now)
			b, err := nullTime.MarshalJSON()
			tb, _ := nullTime.Time.MarshalJSON()
			So(err, ShouldEqual, nil)
			So(string(tb), ShouldEqual, string(b))
			nullTime.Valid = false
			tb, _ = nullTime.MarshalJSON()
			So(string(tb), ShouldEqual, string("null"))
		})
	})
}
