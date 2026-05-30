//go:build dnp3_ffi

// CGO binding to Step Function I/O libdnp3_ffi (Rust core + C FFI).
//
// Layout:
//   - cgo preamble holds C wrapper functions that match the lib's read_handler /
//     association_handler / client_state_listener vtable structs. Each wrapper
//     calls back into Go via the //export funcs at the bottom of this file.
//   - The ctx pointer carried through the lib's callbacks is a cgo.Handle
//     pointing to an *assocCtx (one per association) so we know which
//     outstation a callback belongs to.
//   - Iterators delivered to measurement callbacks are drained synchronously
//     and the resulting dnp3.Measurement values are passed to Handler.OnMeasurement
//     on the master's own goroutine (the lib calls us from its tokio runtime
//     threads; we hand off back to Go without spawning extras).
//
// Build:  go build -tags dnp3_ffi
package dnp3

/*
#include <stdlib.h>
#include <string.h>
#include "dnp3.h"

// Exported Go callback prototypes (definitions emitted by cgo from //export funcs).
extern void goDnp3ClientState(uintptr_t handle, int state);
extern void goDnp3Binary(uintptr_t handle, uint16_t index, _Bool value, uint8_t flags, uint64_t ts_ms, int time_quality, int read_type);
extern void goDnp3DoubleBit(uintptr_t handle, uint16_t index, int value, uint8_t flags, uint64_t ts_ms, int time_quality, int read_type);
extern void goDnp3BinaryOutputStatus(uintptr_t handle, uint16_t index, _Bool value, uint8_t flags, uint64_t ts_ms, int time_quality, int read_type);
extern void goDnp3Counter(uintptr_t handle, uint16_t index, uint32_t value, uint8_t flags, uint64_t ts_ms, int time_quality, int read_type);
extern void goDnp3FrozenCounter(uintptr_t handle, uint16_t index, uint32_t value, uint8_t flags, uint64_t ts_ms, int time_quality, int read_type);
extern void goDnp3Analog(uintptr_t handle, uint16_t index, double value, uint8_t flags, uint64_t ts_ms, int time_quality, int read_type);
extern void goDnp3AnalogOutputStatus(uintptr_t handle, uint16_t index, double value, uint8_t flags, uint64_t ts_ms, int time_quality, int read_type);
extern void goDnp3OctetString(uintptr_t handle, uint16_t index, uint8_t* data, size_t len, int read_type);
extern void goDnp3LibLog(int level, char* msg);

// The current dnp3_read_type_t seen at fragment-begin. Stored on the assocCtx
// via the lib's ctx pointer so individual measurement callbacks can include it.
// We can't read it from inside an iterator callback otherwise.
//
// Simpler approach: pack it into the ctx upper bits. We instead store it via
// a small ThreadLocal-ish mechanism: the lib guarantees begin_fragment / handlers /
// end_fragment are called sequentially on the same thread for one fragment, so we
// stash the read_type on the assocCtx via a setter and read it during measurement
// callbacks. The setter/getter are exposed as Go-side functions to avoid a TLS hack.

// --- read_handler vtable callbacks ---------------------------------------
//
// These are non-static so cgo links them. Each receives the opaque iterator and
// drains it, calling the matching //exported Go func per value. The ctx void*
// is the cgo.Handle uintptr cast.

static void wrap_begin_fragment(dnp3_read_type_t rt, dnp3_response_header_t hdr, void* ctx) {
    // Stash read_type on the assocCtx; goDnp3SetReadType picks it up via the handle.
    extern void goDnp3SetReadType(uintptr_t handle, int rt);
    goDnp3SetReadType((uintptr_t)ctx, (int)rt);
}
static void wrap_end_fragment(dnp3_read_type_t rt, dnp3_response_header_t hdr, void* ctx) {
    (void)rt; (void)hdr; (void)ctx;
}

static void wrap_handle_binary_input(dnp3_header_info_t info, dnp3_binary_input_iterator_t* iter, void* ctx) {
    extern int goDnp3GetReadType(uintptr_t handle);
    int rt = goDnp3GetReadType((uintptr_t)ctx);
    (void)info;
    dnp3_binary_input_t* v;
    while ((v = dnp3_binary_input_iterator_next(iter)) != NULL) {
        goDnp3Binary((uintptr_t)ctx, v->index, v->value, v->flags.value, v->time.value, (int)v->time.quality, rt);
    }
}
static void wrap_handle_double_bit_binary_input(dnp3_header_info_t info, dnp3_double_bit_binary_input_iterator_t* iter, void* ctx) {
    extern int goDnp3GetReadType(uintptr_t handle);
    int rt = goDnp3GetReadType((uintptr_t)ctx);
    (void)info;
    dnp3_double_bit_binary_input_t* v;
    while ((v = dnp3_double_bit_binary_input_iterator_next(iter)) != NULL) {
        goDnp3DoubleBit((uintptr_t)ctx, v->index, (int)v->value, v->flags.value, v->time.value, (int)v->time.quality, rt);
    }
}
static void wrap_handle_binary_output_status(dnp3_header_info_t info, dnp3_binary_output_status_iterator_t* iter, void* ctx) {
    extern int goDnp3GetReadType(uintptr_t handle);
    int rt = goDnp3GetReadType((uintptr_t)ctx);
    (void)info;
    dnp3_binary_output_status_t* v;
    while ((v = dnp3_binary_output_status_iterator_next(iter)) != NULL) {
        goDnp3BinaryOutputStatus((uintptr_t)ctx, v->index, v->value, v->flags.value, v->time.value, (int)v->time.quality, rt);
    }
}
static void wrap_handle_counter(dnp3_header_info_t info, dnp3_counter_iterator_t* iter, void* ctx) {
    extern int goDnp3GetReadType(uintptr_t handle);
    int rt = goDnp3GetReadType((uintptr_t)ctx);
    (void)info;
    dnp3_counter_t* v;
    while ((v = dnp3_counter_iterator_next(iter)) != NULL) {
        goDnp3Counter((uintptr_t)ctx, v->index, v->value, v->flags.value, v->time.value, (int)v->time.quality, rt);
    }
}
static void wrap_handle_frozen_counter(dnp3_header_info_t info, dnp3_frozen_counter_iterator_t* iter, void* ctx) {
    extern int goDnp3GetReadType(uintptr_t handle);
    int rt = goDnp3GetReadType((uintptr_t)ctx);
    (void)info;
    dnp3_frozen_counter_t* v;
    while ((v = dnp3_frozen_counter_iterator_next(iter)) != NULL) {
        goDnp3FrozenCounter((uintptr_t)ctx, v->index, v->value, v->flags.value, v->time.value, (int)v->time.quality, rt);
    }
}
static void wrap_handle_analog_input(dnp3_header_info_t info, dnp3_analog_input_iterator_t* iter, void* ctx) {
    extern int goDnp3GetReadType(uintptr_t handle);
    int rt = goDnp3GetReadType((uintptr_t)ctx);
    (void)info;
    dnp3_analog_input_t* v;
    while ((v = dnp3_analog_input_iterator_next(iter)) != NULL) {
        goDnp3Analog((uintptr_t)ctx, v->index, v->value, v->flags.value, v->time.value, (int)v->time.quality, rt);
    }
}
static void wrap_handle_frozen_analog_input(dnp3_header_info_t info, dnp3_frozen_analog_input_iterator_t* iter, void* ctx) {
    extern int goDnp3GetReadType(uintptr_t handle);
    int rt = goDnp3GetReadType((uintptr_t)ctx);
    (void)info;
    // frozen analog is delivered as a regular analog measurement to Go; we don't
    // distinguish frozen at the gateway since outstation reports them via different
    // groups but the engineering value is the same.
    dnp3_frozen_analog_input_t* v;
    while ((v = dnp3_frozen_analog_input_iterator_next(iter)) != NULL) {
        goDnp3Analog((uintptr_t)ctx, v->index, v->value, v->flags.value, v->time.value, (int)v->time.quality, rt);
    }
}
static void wrap_handle_analog_output_status(dnp3_header_info_t info, dnp3_analog_output_status_iterator_t* iter, void* ctx) {
    extern int goDnp3GetReadType(uintptr_t handle);
    int rt = goDnp3GetReadType((uintptr_t)ctx);
    (void)info;
    dnp3_analog_output_status_t* v;
    while ((v = dnp3_analog_output_status_iterator_next(iter)) != NULL) {
        goDnp3AnalogOutputStatus((uintptr_t)ctx, v->index, v->value, v->flags.value, v->time.value, (int)v->time.quality, rt);
    }
}
static void wrap_handle_octet_string(dnp3_header_info_t info, dnp3_octet_string_iterator_t* iter, void* ctx) {
    extern int goDnp3GetReadType(uintptr_t handle);
    int rt = goDnp3GetReadType((uintptr_t)ctx);
    (void)info;
    dnp3_octet_string_t* v;
    while ((v = dnp3_octet_string_iterator_next(iter)) != NULL) {
        // Drain the byte iterator into a stack buffer (octet strings are at most 255 bytes per DNP3).
        uint8_t buf[256];
        size_t n = 0;
        uint8_t* b;
        while (n < sizeof(buf) && (b = dnp3_byte_iterator_next(v->value)) != NULL) {
            buf[n++] = *b;
        }
        goDnp3OctetString((uintptr_t)ctx, v->index, buf, n, rt);
    }
}

// No-op stubs for callbacks we don't consume. The library calls these regardless;
// providing NULL would crash on some paths.
static void noop_handle_binary_output_command_event(dnp3_header_info_t info, dnp3_binary_output_command_event_iterator_t* iter, void* ctx) { (void)info; (void)iter; (void)ctx; }
static void noop_handle_analog_output_command_event(dnp3_header_info_t info, dnp3_analog_output_command_event_iterator_t* iter, void* ctx) { (void)info; (void)iter; (void)ctx; }
static void noop_handle_unsigned_integer(dnp3_header_info_t info, dnp3_unsigned_integer_iterator_t* iter, void* ctx) { (void)info; (void)iter; (void)ctx; }
static void noop_handle_string_attr(dnp3_header_info_t info, dnp3_string_attr_t a, uint8_t set, uint8_t var, const char* v, void* ctx) { (void)info; (void)a; (void)set; (void)var; (void)v; (void)ctx; }
static void noop_handle_variation_list_attr(dnp3_header_info_t info, dnp3_variation_list_attr_t a, uint8_t set, uint8_t var, dnp3_attr_item_iter_t* it, void* ctx) { (void)info; (void)a; (void)set; (void)var; (void)it; (void)ctx; }
static void noop_handle_uint_attr(dnp3_header_info_t info, dnp3_uint_attr_t a, uint8_t set, uint8_t var, uint32_t v, void* ctx) { (void)info; (void)a; (void)set; (void)var; (void)v; (void)ctx; }
static void noop_handle_bool_attr(dnp3_header_info_t info, dnp3_bool_attr_t a, uint8_t set, uint8_t var, _Bool v, void* ctx) { (void)info; (void)a; (void)set; (void)var; (void)v; (void)ctx; }
static void noop_handle_int_attr(dnp3_header_info_t info, dnp3_int_attr_t a, uint8_t set, uint8_t var, int32_t v, void* ctx) { (void)info; (void)a; (void)set; (void)var; (void)v; (void)ctx; }
static void noop_handle_time_attr(dnp3_header_info_t info, dnp3_time_attr_t a, uint8_t set, uint8_t var, uint64_t v, void* ctx) { (void)info; (void)a; (void)set; (void)var; (void)v; (void)ctx; }
static void noop_handle_float_attr(dnp3_header_info_t info, dnp3_float_attr_t a, uint8_t set, uint8_t var, double v, void* ctx) { (void)info; (void)a; (void)set; (void)var; (void)v; (void)ctx; }
static void noop_handle_octet_string_attr(dnp3_header_info_t info, dnp3_octet_string_attr_t a, uint8_t set, uint8_t var, dnp3_byte_iterator_t* it, void* ctx) { (void)info; (void)a; (void)set; (void)var; (void)it; (void)ctx; }
static void noop_handle_bit_string_attr(dnp3_header_info_t info, dnp3_bit_string_attr_t a, uint8_t set, uint8_t var, dnp3_byte_iterator_t* it, void* ctx) { (void)info; (void)a; (void)set; (void)var; (void)it; (void)ctx; }

// Builds a fully populated read_handler vtable using the wrappers + no-ops above.
static dnp3_read_handler_t build_read_handler(uintptr_t handle) {
    dnp3_read_handler_t h;
    h.begin_fragment = wrap_begin_fragment;
    h.end_fragment = wrap_end_fragment;
    h.handle_binary_input = wrap_handle_binary_input;
    h.handle_double_bit_binary_input = wrap_handle_double_bit_binary_input;
    h.handle_binary_output_status = wrap_handle_binary_output_status;
    h.handle_counter = wrap_handle_counter;
    h.handle_frozen_counter = wrap_handle_frozen_counter;
    h.handle_analog_input = wrap_handle_analog_input;
    h.handle_frozen_analog_input = wrap_handle_frozen_analog_input;
    h.handle_analog_output_status = wrap_handle_analog_output_status;
    h.handle_binary_output_command_event = noop_handle_binary_output_command_event;
    h.handle_analog_output_command_event = noop_handle_analog_output_command_event;
    h.handle_unsigned_integer = noop_handle_unsigned_integer;
    h.handle_octet_string = wrap_handle_octet_string;
    h.handle_string_attr = noop_handle_string_attr;
    h.handle_variation_list_attr = noop_handle_variation_list_attr;
    h.handle_uint_attr = noop_handle_uint_attr;
    h.handle_bool_attr = noop_handle_bool_attr;
    h.handle_int_attr = noop_handle_int_attr;
    h.handle_time_attr = noop_handle_time_attr;
    h.handle_float_attr = noop_handle_float_attr;
    h.handle_octet_string_attr = noop_handle_octet_string_attr;
    h.handle_bit_string_attr = noop_handle_bit_string_attr;
    h.on_destroy = NULL;
    h.ctx = (void*)handle;
    return h;
}

// --- association_handler vtable callbacks --------------------------------

static dnp3_utc_timestamp_t wrap_get_current_time(void* ctx) {
    (void)ctx;
    // We don't drive time-sync (auto_time_sync = NONE), but the lib still calls
    // this. Return invalid to let the lib know we have no opinion.
    return dnp3_utc_timestamp_invalid();
}

static dnp3_association_handler_t build_association_handler(uintptr_t handle) {
    dnp3_association_handler_t h;
    h.get_current_time = wrap_get_current_time;
    h.on_destroy = NULL;
    h.ctx = (void*)handle;
    return h;
}

// --- association_information vtable -- all no-ops for now ----------------
//
// We don't surface task lifecycle (start/success/fail/unsol completion) to the
// gateway yet. The lib has no constructor for this struct, so we build it manually.

static void noop_assoc_task_start(dnp3_task_type_t t, dnp3_function_code_t fc, uint8_t seq, void* ctx) { (void)t; (void)fc; (void)seq; (void)ctx; }
static void noop_assoc_task_success(dnp3_task_type_t t, dnp3_function_code_t fc, uint8_t seq, void* ctx) { (void)t; (void)fc; (void)seq; (void)ctx; }
static void noop_assoc_task_fail(dnp3_task_type_t t, dnp3_task_error_t e, void* ctx) { (void)t; (void)e; (void)ctx; }
static void noop_assoc_unsol_response(_Bool dup, uint8_t seq, void* ctx) { (void)dup; (void)seq; (void)ctx; }

static dnp3_association_information_t build_association_information(uintptr_t handle) {
    dnp3_association_information_t i;
    i.task_start = noop_assoc_task_start;
    i.task_success = noop_assoc_task_success;
    i.task_fail = noop_assoc_task_fail;
    i.unsolicited_response = noop_assoc_unsol_response;
    i.on_destroy = NULL;
    i.ctx = (void*)handle;
    return i;
}

// --- client_state_listener vtable ----------------------------------------

static void wrap_client_state_on_change(dnp3_client_state_t state, void* ctx) {
    goDnp3ClientState((uintptr_t)ctx, (int)state);
}

static dnp3_client_state_listener_t build_client_state_listener(uintptr_t handle) {
    dnp3_client_state_listener_t l;
    l.on_change = wrap_client_state_on_change;
    l.on_destroy = NULL;
    l.ctx = (void*)handle;
    return l;
}

// --- helpers usable from Go ----------------------------------------------

static dnp3_runtime_config_t mk_runtime_config(uint16_t threads) {
    dnp3_runtime_config_t c = dnp3_runtime_config_init();
    c.num_core_threads = threads;
    return c;
}

static dnp3_master_channel_config_t mk_master_channel_config(uint16_t addr, uint16_t tx, uint16_t rx, _Bool verbose) {
    dnp3_master_channel_config_t c = dnp3_master_channel_config_init(addr);
    if (tx >= 249) c.tx_buffer_size = tx;
    if (rx >= 2048) c.rx_buffer_size = rx;
    if (verbose) {
        c.decode_level.application = DNP3_APP_DECODE_LEVEL_OBJECT_VALUES;
        c.decode_level.transport   = DNP3_TRANSPORT_DECODE_LEVEL_HEADER;
        c.decode_level.link        = DNP3_LINK_DECODE_LEVEL_HEADER;
    }
    return c;
}

// --- global logger wiring ------------------------------------------------

static void wrap_logger_on_message(dnp3_log_level_t level, const char* msg, void* ctx) {
    (void)ctx;
    goDnp3LibLog((int)level, (char*)msg);
}

static dnp3_param_error_t install_global_logger(void) {
    dnp3_logging_config_t cfg = dnp3_logging_config_init();
    cfg.level = DNP3_LOG_LEVEL_DEBUG;
    cfg.print_level = true;
    dnp3_logger_t l;
    l.on_message = wrap_logger_on_message;
    l.on_destroy = NULL;
    l.ctx = NULL;
    return dnp3_configure_logging(cfg, l);
}

static dnp3_association_config_t mk_association_config(
    _Bool en_c1, _Bool en_c2, _Bool en_c3,
    _Bool dis_c1, _Bool dis_c2, _Bool dis_c3,
    _Bool integ_c0, _Bool integ_c1, _Bool integ_c2, _Bool integ_c3,
    uint64_t response_timeout_ms,
    uint64_t keep_alive_s)
{
    dnp3_event_classes_t en  = dnp3_event_classes_init(en_c1, en_c2, en_c3);
    dnp3_event_classes_t dis = dnp3_event_classes_init(dis_c1, dis_c2, dis_c3);
    dnp3_classes_t integrity = dnp3_classes_init(integ_c0, integ_c1, integ_c2, integ_c3);
    dnp3_event_classes_t event_scan = dnp3_event_classes_none();
    dnp3_association_config_t cfg = dnp3_association_config_init(dis, en, integrity, event_scan);
    cfg.response_timeout = response_timeout_ms;
    cfg.keep_alive_timeout = keep_alive_s;
    return cfg;
}

static dnp3_connect_strategy_t mk_connect_strategy(void) {
    return dnp3_connect_strategy_init();
}

// Build an "all classes" periodic poll request (class 0+1+2+3 = integrity poll).
static dnp3_request_t* mk_integrity_request(void) {
    return dnp3_request_new_class(true, true, true, true);
}
// Build a single-class periodic poll request.
static dnp3_request_t* mk_class_request(int class_num) {
    return dnp3_request_new_class(false, class_num == 1, class_num == 2, class_num == 3);
}
// Build an "all objects" read for a specific variation (group/variation static read).
// Used to poll outstations that don't flag events; sidesteps the malformed
// Group50 trailers some libraries append to class-0 responses.
static dnp3_request_t* mk_all_objects_request(dnp3_variation_t variation) {
    return dnp3_request_new_all_objects(variation);
}
*/
import "C"

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/cgo"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"goMqttDnp3/config"
)

