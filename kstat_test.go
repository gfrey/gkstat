// +build solaris,cgo

package gkstat

import "testing"

const ModCPUStat = "cpu_stat"
const ModCPUInfo = "cpu_info"

func TestFind(t *testing.T) {
	ks, err := Open()
	if err != nil {
		t.Fatalf("failed to open kstat: %s", err)
	}
	defer func() {
		if err := ks.Close(); err != nil {
			t.Fatalf("failed to close kstat: %s", err)
		}
	}()

	findFirstCPUStat(t, ks)
	// Run this a second time to verify the reset of pointer in ks chain works
	// properly.
	findFirstCPUStat(t, ks)
}

func findFirstCPUStat(t *testing.T, ks *KStat) {
	ksi, err := ks.Find(FilterModule(ModCPUStat))
	switch {
	case err == ErrNotFound:
		t.Fatalf("expected to find at least one cpu_stat module")
	case err != nil:
		t.Fatalf("unexpected error seen: %s", err)
	}

	i := ksi.Instance()
	if i != 0 {
		t.Errorf("expected instance of first finding to be 0, got %d", i)
	}
}

func TestScan(t *testing.T) {
	ks, err := Open()
	if err != nil {
		t.Fatalf("failed to open kstat: %s", err)
	}
	defer func() {
		if err := ks.Close(); err != nil {
			t.Fatalf("failed to close kstat: %s", err)
		}
	}()

	i := 0
	for ks := range ks.Scan(FilterModule(ModCPUStat)) {
		if iID := ks.Instance(); iID != i {
			t.Errorf("expected to find instance ID %d, got %d", i, iID)
		}
		i += 1
	}

	if i == 0 {
		t.Fatalf("didn't find any cpu stats")
	}
}
