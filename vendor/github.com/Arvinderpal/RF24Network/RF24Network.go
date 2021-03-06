/* ----------------------------------------------------------------------------
 * This file was automatically generated by SWIG (http://www.swig.org).
 * Version 3.0.12
 *
 * This file is not intended to be easily readable and contains a number of
 * coding conventions designed to improve portability and efficiency. Do not make
 * changes to this file unless you know what you are doing--modify the SWIG
 * interface file instead.
 * ----------------------------------------------------------------------------- */

// source: RF24Network.i

package RF24Network

/*
#define intgo swig_intgo
typedef void *swig_voidp;

#include <stdint.h>


typedef long long intgo;
typedef unsigned long long uintgo;



typedef struct { char *p; intgo n; } _gostring_;
typedef struct { void* array; intgo len; intgo cap; } _goslice_;


typedef long long swig_type_1;
typedef long long swig_type_2;
typedef long long swig_type_3;
typedef long long swig_type_4;
typedef _gostring_ swig_type_5;
typedef _gostring_ swig_type_6;
typedef _gostring_ swig_type_7;
typedef long long swig_type_8;
typedef long long swig_type_9;
typedef long long swig_type_10;
typedef long long swig_type_11;
typedef _gostring_ swig_type_12;
extern void _wrap_Swig_free_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern uintptr_t _wrap_Swig_malloc_RF24Network_37de5201020d96e2(swig_intgo arg1);
extern uintptr_t _wrap_new_StringVector__SWIG_0_RF24Network_37de5201020d96e2(void);
extern uintptr_t _wrap_new_StringVector__SWIG_1_RF24Network_37de5201020d96e2(swig_type_1 arg1);
extern swig_type_2 _wrap_StringVector_size_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern swig_type_3 _wrap_StringVector_capacity_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_StringVector_reserve_RF24Network_37de5201020d96e2(uintptr_t arg1, swig_type_4 arg2);
extern _Bool _wrap_StringVector_isEmpty_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_StringVector_clear_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_StringVector_add_RF24Network_37de5201020d96e2(uintptr_t arg1, swig_type_5 arg2);
extern swig_type_6 _wrap_StringVector_get_RF24Network_37de5201020d96e2(uintptr_t arg1, swig_intgo arg2);
extern void _wrap_StringVector_set_RF24Network_37de5201020d96e2(uintptr_t arg1, swig_intgo arg2, swig_type_7 arg3);
extern void _wrap_delete_StringVector_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern uintptr_t _wrap_new_ByteVector__SWIG_0_RF24Network_37de5201020d96e2(void);
extern uintptr_t _wrap_new_ByteVector__SWIG_1_RF24Network_37de5201020d96e2(swig_type_8 arg1);
extern swig_type_9 _wrap_ByteVector_size_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern swig_type_10 _wrap_ByteVector_capacity_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_ByteVector_reserve_RF24Network_37de5201020d96e2(uintptr_t arg1, swig_type_11 arg2);
extern _Bool _wrap_ByteVector_isEmpty_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_ByteVector_clear_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_ByteVector_add_RF24Network_37de5201020d96e2(uintptr_t arg1, char arg2);
extern char _wrap_ByteVector_get_RF24Network_37de5201020d96e2(uintptr_t arg1, swig_intgo arg2);
extern void _wrap_ByteVector_set_RF24Network_37de5201020d96e2(uintptr_t arg1, swig_intgo arg2, char arg3);
extern void _wrap_delete_ByteVector_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_RF24NetworkHeader_from_node_set_RF24Network_37de5201020d96e2(uintptr_t arg1, short arg2);
extern short _wrap_RF24NetworkHeader_from_node_get_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_RF24NetworkHeader_to_node_set_RF24Network_37de5201020d96e2(uintptr_t arg1, short arg2);
extern short _wrap_RF24NetworkHeader_to_node_get_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_RF24NetworkHeader_id_set_RF24Network_37de5201020d96e2(uintptr_t arg1, short arg2);
extern short _wrap_RF24NetworkHeader_id_get_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_RF24NetworkHeader_Xtype_set_RF24Network_37de5201020d96e2(uintptr_t arg1, char arg2);
extern char _wrap_RF24NetworkHeader_Xtype_get_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_RF24NetworkHeader_reserved_set_RF24Network_37de5201020d96e2(uintptr_t arg1, char arg2);
extern char _wrap_RF24NetworkHeader_reserved_get_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_RF24NetworkHeader_next_id_set_RF24Network_37de5201020d96e2(short arg1);
extern short _wrap_RF24NetworkHeader_next_id_get_RF24Network_37de5201020d96e2(void);
extern uintptr_t _wrap_new_RF24NetworkHeader__SWIG_0_RF24Network_37de5201020d96e2(void);
extern uintptr_t _wrap_new_RF24NetworkHeader__SWIG_1_RF24Network_37de5201020d96e2(short arg1, char arg2);
extern uintptr_t _wrap_new_RF24NetworkHeader__SWIG_2_RF24Network_37de5201020d96e2(short arg1);
extern swig_type_12 _wrap_RF24NetworkHeader_toString_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_delete_RF24NetworkHeader_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_RF24NetworkFrame_header_set_RF24Network_37de5201020d96e2(uintptr_t arg1, uintptr_t arg2);
extern uintptr_t _wrap_RF24NetworkFrame_header_get_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_RF24NetworkFrame_message_size_set_RF24Network_37de5201020d96e2(uintptr_t arg1, short arg2);
extern short _wrap_RF24NetworkFrame_message_size_get_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_RF24NetworkFrame_message_buffer_set_RF24Network_37de5201020d96e2(uintptr_t arg1, swig_voidp arg2);
extern swig_voidp _wrap_RF24NetworkFrame_message_buffer_get_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern uintptr_t _wrap_new_RF24NetworkFrame__SWIG_0_RF24Network_37de5201020d96e2(void);
extern uintptr_t _wrap_new_RF24NetworkFrame__SWIG_1_RF24Network_37de5201020d96e2(uintptr_t arg1, uintptr_t arg2, short arg3);
extern uintptr_t _wrap_new_RF24NetworkFrame__SWIG_2_RF24Network_37de5201020d96e2(uintptr_t arg1, uintptr_t arg2);
extern uintptr_t _wrap_new_RF24NetworkFrame__SWIG_3_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_delete_RF24NetworkFrame_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern uintptr_t _wrap_new_RF24Network_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_RF24Network_begin__SWIG_0_RF24Network_37de5201020d96e2(uintptr_t arg1, short arg2);
extern char _wrap_RF24Network_update_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern _Bool _wrap_RF24Network_available_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern short _wrap_RF24Network_peek__SWIG_0_RF24Network_37de5201020d96e2(uintptr_t arg1, uintptr_t arg2);
extern void _wrap_RF24Network_peek__SWIG_1_RF24Network_37de5201020d96e2(uintptr_t arg1, uintptr_t arg2, uintptr_t arg3, short arg4);
extern short _wrap_RF24Network_read_RF24Network_37de5201020d96e2(uintptr_t arg1, uintptr_t arg2, uintptr_t arg3, short arg4);
extern _Bool _wrap_RF24Network_write__SWIG_0_RF24Network_37de5201020d96e2(uintptr_t arg1, uintptr_t arg2, uintptr_t arg3, short arg4);
extern void _wrap_RF24Network_multicastLevel_RF24Network_37de5201020d96e2(uintptr_t arg1, char arg2);
extern void _wrap_RF24Network_multicastRelay_set_RF24Network_37de5201020d96e2(uintptr_t arg1, _Bool arg2);
extern _Bool _wrap_RF24Network_multicastRelay_get_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_RF24Network_txTimeout_set_RF24Network_37de5201020d96e2(uintptr_t arg1, swig_intgo arg2);
extern swig_intgo _wrap_RF24Network_txTimeout_get_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_RF24Network_routeTimeout_set_RF24Network_37de5201020d96e2(uintptr_t arg1, short arg2);
extern short _wrap_RF24Network_routeTimeout_get_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern _Bool _wrap_RF24Network_write__SWIG_1_RF24Network_37de5201020d96e2(uintptr_t arg1, uintptr_t arg2, uintptr_t arg3, short arg4, short arg5);
extern short _wrap_RF24Network_parent_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern short _wrap_RF24Network_addressOfPipe_RF24Network_37de5201020d96e2(uintptr_t arg1, short arg2, char arg3);
extern _Bool _wrap_RF24Network_is_valid_address_RF24Network_37de5201020d96e2(uintptr_t arg1, short arg2);
extern void _wrap_RF24Network_begin__SWIG_1_RF24Network_37de5201020d96e2(uintptr_t arg1, char arg2, short arg3);
extern void _wrap_RF24Network_frame_buffer_set_RF24Network_37de5201020d96e2(uintptr_t arg1, swig_voidp arg2);
extern swig_voidp _wrap_RF24Network_frame_buffer_get_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_RF24Network_external_queue_set_RF24Network_37de5201020d96e2(uintptr_t arg1, uintptr_t arg2);
extern uintptr_t _wrap_RF24Network_external_queue_get_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_RF24Network_returnSysMsgs_set_RF24Network_37de5201020d96e2(uintptr_t arg1, _Bool arg2);
extern _Bool _wrap_RF24Network_returnSysMsgs_get_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_RF24Network_networkFlags_set_RF24Network_37de5201020d96e2(uintptr_t arg1, char arg2);
extern char _wrap_RF24Network_networkFlags_get_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_delete_RF24Network_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern uintptr_t _wrap_new_Sync_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_Sync_begin_RF24Network_37de5201020d96e2(uintptr_t arg1, short arg2);
extern void _wrap_Sync_reset_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_Sync_update_RF24Network_37de5201020d96e2(uintptr_t arg1);
extern void _wrap_delete_Sync_RF24Network_37de5201020d96e2(uintptr_t arg1);
#undef intgo
*/
import "C"

