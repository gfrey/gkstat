// +build solaris,cgo

package gkstat

// #cgo LDFLAGS: -lkstat
// #include <sys/sysinfo.h>
// #include <strings.h>
// #include <kstat.h>
// #include <errno.h>
//
// char *lookupKsNamedString(kstat_named_t *ksn) {
//   return ksn->value.str.addr.ptr;
// }
import "C"
import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

var ErrNotFound = errors.New("Not found")
var ErrNotNamedKStat = errors.New("Not a named kstat")

type KStat struct {
	ks       *C.kstat_ctl_t
	ksp      *C.kstat_t
	readData bool
}

// Open kstat for reading.
func Open() (*KStat, error) {
	ks, err := C.kstat_open()
	if ks == nil && err != nil {
		return nil, err
	}
	return &KStat{ks: ks}, nil
}

// Closes the kstat.
func (k *KStat) Close() error {
	_, err := C.kstat_close(k.ks)
	return err
}

// Lookup resets the chain pointer to the chain's first element and starts
// searching.
func (k *KStat) Lookup(module, name string, instance int) error {
	k.readData = false
	k.ksp = k.ks.kc_chain

	for {
		if k.ksp == nil {
			return ErrNotFound
		}

		if instance == -1 || k.ksp.ks_instance == C.int(instance) {
			if module == "" || C.strncmp(&k.ksp.ks_module[0], C.CString(module), C.KSTAT_STRLEN) == C.int(0) {
				if name == "" || C.strncmp(&k.ksp.ks_name[0], C.CString(name), C.KSTAT_STRLEN) == C.int(0) {
					return nil
				}
			}
		}

		k.ksp = k.ksp.ks_next
	}
}

// Reads the found module's data into the pointer given.
func (k *KStat) Read(p unsafe.Pointer) error {
	if k.ksp == nil {
		return ErrNotFound
	}

	_, err := C.kstat_read(k.ks, k.ksp, p)
	k.readData = true
	return err
}

// Returns the instance id of the found module.
func (k *KStat) Instance() int {
	if k.ksp == nil {
		return -1
	}
	return int(k.ksp.ks_instance)
}

// Returns point in time the snapshort was taken, -1 if no kstat was found yet.
// This is a nanosecond counter from some arbitrary point in time (not related
// to real time).
func (k *KStat) Snaptime() int64 {
	if k.ksp == nil {
		return -1
	}
	return int64(k.ksp.ks_snaptime)
}
func (k *KStat) dataLookup(name string) (*C.kstat_named_t, error) {
	if k.ksp == nil {
		return nil, ErrNotFound
	}

	if k.ksp.ks_type != C.KSTAT_TYPE_NAMED {
		return nil, ErrNotNamedKStat
	}

	if !k.readData {
		if err := k.Read(nil); err != nil {
			return nil, err
		}
	}

	raw, err := C.kstat_data_lookup(k.ksp, C.CString(name))
	if raw == nil && err != nil {
		if e, ok := err.(syscall.Errno); ok && e == C.ENOENT {
			return nil, fmt.Errorf("Not found: %s", name)
		}
		return nil, err
	}

	return (*C.kstat_named_t)(unsafe.Pointer(raw)), nil
}

// On named kstat will lookup the given key and return the approprate value.
func (k *KStat) DataLookup(name string) (interface{}, error) {
	val, err := k.dataLookup(name)
	if err != nil {
		return nil, err
	}

	switch val.data_type {
	case C.KSTAT_DATA_INT32:
		return int32(*(*C.int32_t)(unsafe.Pointer(&val.value[0]))), nil
	case C.KSTAT_DATA_UINT32:
		return uint32(*(*C.uint32_t)(unsafe.Pointer(&val.value[0]))), nil
	case C.KSTAT_DATA_INT64:
		return int64(*(*C.int64_t)(unsafe.Pointer(&val.value[0]))), nil
	case C.KSTAT_DATA_UINT64:
		return uint64(*(*C.uint64_t)(unsafe.Pointer(&val.value[0]))), nil
	case C.KSTAT_DATA_CHAR:
		return C.GoString((*C.char)(unsafe.Pointer(&val.value[0]))), nil
	case C.KSTAT_DATA_STRING:
		return C.GoString(C.lookupKsNamedString(val)), nil
	default:
		return nil, fmt.Errorf("unexpected data type: %d", val.data_type)
	}
}

func (k *KStat) DataLookupInt32(name string) (int32, error) {
	switch val, err := k.dataLookup(name); {
	case err != nil:
		return 0, err
	case val.data_type != C.KSTAT_DATA_INT32:
		return 0, fmt.Errorf("invalid data type (not int32) for %q: %s", name, dataTypeToName(val.data_type))
	default:
		return int32(*(*C.int32_t)(unsafe.Pointer(&val.value[0]))), nil
	}
}
func (k *KStat) DataLookupUInt32(name string) (uint32, error) {
	switch val, err := k.dataLookup(name); {
	case err != nil:
		return 0, err
	case val.data_type != C.KSTAT_DATA_UINT32:
		return 0, fmt.Errorf("invalid data type (not uint32) for %q: %s", name, dataTypeToName(val.data_type))
	default:
		return uint32(*(*C.uint32_t)(unsafe.Pointer(&val.value[0]))), nil
	}
}

func (k *KStat) DataLookupInt64(name string) (int64, error) {
	switch val, err := k.dataLookup(name); {
	case err != nil:
		return 0, err
	case val.data_type != C.KSTAT_DATA_INT64:
		return 0, fmt.Errorf("invalid data type (not int64) for %q: %s", name, dataTypeToName(val.data_type))
	default:
		return int64(*(*C.int64_t)(unsafe.Pointer(&val.value[0]))), nil
	}
}

func (k *KStat) DataLookupUInt64(name string) (uint64, error) {
	switch val, err := k.dataLookup(name); {
	case err != nil:
		return 0, err
	case val.data_type != C.KSTAT_DATA_UINT64:
		return 0, fmt.Errorf("invalid data type (not uint64) for %q: %s", name, dataTypeToName(val.data_type))
	default:
		return uint64(*(*C.uint64_t)(unsafe.Pointer(&val.value[0]))), nil
	}
}

func (k *KStat) DataLookupString(name string) (string, error) {
	switch val, err := k.dataLookup(name); {
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
