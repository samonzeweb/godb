package types

import (
	"testing"

	"fmt"

	. "github.com/smartystreets/goconvey/convey"
)

type CompactJSONStrTest struct {
	A string `json:"a" db:"a"`
	B int64  `json:"b" db:"b"`
}

func TestCompactJSONStr(t *testing.T) {
	Convey("Given CompactJSONStr", t, func() {
		Convey("valid value", func() {
			x := CompactJSONStr(`{"a": "test", "b": 2}`)
			v, err := x.Value()
			So(err, ShouldEqual, nil)
			So(v, ShouldNotEqual, nil)
			So(fmt.Sprintf("%s", v.([]byte)), ShouldEqual, `{"a":"test","b":2}`)
			err = (&x).Scan(v)
			So(err, ShouldEqual, nil)
			m := CompactJSONStrTest{}
			err = x.Unmarshal(&m)
			So(err, ShouldEqual, nil)
			So(m.A, ShouldEqual, "test")
			So(m.B, ShouldEqual, 2)
		})

		Convey("nil value", func() {
			x := NullCompactJSONStr{}
			err := x.Scan(`{"a": "test", "b": 2}`)
			So(err, ShouldEqual, nil)
			v, err := x.Value()
			So(err, ShouldEqual, nil)
			So(v, ShouldNotEqual, nil)
			So(fmt.Sprintf("%s", v.([]byte)), ShouldEqual, `{"a":"test","b":2}`)
			err = (&x).Scan(v)
			So(err, ShouldEqual, nil)
			m := CompactJSONStrTest{}
			err = x.Unmarshal(&m)
			So(err, ShouldEqual, nil)
			So(m.A, ShouldEqual, "test")
			So(m.B, ShouldEqual, 2)

			x = NullCompactJSONStr{}
			err = x.Scan(nil)
			So(err, ShouldEqual, nil)
			So(x.Valid, ShouldNotEqual, true)
		})
	})
}