import "unsafe"
import _ "runtime/cgo"
import "sync"


type _ unsafe.Pointer



var Swig_escape_always_false bool
var Swig_escape_val interface{}


type _swig_fnptr *byte
type _swig_memberptr *byte


type _ sync.Mutex


type swig_gostring struct { p uintptr; n int }
func swigCopyString(s string) string {
  p := *(*swig_gostring)(unsafe.Pointer(&s))
  r := string((*[0x7fffffff]byte)(unsafe.Pointer(p.p))[:p.n])
  Swig_free(p.p)
  return r
}

func Swig_free(arg1 uintptr) {
	_swig_i_0 := arg1
	C._wrap_Swig_free_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0))
}

func Swig_malloc(arg1 int) (_swig_ret uintptr) {
	var swig_r uintptr
	_swig_i_0 := arg1
	swig_r = (uintptr)(C._wrap_Swig_malloc_RF24Network_37de5201020d96e2(C.swig_intgo(_swig_i_0)))
	return swig_r
}

type SwigcptrStringVector uintptr

func (p SwigcptrStringVector) Swigcptr() uintptr {
	return (uintptr)(p)
}

func (p SwigcptrStringVector) SwigIsStringVector() {
}

func NewStringVector__SWIG_0() (_swig_ret StringVector) {
	var swig_r StringVector
	swig_r = (StringVector)(SwigcptrStringVector(C._wrap_new_StringVector__SWIG_0_RF24Network_37de5201020d96e2()))
	return swig_r
}

