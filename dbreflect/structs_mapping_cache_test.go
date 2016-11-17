package dbreflect

import (
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type StructToMap struct {
	ID   int    `db:"id"`
	Text string `db:"my_text"`
}

func TestStructsMappingCache(t *testing.T) {
	Convey("Given a struct to map and a cache", t, func() {
		typeStruct := reflect.TypeOf(StructToMap{})
		cache := NewStructsMappingCache()

		Convey("getStructMapping get back a *StructMapping (without error)", func() {
			sm1, err := cache.GetOrCreateStructMapping(typeStruct)
			So(sm1, ShouldNotBeNil)
			So(err, ShouldBeNil)

			Convey("getStructMapping always get back the same *StructMapping", func() {
				sm2, _ := cache.GetOrCreateStructMapping(typeStruct)
				So(sm2, ShouldEqual, sm1)
			})
		})
	})
}

func TestGlobalCache(t *testing.T) {
	Convey("The package initialize has a global cache", t, func() {
		So(Cache, ShouldNotBeNil)
		So(Cache, ShouldHaveSameTypeAs, &StructsMappingCache{})
	})
}
