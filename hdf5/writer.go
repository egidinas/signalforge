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

	h5fcreate   uintptr
	h5fclose    uintptr
	h5gcreate2  uintptr
	h5gclose    uintptr
	h5screate_simple uintptr
	h5sclose    uintptr
	h5dcreate2  uintptr
	h5dwrite    uintptr
	h5dclose    uintptr
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

	purego.RegisterLibFunc(&w.h5fcreate, lib, "H5Fcreate")
	purego.RegisterLibFunc(&w.h5fclose, lib, "H5Fclose")
	purego.RegisterLibFunc(&w.h5gcreate2, lib, "H5Gcreate2")
	purego.RegisterLibFunc(&w.h5gclose, lib, "H5Gclose")
	purego.RegisterLibFunc(&w.h5screate_simple, lib, "H5Screate_simple")
	purego.RegisterLibFunc(&w.h5sclose, lib, "H5Sclose")
	purego.RegisterLibFunc(&w.h5dcreate2, lib, "H5Dcreate2")
	purego.RegisterLibFunc(&w.h5dwrite, lib, "H5Dwrite")
	purego.RegisterLibFunc(&w.h5dclose, lib, "H5Dclose")

	return w, nil
}

const (
	H5F_ACC_TRUNC = 0x0002
	H5P_DEFAULT   = 0
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

	sym, _ := purego.Dlsym(w.lib, "H5T_NATIVE_DOUBLE_g")
	dtype := *(*int64)(unsafe.Pointer(sym))
	namePtr := uintptr(unsafe.Pointer(unsafe.SliceData([]byte(name + "\x00"))))
	
	dset := w.call7(w.h5dcreate2, uintptr(parent), namePtr, uintptr(dtype), space, H5P_DEFAULT, H5P_DEFAULT, H5P_DEFAULT)
	if int64(dset) < 0 {
		return fmt.Errorf("H5Dcreate2 failed")
	}
	defer w.call1(w.h5dclose, dset)

	ret := w.call5(w.h5dwrite, dset, uintptr(dtype), H5P_DEFAULT, H5P_DEFAULT, uintptr(unsafe.Pointer(&data[0])))
	if int(ret) < 0 {
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

	sym, _ := purego.Dlsym(w.lib, "H5T_NATIVE_INT64_g")
	dtype := *(*int64)(unsafe.Pointer(sym))
	namePtr := uintptr(unsafe.Pointer(unsafe.SliceData([]byte(name + "\x00"))))
	
	dset := w.call7(w.h5dcreate2, uintptr(parent), namePtr, uintptr(dtype), space, H5P_DEFAULT, H5P_DEFAULT, H5P_DEFAULT)
	if int64(dset) < 0 {
		return fmt.Errorf("H5Dcreate2 failed")
	}
	defer w.call1(w.h5dclose, dset)

	ret := w.call5(w.h5dwrite, dset, uintptr(dtype), H5P_DEFAULT, H5P_DEFAULT, uintptr(unsafe.Pointer(&data[0])))
	if int(ret) < 0 {
		return fmt.Errorf("H5Dwrite failed")
	}
	return nil
}

func (w *HDF5Writer) call1(fn uintptr, a1 uintptr) uintptr {
	ret, _, _ := purego.SyscallN(fn, a1)
	return ret
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

func (w *HDF5Writer) call7(fn uintptr, a1, a2, a3, a4, a5, a6, a7 uintptr) uintptr {
	ret, _, _ := purego.SyscallN(fn, a1, a2, a3, a4, a5, a6, a7)
	return ret
}