// ffiMaster is the libdnp3_ffi-backed Master.
type ffiMaster struct {
	h Handler

	rt     *C.dnp3_runtime_t
	rtOnce sync.Once

	mu        sync.Mutex
	assocs    map[string]*assocCtx // outstation ID → context
	started   atomic.Bool
}

// assocCtx is the per-association state pinned via cgo.Handle. The lib's
// ctx void* on every callback for this association points back to the handle.
type assocCtx struct {
	outstationID string
	master       *ffiMaster // for Handler dispatch
	cfg          config.DNP3Outstation

	channel *C.dnp3_master_channel_t
	assocID C.dnp3_association_id_t

	// per-fragment state — written by begin_fragment, read by measurement handlers.
	// Lib guarantees sequential dispatch per fragment per thread, so atomic suffices.
	readType atomic.Int32

	handle cgo.Handle
}

func newMaster(h Handler) Master {
	return &ffiMaster{
		h:      h,
		assocs: make(map[string]*assocCtx),
	}
}

func (m *ffiMaster) ensureRuntime() error {
	var firstErr error
	m.rtOnce.Do(func() {
		// Install lib-level logger once per process (returns error on second call).
		_ = C.install_global_logger()
		cfg := C.mk_runtime_config(0) // 0 = auto-size to CPU count
		err := C.dnp3_runtime_create(cfg, &m.rt)
		if err != C.DNP3_PARAM_ERROR_OK {
			firstErr = paramErr("dnp3_runtime_create", err)
		}
	})
	return firstErr
}

