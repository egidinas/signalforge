package hdf5

import (
	"fmt"
	"unsafe"

	"github.com/ebitengine/purego"
)

type HDF5Writer struct {
	filename string
	lib      uintptr
	file     int64

	h5fcreate        uintptr
	h5fclose         uintptr
	h5gcreate2       uintptr
	h5gclose         uintptr
	h5screate_simple uintptr
	h5sclose         uintptr
	h5dcreate2       uintptr
	h5dwrite         uintptr
	h5dclose         uintptr
	memcpy           uintptr
}

func NewHDF5Writer(filename string) (*HDF5Writer, error) {
	libName := "/usr/lib/x86_64-linux-gnu/libhdf5_serial.so.103"
	lib, err := purego.Dlopen(libName, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %v", libName, err)
	}

	w := &HDF5Writer{
		filename: filename,
		lib:      lib,
	}

	if err := w.resolveFunctions(); err != nil {
		_ = purego.Dlclose(lib)
		return nil, err
	}

	return w, nil
}

func (w *HDF5Writer) resolveFunctions() error {
	symbols := []struct {
		name   string
		handle uintptr
		target *uintptr
	}{
		{name: "H5Fcreate", handle: w.lib, target: &w.h5fcreate},
		{name: "H5Fclose", handle: w.lib, target: &w.h5fclose},
		{name: "H5Gcreate2", handle: w.lib, target: &w.h5gcreate2},
		{name: "H5Gclose", handle: w.lib, target: &w.h5gclose},
		{name: "H5Screate_simple", handle: w.lib, target: &w.h5screate_simple},
		{name: "H5Sclose", handle: w.lib, target: &w.h5sclose},
		{name: "H5Dcreate2", handle: w.lib, target: &w.h5dcreate2},
		{name: "H5Dwrite", handle: w.lib, target: &w.h5dwrite},
		{name: "H5Dclose", handle: w.lib, target: &w.h5dclose},
		{name: "memcpy", handle: purego.RTLD_DEFAULT, target: &w.memcpy},
	}
	for _, symbol := range symbols {
		fn, err := purego.Dlsym(symbol.handle, symbol.name)
		if err != nil {
			return fmt.Errorf("failed to resolve %s: %w", symbol.name, err)
		}
		*symbol.target = fn
	}
	return nil
}

const (
	H5F_ACC_TRUNC = 0x0002
	H5P_DEFAULT   = 0
	H5S_ALL       = 0
)

func (w *HDF5Writer) Create() error {
	namePtr := uintptr(unsafe.Pointer(unsafe.SliceData([]byte(w.filename + "\x00"))))
	ret := w.call4(w.h5fcreate, namePtr, uintptr(H5F_ACC_TRUNC), H5P_DEFAULT, H5P_DEFAULT)
	if int64(ret) < 0 {
		return fmt.Errorf("H5Fcreate failed: %d", int64(ret))
	}
	w.file = int64(ret)
	return nil
}

func (w *HDF5Writer) Close() error {
	if w.file != 0 {
		w.call1(w.h5fclose, uintptr(w.file))
		w.file = 0
	}
	return purego.Dlclose(w.lib)
}

func (w *HDF5Writer) CreateGroup(name string) (int64, error) {
	namePtr := uintptr(unsafe.Pointer(unsafe.SliceData([]byte(name + "\x00"))))
	ret := w.call5(w.h5gcreate2, uintptr(w.file), namePtr, H5P_DEFAULT, H5P_DEFAULT, H5P_DEFAULT)
	if int64(ret) < 0 {
		return 0, fmt.Errorf("H5Gcreate2 failed: %d", int64(ret))
	}
	return int64(ret), nil
}

func (w *HDF5Writer) CloseGroup(gid int64) {
	w.call1(w.h5gclose, uintptr(gid))
}