func NewStringVector__SWIG_1(arg1 int64) (_swig_ret StringVector) {
	var swig_r StringVector
	_swig_i_0 := arg1
	swig_r = (StringVector)(SwigcptrStringVector(C._wrap_new_StringVector__SWIG_1_RF24Network_37de5201020d96e2(C.swig_type_1(_swig_i_0))))
	return swig_r
}

func NewStringVector(a ...interface{}) StringVector {
	argc := len(a)
	if argc == 0 {
		return NewStringVector__SWIG_0()
	}
	if argc == 1 {
		return NewStringVector__SWIG_1(a[0].(int64))
	}
	panic("No match for overloaded function call")
}

func (arg1 SwigcptrStringVector) Size() (_swig_ret int64) {
	var swig_r int64
	_swig_i_0 := arg1
	swig_r = (int64)(C._wrap_StringVector_size_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrStringVector) Capacity() (_swig_ret int64) {
	var swig_r int64
	_swig_i_0 := arg1
	swig_r = (int64)(C._wrap_StringVector_capacity_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrStringVector) Reserve(arg2 int64) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_StringVector_reserve_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.swig_type_4(_swig_i_1))
}

func (arg1 SwigcptrStringVector) IsEmpty() (_swig_ret bool) {
	var swig_r bool
	_swig_i_0 := arg1
	swig_r = (bool)(C._wrap_StringVector_isEmpty_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrStringVector) Clear() {
	_swig_i_0 := arg1
	C._wrap_StringVector_clear_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0))
}

func (arg1 SwigcptrStringVector) Add(arg2 string) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_StringVector_add_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), *(*C.swig_type_5)(unsafe.Pointer(&_swig_i_1)))
	if Swig_escape_always_false {
		Swig_escape_val = arg2
	}
}

func (arg1 SwigcptrStringVector) Get(arg2 int) (_swig_ret string) {
	var swig_r string
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	swig_r_p := C._wrap_StringVector_get_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.swig_intgo(_swig_i_1))
	swig_r = *(*string)(unsafe.Pointer(&swig_r_p))
	var swig_r_1 string
 swig_r_1 = swigCopyString(swig_r) 
	return swig_r_1
}

func (arg1 SwigcptrStringVector) Set(arg2 int, arg3 string) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	_swig_i_2 := arg3
	C._wrap_StringVector_set_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.swig_intgo(_swig_i_1), *(*C.swig_type_7)(unsafe.Pointer(&_swig_i_2)))
	if Swig_escape_always_false {
		Swig_escape_val = arg3
	}
}

func DeleteStringVector(arg1 StringVector) {
	_swig_i_0 := arg1.Swigcptr()
	C._wrap_delete_StringVector_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0))
}

type StringVector interface {
	Swigcptr() uintptr
	SwigIsStringVector()
	Size() (_swig_ret int64)
	Capacity() (_swig_ret int64)
	Reserve(arg2 int64)
	IsEmpty() (_swig_ret bool)
	Clear()
	Add(arg2 string)
	Get(arg2 int) (_swig_ret string)
	Set(arg2 int, arg3 string)
}

type SwigcptrByteVector uintptr

func (p SwigcptrByteVector) Swigcptr() uintptr {
	return (uintptr)(p)
}

func (p SwigcptrByteVector) SwigIsByteVector() {
}

func NewByteVector__SWIG_0() (_swig_ret ByteVector) {
	var swig_r ByteVector
	swig_r = (ByteVector)(SwigcptrByteVector(C._wrap_new_ByteVector__SWIG_0_RF24Network_37de5201020d96e2()))
	return swig_r
}

func NewByteVector__SWIG_1(arg1 int64) (_swig_ret ByteVector) {
	var swig_r ByteVector
	_swig_i_0 := arg1
	swig_r = (ByteVector)(SwigcptrByteVector(C._wrap_new_ByteVector__SWIG_1_RF24Network_37de5201020d96e2(C.swig_type_8(_swig_i_0))))
	return swig_r
}

func NewByteVector(a ...interface{}) ByteVector {
	argc := len(a)
	if argc == 0 {
		return NewByteVector__SWIG_0()
	}
	if argc == 1 {
		return NewByteVector__SWIG_1(a[0].(int64))
	}
	panic("No match for overloaded function call")
}

func (arg1 SwigcptrByteVector) Size() (_swig_ret int64) {
	var swig_r int64
	_swig_i_0 := arg1
	swig_r = (int64)(C._wrap_ByteVector_size_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrByteVector) Capacity() (_swig_ret int64) {
	var swig_r int64
	_swig_i_0 := arg1
	swig_r = (int64)(C._wrap_ByteVector_capacity_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrByteVector) Reserve(arg2 int64) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_ByteVector_reserve_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.swig_type_11(_swig_i_1))
}

