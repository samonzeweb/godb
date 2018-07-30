package types

import (
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var valNull = "null"

func TestSqlTypes(t *testing.T) {
	Convey("Given SqlTypes", t, func() {
		Convey("NullString", func() {
			valString := "foo"
			valJSON := fmt.Sprintf(`"%s"`, valString)
			n := NullStringFrom(valString)
			So(n.Valid, ShouldEqual, true)
			So(n.String, ShouldEqual, valString)

			// Marshal valid
			b, err := json.Marshal(n)
			So(err, ShouldEqual, nil)
			So(string(b), ShouldEqual, valJSON)

			// Marshal null
			n = NullStringFrom("")
			n.Valid = false
			b, err = json.Marshal(n)
			So(err, ShouldEqual, nil)
			So(string(b), ShouldEqual, valNull)

			// Unmarshal valid
			n = NullString{}
			err = json.Unmarshal([]byte(valJSON), &n)
			So(err, ShouldEqual, nil)
			So(n.String, ShouldEqual, valString)
			So(n.Valid, ShouldEqual, true)

			// Unmarshal null
			n = NullString{}
			err = json.Unmarshal([]byte(valNull), &n)
			So(err, ShouldEqual, nil)
			So(n.String, ShouldEqual, "")
			So(n.Valid, ShouldEqual, false)
		})

		Convey("NullFloat64", func() {
			valFloat := 99.9
			valJSON := fmt.Sprintf(`%.1f`, valFloat)
			n := NullFloat64From(valFloat)
			So(n.Valid, ShouldEqual, true)
			So(n.Float64, ShouldEqual, valFloat)

			// Marshal valid
			b, err := json.Marshal(n)
			So(err, ShouldEqual, nil)
			So(string(b), ShouldEqual, valJSON)

			// Unmarshal valid
			n = NullFloat64From(0)
			n.Valid = false
			b, err = json.Marshal(n)
			So(err, ShouldEqual, nil)
			So(string(b), ShouldEqual, valNull)

			// Unmarshal valid
			n = NullFloat64From(valFloat)
			err = json.Unmarshal([]byte(valJSON), &n)
			So(err, ShouldEqual, nil)
			So(n.Float64, ShouldEqual, valFloat)
			So(n.Valid, ShouldEqual, true)
			// Unmarshal null
			n = NullFloat64{}
			err = json.Unmarshal([]byte(valNull), &n)
			So(err, ShouldEqual, nil)
			So(n.Float64, ShouldEqual, 0)
			So(n.Valid, ShouldEqual, false)
		})

		Convey("NullBool", func() {
			valBool := true
			valJSON := fmt.Sprintf(`%v`, valBool)
			n := NullBoolFrom(valBool)
			So(n.Valid, ShouldEqual, true)
			So(n.Bool, ShouldEqual, valBool)

			// Marshal valid
			b, err := json.Marshal(n)
			So(err, ShouldEqual, nil)
			So(string(b), ShouldEqual, valJSON)

			// Marshal invalid
			n = NullBoolFrom(false)
			n.Valid = false
			b, err = json.Marshal(n)
			So(err, ShouldEqual, nil)
			So(string(b), ShouldEqual, valNull)

			// Unmarshal valid
			n = NullBoolFrom(false)
			err = json.Unmarshal([]byte(valJSON), &n)
			So(err, ShouldEqual, nil)
			So(n.Bool, ShouldEqual, valBool)
			So(n.Valid, ShouldEqual, true)

			// Unmarshal null
			n = NullBool{}
			err = json.Unmarshal([]byte(valNull), &n)
			So(err, ShouldEqual, nil)
			So(n.Bool, ShouldEqual, false)
			So(n.Valid, ShouldEqual, false)
		})

		Convey("NullInt64", func() {
			valInt64 := int64(99)
			valJSON := fmt.Sprintf(`%d`, valInt64)
			n := NullInt64From(valInt64)
			So(n.Valid, ShouldEqual, true)
			So(n.Int64, ShouldEqual, valInt64)

			// Marshal valid
			b, err := json.Marshal(n)
			So(err, ShouldEqual, nil)
			So(string(b), ShouldEqual, valJSON)

			// Marshal invalid
			n = NullInt64From(0)
			n.Valid = false
			b, err = json.Marshal(n)
			So(err, ShouldEqual, nil)
			So(string(b), ShouldEqual, valNull)

			// Unmarshal valid
			n = NullInt64From(0)
			err = json.Unmarshal([]byte(valJSON), &n)
			So(err, ShouldEqual, nil)
			So(n.Int64, ShouldEqual, valInt64)
			So(n.Valid, ShouldEqual, true)

			// Unmarshal invalid
			n = NullInt64{}
			err = json.Unmarshal([]byte(valNull), &n)
			So(err, ShouldEqual, nil)
			So(n.Int64, ShouldEqual, 0)
			So(n.Valid, ShouldEqual, false)
		})
	})
}