func (m *ffiMaster) AddOutstation(o config.DNP3Outstation) error {
	if m.started.Load() {
		// Adding outstations mid-run requires teardown of the existing channel; not supported yet.
		return errors.New("dnp3_ffi: cannot AddOutstation after Start")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.assocs[o.ID]; exists {
		return fmt.Errorf("dnp3_ffi: outstation %q already added", o.ID)
	}
	ctx := &assocCtx{outstationID: o.ID, master: m, cfg: o}
	ctx.handle = cgo.NewHandle(ctx)
	m.assocs[o.ID] = ctx
	return nil
}

func (m *ffiMaster) RemoveOutstation(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	ctx, ok := m.assocs[id]
	if !ok {
		return nil
	}
	if ctx.channel != nil {
		C.dnp3_master_channel_destroy(ctx.channel)
		ctx.channel = nil
	}
	ctx.handle.Delete()
	delete(m.assocs, id)
	return nil
}

func (m *ffiMaster) Start(_ context.Context) error {
	if m.started.Swap(true) {
		return errors.New("dnp3_ffi: already started")
	}
	if err := m.ensureRuntime(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ctx := range m.assocs {
		if err := m.bringUpLocked(ctx); err != nil {
			return err
		}
	}
	return nil
}

// bringUpLocked must be called with m.mu held. Creates the TCP channel,
// adds the association, registers polls per the outstation config, and enables the channel.
func (m *ffiMaster) bringUpLocked(ctx *assocCtx) error {
	o := ctx.cfg
	addr := fmt.Sprintf("%s:%d", o.Host, defaultPort(o.Port))

	cAddr := C.CString(addr)
	defer C.free(unsafe.Pointer(cAddr))
	endpoints := C.dnp3_endpoint_list_create(cAddr)
	defer C.dnp3_endpoint_list_destroy(endpoints)

	chCfg := C.mk_master_channel_config(C.uint16_t(o.MasterAddress), 2048, 2048, C.bool(true))
	strategy := C.mk_connect_strategy()
	listener := C.build_client_state_listener(C.uintptr_t(ctx.handle))

	var channel *C.dnp3_master_channel_t
	if e := C.dnp3_master_channel_create_tcp(
		m.rt,
		C.DNP3_LINK_ERROR_MODE_CLOSE,
		chCfg,
		endpoints,
		strategy,
		listener,
		&channel,
	); e != C.DNP3_PARAM_ERROR_OK {
		return paramErr("dnp3_master_channel_create_tcp", e)
	}
	ctx.channel = channel

	// Build startup integrity class mask. If none set explicitly we keep the
	// DNP3 default (all classes); otherwise honor exactly what the user picked.
	integ0, integ1, integ2, integ3 := o.IntegrityClass0, o.IntegrityClass1, o.IntegrityClass2, o.IntegrityClass3
	if !o.StartupIntegrity {
		integ0, integ1, integ2, integ3 = false, false, false, false
	} else if !integ0 && !integ1 && !integ2 && !integ3 {
		integ0, integ1, integ2, integ3 = true, true, true, true
	}

	// Build the association.
	assocCfg := C.mk_association_config(
		C.bool(o.UnsolicitedEnabled && o.UnsolicitedClass1),
		C.bool(o.UnsolicitedEnabled && o.UnsolicitedClass2),
		C.bool(o.UnsolicitedEnabled && o.UnsolicitedClass3),
		C.bool(o.DisableUnsolOnStartup),
		C.bool(o.DisableUnsolOnStartup),
		C.bool(o.DisableUnsolOnStartup),
		C.bool(integ0), C.bool(integ1), C.bool(integ2), C.bool(integ3),
		C.uint64_t(defaultMs(o.ResponseTimeoutMs, 5000)),
		C.uint64_t(defaultMs(o.KeepAliveMs, 60000)/1000),
	)
	readHandler := C.build_read_handler(C.uintptr_t(ctx.handle))
	assocHandler := C.build_association_handler(C.uintptr_t(ctx.handle))
	assocInfo := C.build_association_information(C.uintptr_t(ctx.handle))

	if e := C.dnp3_master_channel_add_association(
		channel,
		C.uint16_t(o.OutstationAddress),
		assocCfg,
		readHandler,
		assocHandler,
		assocInfo,
		&ctx.assocID,
	); e != C.DNP3_PARAM_ERROR_OK {
		return paramErr("dnp3_master_channel_add_association", e)
	}

	// Periodic polls.
	if o.IntegrityScanMs > 0 {
		if err := m.addPoll(ctx, C.mk_integrity_request(), o.IntegrityScanMs); err != nil {
			return fmt.Errorf("integrity poll: %w", err)
		}
	}
	if o.Class1ScanMs > 0 {
		if err := m.addPoll(ctx, C.mk_class_request(1), o.Class1ScanMs); err != nil {
			return fmt.Errorf("class1 poll: %w", err)
		}
	}
	if o.Class2ScanMs > 0 {
		if err := m.addPoll(ctx, C.mk_class_request(2), o.Class2ScanMs); err != nil {
			return fmt.Errorf("class2 poll: %w", err)
		}
	}
	if o.Class3ScanMs > 0 {
		if err := m.addPoll(ctx, C.mk_class_request(3), o.Class3ScanMs); err != nil {
			return fmt.Errorf("class3 poll: %w", err)
		}
	}

	// Static (group-specific) polls. One all-objects read per supported point
	// type. The "with flags" variation of each group is used so quality is
	// reported per point.
	if o.StaticPollMs > 0 {
		variations := []C.dnp3_variation_t{
			C.DNP3_VARIATION_GROUP1_VAR2,   // binary input with flags
			C.DNP3_VARIATION_GROUP3_VAR2,   // double-bit binary with flags
			C.DNP3_VARIATION_GROUP10_VAR2,  // binary output status with flags
			C.DNP3_VARIATION_GROUP20_VAR1,  // counter 32-bit with flag
			C.DNP3_VARIATION_GROUP21_VAR1,  // frozen counter 32-bit with flag
			C.DNP3_VARIATION_GROUP30_VAR1,  // analog input 32-bit with flag
			C.DNP3_VARIATION_GROUP40_VAR1,  // analog output status 32-bit with flag
		}
		for _, v := range variations {
			if err := m.addPoll(ctx, C.mk_all_objects_request(v), o.StaticPollMs); err != nil {
				return fmt.Errorf("static poll variation=%d: %w", int(v), err)
			}
		}
	}

	// Enable the channel — kicks off connection + startup integrity.
	if e := C.dnp3_master_channel_enable(channel); e != C.DNP3_PARAM_ERROR_OK {
		return paramErr("dnp3_master_channel_enable", e)
	}
	return nil
}

// addPoll registers a periodic poll request and takes ownership of the request handle.
func (m *ffiMaster) addPoll(ctx *assocCtx, req *C.dnp3_request_t, periodMs int) error {
	var pollID C.dnp3_poll_id_t
	if e := C.dnp3_master_channel_add_poll(ctx.channel, ctx.assocID, req, C.uint64_t(periodMs), &pollID); e != C.DNP3_PARAM_ERROR_OK {
		C.dnp3_request_destroy(req)
		return paramErr("dnp3_master_channel_add_poll", e)
	}
	// add_poll takes ownership of the request per the docs; do NOT destroy here.
	return nil
}

func (m *ffiMaster) Stop() {
	if !m.started.Swap(false) {
		return
	}
	m.mu.Lock()
	for _, ctx := range m.assocs {
		if ctx.channel != nil {
			C.dnp3_master_channel_destroy(ctx.channel)
			ctx.channel = nil
		}
	}
	for id, ctx := range m.assocs {
		ctx.handle.Delete()
		delete(m.assocs, id)
	}
	m.mu.Unlock()
	if m.rt != nil {
		C.dnp3_runtime_set_shutdown_timeout(m.rt, 5)
		C.dnp3_runtime_destroy(m.rt)
		m.rt = nil
	}
}

func (m *ffiMaster) Status() []OutstationStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]OutstationStatus, 0, len(m.assocs))
	for _, ctx := range m.assocs {
		s := ctx.mu()
		s.RLock()
		out = append(out, s.status)
		s.RUnlock()
	}
	return out
}