func (arg1 SwigcptrByteVector) IsEmpty() (_swig_ret bool) {
	var swig_r bool
	_swig_i_0 := arg1
	swig_r = (bool)(C._wrap_ByteVector_isEmpty_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrByteVector) Clear() {
	_swig_i_0 := arg1
	C._wrap_ByteVector_clear_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0))
}

func (arg1 SwigcptrByteVector) Add(arg2 byte) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_ByteVector_add_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.char(_swig_i_1))
}

func (arg1 SwigcptrByteVector) Get(arg2 int) (_swig_ret byte) {
	var swig_r byte
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	swig_r = (byte)(C._wrap_ByteVector_get_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.swig_intgo(_swig_i_1)))
	return swig_r
}

func (arg1 SwigcptrByteVector) Set(arg2 int, arg3 byte) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	_swig_i_2 := arg3
	C._wrap_ByteVector_set_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.swig_intgo(_swig_i_1), C.char(_swig_i_2))
}

func DeleteByteVector(arg1 ByteVector) {
	_swig_i_0 := arg1.Swigcptr()
	C._wrap_delete_ByteVector_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0))
}

type ByteVector interface {
	Swigcptr() uintptr
	SwigIsByteVector()
	Size() (_swig_ret int64)
	Capacity() (_swig_ret int64)
	Reserve(arg2 int64)
	IsEmpty() (_swig_ret bool)
	Clear()
	Add(arg2 byte)
	Get(arg2 int) (_swig_ret byte)
	Set(arg2 int, arg3 byte)
}

const MIN_USER_DEFINED_HEADER_TYPE int = 0
const MAX_USER_DEFINED_HEADER_TYPE int = 127
const NETWORK_ADDR_RESPONSE int = 128
const NETWORK_PING int = 130
const EXTERNAL_DATA_TYPE int = 131
const NETWORK_FIRST_FRAGMENT int = 148
const NETWORK_MORE_FRAGMENTS int = 149
const NETWORK_LAST_FRAGMENT int = 150
const NETWORK_ACK int = 193
const NETWORK_POLL int = 194
const NETWORK_REQ_ADDRESS int = 195
const NETWORK_MORE_FRAGMENTS_NACK int = 200
const TX_NORMAL int = 0
const TX_ROUTED int = 1
const USER_TX_TO_PHYSICAL_ADDRESS int = 2
const USER_TX_TO_LOGICAL_ADDRESS int = 3
const USER_TX_MULTICAST int = 4
const MAX_FRAME_SIZE int = 32
const FRAME_HEADER_SIZE int = 10
const USE_CURRENT_CHANNEL int = 255
const FLAG_HOLD_INCOMING int = 1
const FLAG_BYPASS_HOLDS int = 2
const FLAG_FAST_FRAG int = 4
const FLAG_NO_POLL int = 8
type SwigcptrRF24NetworkHeader uintptr

func (p SwigcptrRF24NetworkHeader) Swigcptr() uintptr {
	return (uintptr)(p)
}

func (p SwigcptrRF24NetworkHeader) SwigIsRF24NetworkHeader() {
}

func (arg1 SwigcptrRF24NetworkHeader) SetFrom_node(arg2 uint16) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_RF24NetworkHeader_from_node_set_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.short(_swig_i_1))
}

func (arg1 SwigcptrRF24NetworkHeader) GetFrom_node() (_swig_ret uint16) {
	var swig_r uint16
	_swig_i_0 := arg1
	swig_r = (uint16)(C._wrap_RF24NetworkHeader_from_node_get_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrRF24NetworkHeader) SetTo_node(arg2 uint16) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_RF24NetworkHeader_to_node_set_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.short(_swig_i_1))
}

func (arg1 SwigcptrRF24NetworkHeader) GetTo_node() (_swig_ret uint16) {
	var swig_r uint16
	_swig_i_0 := arg1
	swig_r = (uint16)(C._wrap_RF24NetworkHeader_to_node_get_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrRF24NetworkHeader) SetId(arg2 uint16) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_RF24NetworkHeader_id_set_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.short(_swig_i_1))
}

func (arg1 SwigcptrRF24NetworkHeader) GetId() (_swig_ret uint16) {
	var swig_r uint16
	_swig_i_0 := arg1
	swig_r = (uint16)(C._wrap_RF24NetworkHeader_id_get_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrRF24NetworkHeader) SetXtype(arg2 byte) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_RF24NetworkHeader_Xtype_set_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.char(_swig_i_1))
}

func (arg1 SwigcptrRF24NetworkHeader) GetXtype() (_swig_ret byte) {
	var swig_r byte
	_swig_i_0 := arg1
	swig_r = (byte)(C._wrap_RF24NetworkHeader_Xtype_get_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrRF24NetworkHeader) SetReserved(arg2 byte) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_RF24NetworkHeader_reserved_set_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.char(_swig_i_1))
}

