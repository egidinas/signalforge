package hdf5

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHDF5WriterSmoke(t *testing.T) {
	if _, err := os.Stat("/usr/lib/x86_64-linux-gnu/libhdf5_serial.so.103"); err != nil {
		t.Skipf("HDF5 runtime library not available: %v", err)
	}

	filename := filepath.Join(t.TempDir(), "smoke.h5")
	writer, err := NewHDF5Writer(filename)
	if err != nil {
		t.Fatalf("NewHDF5Writer failed: %v", err)
	}
	defer func() {
		if err := writer.Close(); err != nil {
			t.Errorf("Close failed: %v", err)
		}
	}()

	if err := writer.Create(); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	group, err := writer.CreateGroup("telemetry")
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}
	defer writer.CloseGroup(group)

	if err := writer.WriteFloat64Dataset(group, "temperatures", []float64{1.25, 2.5}); err != nil {
		t.Fatalf("WriteFloat64Dataset failed: %v", err)
	}
	if err := writer.WriteInt64Dataset(group, "timestamps", []int64{1000, 2000}); err != nil {
		t.Fatalf("WriteInt64Dataset failed: %v", err)
	}

	info, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("expected HDF5 output file: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("expected non-empty HDF5 output file")
	}
}