func (m *ffiMaster) IntegrityPoll(id string) error {
	m.mu.Lock()
	ctx, ok := m.assocs[id]
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("dnp3_ffi: unknown outstation %q", id)
	}
	if ctx.channel == nil {
		return fmt.Errorf("dnp3_ffi: outstation %q not started", id)
	}
	req := C.mk_integrity_request()
	// We use add_poll/demand_poll only for periodic polls. For on-demand reads use channel_read.
	if e := C.dnp3_master_channel_read(ctx.channel, ctx.assocID, req, C.dnp3_read_task_callback_t{}); e != C.DNP3_PARAM_ERROR_OK {
		C.dnp3_request_destroy(req)
		return paramErr("dnp3_master_channel_read", e)
	}
	return nil
}

// --- assocCtx status mirroring ---------------------------------------------

type assocStatusMu struct {
	sync.RWMutex
	status OutstationStatus
}

// statusBacking holds the mu+status for each assocCtx out-of-line to keep the
// callback-facing fields trivially-copyable (cgo struct rules).
var statusBacking sync.Map // *assocCtx → *assocStatusMu

func (a *assocCtx) mu() *assocStatusMu {
	if v, ok := statusBacking.Load(a); ok {
		return v.(*assocStatusMu)
	}
	s := &assocStatusMu{status: OutstationStatus{ID: a.outstationID, Label: a.cfg.Label, Addr: a.cfg.Addr()}}
	actual, _ := statusBacking.LoadOrStore(a, s)
	return actual.(*assocStatusMu)
}