func (arg1 SwigcptrRF24NetworkHeader) GetReserved() (_swig_ret byte) {
	var swig_r byte
	_swig_i_0 := arg1
	swig_r = (byte)(C._wrap_RF24NetworkHeader_reserved_get_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func SetRF24NetworkHeaderNext_id(arg1 uint16) {
	_swig_i_0 := arg1
	C._wrap_RF24NetworkHeader_next_id_set_RF24Network_37de5201020d96e2(C.short(_swig_i_0))
}

func GetRF24NetworkHeaderNext_id() (_swig_ret uint16) {
	var swig_r uint16
	swig_r = (uint16)(C._wrap_RF24NetworkHeader_next_id_get_RF24Network_37de5201020d96e2())
	return swig_r
}

func NewRF24NetworkHeader__SWIG_0() (_swig_ret RF24NetworkHeader) {
	var swig_r RF24NetworkHeader
	swig_r = (RF24NetworkHeader)(SwigcptrRF24NetworkHeader(C._wrap_new_RF24NetworkHeader__SWIG_0_RF24Network_37de5201020d96e2()))
	return swig_r
}

func NewRF24NetworkHeader__SWIG_1(arg1 uint16, arg2 byte) (_swig_ret RF24NetworkHeader) {
	var swig_r RF24NetworkHeader
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	swig_r = (RF24NetworkHeader)(SwigcptrRF24NetworkHeader(C._wrap_new_RF24NetworkHeader__SWIG_1_RF24Network_37de5201020d96e2(C.short(_swig_i_0), C.char(_swig_i_1))))
	return swig_r
}

func NewRF24NetworkHeader__SWIG_2(arg1 uint16) (_swig_ret RF24NetworkHeader) {
	var swig_r RF24NetworkHeader
	_swig_i_0 := arg1
	swig_r = (RF24NetworkHeader)(SwigcptrRF24NetworkHeader(C._wrap_new_RF24NetworkHeader__SWIG_2_RF24Network_37de5201020d96e2(C.short(_swig_i_0))))
	return swig_r
}

func NewRF24NetworkHeader(a ...interface{}) RF24NetworkHeader {
	argc := len(a)
	if argc == 0 {
		return NewRF24NetworkHeader__SWIG_0()
	}
	if argc == 1 {
		return NewRF24NetworkHeader__SWIG_2(a[0].(uint16))
	}
	if argc == 2 {
		return NewRF24NetworkHeader__SWIG_1(a[0].(uint16), a[1].(byte))
	}
	panic("No match for overloaded function call")
}

func (arg1 SwigcptrRF24NetworkHeader) ToString() (_swig_ret string) {
	var swig_r string
	_swig_i_0 := arg1
	swig_r_p := C._wrap_RF24NetworkHeader_toString_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0))
	swig_r = *(*string)(unsafe.Pointer(&swig_r_p))
	var swig_r_1 string
 swig_r_1 = swigCopyString(swig_r) 
	return swig_r_1
}

func DeleteRF24NetworkHeader(arg1 RF24NetworkHeader) {
	_swig_i_0 := arg1.Swigcptr()
	C._wrap_delete_RF24NetworkHeader_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0))
}

type RF24NetworkHeader interface {
	Swigcptr() uintptr
	SwigIsRF24NetworkHeader()
	SetFrom_node(arg2 uint16)
	GetFrom_node() (_swig_ret uint16)
	SetTo_node(arg2 uint16)
	GetTo_node() (_swig_ret uint16)
	SetId(arg2 uint16)
	GetId() (_swig_ret uint16)
	SetXtype(arg2 byte)
	GetXtype() (_swig_ret byte)
	SetReserved(arg2 byte)
	GetReserved() (_swig_ret byte)
	ToString() (_swig_ret string)
}

type SwigcptrRF24NetworkFrame uintptr

func (p SwigcptrRF24NetworkFrame) Swigcptr() uintptr {
	return (uintptr)(p)
}

func (p SwigcptrRF24NetworkFrame) SwigIsRF24NetworkFrame() {
}

func (arg1 SwigcptrRF24NetworkFrame) SetHeader(arg2 RF24NetworkHeader) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2.Swigcptr()
	C._wrap_RF24NetworkFrame_header_set_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.uintptr_t(_swig_i_1))
}

func (arg1 SwigcptrRF24NetworkFrame) GetHeader() (_swig_ret RF24NetworkHeader) {
	var swig_r RF24NetworkHeader
	_swig_i_0 := arg1
	swig_r = (RF24NetworkHeader)(SwigcptrRF24NetworkHeader(C._wrap_RF24NetworkFrame_header_get_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0))))
	return swig_r
}

func (arg1 SwigcptrRF24NetworkFrame) SetMessage_size(arg2 uint16) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_RF24NetworkFrame_message_size_set_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.short(_swig_i_1))
}

func (arg1 SwigcptrRF24NetworkFrame) GetMessage_size() (_swig_ret uint16) {
	var swig_r uint16
	_swig_i_0 := arg1
	swig_r = (uint16)(C._wrap_RF24NetworkFrame_message_size_get_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrRF24NetworkFrame) SetMessage_buffer(arg2 *byte) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_RF24NetworkFrame_message_buffer_set_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.swig_voidp(_swig_i_1))
}

