// +build solaris,cgo

package gkstat

// #cgo LDFLAGS: -lkstat
// #include <stdlib.h>
// #include <kstat.h>
// #include <errno.h>
//
// char *lookupKsNamedString(kstat_named_t *ksn) {
//   return ksn->value.str.addr.ptr;
// }
import "C"

import (
	"fmt"
	"syscall"
	"unsafe"
)

type KStatItem struct {
	ks *C.kstat_t
}

func (ksi *KStatItem) Instance() int {
	return int(ksi.ks.ks_instance)
}

func (ksi *KStatItem) Module() string {
	return C.GoString((*C.char)(unsafe.Pointer(&ksi.ks.ks_module)))
}

func (ksi *KStatItem) Class() string {
	return C.GoString((*C.char)(unsafe.Pointer(&ksi.ks.ks_class)))
}

func (ksi *KStatItem) Name() string {
	return C.GoString((*C.char)(unsafe.Pointer(&ksi.ks.ks_name)))
}

// This is a nanosecond counter from some arbitrary point in time (not related
// to real time).
func (ksi *KStatItem) Snaptime() int {
	return int(ksi.ks.ks_instance)
}

func (ksi *KStatItem) lookup(name string) (*C.kstat_named_t, error) {
	if ksi.ks.ks_type != C.KSTAT_TYPE_NAMED {
		return nil, ErrNotNamedKStat
	}

	cs := C.CString(name)
	defer C.free(unsafe.Pointer(cs))
	raw, err := C.kstat_data_lookup(ksi.ks, cs)
	if raw == nil && err != nil {
		if e, ok := err.(syscall.Errno); ok && e == C.ENOENT {
			return nil, fmt.Errorf("Not found: %s", name)
		}
		return nil, err
	}

	return (*C.kstat_named_t)(unsafe.Pointer(raw)), nil
}

func (ksi *KStatItem) Int32(name string) (int32, error) {
	switch val, err := ksi.lookup(name); {
	case err != nil:
		return 0, err
	case val.data_type != C.KSTAT_DATA_INT32:
		return 0, fmt.Errorf("invalid data type (not int32) for %q: %s", name, dataTypeToName(val.data_type))
	default:
		return int32(*(*C.int32_t)(unsafe.Pointer(&val.value[0]))), nil
	}
}

func (ksi *KStatItem) UInt32(name string) (uint32, error) {
	switch val, err := ksi.lookup(name); {
	case err != nil:
		return 0, err
	case val.data_type != C.KSTAT_DATA_UINT32:
		return 0, fmt.Errorf("invalid data type (not uint32) for %q: %s", name, dataTypeToName(val.data_type))
	default:
		return uint32(*(*C.uint32_t)(unsafe.Pointer(&val.value[0]))), nil
	}
}

func (ksi *KStatItem) Int64(name string) (int64, error) {
	switch val, err := ksi.lookup(name); {
	case err != nil:
		return 0, err
	case val.data_type != C.KSTAT_DATA_INT64:
		return 0, fmt.Errorf("invalid data type (not int64) for %q: %s", name, dataTypeToName(val.data_type))
	default:
		return int64(*(*C.int64_t)(unsafe.Pointer(&val.value[0]))), nil
	}
}

func (ksi *KStatItem) UInt64(name string) (uint64, error) {
	switch val, err := ksi.lookup(name); {
	case err != nil:
		return 0, err
	case val.data_type != C.KSTAT_DATA_UINT64:
		return 0, fmt.Errorf("invalid data type (not uint64) for %q: %s", name, dataTypeToName(val.data_type))
	default:
		return uint64(*(*C.uint64_t)(unsafe.Pointer(&val.value[0]))), nil
	}
}

func (ksi *KStatItem) String(name string) (string, error) {
	switch val, err := ksi.lookup(name); {
	case err != nil:
		return "", err
	case val.data_type == C.KSTAT_DATA_CHAR:
		return C.GoString((*C.char)(unsafe.Pointer(&val.value[0]))), nil
	case val.data_type == C.KSTAT_DATA_STRING:
		return C.GoString(C.lookupKsNamedString(val)), nil
	default:
		return "", fmt.Errorf("invalid data type (not string) for %q: %s", name, dataTypeToName(val.data_type))
	}
}

func dataTypeToName(typ C.uchar_t) string {
	switch typ {
	case C.KSTAT_DATA_INT32:
		return "int32"
	case C.KSTAT_DATA_UINT32:
		return "uint32"
	case C.KSTAT_DATA_INT64:
		return "int64"
	case C.KSTAT_DATA_UINT64:
		return "uint64"
	case C.KSTAT_DATA_CHAR:
		return "string"
	case C.KSTAT_DATA_STRING:
		return "string"
	default:
		return fmt.Sprintf("unknown data type: %d", typ)
	}
}