// --- helpers ---------------------------------------------------------------

func defaultPort(p int) int {
	if p == 0 {
		return 20000
	}
	return p
}
func defaultMs(v, def int) int {
	if v <= 0 {
		return def
	}
	return v
}

func paramErr(call string, e C.dnp3_param_error_t) error {
	cs := C.dnp3_param_error_to_string(e)
	return fmt.Errorf("%s: %s (code=%d)", call, C.GoString(cs), int(e))
}

// makeMeasurement builds the gateway-facing Measurement from raw callback fields.
func makeMeasurement(ctx *assocCtx, pt PointType, idx uint16, flags uint8, tsMs uint64, tq int, rt int) Measurement {
	var t time.Time
	// dnp3_time_quality_t: 0=SYNC, 1=UNSYNC, 2=INVALID
	if tq == 2 {
		t = time.Now()
	} else {
		t = time.UnixMilli(int64(tsMs))
	}
	_ = rt
	return Measurement{
		OutstationID: ctx.outstationID,
		PointType:    pt,
		Index:        idx,
		Time:         t,
		Quality:      Quality(flags),
	}
}

// --- //export funcs called by the C wrappers ------------------------------

//export goDnp3SetReadType
func goDnp3SetReadType(handle C.uintptr_t, rt C.int) {
	ctx := cgo.Handle(handle).Value().(*assocCtx)
	ctx.readType.Store(int32(rt))
}