func (arg1 SwigcptrRF24NetworkFrame) GetMessage_buffer() (_swig_ret *byte) {
	var swig_r *byte
	_swig_i_0 := arg1
	swig_r = (*byte)(C._wrap_RF24NetworkFrame_message_buffer_get_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func NewRF24NetworkFrame__SWIG_0() (_swig_ret RF24NetworkFrame) {
	var swig_r RF24NetworkFrame
	swig_r = (RF24NetworkFrame)(SwigcptrRF24NetworkFrame(C._wrap_new_RF24NetworkFrame__SWIG_0_RF24Network_37de5201020d96e2()))
	return swig_r
}

func NewRF24NetworkFrame__SWIG_1(arg1 RF24NetworkHeader, arg2 uintptr, arg3 uint16) (_swig_ret RF24NetworkFrame) {
	var swig_r RF24NetworkFrame
	_swig_i_0 := arg1.Swigcptr()
	_swig_i_1 := arg2
	_swig_i_2 := arg3
	swig_r = (RF24NetworkFrame)(SwigcptrRF24NetworkFrame(C._wrap_new_RF24NetworkFrame__SWIG_1_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.uintptr_t(_swig_i_1), C.short(_swig_i_2))))
	return swig_r
}

func NewRF24NetworkFrame__SWIG_2(arg1 RF24NetworkHeader, arg2 uintptr) (_swig_ret RF24NetworkFrame) {
	var swig_r RF24NetworkFrame
	_swig_i_0 := arg1.Swigcptr()
	_swig_i_1 := arg2
	swig_r = (RF24NetworkFrame)(SwigcptrRF24NetworkFrame(C._wrap_new_RF24NetworkFrame__SWIG_2_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.uintptr_t(_swig_i_1))))
	return swig_r
}

func NewRF24NetworkFrame__SWIG_3(arg1 RF24NetworkHeader) (_swig_ret RF24NetworkFrame) {
	var swig_r RF24NetworkFrame
	_swig_i_0 := arg1.Swigcptr()
	swig_r = (RF24NetworkFrame)(SwigcptrRF24NetworkFrame(C._wrap_new_RF24NetworkFrame__SWIG_3_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0))))
	return swig_r
}

func NewRF24NetworkFrame(a ...interface{}) RF24NetworkFrame {
	argc := len(a)
	if argc == 0 {
		return NewRF24NetworkFrame__SWIG_0()
	}
	if argc == 1 {
		return NewRF24NetworkFrame__SWIG_3(a[0].(RF24NetworkHeader))
	}
	if argc == 2 {
		return NewRF24NetworkFrame__SWIG_2(a[0].(RF24NetworkHeader), a[1].(uintptr))
	}
	if argc == 3 {
		return NewRF24NetworkFrame__SWIG_1(a[0].(RF24NetworkHeader), a[1].(uintptr), a[2].(uint16))
	}
	panic("No match for overloaded function call")
}

func DeleteRF24NetworkFrame(arg1 RF24NetworkFrame) {
	_swig_i_0 := arg1.Swigcptr()
	C._wrap_delete_RF24NetworkFrame_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0))
}

type RF24NetworkFrame interface {
	Swigcptr() uintptr
	SwigIsRF24NetworkFrame()
	SetHeader(arg2 RF24NetworkHeader)
	GetHeader() (_swig_ret RF24NetworkHeader)
	SetMessage_size(arg2 uint16)
	GetMessage_size() (_swig_ret uint16)
	SetMessage_buffer(arg2 *byte)
	GetMessage_buffer() (_swig_ret *byte)
}

type SwigcptrRF24Network uintptr

func (p SwigcptrRF24Network) Swigcptr() uintptr {
	return (uintptr)(p)
}

func (p SwigcptrRF24Network) SwigIsRF24Network() {
}

func NewRF24Network(arg1 RF24) (_swig_ret RF24Network) {
	var swig_r RF24Network
	_swig_i_0 := arg1.Swigcptr()
	swig_r = (RF24Network)(SwigcptrRF24Network(C._wrap_new_RF24Network_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0))))
	return swig_r
}

func (arg1 SwigcptrRF24Network) Begin__SWIG_0(arg2 uint16) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_RF24Network_begin__SWIG_0_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.short(_swig_i_1))
}

func (arg1 SwigcptrRF24Network) Update() (_swig_ret byte) {
	var swig_r byte
	_swig_i_0 := arg1
	swig_r = (byte)(C._wrap_RF24Network_update_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrRF24Network) Available() (_swig_ret bool) {
	var swig_r bool
	_swig_i_0 := arg1
	swig_r = (bool)(C._wrap_RF24Network_available_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrRF24Network) Peek__SWIG_0(arg2 RF24NetworkHeader) (_swig_ret uint16) {
	var swig_r uint16
	_swig_i_0 := arg1
	_swig_i_1 := arg2.Swigcptr()
	swig_r = (uint16)(C._wrap_RF24Network_peek__SWIG_0_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.uintptr_t(_swig_i_1)))
	return swig_r
}

func (arg1 SwigcptrRF24Network) Peek__SWIG_1(arg2 RF24NetworkHeader, arg3 uintptr, arg4 uint16) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2.Swigcptr()
	_swig_i_2 := arg3
	_swig_i_3 := arg4
	C._wrap_RF24Network_peek__SWIG_1_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.uintptr_t(_swig_i_1), C.uintptr_t(_swig_i_2), C.short(_swig_i_3))
}

func (p SwigcptrRF24Network) Peek(a ...interface{}) interface{} {
	argc := len(a)
	if argc == 1 {
		return p.Peek__SWIG_0(a[0].(RF24NetworkHeader))
	}
	if argc == 3 {
		p.Peek__SWIG_1(a[0].(RF24NetworkHeader), a[1].(uintptr), a[2].(uint16))
		return 0
	}
	panic("No match for overloaded function call")
}

