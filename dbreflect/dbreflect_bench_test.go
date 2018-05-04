package dbreflect

import (
	"reflect"
	"testing"
)

// Record will be used to bench dbreflect (no meaning by itself)
type Record struct {
	ID     int    `db:"id,key,auto"`
	Dummy1 string `db:"dummy1"`
	Dummy2 string `db:"dummy2"`
	Dummy3 string `db:"dummy3"`
	Dummy4 string `db:"dummy4"`
	Dummy5 string `db:"dummy5"`
}

// Tips to benchmark :
//
// Do a simple bencmark, without profiling :
//   go test -bench=GetAllFieldsPointers -run=^$ ./dbreflect/ -benchmem -benchtime=10s
//
// CPU profiling, and analysis using a browser :
//   go test -bench=GetAllFieldsPointers -run=^$ ./dbreflect/ -benchtime=10s -cpuprofile cpu.out
//   $GOPATH/bin/pprof -http=localhost:7777 cpu.out
//
// Memory profiling, and analysis using a browser :
//   go test -bench=GetAllFieldsPointers -run=^$ ./dbreflect/ -benchtime=10s -benchmem -memprofilerate=1 -memprofile mem.out
//   $GOPATH/bin/pprof -http=localhost:7777 -alloc_space mem.out
//
// Escape analysis : go test -gcflags="-m -m"...
//   go test -gcflags="-m -m" ./dbreflect 2>&1 | grep dbreflect.go
//
// Note : profiling is done here with latest pprof.
// Install it with :
//   go get -u github.com/google/pprof

func BenchmarkGetAllFieldsPointers(b *testing.B) {
	r := Record{}
	sm, _ := NewStructMapping(reflect.TypeOf(r))

	for n := 0; n < b.N; n++ {
		_ = sm.GetAllFieldsPointers(&r)
	}
}
