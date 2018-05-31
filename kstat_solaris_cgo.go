// +build solaris,cgo

package gkstat

// #cgo LDFLAGS: -lkstat
// #include <sys/sysinfo.h>
// #include <kstat.h>
import "C"
import (
	"errors"
	"fmt"
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

func (k *KStat) Find(filters ...filter) (*KStatItem, error) {
	defer closeFilters(filters)

	ks := k.ks.kc_chain
	for {
		if ks == nil {
			return nil, ErrNotFound
		}

		if matchFilters(ks, filters) {
			return k.KStatItem(ks)
		}
		ks = ks.ks_next
	}
}

func (k *KStat) Scan(filters ...filter) <-chan *KStatItem {
	iC := make(chan *KStatItem)
	go func(iC chan<- *KStatItem, filters []filter) {
		defer closeFilters(filters)

		ks := k.ks.kc_chain
		for {
			if ks == nil {
				close(iC)
				return
			}

			if matchFilters(ks, filters) {
				if ksi, err := k.KStatItem(ks); err == nil {
					iC <- ksi
				} else {
					fmt.Printf("failed to create kstat item: %s\n", err)
				}
			}
			ks = ks.ks_next
		}
	}(iC, filters)
	return iC
}

func (k *KStat) KStatItem(ksi *C.kstat_t) (*KStatItem, error) {
	_, err := C.kstat_read(k.ks, ksi, nil)
	return &KStatItem{ks: ksi}, err
}
