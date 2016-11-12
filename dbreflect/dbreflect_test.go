package dbreflect

import (
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type simpleStruct struct {
	ID    int    `db:"id,key,auto"`
	Text  string `db:"my_text"`
	Other string
}

func TestStructMapping(t *testing.T) {
	Convey("NewStructMapping with a struct type", t, func() {
		structMap, _ := NewStructMapping(reflect.TypeOf(simpleStruct{}))

		Convey("It stores the struct name ", func() {
			So(structMap.Name, ShouldEndWith, "simpleStruct")
		})

		Convey("It store data about all tagged fields ", func() {
			So(len(structMap.GetAllColumnsNames()), ShouldEqual, 2)
			So(structMap.GetAllColumnsNames(), ShouldContain, "id")
			So(structMap.GetAllColumnsNames(), ShouldContain, "my_text")
		})
	})
}

func TestNewStructMappingErrors(t *testing.T) {
	Convey("Calling NewStructMapping without a struct ", t, func() {
		dummy := true
		_, err := NewStructMapping(reflect.TypeOf(dummy))

		Convey("NewStructMapping return an error", func() {
			So(err, ShouldNotBeNil)
		})
	})
}
