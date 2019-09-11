package main

/*
#include "../sphinxudf.h"
#include <string.h>
#include <stdlib.h>
size_t _GoStringLen(_GoString_ s);
const char *_GoStringPtr(_GoString_ s);
static void cmsg (char * sDst, _GoString_ s)
{
	size_t n = _GoStringLen(s);
	if ( n>SPH_UDF_ERROR_LEN-1 )
		n = SPH_UDF_ERROR_LEN-1;
	strncpy ( sDst, (const char*) _GoStringPtr(s), n);
	sDst[n] = '\0';
}
static char* retmsg ( _GoString_ msg, sphinx_malloc_fn f )
{
	size_t iLen = _GoStringLen(msg);
	char* sRes = f(iLen+1);
	strncpy ( sRes, (const char*) _GoStringPtr(msg), iLen);
	sRes[iLen] = '\0';
	return sRes;
}
static void logmsg ( _GoString_ msg, sphinx_log_fn f )
{
	if (f)
		f ( _GoStringPtr(msg), _GoStringLen(msg) );
}
*/
import "C"
import (
	"reflect"
	"unsafe"
)

// Common constants for daemon and client.
const (
	/// current udf version
	SPH_UDF_VERSION = C.SPH_UDF_VERSION
)

type SPH_UDF_TYPE uint32

/// UDF argument and result value types
const (
	SPH_UDF_TYPE_UINT32    = SPH_UDF_TYPE(C.SPH_UDF_TYPE_UINT32)    ///< unsigned 32-bit integer
	SPH_UDF_TYPE_UINT32SET = SPH_UDF_TYPE(C.SPH_UDF_TYPE_UINT32SET) ///< sorted set of unsigned 32-bit integers
	SPH_UDF_TYPE_INT64     = SPH_UDF_TYPE(C.SPH_UDF_TYPE_INT64)     ///< signed 64-bit integer
	SPH_UDF_TYPE_FLOAT     = SPH_UDF_TYPE(C.SPH_UDF_TYPE_FLOAT)     ///< single-precision IEEE 754 float
	SPH_UDF_TYPE_STRING    = SPH_UDF_TYPE(C.SPH_UDF_TYPE_STRING)    ///< non-ASCIIZ string, with a separately stored length
	SPH_UDF_TYPE_INT64SET  = SPH_UDF_TYPE(C.SPH_UDF_TYPE_INT64SET)  ///< sorted set of signed 64-bit integers
	SPH_UDF_TYPE_FACTORS   = SPH_UDF_TYPE(C.SPH_UDF_TYPE_FACTORS)   ///< packed ranking factors
	SPH_UDF_TYPE_JSON      = SPH_UDF_TYPE(C.SPH_UDF_TYPE_JSON)      ///< whole json or particular field as a string
)

// ERR_MSG is the buffer for returning error messages from _init functions.
type ERR_MSG C.char

// Report packs message from go string into C string buffer to be returned from UDF function.
// It returns 1 in order to be used as shortcut (i.e. 'return msg.Report(...)' instead of 'msg.Report(...); return 1'
func (errmsg *ERR_MSG) say(message string) int32 {
	putstr((*C.char)(errmsg), message)
	return 1
}

// ERR_FLAG points to success flag and may be used to indicate critical errors
type ERR_FLAG C.char

// fail set ERR_FLAG to 1
func (errflag *ERR_FLAG) fail() {
	*errflag = 1
}

// SPH_UDF_ARGS contain arguments passed to the function.
/*  -godefs shows this structure here:
type SPH_UDF_ARGS struct {
        Arg_count       int32
        Arg_types       *uint32
        Arg_values      **int8
        Arg_names       **int8
        Str_lengths     *int32
        Fn_malloc       *[0]byte
}
*/
type SPH_UDF_ARGS C.SPH_UDF_ARGS

func (args *SPH_UDF_ARGS) Arg_count() int32 {
	return int32(args.arg_count)
}

// internal: returns pointer go arg_type
func (args *SPH_UDF_ARGS) typeptr(idx int) *SPH_UDF_TYPE {
	return (*SPH_UDF_TYPE)(unsafe.Pointer(uintptr(unsafe.Pointer(args.arg_types)) +
		unsafe.Sizeof(*args.arg_types)*uintptr(idx)))
}

// internal: returns unsafe pointer go arg_value
func (args *SPH_UDF_ARGS) valueptr(idx int) unsafe.Pointer {
	ptr := unsafe.Pointer(uintptr(unsafe.Pointer(args.arg_values)) +
		unsafe.Sizeof(*args.arg_values)*uintptr(idx))
	return *(*unsafe.Pointer)(ptr)
}

// internal: returns unsafe pointer go arg_name
func (args *SPH_UDF_ARGS) nameptr(idx int) unsafe.Pointer {
	base := uintptr(unsafe.Pointer(args.arg_names))
	if base == 0 {
		return nil
	}

	ptr := base + unsafe.Sizeof(*args.arg_names)*uintptr(idx)
	return *(*unsafe.Pointer)(unsafe.Pointer(ptr))
}

