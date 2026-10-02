package io

/*
#include <stdlib.h>
#include <string.h>
#include "hdf5.h"

// The pinned gonum binding can create and read attributes but cannot replace
// one, and its string handling does not cover arrays, so attributes go
// through these helpers instead. loc is any object id: a file id addresses the
// root group.

// ow_write_str_attr writes a scalar variable-length UTF-8 string, replacing
// any existing attribute of that name. This is the form h5py produces for
// `obj.attrs[name] = "text"`, so files stamped from Go and from Python read
// back identically.
static herr_t ow_write_str_attr(hid_t loc, const char* name, const char* value) {
    htri_t exists = H5Aexists(loc, name);
    if (exists < 0) return -1;
    if (exists > 0 && H5Adelete(loc, name) < 0) return -1;

    hid_t type = H5Tcopy(H5T_C_S1);
    H5Tset_size(type, H5T_VARIABLE);
    H5Tset_cset(type, H5T_CSET_UTF8);
    hid_t space = H5Screate(H5S_SCALAR);

    herr_t rc = -1;
    hid_t attr = H5Acreate2(loc, name, type, space, H5P_DEFAULT, H5P_DEFAULT);
    if (attr >= 0) {
        rc = H5Awrite(attr, type, &value);
        H5Aclose(attr);
    }
    H5Sclose(space);
    H5Tclose(type);
    return rc;
}

// ow_read_str_attr reads a scalar string attribute, variable- or fixed-length.
// Returns a malloc'd NUL-terminated copy, or NULL when the attribute is
// missing or is not a single string.
static char* ow_read_str_attr(hid_t loc, const char* name) {
    if (H5Aexists(loc, name) <= 0) return NULL;
    hid_t attr = H5Aopen(loc, name, H5P_DEFAULT);
    if (attr < 0) return NULL;

    char* result = NULL;
    hid_t type = H5Aget_type(attr);
    hid_t space = H5Aget_space(attr);
    if (H5Tget_class(type) == H5T_STRING && H5Sget_simple_extent_npoints(space) == 1) {
        if (H5Tis_variable_str(type) > 0) {
            char* tmp = NULL;
            hid_t memtype = H5Tcopy(H5T_C_S1);
            H5Tset_size(memtype, H5T_VARIABLE);
            H5Tset_cset(memtype, H5Tget_cset(type));
            if (H5Aread(attr, memtype, &tmp) >= 0 && tmp != NULL) {
                result = strdup(tmp);
                H5free_memory(tmp);
            }
            H5Tclose(memtype);
        } else {
            size_t n = H5Tget_size(type);
            result = (char*)calloc(n + 1, 1);
            if (result != NULL && H5Aread(attr, type, result) < 0) {
                free(result);
                result = NULL;
            }
        }
    }
    H5Sclose(space);
    H5Tclose(type);
    H5Aclose(attr);
    return result;
}

// ow_delete_attr removes an attribute if it is there. Not an error if it isn't.
static herr_t ow_delete_attr(hid_t loc, const char* name) {
    htri_t exists = H5Aexists(loc, name);
    if (exists < 0) return -1;
    if (exists == 0) return 0;
    return H5Adelete(loc, name);
}

// ow_write_str_array_attr writes a 1-D array of fixed-length, NUL-padded
// strings, replacing any existing attribute of that name. packed holds n
// strings of width bytes each. This is the form h5py produces for a list of
// numpy.bytes_, which is how the model file's DIMS and PROCESSES attributes
// are already stored.
static herr_t ow_write_str_array_attr(hid_t loc, const char* name, const char* packed, size_t width, hsize_t n) {
    htri_t exists = H5Aexists(loc, name);
    if (exists < 0) return -1;
    if (exists > 0 && H5Adelete(loc, name) < 0) return -1;

    hid_t type = H5Tcopy(H5T_C_S1);
    H5Tset_size(type, width);
    H5Tset_strpad(type, H5T_STR_NULLPAD);
    hid_t space = H5Screate_simple(1, &n, NULL);

    herr_t rc = -1;
    hid_t attr = H5Acreate2(loc, name, type, space, H5P_DEFAULT, H5P_DEFAULT);
    if (attr >= 0) {
        rc = (n == 0) ? 0 : H5Awrite(attr, type, packed);
        H5Aclose(attr);
    }
    H5Sclose(space);
    H5Tclose(type);
    return rc;
}

// ow_read_str_array_attr reads a 1-D array of fixed-length strings. Returns a
// malloc'd buffer of (*n) * (*width) bytes (NULL when *n is zero), and sets
// *found to 0 when the attribute is missing, 1 when it was read, or -1 when
// it exists but is not an array of fixed-length strings.
static char* ow_read_str_array_attr(hid_t loc, const char* name, size_t* width, hsize_t* n, int* found) {
    *found = 0;
    *width = 0;
    *n = 0;
    if (H5Aexists(loc, name) <= 0) return NULL;
    hid_t attr = H5Aopen(loc, name, H5P_DEFAULT);
    if (attr < 0) return NULL;

    char* result = NULL;
    *found = -1;
    hid_t type = H5Aget_type(attr);
    hid_t space = H5Aget_space(attr);
    if (H5Tget_class(type) == H5T_STRING && H5Tis_variable_str(type) == 0) {
        hssize_t points = H5Sget_simple_extent_npoints(space);
        if (points >= 0) {
            *width = H5Tget_size(type);
            *n = (hsize_t)points;
            if (points == 0) {
                *found = 1;
            } else {
                result = (char*)calloc((size_t)points * (*width), 1);
                if (result != NULL && H5Aread(attr, type, result) >= 0) {
                    *found = 1;
                } else {
                    free(result);
                    result = NULL;
                }
            }
        }
    }
    H5Sclose(space);
    H5Tclose(type);
    H5Aclose(attr);
    return result;
}
*/
import "C"