//export goDnp3GetReadType
func goDnp3GetReadType(handle C.uintptr_t) C.int {
	ctx := cgo.Handle(handle).Value().(*assocCtx)
	return C.int(ctx.readType.Load())
}

//export goDnp3ClientState
func goDnp3ClientState(handle C.uintptr_t, state C.int) {
	ctx := cgo.Handle(handle).Value().(*assocCtx)
	connected := state == C.DNP3_CLIENT_STATE_CONNECTED
	s := ctx.mu()
	s.Lock()
	s.status.Connected = connected
	switch state {
	case C.DNP3_CLIENT_STATE_DISABLED:
		s.status.LastError = "disabled"
	case C.DNP3_CLIENT_STATE_CONNECTING:
		s.status.LastError = "connecting"
	case C.DNP3_CLIENT_STATE_CONNECTED:
		s.status.LastError = ""
	case C.DNP3_CLIENT_STATE_WAIT_AFTER_FAILED_CONNECT:
		s.status.LastError = "wait after failed connect"
	case C.DNP3_CLIENT_STATE_WAIT_AFTER_DISCONNECT:
		s.status.LastError = "wait after disconnect"
	case C.DNP3_CLIENT_STATE_SHUTDOWN:
		s.status.LastError = "shutdown"
	}
	snap := s.status
	s.Unlock()
	if ctx.master.h != nil {
		ctx.master.h.OnStatusChange(snap)
	}
}