// internal: returns len of arg_value
func (args *SPH_UDF_ARGS) lenval(idx int) int {
	return int(*(*int32)(unsafe.Pointer(uintptr(unsafe.Pointer(args.str_lengths)) +
		unsafe.Sizeof(*args.str_lengths)*uintptr(idx))))
}

// return name of the arg by idx
func (args *SPH_UDF_ARGS) arg_name(idx int) string {
	return GoString((*C.char)(args.nameptr(idx)))
}

// return type of the arg by idx
func (args *SPH_UDF_ARGS) arg_type(idx int) SPH_UDF_TYPE {
	return *args.typeptr(idx)
}

// return string value by idx
// it not copies value, but use backend C string instead
func (args *SPH_UDF_ARGS) stringval(idx int) string {
	return GoStringN((*C.char)(args.valueptr(idx)), args.lenval(idx))
}

// return slice value by idx
// it not copies value, but use backend C string instead
func (args *SPH_UDF_ARGS) mva32(idx int) []uint32 {
	return GoSliceUint32(args.valueptr(idx), args.lenval(idx))
}

// return slice value by idx
// it not copies value, but use backend C string instead
func (args *SPH_UDF_ARGS) mva64(idx int) []int64 {
	return GoSliceInt64(args.valueptr(idx), args.lenval(idx))
}

// convert Go string into C string result and return it
func (args *SPH_UDF_ARGS) return_string(result string) uintptr {
	return uintptr(unsafe.Pointer(C.retmsg(result, args.fn_malloc)))
}

/* UDF initialization
// -godefs shows this structure here:
type SPH_UDF_INIT struct {
        Func_data       *byte
        Is_const        int8
        Pad_cgo_0       [7]byte
}
*/
type SPH_UDF_INIT C.SPH_UDF_INIT

// set func_data to given value
// note that according CGO spec you can't provide any pointer, but only integer value,
// since gc manages all go pointers
func (init *SPH_UDF_INIT) setvalue(value uintptr) {
	*(*uintptr)(unsafe.Pointer(&init.func_data)) = value
}

func (init *SPH_UDF_INIT) setuint32(value uint32) {
	*(*uint32)(unsafe.Pointer(&init.func_data)) = value
}

// get func_data
func (init *SPH_UDF_INIT) getvalue() uintptr {
	return *(*uintptr)(unsafe.Pointer(&init.func_data))
}

func (init *SPH_UDF_INIT) getuint32() uint32 {
	return *(*uint32)(unsafe.Pointer(&init.func_data))
}

// generic stuff

type SPH_RANKER_INIT C.SPH_RANKER_INIT
type SPH_RANKER_HIT C.SPH_RANKER_HIT

// push warning message back to daemon
var cblog *C.sphinx_log_fn

func sphWarning(msg string) {
	C.logmsg(msg, cblog)
}

// C.GoString works by copying backend data.
// We, in turn, make just a tiny wrapper
func GoString(str *C.char) string {
	var rawstring string
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&rawstring))
	hdr.Data = uintptr(unsafe.Pointer(str))
	hdr.Len = int(C.strlen(str))
	return rawstring
}

func GoStringN(str *C.char, len int) string {
	var rawstring string
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&rawstring))
	hdr.Data = uintptr(unsafe.Pointer(str))
	hdr.Len = len
	return rawstring
}

// return slice value from raw C pointer and length
func GoSliceInt32(data unsafe.Pointer, len int) []int32 {
	var slice []int32
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	sliceHeader.Cap = len
	sliceHeader.Len = len
	sliceHeader.Data = uintptr(data)
	return slice
}

func GoSliceUint32(data unsafe.Pointer, len int) []uint32 {
	var slice []uint32
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	sliceHeader.Cap = len
	sliceHeader.Len = len
	sliceHeader.Data = uintptr(data)
	return slice
}

func GoSliceInt64(data unsafe.Pointer, len int) []int64 {
	var slice []int64
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	sliceHeader.Cap = len
	sliceHeader.Len = len
	sliceHeader.Data = uintptr(data)
	return slice
}

// put given mesage into C string destination
func putstr(dst *C.char, message string) {
	C.cmsg(dst, message)
}

// wrappers over C malloc and free.
func malloc(param int) uintptr {
	return uintptr(C.malloc((C.ulong)(param)))
}

func free(param uintptr) {
	C.free(unsafe.Pointer(param))
}

// global functions that must be in any udf plugin library

/// UDF version control. Named as LIBRARYNAME_ver (i.e. goudf_ver in the case)
/// gets called once when the library is loaded
//export goudf_ver
func goudf_ver() int32 {
	return SPH_UDF_VERSION
}

/// Reinit. Was called in workers=prefork, now it is just necessary stub.
//export goudf_reinit
func goudf_reinit() {
}

/// Set warnings gallback. Allows to report messages to daemon's log by calling 'sphWarning'.
//export goudf_setlogcb
func goudf_setlogcb(logfn *C.sphinx_log_fn) {
	cblog = logfn
}

func main() {
	// We need the main function to make possible
	// CGO compiler to compile the package as C shared library
}
