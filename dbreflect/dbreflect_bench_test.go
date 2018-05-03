package dbreflect

import (
	"reflect"
	"testing"
)

type Record struct {
	ID     int    `db:"id,key,auto"`
	Dummy1 string `db:"dummy1"`
	Dummy2 string `db:"dummy2"`
	Dummy3 string `db:"dummy3"`
	Dummy4 string `db:"dummy4"`
	Dummy5 string `db:"dummy5"`
}

// go test -bench=GetAllFieldsPointers -run=none ./dbreflect/ -benchtime=10s -cpuprofile cpu.out
// $GOPATH/bin/pprof -http=localhost:7777 cpu.out

// go test -bench=GetAllFieldsPointers -run=nore ./dbreflect/ -benchtime=10s -benchmem -memprofilerate=1 -memprofile mem.out
// $GOPATH/bin/pprof -http=localhost:7777 -alloc_space mem.out

// Escape analysis : go test -gcglags "-m"... (ou go build)
// gcglags informations : go tool compile -help

func BenchmarkGetAllFieldsPointers(b *testing.B) {
	r := Record{}
	sm, _ := NewStructMapping(reflect.TypeOf(r))

	for n := 0; n < b.N; n++ {
		_ = sm.GetAllFieldsPointers(&r)
	}
}