func deliver(ctx *assocCtx, m Measurement) {
	s := ctx.mu()
	s.Lock()
	s.status.MeasurementsRx++
	s.status.LastReadAt = time.Now()
	s.Unlock()
	if ctx.master.h != nil {
		ctx.master.h.OnMeasurement(m)
	}
}

//export goDnp3Binary
func goDnp3Binary(handle C.uintptr_t, index C.uint16_t, value C.bool, flags C.uint8_t, ts C.uint64_t, tq C.int, rt C.int) {
	ctx := cgo.Handle(handle).Value().(*assocCtx)
	m := makeMeasurement(ctx, PointBinary, uint16(index), uint8(flags), uint64(ts), int(tq), int(rt))
	m.BoolValue = bool(value)
	deliver(ctx, m)
}

//export goDnp3DoubleBit
func goDnp3DoubleBit(handle C.uintptr_t, index C.uint16_t, value C.int, flags C.uint8_t, ts C.uint64_t, tq C.int, rt C.int) {
	ctx := cgo.Handle(handle).Value().(*assocCtx)
	m := makeMeasurement(ctx, PointDoubleBitBinary, uint16(index), uint8(flags), uint64(ts), int(tq), int(rt))
	m.DBBValue = DoubleBitState(value)
	deliver(ctx, m)
}

