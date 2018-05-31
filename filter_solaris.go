// +build solaris,cgo

package gkstat

// #cgo LDFLAGS: -lkstat
// #include <kstat.h>
// #include <stdlib.h>
// #include <strings.h>
import "C"
import (
	"unsafe"
)

// Lookup resets the chain pointer to the chain's first element and starts
// searching.
type filter interface {
	Do(ks *C.kstat_t) bool
	Close() error
}

func closeFilters(filters []filter) {
	for i := range filters {
		_ = filters[i].Close()
	}
}

func matchFilters(ks *C.kstat_t, filters []filter) bool {
	for i := range filters {
		if !filters[i].Do(ks) {
			return false
		}
	}
	return true
}

type classFilter struct {
	cs *C.char
}

func FilterClass(class string) filter {
	return &classFilter{cs: C.CString(class)}
}

func (f *classFilter) Do(ks *C.kstat_t) bool {
	return C.strncmp(&ks.ks_class[0], f.cs, C.KSTAT_STRLEN) == C.int(0)
}

func (f *classFilter) Close() error {
	_, err := C.free(unsafe.Pointer(f.cs))
	return err
}

type moduleFilter struct {
	cs *C.char
}

func FilterModule(module string) filter {
	return &moduleFilter{cs: C.CString(module)}
}

func (f *moduleFilter) Do(ks *C.kstat_t) bool {
	return C.strncmp(&ks.ks_module[0], f.cs, C.KSTAT_STRLEN) == C.int(0)
}

func (f *moduleFilter) Close() error {
	_, err := C.free(unsafe.Pointer(f.cs))
	return err
}

type nameFilter struct {
	cs *C.char
}

func FilterName(name string) filter {
	return &nameFilter{cs: C.CString(name)}
}

func (f *nameFilter) Do(ks *C.kstat_t) bool {
	return C.strncmp(&ks.ks_name[0], f.cs, C.KSTAT_STRLEN) == C.int(0)
}

func (f *nameFilter) Close() error {
	_, err := C.free(unsafe.Pointer(f.cs))
	return err
}

type instanceFilter struct {
	id C.int
}

func FilterInstance(id int) filter {
	return &instanceFilter{id: C.int(id)}
}

func (f *instanceFilter) Do(ks *C.kstat_t) bool {
	return ks.ks_instance == f.id
}

func (f *instanceFilter) Close() error {
	return nil
}
