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

LOOP:
	for i := 1; ; i++ {
		err := ks.FindNext(ModCPUStat)
		switch {
		case err == ErrNotFound:
			t.Logf("found %d cpu_stat modules", i)
			break LOOP
		case err != nil:
			t.Fatalf("failed to find next %q module: %s", ModCPUStat, err)
		}
	}
}

func findFirstCPUStat(t *testing.T, ks *KStat) {
	err := ks.Find(ModCPUStat)
	switch {
	case err == ErrNotFound:
		t.Fatalf("expected to find at least one cpu_stat module")
	case err != nil:
		t.Fatalf("unexpected error seen: %s", err)
	}

	i := ks.Instance()
	if i != 0 {
		t.Errorf("expected instance of first finding to be 0, got %d", i)
	}
}

func TestFindNext(t *testing.T) {
	ks, err := Open()
	if err != nil {
		t.Fatalf("failed to open kstat: %s", err)
	}
	defer func() {
		if err := ks.Close(); err != nil {
			t.Fatalf("failed to close kstat: %s", err)
		}
	}()

	// FindNext should properly initialize to the first element of the ks chain.
	switch err := ks.FindNext(ModCPUStat); {
	case err == ErrNotFound:
		t.Fatalf("expected to find at least one %q module, got none", ModCPUStat)
	case err != nil:
		t.Fatalf("failed to find next %q module: %s", ModCPUStat, err)
	}

	if iID := ks.Instance(); iID != 0 {
		t.Errorf("expected to find instance ID 0, got %d", iID)
	}

LOOP:
	for i := 1; ; i++ {
		switch err := ks.FindNext(ModCPUStat); {
		case err == ErrNotFound:
			t.Logf("found %d %q modules", i, ModCPUStat)
			break LOOP
		case err != nil:
			t.Fatalf("failed to find next %q module: %s", ModCPUStat, err)
		}

		if iID := ks.Instance(); iID != i {
			t.Errorf("expected to find instance ID %d, got %d", i, iID)
		}
	}

	// The next FindNext should again find the first (we're at the end of the
	// ks chain, FindNext should reinitialize to the beginning and restart the
	// search.
	switch err := ks.FindNext(ModCPUStat); {
	case err == ErrNotFound:
		t.Fatalf("expected to find at least one %q module, got none", ModCPUStat)
	case err != nil:
		t.Fatalf("failed to find next %q module: %s", ModCPUStat, err)
	}

	if iID := ks.Instance(); iID != 0 {
		t.Errorf("expected to find instance ID 0, got %d", iID)
	}
}

func TestRead(t *testing.T) {
	ks, err := Open()
	if err != nil {
		t.Fatalf("failed to open kstat: %s", err)
	}
	defer func() {
		if err := ks.Close(); err != nil {
			t.Fatalf("failed to close kstat: %s", err)
		}
	}()

	if err := ks.Read(nil); err == nil {
		t.Fatalf("expected to fail to read current kstat, succeeded", err)
	} else if err != ErrNotFound {
		t.Fatalf("expected NotFound error, got: %s", err)
	}

	// FindNext should properly initialize to the first element of the ks chain.
	switch err := ks.FindNext(ModCPUStat); {
	case err == ErrNotFound:
		t.Fatalf("expected to find at least one %q module, got none", ModCPUStat)
	case err != nil:
		t.Fatalf("failed to find next %q module: %s", ModCPUStat, err)
	}

	if err := ks.Read(nil); err != nil {
		t.Fatalf("failed to read current kstat, got: %s", err)
	}
}

func TestReadNamedWithNonNamedModule(t *testing.T) {
	ks, err := Open()
	if err != nil {
		t.Fatalf("failed to open kstat: %s", err)
	}
	defer func() {
		if err := ks.Close(); err != nil {
			t.Fatalf("failed to close kstat: %s", err)
		}
	}()

	if _, err := ks.DataLookup("foobar"); err == nil {
		t.Fatalf("expected to fail to read current kstat, succeeded", err)
	} else if err != ErrNotFound {
		t.Fatalf("expected NotFound error, got: %s", err)
	}

	// FindNext should properly initialize to the first element of the ks chain.
	switch err := ks.FindNext(ModCPUStat); {
	case err == ErrNotFound:
		t.Fatalf("expected to find at least one %q module, got none", ModCPUStat)
	case err != nil:
		t.Fatalf("failed to find next %q module: %s", ModCPUStat, err)
	}

	if _, err := ks.DataLookup("foobar"); err == nil {
		t.Fatalf("expected to fail to read current kstat, succeeded", err)
	} else if err != ErrNotNamedKStat {
		t.Fatalf("expected ErrNotNamedKStat error, got: %s", err)
	}
}
func TestReadNamed(t *testing.T) {
	ks, err := Open()
	if err != nil {
		t.Fatalf("failed to open kstat: %s", err)
	}
	defer func() {
		if err := ks.Close(); err != nil {
			t.Fatalf("failed to close kstat: %s", err)
		}
	}()

	if _, err := ks.DataLookup("foobar"); err == nil {
		t.Fatalf("expected to fail to read current kstat, succeeded", err)
	} else if err != ErrNotFound {
		t.Fatalf("expected NotFound error, got: %s", err)
	}

	switch err := ks.FindNext(ModCPUInfo); {
	case err == ErrNotFound:
		t.Fatalf("expected to find at least one %q module, got none", ModCPUStat)
	case err != nil:
		t.Fatalf("failed to find next %q module: %s", ModCPUStat, err)
	}

	// Lookup of inexistent value must fail with ErrNotFound.
	if _, err := ks.DataLookup("foobar"); err == nil {
		t.Fatalf("expected to fail to read current kstat, succeeded", err)
	} else if err != ErrNotFound {
		t.Fatalf("expected NotFound error, got: %s", err)
	}

	// Lookup of existent data must return a proper value (expected core_id to
	// be 0 for first cpu).
	lookupInt64(t, ks, "core_id", func(got int64) bool { return got == 0 })
	lookupInt64(t, ks, "clock_MHz", func(got int64) bool { return got > 0 })
	lookupString(t, ks, "cpu_type", func(got string) bool { return len(got) > 0 })
	lookupString(t, ks, "brand", func(got string) bool { return len(got) > 0 })
}

func lookupInt64(t *testing.T, ks *KStat, name string, test func(got int64) bool) {
	val, err := ks.DataLookup(name)
	if err != nil {
		t.Fatalf("didn't expect an error, got: %s", err)
	}
	v, ok := val.(int64)
	if !ok {
		t.Fatalf("expected %s to be int64, got %T", name, val)
	}
	if !test(v) {
		t.Fatalf("test for %s with value %d failed test", name, v)
	}
	t.Logf("%s = %d", name, v)
}

func lookupString(t *testing.T, ks *KStat, name string, test func(got string) bool) {
	val, err := ks.DataLookup(name)
	if err != nil {
		t.Fatalf("didn't expect an error, got: %s", err)
	}
	v, ok := val.(string)
	if !ok {
		t.Fatalf("expected %s to be string, got %T", name, val)
	}
	if !test(v) {
		t.Fatalf("test for %s with value %q failed test", name, v)
	}
	t.Logf("%s = %q", name, v)
}
