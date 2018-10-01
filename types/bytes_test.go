package types

import (
	"bytes"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNullBytes(t *testing.T) {
	Convey("Given NullBytes", t, func() {
		Convey("valid value", func() {
			val := []byte("Test bytes")
			nullBytes := ToNullBytes(val)
			So(nullBytes.Valid, ShouldEqual, true)
			So(bytes.Compare(nullBytes.Bytes, val), ShouldEqual, 0)
			v, err := nullBytes.Value()
			So(err, ShouldEqual, nil)
			So(bytes.Compare(v.([]byte), val), ShouldEqual, 0)
		})

		Convey("nil value", func() {
			nullBytes := ToNullBytes(nil)
			So(nullBytes.Valid, ShouldEqual, false)
			So(nullBytes.Bytes, ShouldEqual, nil)
		})
	})
}
