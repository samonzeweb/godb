package types

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMap(t *testing.T) {
	Convey("Given Map", t, func() {
		Convey("nil value", func() {
			var dbMap Map
			err := dbMap.Scan(nil)
			So(err, ShouldEqual, nil)
			So(len(dbMap), ShouldEqual, 0)
		})

		Convey("invalid value", func() {
			var dbMap Map
			err := dbMap.Scan("1")
			So(err, ShouldNotEqual, nil)
			So(len(dbMap), ShouldEqual, 0)
		})

		Convey("parse null", func() {
			var dbMap Map
			err := dbMap.Scan([]byte("null"))
			So(err, ShouldEqual, nil)
			So(dbMap, ShouldEqual, nil)
			So(len(dbMap), ShouldEqual, 0)
		})
		Convey("to JSON", func() {
			dbMap := Map{}
			dbMap["a"] = 1
			dbMap["b"] = "abc"
			dbMap["c"] = true
			v, err := dbMap.Value()
			So(err, ShouldEqual, nil)
			So(string(v.([]byte)), ShouldEqual, `{"a":1,"b":"abc","c":true}`)
		})
		Convey("parse from JSON", func() {
			var dbMap Map
			err := dbMap.Scan([]byte(`{"a":1,"b":"abc","c":true}`))
			So(err, ShouldEqual, nil)
			numVal, ok := dbMap["a"].(float64)
			So(ok, ShouldEqual, true)
			So(numVal, ShouldEqual, 1)

			strVal, ok := dbMap["b"].(string)
			So(ok, ShouldEqual, true)
			So(strVal, ShouldEqual, "abc")

			boolVal, ok := dbMap["c"].(bool)
			So(ok, ShouldEqual, true)
			So(boolVal, ShouldEqual, true)
		})
	})
}