func (w *HDF5Writer) WriteFloat64Dataset(parent int64, name string, data []float64) error {
	if len(data) == 0 {
		return nil
	}
	dims := []uint64{uint64(len(data))}
	space := w.call3(w.h5screate_simple, 1, uintptr(unsafe.Pointer(&dims[0])), 0)
	if int64(space) < 0 {
		return fmt.Errorf("H5Screate_simple failed")
	}
	defer w.call1(w.h5sclose, space)

	dtype, err := w.nativeDatatype("H5T_NATIVE_DOUBLE_g")
	if err != nil {
		return err
	}
	namePtr := uintptr(unsafe.Pointer(unsafe.SliceData([]byte(name + "\x00"))))

	dset := w.call7(w.h5dcreate2, uintptr(parent), namePtr, dtype, space, H5P_DEFAULT, H5P_DEFAULT, H5P_DEFAULT)
	if int64(dset) < 0 {
		return fmt.Errorf("H5Dcreate2 failed")
	}
	defer w.call1(w.h5dclose, dset)

	ret := w.call6(w.h5dwrite, dset, dtype, H5S_ALL, H5S_ALL, H5P_DEFAULT, uintptr(unsafe.Pointer(&data[0])))
	if int32(ret) < 0 {
		return fmt.Errorf("H5Dwrite failed")
	}
	return nil
}

func (w *HDF5Writer) WriteInt64Dataset(parent int64, name string, data []int64) error {
	if len(data) == 0 {
		return nil
	}
	dims := []uint64{uint64(len(data))}
	space := w.call3(w.h5screate_simple, 1, uintptr(unsafe.Pointer(&dims[0])), 0)
	if int64(space) < 0 {
		return fmt.Errorf("H5Screate_simple failed")
	}
	defer w.call1(w.h5sclose, space)

	dtype, err := w.nativeDatatype("H5T_NATIVE_INT64_g")
	if err != nil {
		return err
	}
	namePtr := uintptr(unsafe.Pointer(unsafe.SliceData([]byte(name + "\x00"))))

	dset := w.call7(w.h5dcreate2, uintptr(parent), namePtr, dtype, space, H5P_DEFAULT, H5P_DEFAULT, H5P_DEFAULT)
	if int64(dset) < 0 {
		return fmt.Errorf("H5Dcreate2 failed")
	}
	defer w.call1(w.h5dclose, dset)

	ret := w.call6(w.h5dwrite, dset, dtype, H5S_ALL, H5S_ALL, H5P_DEFAULT, uintptr(unsafe.Pointer(&data[0])))
	if int32(ret) < 0 {
		return fmt.Errorf("H5Dwrite failed")
	}
	return nil
}

func (w *HDF5Writer) call1(fn uintptr, a1 uintptr) uintptr {
	ret, _, _ := purego.SyscallN(fn, a1)
	return ret
}

func (w *HDF5Writer) nativeDatatype(symbol string) (uintptr, error) {
	sym, err := purego.Dlsym(w.lib, symbol)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve %s: %w", symbol, err)
	}

	var dtype int64
	ret := w.call3(w.memcpy, uintptr(unsafe.Pointer(&dtype)), sym, unsafe.Sizeof(dtype))
	if ret == 0 {
		return 0, fmt.Errorf("failed to read %s", symbol)
	}
	if dtype < 0 {
		return 0, fmt.Errorf("invalid datatype %s: %d", symbol, dtype)
	}
	return uintptr(dtype), nil
}

func (w *HDF5Writer) call3(fn uintptr, a1, a2, a3 uintptr) uintptr {
	ret, _, _ := purego.SyscallN(fn, a1, a2, a3)
	return ret
}

func (w *HDF5Writer) call4(fn uintptr, a1, a2, a3, a4 uintptr) uintptr {
	ret, _, _ := purego.SyscallN(fn, a1, a2, a3, a4)
	return ret
}

func (w *HDF5Writer) call5(fn uintptr, a1, a2, a3, a4, a5 uintptr) uintptr {
	ret, _, _ := purego.SyscallN(fn, a1, a2, a3, a4, a5)
	return ret
}

func (w *HDF5Writer) call6(fn uintptr, a1, a2, a3, a4, a5, a6 uintptr) uintptr {
	ret, _, _ := purego.SyscallN(fn, a1, a2, a3, a4, a5, a6)
	return ret
}

func (w *HDF5Writer) call7(fn uintptr, a1, a2, a3, a4, a5, a6, a7 uintptr) uintptr {
	ret, _, _ := purego.SyscallN(fn, a1, a2, a3, a4, a5, a6, a7)
	return ret
}
