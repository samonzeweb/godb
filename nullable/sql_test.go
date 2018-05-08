
package nullable
import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)


func TestSqlTypes(t *testing.T) {
	Convey("Given SqlTypes", t, func() {
		Convey("NullStringFrom", func() {
			v := "foo"
			n := NullStringFrom(v)
			So(n.Valid, ShouldEqual, true)
			So(n.String, ShouldEqual, v)
		})
		Convey("NullFloat64From", func() {
			v := 99.9
			n := NullFloat64From(v)
			So(n.Valid, ShouldEqual, true)
			So(n.Float64, ShouldEqual, v)
		})
		Convey("NullBoolFrom", func() {
			v := true
			n := NullBoolFrom(v)
			So(n.Valid, ShouldEqual, true)
			So(n.Bool, ShouldEqual, v)
		})
		Convey("NullInt64From", func() {
			v := int64(99)
			n := NullInt64From(v)
			So(n.Valid, ShouldEqual, true)
			So(n.Int64, ShouldEqual, v)
		})
	})
}