//export goDnp3BinaryOutputStatus
func goDnp3BinaryOutputStatus(handle C.uintptr_t, index C.uint16_t, value C.bool, flags C.uint8_t, ts C.uint64_t, tq C.int, rt C.int) {
	ctx := cgo.Handle(handle).Value().(*assocCtx)
	m := makeMeasurement(ctx, PointBinaryOutputStatus, uint16(index), uint8(flags), uint64(ts), int(tq), int(rt))
	m.BoolValue = bool(value)
	deliver(ctx, m)
}

//export goDnp3Counter
func goDnp3Counter(handle C.uintptr_t, index C.uint16_t, value C.uint32_t, flags C.uint8_t, ts C.uint64_t, tq C.int, rt C.int) {
	ctx := cgo.Handle(handle).Value().(*assocCtx)
	m := makeMeasurement(ctx, PointCounter, uint16(index), uint8(flags), uint64(ts), int(tq), int(rt))
	m.UintValue = uint32(value)
	deliver(ctx, m)
}

//export goDnp3FrozenCounter
func goDnp3FrozenCounter(handle C.uintptr_t, index C.uint16_t, value C.uint32_t, flags C.uint8_t, ts C.uint64_t, tq C.int, rt C.int) {
	ctx := cgo.Handle(handle).Value().(*assocCtx)
	m := makeMeasurement(ctx, PointFrozenCounter, uint16(index), uint8(flags), uint64(ts), int(tq), int(rt))
	m.UintValue = uint32(value)
	deliver(ctx, m)
}

//export goDnp3Analog
func goDnp3Analog(handle C.uintptr_t, index C.uint16_t, value C.double, flags C.uint8_t, ts C.uint64_t, tq C.int, rt C.int) {
	ctx := cgo.Handle(handle).Value().(*assocCtx)
	m := makeMeasurement(ctx, PointAnalog, uint16(index), uint8(flags), uint64(ts), int(tq), int(rt))
	m.FloatValue = float64(value)
	deliver(ctx, m)
}

//export goDnp3AnalogOutputStatus
func goDnp3AnalogOutputStatus(handle C.uintptr_t, index C.uint16_t, value C.double, flags C.uint8_t, ts C.uint64_t, tq C.int, rt C.int) {
	ctx := cgo.Handle(handle).Value().(*assocCtx)
	m := makeMeasurement(ctx, PointAnalogOutputStatus, uint16(index), uint8(flags), uint64(ts), int(tq), int(rt))
	m.FloatValue = float64(value)
	deliver(ctx, m)
}

//export goDnp3OctetString
func goDnp3OctetString(handle C.uintptr_t, index C.uint16_t, data *C.uint8_t, length C.size_t, rt C.int) {
	ctx := cgo.Handle(handle).Value().(*assocCtx)
	m := makeMeasurement(ctx, PointOctetString, uint16(index), 0, uint64(time.Now().UnixMilli()), 0, int(rt))
	if length > 0 {
		m.BytesValue = C.GoBytes(unsafe.Pointer(data), C.int(length))
	}
	deliver(ctx, m)
}

//export goDnp3LibLog
func goDnp3LibLog(level C.int, msg *C.char) {
	s := C.GoString(msg)
	// dnp3_log_level_t: 0=ERROR 1=WARN 2=INFO 3=DEBUG 4=TRACE
	switch level {
	case 0:
		slog.Error("dnp3-lib: " + s)
	case 1:
		slog.Warn("dnp3-lib: " + s)
	default:
		slog.Info("dnp3-lib: " + s)
	}
}
