package godb

import (
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type StructToMap struct {
	ID   int    `db:"id"`
	Text string `db:"my_text"`
}

func TestStructMappingGlobalCache(t *testing.T) {
	Convey("Given a struct to map", t, func() {
		typeStruct := reflect.TypeOf(StructToMap{})

		Convey("getStructMapping get back a *StructMapping (without error)", func() {
			sm1, err := getOrCreateStructMapping(typeStruct)
			So(sm1, ShouldNotBeNil)
			So(err, ShouldBeNil)

			Convey("getStructMapping always get back the same *StructMapping", func() {
				sm2, _ := getOrCreateStructMapping(typeStruct)
				So(sm2, ShouldEqual, sm1)
			})
		})
	})
}