func (arg1 SwigcptrRF24Network) Read(arg2 RF24NetworkHeader, arg3 uintptr, arg4 uint16) (_swig_ret uint16) {
	var swig_r uint16
	_swig_i_0 := arg1
	_swig_i_1 := arg2.Swigcptr()
	_swig_i_2 := arg3
	_swig_i_3 := arg4
	swig_r = (uint16)(C._wrap_RF24Network_read_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.uintptr_t(_swig_i_1), C.uintptr_t(_swig_i_2), C.short(_swig_i_3)))
	return swig_r
}

func (arg1 SwigcptrRF24Network) Write__SWIG_0(arg2 RF24NetworkHeader, arg3 uintptr, arg4 uint16) (_swig_ret bool) {
	var swig_r bool
	_swig_i_0 := arg1
	_swig_i_1 := arg2.Swigcptr()
	_swig_i_2 := arg3
	_swig_i_3 := arg4
	swig_r = (bool)(C._wrap_RF24Network_write__SWIG_0_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.uintptr_t(_swig_i_1), C.uintptr_t(_swig_i_2), C.short(_swig_i_3)))
	return swig_r
}

func (arg1 SwigcptrRF24Network) MulticastLevel(arg2 byte) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_RF24Network_multicastLevel_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.char(_swig_i_1))
}

func (arg1 SwigcptrRF24Network) SetMulticastRelay(arg2 bool) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_RF24Network_multicastRelay_set_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C._Bool(_swig_i_1))
}

func (arg1 SwigcptrRF24Network) GetMulticastRelay() (_swig_ret bool) {
	var swig_r bool
	_swig_i_0 := arg1
	swig_r = (bool)(C._wrap_RF24Network_multicastRelay_get_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrRF24Network) SetTxTimeout(arg2 uint) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_RF24Network_txTimeout_set_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.swig_intgo(_swig_i_1))
}

func (arg1 SwigcptrRF24Network) GetTxTimeout() (_swig_ret uint) {
	var swig_r uint
	_swig_i_0 := arg1
	swig_r = (uint)(C._wrap_RF24Network_txTimeout_get_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrRF24Network) SetRouteTimeout(arg2 uint16) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_RF24Network_routeTimeout_set_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.short(_swig_i_1))
}

func (arg1 SwigcptrRF24Network) GetRouteTimeout() (_swig_ret uint16) {
	var swig_r uint16
	_swig_i_0 := arg1
	swig_r = (uint16)(C._wrap_RF24Network_routeTimeout_get_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrRF24Network) Write__SWIG_1(arg2 RF24NetworkHeader, arg3 uintptr, arg4 uint16, arg5 uint16) (_swig_ret bool) {
	var swig_r bool
	_swig_i_0 := arg1
	_swig_i_1 := arg2.Swigcptr()
	_swig_i_2 := arg3
	_swig_i_3 := arg4
	_swig_i_4 := arg5
	swig_r = (bool)(C._wrap_RF24Network_write__SWIG_1_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.uintptr_t(_swig_i_1), C.uintptr_t(_swig_i_2), C.short(_swig_i_3), C.short(_swig_i_4)))
	return swig_r
}

func (p SwigcptrRF24Network) Write(a ...interface{}) bool {
	argc := len(a)
	if argc == 3 {
		return p.Write__SWIG_0(a[0].(RF24NetworkHeader), a[1].(uintptr), a[2].(uint16))
	}
	if argc == 4 {
		return p.Write__SWIG_1(a[0].(RF24NetworkHeader), a[1].(uintptr), a[2].(uint16), a[3].(uint16))
	}
	panic("No match for overloaded function call")
}

func (arg1 SwigcptrRF24Network) Parent() (_swig_ret uint16) {
	var swig_r uint16
	_swig_i_0 := arg1
	swig_r = (uint16)(C._wrap_RF24Network_parent_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrRF24Network) AddressOfPipe(arg2 uint16, arg3 byte) (_swig_ret uint16) {
	var swig_r uint16
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	_swig_i_2 := arg3
	swig_r = (uint16)(C._wrap_RF24Network_addressOfPipe_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.short(_swig_i_1), C.char(_swig_i_2)))
	return swig_r
}

func (arg1 SwigcptrRF24Network) Is_valid_address(arg2 uint16) (_swig_ret bool) {
	var swig_r bool
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	swig_r = (bool)(C._wrap_RF24Network_is_valid_address_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.short(_swig_i_1)))
	return swig_r
}

func (arg1 SwigcptrRF24Network) Begin__SWIG_1(arg2 byte, arg3 uint16) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	_swig_i_2 := arg3
	C._wrap_RF24Network_begin__SWIG_1_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.char(_swig_i_1), C.short(_swig_i_2))
}

func (p SwigcptrRF24Network) Begin(a ...interface{}) {
	argc := len(a)
	if argc == 1 {
		p.Begin__SWIG_0(a[0].(uint16))
		return
	}
	if argc == 2 {
		p.Begin__SWIG_1(a[0].(byte), a[1].(uint16))
		return
	}
	panic("No match for overloaded function call")
}

func (arg1 SwigcptrRF24Network) SetFrame_buffer(arg2 *byte) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_RF24Network_frame_buffer_set_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.swig_voidp(_swig_i_1))
}