import (
	"errors"
	"strings"
	"unsafe"

	"gonum.org/v1/hdf5"
)

func writeStringAttribute(loc int64, name, value string) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	if C.ow_write_str_attr(C.hid_t(loc), cName, cValue) < 0 {
		return errors.New("Cannot write attribute " + name)
	}
	return nil
}

func readStringAttribute(loc int64, name string) (string, bool) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	cValue := C.ow_read_str_attr(C.hid_t(loc), cName)
	if cValue == nil {
		return "", false
	}
	defer C.free(unsafe.Pointer(cValue))
	// A fixed-length string may be space-padded rather than NUL-padded.
	return strings.TrimRight(C.GoString(cValue), " "), true
}

// SetFileAttributesInFile writes string attributes at the root of an open
// HDF5 file, replacing any that already exist. The caller must hold the
// global HDF5 mutex and manage the file lifecycle.
func SetFileAttributesInFile(f *hdf5.File, attrs map[string]string) error {
	for name, value := range attrs {
		if err := writeStringAttribute(f.ID(), name, value); err != nil {
			return prefix("In "+f.FileName()+": ", err)
		}
	}
	return nil
}

// SetFileAttributes writes string attributes at the root of an HDF5 file,
// creating the file if it doesn't exist and replacing any attributes of the
// same name. Values are stored as h5py stores a Python str, so they read back
// the same whichever side wrote them.
func SetFileAttributes(filename string, attrs map[string]string) error {
	lockHDF5(filename)
	defer unlockHDF5(filename)

	f, err := openWriteOrCreate(filename, true)
	if err != nil {
		return err
	}
	defer f.Close()

	return SetFileAttributesInFile(f, attrs)
}

// RemoveFileAttributes deletes the named attributes from the root of an
// existing HDF5 file. Names the file doesn't have are ignored.
func RemoveFileAttributes(filename string, names ...string) error {
	lockHDF5(filename)
	defer unlockHDF5(filename)

	f, err := openWriteOrCreate(filename, false)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, name := range names {
		cName := C.CString(name)
		rc := C.ow_delete_attr(C.hid_t(f.ID()), cName)
		C.free(unsafe.Pointer(cName))
		if rc < 0 {
			return errors.New("Cannot remove attribute " + name + " from " + filename)
		}
	}
	return nil
}