func (arg1 SwigcptrRF24Network) GetFrame_buffer() (_swig_ret *byte) {
	var swig_r *byte
	_swig_i_0 := arg1
	swig_r = (*byte)(C._wrap_RF24Network_frame_buffer_get_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrRF24Network) SetExternal_queue(arg2 Std_queue_Sl_RF24NetworkFrame_Sg_) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2.Swigcptr()
	C._wrap_RF24Network_external_queue_set_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.uintptr_t(_swig_i_1))
}

func (arg1 SwigcptrRF24Network) GetExternal_queue() (_swig_ret Std_queue_Sl_RF24NetworkFrame_Sg_) {
	var swig_r Std_queue_Sl_RF24NetworkFrame_Sg_
	_swig_i_0 := arg1
	swig_r = (Std_queue_Sl_RF24NetworkFrame_Sg_)(SwigcptrStd_queue_Sl_RF24NetworkFrame_Sg_(C._wrap_RF24Network_external_queue_get_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0))))
	return swig_r
}

func (arg1 SwigcptrRF24Network) SetReturnSysMsgs(arg2 bool) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_RF24Network_returnSysMsgs_set_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C._Bool(_swig_i_1))
}

func (arg1 SwigcptrRF24Network) GetReturnSysMsgs() (_swig_ret bool) {
	var swig_r bool
	_swig_i_0 := arg1
	swig_r = (bool)(C._wrap_RF24Network_returnSysMsgs_get_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrRF24Network) SetNetworkFlags(arg2 byte) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_RF24Network_networkFlags_set_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.char(_swig_i_1))
}

func (arg1 SwigcptrRF24Network) GetNetworkFlags() (_swig_ret byte) {
	var swig_r byte
	_swig_i_0 := arg1
	swig_r = (byte)(C._wrap_RF24Network_networkFlags_get_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func DeleteRF24Network(arg1 RF24Network) {
	_swig_i_0 := arg1.Swigcptr()
	C._wrap_delete_RF24Network_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0))
}

type RF24Network interface {
	Swigcptr() uintptr
	SwigIsRF24Network()
	Update() (_swig_ret byte)
	Available() (_swig_ret bool)
	Peek(a ...interface{}) interface{}
	Read(arg2 RF24NetworkHeader, arg3 uintptr, arg4 uint16) (_swig_ret uint16)
	MulticastLevel(arg2 byte)
	SetMulticastRelay(arg2 bool)
	GetMulticastRelay() (_swig_ret bool)
	SetTxTimeout(arg2 uint)
	GetTxTimeout() (_swig_ret uint)
	SetRouteTimeout(arg2 uint16)
	GetRouteTimeout() (_swig_ret uint16)
	Write(a ...interface{}) bool
	Parent() (_swig_ret uint16)
	AddressOfPipe(arg2 uint16, arg3 byte) (_swig_ret uint16)
	Is_valid_address(arg2 uint16) (_swig_ret bool)
	Begin(a ...interface{})
	SetFrame_buffer(arg2 *byte)
	GetFrame_buffer() (_swig_ret *byte)
	SetExternal_queue(arg2 Std_queue_Sl_RF24NetworkFrame_Sg_)
	GetExternal_queue() (_swig_ret Std_queue_Sl_RF24NetworkFrame_Sg_)
	SetReturnSysMsgs(arg2 bool)
	GetReturnSysMsgs() (_swig_ret bool)
	SetNetworkFlags(arg2 byte)
	GetNetworkFlags() (_swig_ret byte)
}

const NETWORK_DEFAULT_ADDRESS int = 04444
const MAIN_BUFFER_SIZE int = 144+10
const MAX_PAYLOAD_SIZE int = 144+10-10
type SwigcptrSync uintptr

func (p SwigcptrSync) Swigcptr() uintptr {
	return (uintptr)(p)
}

func (p SwigcptrSync) SwigIsSync() {
}

func NewSync(arg1 RF24Network) (_swig_ret Sync) {
	var swig_r Sync
	_swig_i_0 := arg1.Swigcptr()
	swig_r = (Sync)(SwigcptrSync(C._wrap_new_Sync_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0))))
	return swig_r
}

func (arg1 SwigcptrSync) Begin(arg2 uint16) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_Sync_begin_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0), C.short(_swig_i_1))
}

func (arg1 SwigcptrSync) Reset() {
	_swig_i_0 := arg1
	C._wrap_Sync_reset_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0))
}

func (arg1 SwigcptrSync) Update() {
	_swig_i_0 := arg1
	C._wrap_Sync_update_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0))
}

func DeleteSync(arg1 Sync) {
	_swig_i_0 := arg1.Swigcptr()
	C._wrap_delete_Sync_RF24Network_37de5201020d96e2(C.uintptr_t(_swig_i_0))
}

type Sync interface {
	Swigcptr() uintptr
	SwigIsSync()
	Begin(arg2 uint16)
	Reset()
	Update()
}


type SwigcptrStd_queue_Sl_RF24NetworkFrame_Sg_ uintptr
type Std_queue_Sl_RF24NetworkFrame_Sg_ interface {
	Swigcptr() uintptr;
}
func (p SwigcptrStd_queue_Sl_RF24NetworkFrame_Sg_) Swigcptr() uintptr {
	return uintptr(p)
}

type SwigcptrRF24 uintptr
type RF24 interface {
	Swigcptr() uintptr;
}
func (p SwigcptrRF24) Swigcptr() uintptr {
	return uintptr(p)
}