// FileAttributes reads the named string attributes from the root of an HDF5
// file. Attributes that are missing, or that aren't a single string, are left
// out of the result, so a file with none of them yields an empty map rather
// than an error.
func FileAttributes(filename string, names ...string) (map[string]string, error) {
	rLockHDF5(filename)
	defer rUnlockHDF5(filename)

	f, err := hdf5.OpenFile(filename, hdf5.F_ACC_RDONLY)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := make(map[string]string)
	for _, name := range names {
		if value, ok := readStringAttribute(f.ID(), name); ok {
			result[name] = value
		}
	}
	return result, nil
}

// SetTextAttributeInFile attaches a list of strings to this dataset as an
// attribute, in an already-open file. The caller must hold the global HDF5
// mutex and manage the file lifecycle.
func (h H5Ref[T]) SetTextAttributeInFile(f *hdf5.File, name string, values []string) error {
	ds, err := f.OpenDataset(h.Dataset)
	if err != nil {
		return prefix("Cannot open dataset "+h.Dataset+": ", err)
	}
	defer ds.Close()

	width := 1
	for _, v := range values {
		if len(v) > width {
			width = len(v)
		}
	}

	// Never hand C a pointer into a zero-length slice.
	packed := make([]byte, len(values)*width+1)
	for i, v := range values {
		copy(packed[i*width:(i+1)*width], v)
	}

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	rc := C.ow_write_str_array_attr(C.hid_t(ds.ID()), cName,
		(*C.char)(unsafe.Pointer(&packed[0])), C.size_t(width), C.hsize_t(len(values)))
	if rc < 0 {
		return errors.New("Cannot write attribute " + name + " on " + h.Dataset)
	}
	return nil
}

// SetTextAttribute attaches a list of strings to this dataset as an
// attribute, replacing any existing attribute of that name. The dataset must
// already exist. Values are stored as fixed-length strings, the layout h5py
// uses for a list of bytes and the one the model file's DIMS attribute has.
func (h H5Ref[T]) SetTextAttribute(name string, values []string) error {
	lockHDF5(h.Filename)
	defer unlockHDF5(h.Filename)

	f, err := openWriteOrCreate(h.Filename, true)
	if err != nil {
		return err
	}
	defer f.Close()

	return h.SetTextAttributeInFile(f, name, values)
}

// TextAttribute reads a list-of-strings attribute from this dataset. The
// second return value is false when the dataset has no such attribute.
func (h H5Ref[T]) TextAttribute(name string) ([]string, bool, error) {
	rLockHDF5(h.Filename)
	defer rUnlockHDF5(h.Filename)

	f, err := hdf5.OpenFile(h.Filename, hdf5.F_ACC_RDONLY)
	if err != nil {
		return nil, false, err
	}
	defer f.Close()

	ds, err := f.OpenDataset(h.Dataset)
	if err != nil {
		return nil, false, prefix("Cannot open dataset "+h.Dataset+": ", err)
	}
	defer ds.Close()

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var width C.size_t
	var n C.hsize_t
	var found C.int
	buf := C.ow_read_str_array_attr(C.hid_t(ds.ID()), cName, &width, &n, &found)
	if buf != nil {
		defer C.free(unsafe.Pointer(buf))
	}

	switch {
	case found == 0:
		return nil, false, nil
	case found < 0:
		return nil, true, errors.New("Attribute " + name + " on " + h.Dataset + " is not a list of fixed-length strings")
	}

	result := make([]string, int(n))
	if n > 0 {
		packed := C.GoBytes(unsafe.Pointer(buf), C.int(int(n)*int(width)))
		for i := range result {
			result[i] = strings.TrimRight(string(packed[i*int(width):(i+1)*int(width)]), "\x00")
		}
	}
	return result, true, nil
}
