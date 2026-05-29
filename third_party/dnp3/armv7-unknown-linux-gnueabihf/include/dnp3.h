// This library is provided under the terms of a non-commercial license.
// 
// Please refer to the source repository for details:
// 
// https://github.com/stepfunc/dnp3/blob/master/LICENSE.txt
// 
// Please contact Step Function I/O if you are interested in commercial license:
// 
// info@stepfunc.io
#pragma once

#ifdef __cplusplus
extern "C" {
#endif

#define DNP3_VERSION_MAJOR 1
#define DNP3_VERSION_MINOR 6
#define DNP3_VERSION_PATCH 0
#define DNP3_VERSION_STRING "1.6.0"

#include <stdbool.h>
#include <stdint.h>

/// @file dnp3.h C API for the dnp3 library
/// 
/// @mainpage
/// 
/// Safe and fast DNP3 library
/// 
/// For complete documentation, see @link dnp3.h @endlink
/// 
/// @section license License
/// 
/// This library is provided under the terms of a non-commercial license.
/// 
/// Please refer to the source repository for details:
/// 
/// https://github.com/stepfunc/dnp3/blob/master/LICENSE.txt
/// 
/// Please contact Step Function I/O if you are interested in commercial license:
/// 
/// info@stepfunc.io

/// @brief By default, TCP_NODELAY is set to true for all client TCP/TLS connections. This disables Nagle's algorithm causing the OS to send data written to socket ASAP without waiting. This reduces latency and is usually the appropriate setting for DNP3. This library always writes data in units of link-layer frames so the default setting might cause more TCP fragmentation if clients send requests that exceed a single link-layer frame
/// 
/// Calling this function will enable Nagle's algorithm for all future outbound TCP connections.
/// 
/// This would typically be called prior to creating any TCP/TLS clients. In a future 2.0 release, this flag will likely be settable on a per-session basis but is done globally to preserve API compatibility
void dnp3_disable_client_tcp_no_delay();

/// @brief By default, TCP_NODELAY is set to true for all server TCP/TLS connections. This disables Nagle's algorithm causing the OS to send data written to socket ASAP without waiting. This reduces latency and is usually the appropriate setting for DNP3. This library always writes data in units of link-layer frames so the default setting might cause more TCP fragmentation if clients send requests that exceed a single link-layer frame
/// 
/// Calling this function will enable Nagle's algorithm for all future TCP/TLS connections accepted by servers
/// 
/// This would typically be called prior to creating any TCP/TLS clients. In a future 2.0 release, this flag will likely be settable on a per-session basis but is done globally to preserve API compatibility
void dnp3_disable_server_tcp_no_delay();


/// @brief Error type used throughout the library
typedef enum dnp3_param_error_t
{
    /// @brief Success, i.e. no error occurred
    DNP3_PARAM_ERROR_OK = 0,
    /// @brief The supplied timeout value is too small or too large
    DNP3_PARAM_ERROR_INVALID_TIMEOUT = 1,
    /// @brief Null parameter
    DNP3_PARAM_ERROR_NULL_PARAMETER = 2,
    /// @brief Provided string argument is not UTF-8
    DNP3_PARAM_ERROR_STRING_NOT_UTF8 = 3,
    /// @brief Native library was compiled without support for this feature
    DNP3_PARAM_ERROR_NO_SUPPORT = 4,
    /// @brief The specified association does not exist
    DNP3_PARAM_ERROR_ASSOCIATION_DOES_NOT_EXIST = 5,
    /// @brief Duplicate association address
    DNP3_PARAM_ERROR_ASSOCIATION_DUPLICATE_ADDRESS = 6,
    /// @brief Invalid socket address
    DNP3_PARAM_ERROR_INVALID_SOCKET_ADDRESS = 7,
    /// @brief Invalid link-layer DNP3 address
    DNP3_PARAM_ERROR_INVALID_DNP3_ADDRESS = 8,
    /// @brief Invalid buffer size
    DNP3_PARAM_ERROR_INVALID_BUFFER_SIZE = 9,
    /// @brief Conflict in the address filter specification
    DNP3_PARAM_ERROR_ADDRESS_FILTER_CONFLICT = 10,
    /// @brief Server already started
    DNP3_PARAM_ERROR_SERVER_ALREADY_STARTED = 11,
    /// @brief Server failed to bind to the specified port
    DNP3_PARAM_ERROR_SERVER_BIND_ERROR = 12,
    /// @brief Master was already shutdown
    DNP3_PARAM_ERROR_MASTER_ALREADY_SHUTDOWN = 13,
    /// @brief Failed to create Tokio runtime
    DNP3_PARAM_ERROR_RUNTIME_CREATION_FAILURE = 14,
    /// @brief Runtime has already been disposed
    DNP3_PARAM_ERROR_RUNTIME_DESTROYED = 15,
    /// @brief Runtime cannot execute blocking call within asynchronous context
    DNP3_PARAM_ERROR_RUNTIME_CANNOT_BLOCK_WITHIN_ASYNC = 16,
    /// @brief Logging can only be configured once
    DNP3_PARAM_ERROR_LOGGING_ALREADY_CONFIGURED = 17,
    /// @brief Point does not exist
    DNP3_PARAM_ERROR_POINT_DOES_NOT_EXIST = 18,
    /// @brief Invalid peer certificate file
    DNP3_PARAM_ERROR_INVALID_PEER_CERTIFICATE = 19,
    /// @brief Invalid local certificate file
    DNP3_PARAM_ERROR_INVALID_LOCAL_CERTIFICATE = 20,
    /// @brief Invalid private key file
    DNP3_PARAM_ERROR_INVALID_PRIVATE_KEY = 21,
    /// @brief Invalid DNS name
    DNP3_PARAM_ERROR_INVALID_DNS_NAME = 22,
    /// @brief Other TLS error
    DNP3_PARAM_ERROR_OTHER_TLS_ERROR = 23,
    /// @brief This operation cannot be performed on this channel type
    DNP3_PARAM_ERROR_WRONG_CHANNEL_TYPE = 24,
    /// @brief This object is consumed and cannot be used again
    DNP3_PARAM_ERROR_CONSUMED = 25,
} dnp3_param_error_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_param_error_to_string(dnp3_param_error_t value)
{
    switch (value)
    {
        case DNP3_PARAM_ERROR_OK: return "ok";
        case DNP3_PARAM_ERROR_INVALID_TIMEOUT: return "invalid_timeout";
        case DNP3_PARAM_ERROR_NULL_PARAMETER: return "null_parameter";
        case DNP3_PARAM_ERROR_STRING_NOT_UTF8: return "string_not_utf8";
        case DNP3_PARAM_ERROR_NO_SUPPORT: return "no_support";
        case DNP3_PARAM_ERROR_ASSOCIATION_DOES_NOT_EXIST: return "association_does_not_exist";
        case DNP3_PARAM_ERROR_ASSOCIATION_DUPLICATE_ADDRESS: return "association_duplicate_address";
        case DNP3_PARAM_ERROR_INVALID_SOCKET_ADDRESS: return "invalid_socket_address";
        case DNP3_PARAM_ERROR_INVALID_DNP3_ADDRESS: return "invalid_dnp3_address";
        case DNP3_PARAM_ERROR_INVALID_BUFFER_SIZE: return "invalid_buffer_size";
        case DNP3_PARAM_ERROR_ADDRESS_FILTER_CONFLICT: return "address_filter_conflict";
        case DNP3_PARAM_ERROR_SERVER_ALREADY_STARTED: return "server_already_started";
        case DNP3_PARAM_ERROR_SERVER_BIND_ERROR: return "server_bind_error";
        case DNP3_PARAM_ERROR_MASTER_ALREADY_SHUTDOWN: return "master_already_shutdown";
        case DNP3_PARAM_ERROR_RUNTIME_CREATION_FAILURE: return "runtime_creation_failure";
        case DNP3_PARAM_ERROR_RUNTIME_DESTROYED: return "runtime_destroyed";
        case DNP3_PARAM_ERROR_RUNTIME_CANNOT_BLOCK_WITHIN_ASYNC: return "runtime_cannot_block_within_async";
        case DNP3_PARAM_ERROR_LOGGING_ALREADY_CONFIGURED: return "logging_already_configured";
        case DNP3_PARAM_ERROR_POINT_DOES_NOT_EXIST: return "point_does_not_exist";
        case DNP3_PARAM_ERROR_INVALID_PEER_CERTIFICATE: return "invalid_peer_certificate";
        case DNP3_PARAM_ERROR_INVALID_LOCAL_CERTIFICATE: return "invalid_local_certificate";
        case DNP3_PARAM_ERROR_INVALID_PRIVATE_KEY: return "invalid_private_key";
        case DNP3_PARAM_ERROR_INVALID_DNS_NAME: return "invalid_dns_name";
        case DNP3_PARAM_ERROR_OTHER_TLS_ERROR: return "other_tls_error";
        case DNP3_PARAM_ERROR_WRONG_CHANNEL_TYPE: return "wrong_channel_type";
        case DNP3_PARAM_ERROR_CONSUMED: return "consumed";
        default: return "unknown param_error value";
    }
}


/// @brief Enumeration received from an outstation in response to command request
typedef enum dnp3_command_status_t
{
    /// @brief command was accepted, initiated, or queued (value == 0)
    DNP3_COMMAND_STATUS_SUCCESS = 0,
    /// @brief command timed out before completing (value == 1)
    DNP3_COMMAND_STATUS_TIMEOUT = 1,
    /// @brief command requires being selected before operate, configuration issue (value == 2)
    DNP3_COMMAND_STATUS_NO_SELECT = 2,
    /// @brief bad control code or timing values (value == 3)
    DNP3_COMMAND_STATUS_FORMAT_ERROR = 3,
    /// @brief command is not implemented (value == 4)
    DNP3_COMMAND_STATUS_NOT_SUPPORTED = 4,
    /// @brief command is all ready in progress or its all ready in that mode (value == 5)
    DNP3_COMMAND_STATUS_ALREADY_ACTIVE = 5,
    /// @brief something is stopping the command, often a local/remote interlock (value == 6)
    DNP3_COMMAND_STATUS_HARDWARE_ERROR = 6,
    /// @brief the function governed by the control is in local only control (value == 7)
    DNP3_COMMAND_STATUS_LOCAL = 7,
    /// @brief the command has been done too often and has been throttled (value == 8)
    DNP3_COMMAND_STATUS_TOO_MANY_OPS = 8,
    /// @brief the command was rejected because the device denied it or an RTU intercepted it (value == 9)
    DNP3_COMMAND_STATUS_NOT_AUTHORIZED = 9,
    /// @brief command not accepted because it was prevented or inhibited by a local automation process, such as interlocking logic or synchrocheck (value == 10)
    DNP3_COMMAND_STATUS_AUTOMATION_INHIBIT = 10,
    /// @brief command not accepted because the device cannot process any more activities than are presently in progress (value == 11)
    DNP3_COMMAND_STATUS_PROCESSING_LIMITED = 11,
    /// @brief command not accepted because the value is outside the acceptable range permitted for this point (value == 12)
    DNP3_COMMAND_STATUS_OUT_OF_RANGE = 12,
    /// @brief command not accepted because the outstation is forwarding the request to another downstream device which reported LOCAL (value == 13)
    DNP3_COMMAND_STATUS_DOWNSTREAM_LOCAL = 13,
    /// @brief command not accepted because the outstation has already completed the requested operation (value == 14)
    DNP3_COMMAND_STATUS_ALREADY_COMPLETE = 14,
    /// @brief command not accepted because the requested function is specifically blocked at the outstation (value == 15)
    DNP3_COMMAND_STATUS_BLOCKED = 15,
    /// @brief command not accepted because the operation was cancelled (value == 16)
    DNP3_COMMAND_STATUS_CANCELED = 16,
    /// @brief command not accepted because another master is communicating with the outstation and has exclusive rights to operate this control point (value == 17)
    DNP3_COMMAND_STATUS_BLOCKED_OTHER_MASTER = 17,
    /// @brief command not accepted because the outstation is forwarding the request to another downstream device which cannot be reached or is otherwise incapable of performing the request (value == 18)
    DNP3_COMMAND_STATUS_DOWNSTREAM_FAIL = 18,
    /// @brief (deprecated) indicates the outstation shall not issue or perform the control operation (value == 126)
    DNP3_COMMAND_STATUS_NON_PARTICIPATING = 19,
    /// @brief captures any value not defined in the enumeration
    DNP3_COMMAND_STATUS_UNKNOWN = 20,
} dnp3_command_status_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_command_status_to_string(dnp3_command_status_t value)
{
    switch (value)
    {
        case DNP3_COMMAND_STATUS_SUCCESS: return "success";
        case DNP3_COMMAND_STATUS_TIMEOUT: return "timeout";
        case DNP3_COMMAND_STATUS_NO_SELECT: return "no_select";
        case DNP3_COMMAND_STATUS_FORMAT_ERROR: return "format_error";
        case DNP3_COMMAND_STATUS_NOT_SUPPORTED: return "not_supported";
        case DNP3_COMMAND_STATUS_ALREADY_ACTIVE: return "already_active";
        case DNP3_COMMAND_STATUS_HARDWARE_ERROR: return "hardware_error";
        case DNP3_COMMAND_STATUS_LOCAL: return "local";
        case DNP3_COMMAND_STATUS_TOO_MANY_OPS: return "too_many_ops";
        case DNP3_COMMAND_STATUS_NOT_AUTHORIZED: return "not_authorized";
        case DNP3_COMMAND_STATUS_AUTOMATION_INHIBIT: return "automation_inhibit";
        case DNP3_COMMAND_STATUS_PROCESSING_LIMITED: return "processing_limited";
        case DNP3_COMMAND_STATUS_OUT_OF_RANGE: return "out_of_range";
        case DNP3_COMMAND_STATUS_DOWNSTREAM_LOCAL: return "downstream_local";
        case DNP3_COMMAND_STATUS_ALREADY_COMPLETE: return "already_complete";
        case DNP3_COMMAND_STATUS_BLOCKED: return "blocked";
        case DNP3_COMMAND_STATUS_CANCELED: return "canceled";
        case DNP3_COMMAND_STATUS_BLOCKED_OTHER_MASTER: return "blocked_other_master";
        case DNP3_COMMAND_STATUS_DOWNSTREAM_FAIL: return "downstream_fail";
        case DNP3_COMMAND_STATUS_NON_PARTICIPATING: return "non_participating";
        case DNP3_COMMAND_STATUS_UNKNOWN: return "unknown";
        default: return "unknown command_status value";
    }
}

/// @brief Object value is 'good' / 'valid' / 'nominal'
#define DNP3_FLAG_ONLINE 0x01
/// @brief Object value has not been updated since device restart
#define DNP3_FLAG_RESTART 0x02
/// @brief Object value represents the last value available before a communication failure occurred. Should never be set by originating devices
#define DNP3_FLAG_COMM_LOST 0x04
/// @brief Object value is overridden in a downstream reporting device
#define DNP3_FLAG_REMOTE_FORCED 0x08
/// @brief Object value is overridden by the device reporting this flag
#define DNP3_FLAG_LOCAL_FORCED 0x10
/// @brief Object value is changing state rapidly (device dependent meaning)
#define DNP3_FLAG_CHATTER_FILTER 0x20
/// @brief Object's true exceeds the measurement range of the reported variation
#define DNP3_FLAG_OVER_RANGE 0x20
/// @brief Reported counter value cannot be compared against a prior value to obtain the correct count difference
#define DNP3_FLAG_DISCONTINUITY 0x40
/// @brief Object's value might not have the expected level of accuracy
#define DNP3_FLAG_REFERENCE_ERR 0x40

/// @brief Controls how transmitted and received application-layer fragments are decoded at the INFO log level
typedef enum dnp3_app_decode_level_t
{
    /// @brief Decode nothing
    DNP3_APP_DECODE_LEVEL_NOTHING = 0,
    /// @brief  Decode the header-only
    DNP3_APP_DECODE_LEVEL_HEADER = 1,
    /// @brief Decode the header and the object headers
    DNP3_APP_DECODE_LEVEL_OBJECT_HEADERS = 2,
    /// @brief Decode the header, the object headers, and the object values
    DNP3_APP_DECODE_LEVEL_OBJECT_VALUES = 3,
} dnp3_app_decode_level_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_app_decode_level_to_string(dnp3_app_decode_level_t value)
{
    switch (value)
    {
        case DNP3_APP_DECODE_LEVEL_NOTHING: return "nothing";
        case DNP3_APP_DECODE_LEVEL_HEADER: return "header";
        case DNP3_APP_DECODE_LEVEL_OBJECT_HEADERS: return "object_headers";
        case DNP3_APP_DECODE_LEVEL_OBJECT_VALUES: return "object_values";
        default: return "unknown app_decode_level value";
    }
}

/// @brief Controls how transmitted and received transport segments are decoded at the INFO log level
typedef enum dnp3_transport_decode_level_t
{
    /// @brief Decode nothing
    DNP3_TRANSPORT_DECODE_LEVEL_NOTHING = 0,
    /// @brief  Decode the header
    DNP3_TRANSPORT_DECODE_LEVEL_HEADER = 1,
    /// @brief Decode the header and the raw payload as hexadecimal
    DNP3_TRANSPORT_DECODE_LEVEL_PAYLOAD = 2,
} dnp3_transport_decode_level_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_transport_decode_level_to_string(dnp3_transport_decode_level_t value)
{
    switch (value)
    {
        case DNP3_TRANSPORT_DECODE_LEVEL_NOTHING: return "nothing";
        case DNP3_TRANSPORT_DECODE_LEVEL_HEADER: return "header";
        case DNP3_TRANSPORT_DECODE_LEVEL_PAYLOAD: return "payload";
        default: return "unknown transport_decode_level value";
    }
}

/// @brief Controls how transmitted and received link frames are decoded at the INFO log level
typedef enum dnp3_link_decode_level_t
{
    /// @brief Decode nothing
    DNP3_LINK_DECODE_LEVEL_NOTHING = 0,
    /// @brief  Decode the header
    DNP3_LINK_DECODE_LEVEL_HEADER = 1,
    /// @brief Decode the header and the raw payload as hexadecimal
    DNP3_LINK_DECODE_LEVEL_PAYLOAD = 2,
} dnp3_link_decode_level_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_link_decode_level_to_string(dnp3_link_decode_level_t value)
{
    switch (value)
    {
        case DNP3_LINK_DECODE_LEVEL_NOTHING: return "nothing";
        case DNP3_LINK_DECODE_LEVEL_HEADER: return "header";
        case DNP3_LINK_DECODE_LEVEL_PAYLOAD: return "payload";
        default: return "unknown link_decode_level value";
    }
}

/// @brief Controls how data transmitted at the physical layer (TCP, serial, etc) is logged
typedef enum dnp3_phys_decode_level_t
{
    /// @brief Log nothing
    DNP3_PHYS_DECODE_LEVEL_NOTHING = 0,
    /// @brief Log only the length of data that is sent and received
    DNP3_PHYS_DECODE_LEVEL_LENGTH = 1,
    /// @brief Log the length and the actual data that is sent and received
    DNP3_PHYS_DECODE_LEVEL_DATA = 2,
} dnp3_phys_decode_level_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_phys_decode_level_to_string(dnp3_phys_decode_level_t value)
{
    switch (value)
    {
        case DNP3_PHYS_DECODE_LEVEL_NOTHING: return "nothing";
        case DNP3_PHYS_DECODE_LEVEL_LENGTH: return "length";
        case DNP3_PHYS_DECODE_LEVEL_DATA: return "data";
        default: return "unknown phys_decode_level value";
    }
}

typedef struct dnp3_decode_level_t dnp3_decode_level_t;

/// @brief Controls the decoding of transmitted and received data at the application, transport, link, and physical layers
typedef struct dnp3_decode_level_t
{
    /// @brief Controls application fragment decoding
    dnp3_app_decode_level_t application;
    /// @brief Controls transport segment layer decoding
    dnp3_transport_decode_level_t transport;
    /// @brief Controls link frame decoding
    dnp3_link_decode_level_t link;
    /// @brief Controls the logging of physical layer read/write
    dnp3_phys_decode_level_t physical;
} dnp3_decode_level_t;

/// @brief Initialize log levels to defaults
/// 
/// @note Values are initialized to:
/// - @ref dnp3_decode_level_t.application : @ref DNP3_APP_DECODE_LEVEL_NOTHING
/// - @ref dnp3_decode_level_t.transport : @ref DNP3_TRANSPORT_DECODE_LEVEL_NOTHING
/// - @ref dnp3_decode_level_t.link : @ref DNP3_LINK_DECODE_LEVEL_NOTHING
/// - @ref dnp3_decode_level_t.physical : @ref DNP3_PHYS_DECODE_LEVEL_NOTHING
/// 
/// @returns New instance of @ref dnp3_decode_level_t
static dnp3_decode_level_t dnp3_decode_level_init()
{
    dnp3_decode_level_t _return_value = {
        DNP3_APP_DECODE_LEVEL_NOTHING,
        DNP3_TRANSPORT_DECODE_LEVEL_NOTHING,
        DNP3_LINK_DECODE_LEVEL_NOTHING,
        DNP3_PHYS_DECODE_LEVEL_NOTHING
    };
    return _return_value;
}

/// @brief Initialize log levels to nothing
/// 
/// @note Values are initialized to:
/// - @ref dnp3_decode_level_t.application : @ref DNP3_APP_DECODE_LEVEL_NOTHING
/// - @ref dnp3_decode_level_t.transport : @ref DNP3_TRANSPORT_DECODE_LEVEL_NOTHING
/// - @ref dnp3_decode_level_t.link : @ref DNP3_LINK_DECODE_LEVEL_NOTHING
/// - @ref dnp3_decode_level_t.physical : @ref DNP3_PHYS_DECODE_LEVEL_NOTHING
/// 
/// @returns New instance of @ref dnp3_decode_level_t
static dnp3_decode_level_t dnp3_decode_level_nothing()
{
    dnp3_decode_level_t _return_value = {
        DNP3_APP_DECODE_LEVEL_NOTHING,
        DNP3_TRANSPORT_DECODE_LEVEL_NOTHING,
        DNP3_LINK_DECODE_LEVEL_NOTHING,
        DNP3_PHYS_DECODE_LEVEL_NOTHING
    };
    return _return_value;
}


/// @brief Handle to the underlying runtime
typedef struct dnp3_runtime_t dnp3_runtime_t;

typedef struct dnp3_runtime_config_t dnp3_runtime_config_t;

/// @brief Runtime configuration
typedef struct dnp3_runtime_config_t
{
    /// @brief Number of runtime threads to spawn. For a guess of the number of CPU cores, use 0.
    /// 
    /// Even if tons of connections are expected, it is preferred to use a value around the number of CPU cores for better performances. The library uses an efficient thread pool polling mechanism.
    uint16_t num_core_threads;
} dnp3_runtime_config_t;

/// @brief Initialize the configuration to default values
/// 
/// @note Values are initialized to:
/// - @ref dnp3_runtime_config_t.num_core_threads : 0
/// 
/// @returns New instance of @ref dnp3_runtime_config_t
static dnp3_runtime_config_t dnp3_runtime_config_init()
{
    dnp3_runtime_config_t _return_value = {
        0
    };
    return _return_value;
}


/// @brief Creates a new runtime for running the protocol stack.
/// 
/// @warning The runtime should be kept alive for as long as it's needed and it should be released with @ref dnp3_runtime_destroy
/// @param config Runtime configuration
/// @param out Instance of @ref dnp3_runtime_t
/// @return Error code
dnp3_param_error_t dnp3_runtime_create(dnp3_runtime_config_t config, dnp3_runtime_t** out);

/// @brief Destroy a runtime.
/// 
/// This method will gracefully wait for all asynchronous operation to end before returning
/// @param instance Instance of @ref dnp3_runtime_t to destroy
void dnp3_runtime_destroy(dnp3_runtime_t* instance);

/// @brief By default, when the runtime shuts down, it does so without a timeout and waits indefinitely for all spawned tasks to yield.
/// 
/// Setting this value will put a maximum time bound on the eventual shutdown. Threads that have not exited within this timeout will be terminated.
/// 
/// @warning This can leak memory. This method should only be used if the the entire application is being shut down so that memory can be cleaned up by the OS.
/// @param instance Instance of @ref dnp3_runtime_t
/// @param timeout Maximum number of seconds to wait for the runtime to shut down (seconds)
void dnp3_runtime_set_shutdown_timeout(dnp3_runtime_t* instance, uint64_t timeout);


typedef struct dnp3_control_field_t dnp3_control_field_t;

/// @brief APDU Control field
typedef struct dnp3_control_field_t
{
    /// @brief First fragment in the message
    bool fir;
    /// @brief Final fragment of the message
    bool fin;
    /// @brief Requires confirmation
    bool con;
    /// @brief Unsolicited response
    bool uns;
    /// @brief Sequence number
    uint8_t seq;
} dnp3_control_field_t;


/// @brief Trip-Close Code field, used in conjunction with @ref dnp3_op_type_t to specify a control operation
typedef enum dnp3_trip_close_code_t
{
    /// @brief NUL (0)
    DNP3_TRIP_CLOSE_CODE_NUL = 0,
    /// @brief CLOSE (1)
    DNP3_TRIP_CLOSE_CODE_CLOSE = 1,
    /// @brief TRIP (2)
    DNP3_TRIP_CLOSE_CODE_TRIP = 2,
    /// @brief RESERVED (3)
    DNP3_TRIP_CLOSE_CODE_RESERVED = 3,
} dnp3_trip_close_code_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_trip_close_code_to_string(dnp3_trip_close_code_t value)
{
    switch (value)
    {
        case DNP3_TRIP_CLOSE_CODE_NUL: return "nul";
        case DNP3_TRIP_CLOSE_CODE_CLOSE: return "close";
        case DNP3_TRIP_CLOSE_CODE_TRIP: return "trip";
        case DNP3_TRIP_CLOSE_CODE_RESERVED: return "reserved";
        default: return "unknown trip_close_code value";
    }
}

/// @brief Operation Type field, used in conjunction with @ref dnp3_trip_close_code_t to specify a control operation
typedef enum dnp3_op_type_t
{
    /// @brief NUL (0)
    DNP3_OP_TYPE_NUL = 0,
    /// @brief PULSE_ON (1)
    DNP3_OP_TYPE_PULSE_ON = 1,
    /// @brief PULSE_OFF (2)
    DNP3_OP_TYPE_PULSE_OFF = 2,
    /// @brief LATCH_ON (3)
    DNP3_OP_TYPE_LATCH_ON = 3,
    /// @brief LATCH_OFF(4)
    DNP3_OP_TYPE_LATCH_OFF = 4,
} dnp3_op_type_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_op_type_to_string(dnp3_op_type_t value)
{
    switch (value)
    {
        case DNP3_OP_TYPE_NUL: return "nul";
        case DNP3_OP_TYPE_PULSE_ON: return "pulse_on";
        case DNP3_OP_TYPE_PULSE_OFF: return "pulse_off";
        case DNP3_OP_TYPE_LATCH_ON: return "latch_on";
        case DNP3_OP_TYPE_LATCH_OFF: return "latch_off";
        default: return "unknown op_type value";
    }
}

typedef struct dnp3_control_code_t dnp3_control_code_t;

/// @brief CROB (@ref dnp3_group12_var1_t) control code
typedef struct dnp3_control_code_t
{
    /// @brief This field is used in conjunction with @ref dnp3_control_code_t.op_type to specify a control operation
    dnp3_trip_close_code_t tcc;
    /// @brief Support for this field is optional. When the clear bit is set, the device shall remove pending control commands for that index and stop any control operation that is in progress for that index. The indexed point shall go to the state that it would have if the command were allowed to complete normally.
    bool clear;
    /// @brief This field is obsolete and should always be 0
    bool queue;
    /// @brief This field is used in conjunction with the @ref dnp3_control_code_t.tcc field to specify a control operation
    dnp3_op_type_t op_type;
} dnp3_control_code_t;

/// @brief Initialize a @ref dnp3_control_code_t instance
/// 
/// @note Values are initialized to:
/// - @ref dnp3_control_code_t.queue : @p false
/// 
/// @param tcc This field is used in conjunction with @ref dnp3_control_code_t.op_type to specify a control operation
/// @param clear Support for this field is optional. When the clear bit is set, the device shall remove pending control commands for that index and stop any control operation that is in progress for that index. The indexed point shall go to the state that it would have if the command were allowed to complete normally.
/// @param op_type This field is used in conjunction with the @ref dnp3_control_code_t.tcc field to specify a control operation
/// @returns New instance of @ref dnp3_control_code_t
static dnp3_control_code_t dnp3_control_code_init(dnp3_trip_close_code_t tcc, bool clear, dnp3_op_type_t op_type)
{
    dnp3_control_code_t _return_value = {
        tcc,
        clear,
        false,
        op_type
    };
    return _return_value;
}

/// @brief Initialize a @ref dnp3_control_code_t instance from a @ref dnp3_op_type_t
/// 
/// @note Values are initialized to:
/// - @ref dnp3_control_code_t.tcc : @ref DNP3_TRIP_CLOSE_CODE_NUL
/// - @ref dnp3_control_code_t.clear : @p false
/// - @ref dnp3_control_code_t.queue : @p false
/// 
/// @param op_type This field is used in conjunction with the @ref dnp3_control_code_t.tcc field to specify a control operation
/// @returns New instance of @ref dnp3_control_code_t
static dnp3_control_code_t dnp3_control_code_from_op_type(dnp3_op_type_t op_type)
{
    dnp3_control_code_t _return_value = {
        DNP3_TRIP_CLOSE_CODE_NUL,
        false,
        false,
        op_type
    };
    return _return_value;
}

/// @brief Initialize a @ref dnp3_control_code_t instance from a @ref dnp3_trip_close_code_t and a @ref dnp3_op_type_t.
/// 
/// @note Values are initialized to:
/// - @ref dnp3_control_code_t.clear : @p false
/// - @ref dnp3_control_code_t.queue : @p false
/// 
/// @param tcc This field is used in conjunction with @ref dnp3_control_code_t.op_type to specify a control operation
/// @param op_type This field is used in conjunction with the @ref dnp3_control_code_t.tcc field to specify a control operation
/// @returns New instance of @ref dnp3_control_code_t
static dnp3_control_code_t dnp3_control_code_from_tcc_and_op_type(dnp3_trip_close_code_t tcc, dnp3_op_type_t op_type)
{
    dnp3_control_code_t _return_value = {
        tcc,
        false,
        false,
        op_type
    };
    return _return_value;
}


typedef struct dnp3_group12_var1_t dnp3_group12_var1_t;

/// @brief Control Relay Output Block
typedef struct dnp3_group12_var1_t
{
    /// @brief Control code
    dnp3_control_code_t code;
    /// @brief Count
    uint8_t count;
    /// @brief Duration the output drive remains active (in milliseconds)
    uint32_t on_time;
    /// @brief Duration the output drive remains non-active (in milliseconds)
    uint32_t off_time;
} dnp3_group12_var1_t;

/// @brief Fully construct @ref dnp3_group12_var1_t specifying the value of each field
/// @param code Control code
/// @param count Count
/// @param on_time Duration the output drive remains active (in milliseconds)
/// @param off_time Duration the output drive remains non-active (in milliseconds)
/// @returns New instance of @ref dnp3_group12_var1_t
static dnp3_group12_var1_t dnp3_group12_var1_init(dnp3_control_code_t code, uint8_t count, uint32_t on_time, uint32_t off_time)
{
    dnp3_group12_var1_t _return_value = {
        code,
        count,
        on_time,
        off_time
    };
    return _return_value;
}

/// @brief Construct a @ref dnp3_group12_var1_t from a @ref dnp3_control_code_t.
/// 
/// @note Values are initialized to:
/// - @ref dnp3_group12_var1_t.count : 1
/// - @ref dnp3_group12_var1_t.on_time : 1000
/// - @ref dnp3_group12_var1_t.off_time : 1000
/// 
/// @param code Control code
/// @returns New instance of @ref dnp3_group12_var1_t
static dnp3_group12_var1_t dnp3_group12_var1_from_code(dnp3_control_code_t code)
{
    dnp3_group12_var1_t _return_value = {
        code,
        1,
        1000,
        1000
    };
    return _return_value;
}


typedef struct dnp3_flags_t dnp3_flags_t;

/// @brief Collection of individual flag bits represented by an underlying mask value
typedef struct dnp3_flags_t
{
    /// @brief bit-mask representing a set of individual flag bits
    uint8_t value;
} dnp3_flags_t;

/// @brief Fully construct @ref dnp3_flags_t specifying the value of each field
/// @param value bit-mask representing a set of individual flag bits
/// @returns New instance of @ref dnp3_flags_t
static dnp3_flags_t dnp3_flags_init(uint8_t value)
{
    dnp3_flags_t _return_value = {
        value
    };
    return _return_value;
}


/// @brief Timestamp quality
typedef enum dnp3_time_quality_t
{
    /// @brief The timestamp is UTC synchronized at the remote device
    DNP3_TIME_QUALITY_SYNCHRONIZED_TIME = 0,
    /// @brief The device indicates the timestamp may be not be synchronized
    DNP3_TIME_QUALITY_UNSYNCHRONIZED_TIME = 1,
    /// @brief Timestamp is not valid, ignore the value and use a local timestamp
    DNP3_TIME_QUALITY_INVALID_TIME = 2,
} dnp3_time_quality_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_time_quality_to_string(dnp3_time_quality_t value)
{
    switch (value)
    {
        case DNP3_TIME_QUALITY_SYNCHRONIZED_TIME: return "synchronized_time";
        case DNP3_TIME_QUALITY_UNSYNCHRONIZED_TIME: return "unsynchronized_time";
        case DNP3_TIME_QUALITY_INVALID_TIME: return "invalid_time";
        default: return "unknown time_quality value";
    }
}

typedef struct dnp3_timestamp_t dnp3_timestamp_t;

/// @brief Timestamp associated with particular measurement from the outstation. The validity of the value depends on the quality.
typedef struct dnp3_timestamp_t
{
    /// @brief Count of milliseconds since UNIX epoch
    /// 
    /// @warning Only the lower 48-bits are used in DNP3 timestamps and time synchronization
    uint64_t value;
    /// @brief Enumeration that indicates the timestamp's validity
    dnp3_time_quality_t quality;
} dnp3_timestamp_t;

/// @brief Creates an invalid timestamp struct
/// 
/// @note Values are initialized to:
/// - @ref dnp3_timestamp_t.value : 0
/// - @ref dnp3_timestamp_t.quality : @ref DNP3_TIME_QUALITY_INVALID_TIME
/// 
/// @returns New instance of @ref dnp3_timestamp_t
static dnp3_timestamp_t dnp3_timestamp_invalid_timestamp()
{
    dnp3_timestamp_t _return_value = {
        0,
        DNP3_TIME_QUALITY_INVALID_TIME
    };
    return _return_value;
}

/// @brief Creates a synchronized timestamp struct
/// 
/// @note Values are initialized to:
/// - @ref dnp3_timestamp_t.quality : @ref DNP3_TIME_QUALITY_SYNCHRONIZED_TIME
/// 
/// @param value Count of milliseconds since UNIX epoch
/// @returns New instance of @ref dnp3_timestamp_t
static dnp3_timestamp_t dnp3_timestamp_synchronized_timestamp(uint64_t value)
{
    dnp3_timestamp_t _return_value = {
        value,
        DNP3_TIME_QUALITY_SYNCHRONIZED_TIME
    };
    return _return_value;
}

/// @brief Creates an unsynchronized timestamp struct
/// 
/// @note Values are initialized to:
/// - @ref dnp3_timestamp_t.quality : @ref DNP3_TIME_QUALITY_UNSYNCHRONIZED_TIME
/// 
/// @param value Count of milliseconds since UNIX epoch
/// @returns New instance of @ref dnp3_timestamp_t
static dnp3_timestamp_t dnp3_timestamp_unsynchronized_timestamp(uint64_t value)
{
    dnp3_timestamp_t _return_value = {
        value,
        DNP3_TIME_QUALITY_UNSYNCHRONIZED_TIME
    };
    return _return_value;
}


/// @brief Double-bit binary input value
typedef enum dnp3_double_bit_t
{
    /// @brief Transition between conditions
    DNP3_DOUBLE_BIT_INTERMEDIATE = 0,
    /// @brief Determined to be OFF
    DNP3_DOUBLE_BIT_DETERMINED_OFF = 1,
    /// @brief Determined to be ON
    DNP3_DOUBLE_BIT_DETERMINED_ON = 2,
    /// @brief Abnormal or custom condition
    DNP3_DOUBLE_BIT_INDETERMINATE = 3,
} dnp3_double_bit_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_double_bit_to_string(dnp3_double_bit_t value)
{
    switch (value)
    {
        case DNP3_DOUBLE_BIT_INTERMEDIATE: return "intermediate";
        case DNP3_DOUBLE_BIT_DETERMINED_OFF: return "determined_off";
        case DNP3_DOUBLE_BIT_DETERMINED_ON: return "determined_on";
        case DNP3_DOUBLE_BIT_INDETERMINATE: return "indeterminate";
        default: return "unknown double_bit value";
    }
}

typedef struct dnp3_binary_input_t dnp3_binary_input_t;

/// @brief binary_input point
typedef struct dnp3_binary_input_t
{
    /// @brief Point index
    uint16_t index;
    /// @brief Point value
    bool value;
    /// @brief Point flags
    dnp3_flags_t flags;
    /// @brief Point timestamp
    dnp3_timestamp_t time;
} dnp3_binary_input_t;

/// @brief Fully construct @ref dnp3_binary_input_t specifying the value of each field
/// @param index Point index
/// @param value Point value
/// @param flags Point flags
/// @param time Point timestamp
/// @returns New instance of @ref dnp3_binary_input_t
static dnp3_binary_input_t dnp3_binary_input_init(uint16_t index, bool value, dnp3_flags_t flags, dnp3_timestamp_t time)
{
    dnp3_binary_input_t _return_value = {
        index,
        value,
        flags,
        time
    };
    return _return_value;
}


/// @brief Iterator of binary_input
typedef struct dnp3_binary_input_iterator_t dnp3_binary_input_iterator_t;

/// @brief returns a pointer to the next value or NULL
/// @param iter opaque iterator on which to retrieve the next value
/// @return next value or NULL
dnp3_binary_input_t* dnp3_binary_input_iterator_next(dnp3_binary_input_iterator_t* iter);


typedef struct dnp3_double_bit_binary_input_t dnp3_double_bit_binary_input_t;

/// @brief double_bit_binary_input point
typedef struct dnp3_double_bit_binary_input_t
{
    /// @brief Point index
    uint16_t index;
    /// @brief Point value
    dnp3_double_bit_t value;
    /// @brief Point flags
    dnp3_flags_t flags;
    /// @brief Point timestamp
    dnp3_timestamp_t time;
} dnp3_double_bit_binary_input_t;

/// @brief Fully construct @ref dnp3_double_bit_binary_input_t specifying the value of each field
/// @param index Point index
/// @param value Point value
/// @param flags Point flags
/// @param time Point timestamp
/// @returns New instance of @ref dnp3_double_bit_binary_input_t
static dnp3_double_bit_binary_input_t dnp3_double_bit_binary_input_init(uint16_t index, dnp3_double_bit_t value, dnp3_flags_t flags, dnp3_timestamp_t time)
{
    dnp3_double_bit_binary_input_t _return_value = {
        index,
        value,
        flags,
        time
    };
    return _return_value;
}


/// @brief Iterator of double_bit_binary_input
typedef struct dnp3_double_bit_binary_input_iterator_t dnp3_double_bit_binary_input_iterator_t;

/// @brief returns a pointer to the next value or NULL
/// @param iter opaque iterator on which to retrieve the next value
/// @return next value or NULL
dnp3_double_bit_binary_input_t* dnp3_double_bit_binary_input_iterator_next(dnp3_double_bit_binary_input_iterator_t* iter);


typedef struct dnp3_binary_output_status_t dnp3_binary_output_status_t;

/// @brief binary_output_status point
typedef struct dnp3_binary_output_status_t
{
    /// @brief Point index
    uint16_t index;
    /// @brief Point value
    bool value;
    /// @brief Point flags
    dnp3_flags_t flags;
    /// @brief Point timestamp
    dnp3_timestamp_t time;
} dnp3_binary_output_status_t;

/// @brief Fully construct @ref dnp3_binary_output_status_t specifying the value of each field
/// @param index Point index
/// @param value Point value
/// @param flags Point flags
/// @param time Point timestamp
/// @returns New instance of @ref dnp3_binary_output_status_t
static dnp3_binary_output_status_t dnp3_binary_output_status_init(uint16_t index, bool value, dnp3_flags_t flags, dnp3_timestamp_t time)
{
    dnp3_binary_output_status_t _return_value = {
        index,
        value,
        flags,
        time
    };
    return _return_value;
}


/// @brief Iterator of binary_output_status
typedef struct dnp3_binary_output_status_iterator_t dnp3_binary_output_status_iterator_t;

/// @brief returns a pointer to the next value or NULL
/// @param iter opaque iterator on which to retrieve the next value
/// @return next value or NULL
dnp3_binary_output_status_t* dnp3_binary_output_status_iterator_next(dnp3_binary_output_status_iterator_t* iter);


typedef struct dnp3_counter_t dnp3_counter_t;

/// @brief counter point
typedef struct dnp3_counter_t
{
    /// @brief Point index
    uint16_t index;
    /// @brief Point value
    uint32_t value;
    /// @brief Point flags
    dnp3_flags_t flags;
    /// @brief Point timestamp
    dnp3_timestamp_t time;
} dnp3_counter_t;

/// @brief Fully construct @ref dnp3_counter_t specifying the value of each field
/// @param index Point index
/// @param value Point value
/// @param flags Point flags
/// @param time Point timestamp
/// @returns New instance of @ref dnp3_counter_t
static dnp3_counter_t dnp3_counter_init(uint16_t index, uint32_t value, dnp3_flags_t flags, dnp3_timestamp_t time)
{
    dnp3_counter_t _return_value = {
        index,
        value,
        flags,
        time
    };
    return _return_value;
}


/// @brief Iterator of counter
typedef struct dnp3_counter_iterator_t dnp3_counter_iterator_t;

/// @brief returns a pointer to the next value or NULL
/// @param iter opaque iterator on which to retrieve the next value
/// @return next value or NULL
dnp3_counter_t* dnp3_counter_iterator_next(dnp3_counter_iterator_t* iter);


typedef struct dnp3_frozen_counter_t dnp3_frozen_counter_t;

/// @brief frozen_counter point
typedef struct dnp3_frozen_counter_t
{
    /// @brief Point index
    uint16_t index;
    /// @brief Point value
    uint32_t value;
    /// @brief Point flags
    dnp3_flags_t flags;
    /// @brief Point timestamp
    dnp3_timestamp_t time;
} dnp3_frozen_counter_t;

/// @brief Fully construct @ref dnp3_frozen_counter_t specifying the value of each field
/// @param index Point index
/// @param value Point value
/// @param flags Point flags
/// @param time Point timestamp
/// @returns New instance of @ref dnp3_frozen_counter_t
static dnp3_frozen_counter_t dnp3_frozen_counter_init(uint16_t index, uint32_t value, dnp3_flags_t flags, dnp3_timestamp_t time)
{
    dnp3_frozen_counter_t _return_value = {
        index,
        value,
        flags,
        time
    };
    return _return_value;
}


/// @brief Iterator of frozen_counter
typedef struct dnp3_frozen_counter_iterator_t dnp3_frozen_counter_iterator_t;

/// @brief returns a pointer to the next value or NULL
/// @param iter opaque iterator on which to retrieve the next value
/// @return next value or NULL
dnp3_frozen_counter_t* dnp3_frozen_counter_iterator_next(dnp3_frozen_counter_iterator_t* iter);


typedef struct dnp3_analog_input_t dnp3_analog_input_t;

/// @brief analog_input point
typedef struct dnp3_analog_input_t
{
    /// @brief Point index
    uint16_t index;
    /// @brief Point value
    double value;
    /// @brief Point flags
    dnp3_flags_t flags;
    /// @brief Point timestamp
    dnp3_timestamp_t time;
} dnp3_analog_input_t;

/// @brief Fully construct @ref dnp3_analog_input_t specifying the value of each field
/// @param index Point index
/// @param value Point value
/// @param flags Point flags
/// @param time Point timestamp
/// @returns New instance of @ref dnp3_analog_input_t
static dnp3_analog_input_t dnp3_analog_input_init(uint16_t index, double value, dnp3_flags_t flags, dnp3_timestamp_t time)
{
    dnp3_analog_input_t _return_value = {
        index,
        value,
        flags,
        time
    };
    return _return_value;
}


/// @brief Iterator of analog_input
typedef struct dnp3_analog_input_iterator_t dnp3_analog_input_iterator_t;

/// @brief returns a pointer to the next value or NULL
/// @param iter opaque iterator on which to retrieve the next value
/// @return next value or NULL
dnp3_analog_input_t* dnp3_analog_input_iterator_next(dnp3_analog_input_iterator_t* iter);


typedef struct dnp3_frozen_analog_input_t dnp3_frozen_analog_input_t;

/// @brief frozen_analog_input point
typedef struct dnp3_frozen_analog_input_t
{
    /// @brief Point index
    uint16_t index;
    /// @brief Point value
    double value;
    /// @brief Point flags
    dnp3_flags_t flags;
    /// @brief Point timestamp
    dnp3_timestamp_t time;
} dnp3_frozen_analog_input_t;

/// @brief Fully construct @ref dnp3_frozen_analog_input_t specifying the value of each field
/// @param index Point index
/// @param value Point value
/// @param flags Point flags
/// @param time Point timestamp
/// @returns New instance of @ref dnp3_frozen_analog_input_t
static dnp3_frozen_analog_input_t dnp3_frozen_analog_input_init(uint16_t index, double value, dnp3_flags_t flags, dnp3_timestamp_t time)
{
    dnp3_frozen_analog_input_t _return_value = {
        index,
        value,
        flags,
        time
    };
    return _return_value;
}


/// @brief Iterator of frozen_analog_input
typedef struct dnp3_frozen_analog_input_iterator_t dnp3_frozen_analog_input_iterator_t;

/// @brief returns a pointer to the next value or NULL
/// @param iter opaque iterator on which to retrieve the next value
/// @return next value or NULL
dnp3_frozen_analog_input_t* dnp3_frozen_analog_input_iterator_next(dnp3_frozen_analog_input_iterator_t* iter);


typedef struct dnp3_analog_output_status_t dnp3_analog_output_status_t;

/// @brief analog_output_status point
typedef struct dnp3_analog_output_status_t
{
    /// @brief Point index
    uint16_t index;
    /// @brief Point value
    double value;
    /// @brief Point flags
    dnp3_flags_t flags;
    /// @brief Point timestamp
    dnp3_timestamp_t time;
} dnp3_analog_output_status_t;

/// @brief Fully construct @ref dnp3_analog_output_status_t specifying the value of each field
/// @param index Point index
/// @param value Point value
/// @param flags Point flags
/// @param time Point timestamp
/// @returns New instance of @ref dnp3_analog_output_status_t
static dnp3_analog_output_status_t dnp3_analog_output_status_init(uint16_t index, double value, dnp3_flags_t flags, dnp3_timestamp_t time)
{
    dnp3_analog_output_status_t _return_value = {
        index,
        value,
        flags,
        time
    };
    return _return_value;
}


/// @brief Iterator of analog_output_status
typedef struct dnp3_analog_output_status_iterator_t dnp3_analog_output_status_iterator_t;

/// @brief returns a pointer to the next value or NULL
/// @param iter opaque iterator on which to retrieve the next value
/// @return next value or NULL
dnp3_analog_output_status_t* dnp3_analog_output_status_iterator_next(dnp3_analog_output_status_iterator_t* iter);


typedef struct dnp3_binary_output_command_event_t dnp3_binary_output_command_event_t;

/// @brief Event transferred from master to outstation when the outstation receives a corresponding command.
/// 
/// Maps to group 13 variations 1 and 2.
/// 
/// These objects are part of subset level 4 and are not commonly used.
typedef struct dnp3_binary_output_command_event_t
{
    /// @brief Index of the binary command event
    uint16_t index;
    /// @brief Status from processing the command that triggered this event
    dnp3_command_status_t status;
    /// @brief Commanded state of the binary output
    /// 
    /// From the spec:  0 = Latch Off / Trip / NULL, 1 = Latch On / Close. Where the commanded state is unknown, the commanded state flag shall be 0.
    bool commanded_state;
    /// @brief Associated timestamp
    dnp3_timestamp_t time;
} dnp3_binary_output_command_event_t;

/// @brief Fully construct @ref dnp3_binary_output_command_event_t specifying the value of each field
/// @param index Index of the binary command event
/// @param status Status from processing the command that triggered this event
/// @param commanded_state Commanded state of the binary output
/// @param time Associated timestamp
/// @returns New instance of @ref dnp3_binary_output_command_event_t
static dnp3_binary_output_command_event_t dnp3_binary_output_command_event_init(uint16_t index, dnp3_command_status_t status, bool commanded_state, dnp3_timestamp_t time)
{
    dnp3_binary_output_command_event_t _return_value = {
        index,
        status,
        commanded_state,
        time
    };
    return _return_value;
}


/// @brief Iterator of binary_output_command_event
typedef struct dnp3_binary_output_command_event_iterator_t dnp3_binary_output_command_event_iterator_t;

/// @brief returns a pointer to the next value or NULL
/// @param iter opaque iterator on which to retrieve the next value
/// @return next value or NULL
dnp3_binary_output_command_event_t* dnp3_binary_output_command_event_iterator_next(dnp3_binary_output_command_event_iterator_t* iter);


/// @brief Describes the encoding of the commanded value
typedef enum dnp3_analog_command_type_t
{
    /// @brief 16-bit integer
    DNP3_ANALOG_COMMAND_TYPE_I16 = 0,
    /// @brief 16-bit integer
    DNP3_ANALOG_COMMAND_TYPE_I32 = 1,
    /// @brief single-precision floating point
    DNP3_ANALOG_COMMAND_TYPE_F32 = 2,
    /// @brief double-precision floating point
    DNP3_ANALOG_COMMAND_TYPE_F64 = 3,
} dnp3_analog_command_type_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_analog_command_type_to_string(dnp3_analog_command_type_t value)
{
    switch (value)
    {
        case DNP3_ANALOG_COMMAND_TYPE_I16: return "i16";
        case DNP3_ANALOG_COMMAND_TYPE_I32: return "i32";
        case DNP3_ANALOG_COMMAND_TYPE_F32: return "f32";
        case DNP3_ANALOG_COMMAND_TYPE_F64: return "f64";
        default: return "unknown analog_command_type value";
    }
}

typedef struct dnp3_analog_output_command_event_t dnp3_analog_output_command_event_t;

/// @brief Event transferred from master to outstation when the outstation receives a corresponding command.
/// 
/// Maps to group 43 variations 1 to 8.
/// 
/// These objects are part of subset level 4 and are not commonly used.
typedef struct dnp3_analog_output_command_event_t
{
    /// @brief Index of the command event
    uint16_t index;
    /// @brief Status from processing the command that triggered this event
    dnp3_command_status_t status;
    /// @brief Commanded state of the binary output
    /// 
    /// All of the variations in group 43 are mapped to double-precision floats
    double commanded_value;
    /// @brief Describes how the value was encoded in the protocol
    dnp3_analog_command_type_t command_type;
    /// @brief Associated timestamp
    dnp3_timestamp_t time;
} dnp3_analog_output_command_event_t;

/// @brief Fully construct @ref dnp3_analog_output_command_event_t specifying the value of each field
/// @param index Index of the command event
/// @param status Status from processing the command that triggered this event
/// @param commanded_value Commanded state of the binary output
/// @param command_type Describes how the value was encoded in the protocol
/// @param time Associated timestamp
/// @returns New instance of @ref dnp3_analog_output_command_event_t
static dnp3_analog_output_command_event_t dnp3_analog_output_command_event_init(uint16_t index, dnp3_command_status_t status, double commanded_value, dnp3_analog_command_type_t command_type, dnp3_timestamp_t time)
{
    dnp3_analog_output_command_event_t _return_value = {
        index,
        status,
        commanded_value,
        command_type,
        time
    };
    return _return_value;
}


/// @brief Iterator of analog_output_command_event
typedef struct dnp3_analog_output_command_event_iterator_t dnp3_analog_output_command_event_iterator_t;

/// @brief returns a pointer to the next value or NULL
/// @param iter opaque iterator on which to retrieve the next value
/// @return next value or NULL
dnp3_analog_output_command_event_t* dnp3_analog_output_command_event_iterator_next(dnp3_analog_output_command_event_iterator_t* iter);


typedef struct dnp3_unsigned_integer_t dnp3_unsigned_integer_t;

/// @brief Unsigned byte corresponding to group 102 variation 1
/// 
/// These objects are not part of any subset level and are not commonly used.
typedef struct dnp3_unsigned_integer_t
{
    /// @brief Index of the object
    uint16_t index;
    /// @brief Value of the object
    uint8_t value;
} dnp3_unsigned_integer_t;

/// @brief Fully construct @ref dnp3_unsigned_integer_t specifying the value of each field
/// @param index Index of the object
/// @param value Value of the object
/// @returns New instance of @ref dnp3_unsigned_integer_t
static dnp3_unsigned_integer_t dnp3_unsigned_integer_init(uint16_t index, uint8_t value)
{
    dnp3_unsigned_integer_t _return_value = {
        index,
        value
    };
    return _return_value;
}


/// @brief Iterator of unsigned_integer
typedef struct dnp3_unsigned_integer_iterator_t dnp3_unsigned_integer_iterator_t;

/// @brief returns a pointer to the next value or NULL
/// @param iter opaque iterator on which to retrieve the next value
/// @return next value or NULL
dnp3_unsigned_integer_t* dnp3_unsigned_integer_iterator_next(dnp3_unsigned_integer_iterator_t* iter);


/// @brief Iterator of uint8_t
typedef struct dnp3_byte_iterator_t dnp3_byte_iterator_t;

/// @brief returns a pointer to the next value or NULL
/// @param iter opaque iterator on which to retrieve the next value
/// @return next value or NULL
uint8_t* dnp3_byte_iterator_next(dnp3_byte_iterator_t* iter);


typedef struct dnp3_octet_string_t dnp3_octet_string_t;

/// @brief Octet String point
typedef struct dnp3_octet_string_t
{
    /// @brief Point index
    uint16_t index;
    /// @brief Point value
    dnp3_byte_iterator_t* value;
} dnp3_octet_string_t;


/// @brief Iterator of octet_string
typedef struct dnp3_octet_string_iterator_t dnp3_octet_string_iterator_t;

/// @brief returns a pointer to the next value or NULL
/// @param iter opaque iterator on which to retrieve the next value
/// @return next value or NULL
dnp3_octet_string_t* dnp3_octet_string_iterator_next(dnp3_octet_string_iterator_t* iter);


/// @brief Options that control how TCP connections are established
typedef struct dnp3_connect_options_t dnp3_connect_options_t;

/// @brief Initialize to the defaults
/// @return Instance of @ref dnp3_connect_options_t
dnp3_connect_options_t* dnp3_connect_options_create();

/// @brief Destroy an instance
/// @param instance Instance of @ref dnp3_connect_options_t to destroy
void dnp3_connect_options_destroy(dnp3_connect_options_t* instance);

/// @brief Set a timeout for the TCP connection that might be less than the default for the OS
/// @param instance Instance of @ref dnp3_connect_options_t
/// @param timeout Timeout value (seconds)
void dnp3_connect_options_set_timeout(dnp3_connect_options_t* instance, uint64_t timeout);

/// @brief Set the local address to which the socket is bound
/// 
/// If not specified, then any available adapter may be used with an OS assigned port.
/// @param instance Instance of @ref dnp3_connect_options_t
/// @param endpoint String in 'address:port' format, where address can be IPv4 or IPv6. Using 0 for the port results in an OS assigned port
/// @return Error code
dnp3_param_error_t dnp3_connect_options_set_local_endpoint(dnp3_connect_options_t* instance, const char* endpoint);


/// @brief List of IP endpoints.
/// 
/// You can write IP addresses or DNS names and the port to connect to. e.g. "127.0.0.1:20000" or "dnp3.myorg.com:20000".
typedef struct dnp3_endpoint_list_t dnp3_endpoint_list_t;

/// @brief Create a new list of IP endpoints.
/// 
/// You can write IP addresses or DNS names and the port to connect to. e.g. "127.0.0.1:20000" or "dnp3.myorg.com:20000".
/// @param main_endpoint Main endpoint
/// @return Instance of @ref dnp3_endpoint_list_t
dnp3_endpoint_list_t* dnp3_endpoint_list_create(const char* main_endpoint);

/// @brief Destroy a previously allocated endpoint list
/// @param instance Instance of @ref dnp3_endpoint_list_t to destroy
void dnp3_endpoint_list_destroy(dnp3_endpoint_list_t* instance);

/// @brief Add an IP endpoint to the list.
/// 
/// You can write IP addresses or DNS names and the port to connect to. e.g. "127.0.0.1:20000" or "dnp3.myorg.com:20000".
/// @param instance Instance of @ref dnp3_endpoint_list_t
/// @param endpoint Endpoint to add to the list
void dnp3_endpoint_list_add(dnp3_endpoint_list_t* instance, const char* endpoint);


typedef struct dnp3_connect_strategy_t dnp3_connect_strategy_t;

/// @brief Timing parameters for connection attempts
typedef struct dnp3_connect_strategy_t
{
    /// @brief Minimum delay between two connection attempts, doubles up to the maximum delay
    /// @note The unit is milliseconds
    uint64_t min_connect_delay;
    /// @brief Maximum delay between two connection attempts
    /// @note The unit is milliseconds
    uint64_t max_connect_delay;
    /// @brief Delay before attempting a connection after a disconnect
    /// @note The unit is milliseconds
    uint64_t reconnect_delay;
} dnp3_connect_strategy_t;

/// @brief Initialize to default values
/// 
/// @note Values are initialized to:
/// - @ref dnp3_connect_strategy_t.min_connect_delay : 1000ms
/// - @ref dnp3_connect_strategy_t.max_connect_delay : 10000ms
/// - @ref dnp3_connect_strategy_t.reconnect_delay : 1000ms
/// 
/// @returns New instance of @ref dnp3_connect_strategy_t
static dnp3_connect_strategy_t dnp3_connect_strategy_init()
{
    dnp3_connect_strategy_t _return_value = {
        1000,
        10000,
        1000
    };
    return _return_value;
}


/// @brief Minimum TLS version to allow
typedef enum dnp3_min_tls_version_t
{
    /// @brief Allow TLS 1.2 and 1.3
    DNP3_MIN_TLS_VERSION_V12 = 0,
    /// @brief Only allow TLS 1.3
    DNP3_MIN_TLS_VERSION_V13 = 1,
} dnp3_min_tls_version_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_min_tls_version_to_string(dnp3_min_tls_version_t value)
{
    switch (value)
    {
        case DNP3_MIN_TLS_VERSION_V12: return "v12";
        case DNP3_MIN_TLS_VERSION_V13: return "v13";
        default: return "unknown min_tls_version value";
    }
}

/// @brief Determines how the certificate(s) presented by the peer are validated
/// 
/// This validation always occurs **after** the handshake signature has been verified.
typedef enum dnp3_certificate_mode_t
{
    /// @brief Validates the peer certificate against one or more configured trust anchors
    /// 
    /// This mode uses the default certificate verifier in `rustls` to ensure that the chain of certificates presented by the peer is valid against one of the configured trust anchors.
    /// 
    /// The name verification is relaxed to allow for certificates that do not contain the SAN extension. In these cases the name is verified using the Common Name instead.
    DNP3_CERTIFICATE_MODE_AUTHORITY_BASED = 0,
    /// @brief Validates that the peer presents a single certificate which is a byte-for-byte match against the configured peer certificate
    /// 
    /// The certificate is parsed only to ensure that the `NotBefore` and `NotAfter` are valid for the current system time.
    DNP3_CERTIFICATE_MODE_SELF_SIGNED = 1,
} dnp3_certificate_mode_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_certificate_mode_to_string(dnp3_certificate_mode_t value)
{
    switch (value)
    {
        case DNP3_CERTIFICATE_MODE_AUTHORITY_BASED: return "authority_based";
        case DNP3_CERTIFICATE_MODE_SELF_SIGNED: return "self_signed";
        default: return "unknown certificate_mode value";
    }
}

typedef struct dnp3_tls_client_config_t dnp3_tls_client_config_t;

/// @brief TLS client configuration
typedef struct dnp3_tls_client_config_t
{
    /// @brief Subject name which is verified in the presented server certificate, from the SAN extension or in the common name field.
    /// 
    /// @warning This argument is only used when used with @ref DNP3_CERTIFICATE_MODE_AUTHORITY_BASED
    const char* dns_name;
    /// @brief Path to the PEM-encoded certificate of the peer
    const char* peer_cert_path;
    /// @brief Path to the PEM-encoded local certificate
    const char* local_cert_path;
    /// @brief Path to the the PEM-encoded private key
    const char* private_key_path;
    /// @brief Optional password if the private key file is encrypted
    /// 
    /// Only PKCS#8 encrypted files are supported.
    /// 
    /// Pass empty string if the file is not encrypted.
    const char* password;
    /// @brief Minimum TLS version allowed
    dnp3_min_tls_version_t min_tls_version;
    /// @brief Certificate validation mode
    dnp3_certificate_mode_t certificate_mode;
    /// @brief If set to true, a '*' may be used for @ref dnp3_tls_client_config_t.dns_name to bypass server name validation
    bool allow_server_name_wildcard;
} dnp3_tls_client_config_t;

/// @brief construct the configuration with defaults
/// 
/// @note Values are initialized to:
/// - @ref dnp3_tls_client_config_t.min_tls_version : @ref DNP3_MIN_TLS_VERSION_V12
/// - @ref dnp3_tls_client_config_t.certificate_mode : @ref DNP3_CERTIFICATE_MODE_AUTHORITY_BASED
/// - @ref dnp3_tls_client_config_t.allow_server_name_wildcard : @p false
/// 
/// @param dns_name Subject name which is verified in the presented server certificate, from the SAN extension or in the common name field.
/// @param peer_cert_path Path to the PEM-encoded certificate of the peer
/// @param local_cert_path Path to the PEM-encoded local certificate
/// @param private_key_path Path to the the PEM-encoded private key
/// @param password Optional password if the private key file is encrypted
/// @returns New instance of @ref dnp3_tls_client_config_t
static dnp3_tls_client_config_t dnp3_tls_client_config_init(const char* dns_name, const char* peer_cert_path, const char* local_cert_path, const char* private_key_path, const char* password)
{
    dnp3_tls_client_config_t _return_value = {
        dns_name,
        peer_cert_path,
        local_cert_path,
        private_key_path,
        password,
        DNP3_MIN_TLS_VERSION_V12,
        DNP3_CERTIFICATE_MODE_AUTHORITY_BASED,
        false
    };
    return _return_value;
}


/// @brief Configuration id
#define DNP3_ATTRIBUTE_VARIATIONS_CONFIG_ID 0xC4
/// @brief Configuration version
#define DNP3_ATTRIBUTE_VARIATIONS_CONFIG_VERSION 0xC5
/// @brief Time and date that the outstation's current configuration was built defined
#define DNP3_ATTRIBUTE_VARIATIONS_CONFIG_BUILD_DATE 0xC6
/// @brief Time and date that the outstation's configuration was last modified
#define DNP3_ATTRIBUTE_VARIATIONS_CONFIG_LAST_CHANGE_DATE 0xC7
/// @brief Digest (aka fingerprint) of the configuration using a CRC, HASH, MAC, or public key signature
#define DNP3_ATTRIBUTE_VARIATIONS_CONFIG_DIGEST 0xC8
/// @brief Configuration digest algorithm
#define DNP3_ATTRIBUTE_VARIATIONS_CONFIG_DIGEST_ALGORITHM 0xC9
/// @brief Master resource id (mRID)
#define DNP3_ATTRIBUTE_VARIATIONS_MASTER_RESOURCE_ID 0xCA
/// @brief Altitude of the device
#define DNP3_ATTRIBUTE_VARIATIONS_DEVICE_LOCATION_ALTITUDE 0xCB
/// @brief Longitude of the device from reference meridian (-180.0 to 180.0 deg)
#define DNP3_ATTRIBUTE_VARIATIONS_DEVICE_LOCATION_LONGITUDE 0xCC
/// @brief Latitude of the device from the equator (90.0 to -90.0 deg)
#define DNP3_ATTRIBUTE_VARIATIONS_DEVICE_LOCATION_LATITUDE 0xCD
/// @brief User-assigned secondary operator name
#define DNP3_ATTRIBUTE_VARIATIONS_USER_ASSIGNED_SECONDARY_OPERATOR_NAME 0xCE
/// @brief User-assigned primary operator name
#define DNP3_ATTRIBUTE_VARIATIONS_USER_ASSIGNED_PRIMARY_OPERATOR_NAME 0xCF
/// @brief User-assigned system name
#define DNP3_ATTRIBUTE_VARIATIONS_USER_ASSIGNED_SYSTEM_NAME 0xD0
/// @brief Secure authentication version
#define DNP3_ATTRIBUTE_VARIATIONS_SECURE_AUTH_VERSION 0xD1
/// @brief Number of security statistics per association
#define DNP3_ATTRIBUTE_VARIATIONS_NUM_SECURITY_STATS_PER_ASSOC 0xD2
/// @brief Identification of user-specific attributes
#define DNP3_ATTRIBUTE_VARIATIONS_USER_SPECIFIC_ATTRIBUTES 0xD3
/// @brief Number of master defined data-set prototypes
#define DNP3_ATTRIBUTE_VARIATIONS_NUM_MASTER_DEFINED_DATA_SET_PROTO 0xD4
/// @brief Number of outstation defined data-set prototypes
#define DNP3_ATTRIBUTE_VARIATIONS_NUM_OUTSTATION_DEFINED_DATA_SET_PROTO 0xD5
/// @brief Number of master defined data-sets
#define DNP3_ATTRIBUTE_VARIATIONS_NUM_MASTER_DEFINED_DATA_SETS 0xD6
/// @brief Number of outstation defined data-sets
#define DNP3_ATTRIBUTE_VARIATIONS_NUM_OUTSTATION_DEFINED_DATA_SETS 0xD7
/// @brief Maximum number of binary outputs per request
#define DNP3_ATTRIBUTE_VARIATIONS_MAX_BINARY_OUTPUT_PER_REQUEST 0xD8
/// @brief Local timing accuracy (microseconds)
#define DNP3_ATTRIBUTE_VARIATIONS_LOCAL_TIMING_ACCURACY 0xD9
/// @brief Duration of time accuracy (seconds)
#define DNP3_ATTRIBUTE_VARIATIONS_DURATION_OF_TIME_ACCURACY 0xDA
/// @brief Supports analog output events
#define DNP3_ATTRIBUTE_VARIATIONS_SUPPORTS_ANALOG_OUTPUT_EVENTS 0xDB
/// @brief Maximum analog output index
#define DNP3_ATTRIBUTE_VARIATIONS_MAX_ANALOG_OUTPUT_INDEX 0xDC
/// @brief Number of analog outputs
#define DNP3_ATTRIBUTE_VARIATIONS_NUM_ANALOG_OUTPUTS 0xDD
/// @brief Supports binary output events
#define DNP3_ATTRIBUTE_VARIATIONS_SUPPORTS_BINARY_OUTPUT_EVENTS 0xDE
/// @brief Maximum binary output index
#define DNP3_ATTRIBUTE_VARIATIONS_MAX_BINARY_OUTPUT_INDEX 0xDF
/// @brief Number of binary outputs
#define DNP3_ATTRIBUTE_VARIATIONS_NUM_BINARY_OUTPUTS 0xE0
/// @brief Supports frozen counter events
#define DNP3_ATTRIBUTE_VARIATIONS_SUPPORTS_FROZEN_COUNTER_EVENTS 0xE1
/// @brief Supports frozen counters
#define DNP3_ATTRIBUTE_VARIATIONS_SUPPORTS_FROZEN_COUNTERS 0xE2
/// @brief Supports counter events
#define DNP3_ATTRIBUTE_VARIATIONS_SUPPORTS_COUNTER_EVENTS 0xE3
/// @brief Maximum counter point index
#define DNP3_ATTRIBUTE_VARIATIONS_MAX_COUNTER_INDEX 0xE4
/// @brief Number of counter points
#define DNP3_ATTRIBUTE_VARIATIONS_NUM_COUNTER 0xE5
/// @brief Supports frozen analog input events
#define DNP3_ATTRIBUTE_VARIATIONS_SUPPORTS_FROZEN_ANALOG_INPUTS 0xE6
/// @brief Supports analog input events
#define DNP3_ATTRIBUTE_VARIATIONS_SUPPORTS_ANALOG_INPUT_EVENTS 0xE7
/// @brief Maximum analog input point index
#define DNP3_ATTRIBUTE_VARIATIONS_MAX_ANALOG_INPUT_INDEX 0xE8
/// @brief Number of analog input points
#define DNP3_ATTRIBUTE_VARIATIONS_NUM_ANALOG_INPUT 0xE9
/// @brief Supports double-bit binary input events
#define DNP3_ATTRIBUTE_VARIATIONS_SUPPORTS_DOUBLE_BIT_BINARY_INPUT_EVENTS 0xEA
/// @brief Maximum double-bit binary input point index
#define DNP3_ATTRIBUTE_VARIATIONS_MAX_DOUBLE_BIT_BINARY_INPUT_INDEX 0xEB
/// @brief Number of double-bit binary input points
#define DNP3_ATTRIBUTE_VARIATIONS_NUM_DOUBLE_BIT_BINARY_INPUT 0xEC
/// @brief Support binary input events
#define DNP3_ATTRIBUTE_VARIATIONS_SUPPORTS_BINARY_INPUT_EVENTS 0xED
/// @brief Maximum binary input point index
#define DNP3_ATTRIBUTE_VARIATIONS_MAX_BINARY_INPUT_INDEX 0xEE
/// @brief Number of binary input points
#define DNP3_ATTRIBUTE_VARIATIONS_NUM_BINARY_INPUT 0xEF
/// @brief Maximum transmit fragment size
#define DNP3_ATTRIBUTE_VARIATIONS_MAX_TX_FRAGMENT_SIZE 0xF0
/// @brief Maximum receive fragment size
#define DNP3_ATTRIBUTE_VARIATIONS_MAX_RX_FRAGMENT_SIZE 0xF1
/// @brief Device manufacturer software version
#define DNP3_ATTRIBUTE_VARIATIONS_DEVICE_MANUFACTURER_SOFTWARE_VERSION 0xF2
/// @brief Device manufacturer hardware version
#define DNP3_ATTRIBUTE_VARIATIONS_DEVICE_MANUFACTURER_HARDWARE_VERSION 0xF3
/// @brief User-assigned owner name
#define DNP3_ATTRIBUTE_VARIATIONS_USER_ASSIGNED_OWNER_NAME 0xF4
/// @brief User assigned location name
#define DNP3_ATTRIBUTE_VARIATIONS_USER_ASSIGNED_LOCATION 0xF5
/// @brief User assigned ID code/number
#define DNP3_ATTRIBUTE_VARIATIONS_USER_ASSIGNED_ID 0xF6
/// @brief User assigned device name
#define DNP3_ATTRIBUTE_VARIATIONS_USER_ASSIGNED_DEVICE_NAME 0xF7
/// @brief Device serial number
#define DNP3_ATTRIBUTE_VARIATIONS_DEVICE_SERIAL_NUMBER 0xF8
/// @brief DNP3 subset and conformance
#define DNP3_ATTRIBUTE_VARIATIONS_DEVICE_SUBSET_AND_CONFORMANCE 0xF9
/// @brief Device manufacturer's product name and model
#define DNP3_ATTRIBUTE_VARIATIONS_PRODUCT_NAME_AND_MODEL 0xFA
/// @brief Device manufacturer's name
#define DNP3_ATTRIBUTE_VARIATIONS_DEVICE_MANUFACTURERS_NAME 0xFC
/// @brief Non-specific all attributes request
#define DNP3_ATTRIBUTE_VARIATIONS_ALL_ATTRIBUTES_REQUEST 0xFE
/// @brief List of attribute variations
#define DNP3_ATTRIBUTE_VARIATIONS_LIST_OF_VARIATIONS 0xFF

/// @brief Enumeration of all the variation list attributes
typedef enum dnp3_variation_list_attr_t
{
    /// @brief The attribute variation is not defined or is not part of the default set
    DNP3_VARIATION_LIST_ATTR_UNKNOWN = 0,
    /// @brief Variation 255 - List of attribute variations
    DNP3_VARIATION_LIST_ATTR_LIST_OF_VARIATIONS = 1,
} dnp3_variation_list_attr_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_variation_list_attr_to_string(dnp3_variation_list_attr_t value)
{
    switch (value)
    {
        case DNP3_VARIATION_LIST_ATTR_UNKNOWN: return "unknown";
        case DNP3_VARIATION_LIST_ATTR_LIST_OF_VARIATIONS: return "list_of_variations";
        default: return "unknown variation_list_attr value";
    }
}

/// @brief Enumeration of all the default string attributes
typedef enum dnp3_string_attr_t
{
    /// @brief The attribute variation is not defined or is not part of the default set
    DNP3_STRING_ATTR_UNKNOWN = 0,
    /// @brief Variation 196 - Configuration id
    DNP3_STRING_ATTR_CONFIG_ID = 1,
    /// @brief Variation 197 - Configuration version
    DNP3_STRING_ATTR_CONFIG_VERSION = 2,
    /// @brief Variation 201 - Configuration digest algorithm
    DNP3_STRING_ATTR_CONFIG_DIGEST_ALGORITHM = 3,
    /// @brief Variation 202 - Master resource id (mRID)
    DNP3_STRING_ATTR_MASTER_RESOURCE_ID = 4,
    /// @brief Variation 206 - User-assigned secondary operator name
    DNP3_STRING_ATTR_USER_ASSIGNED_SECONDARY_OPERATOR_NAME = 5,
    /// @brief Variation 207 - User-assigned primary operator name
    DNP3_STRING_ATTR_USER_ASSIGNED_PRIMARY_OPERATOR_NAME = 6,
    /// @brief Variation 208 - User-assigned system name
    DNP3_STRING_ATTR_USER_ASSIGNED_SYSTEM_NAME = 7,
    /// @brief Variation 211 - Identification of user-specific attributes
    DNP3_STRING_ATTR_USER_SPECIFIC_ATTRIBUTES = 8,
    /// @brief Variation 242 - Device manufacturer software version
    DNP3_STRING_ATTR_DEVICE_MANUFACTURER_SOFTWARE_VERSION = 9,
    /// @brief Variation 243 - Device manufacturer hardware version
    DNP3_STRING_ATTR_DEVICE_MANUFACTURER_HARDWARE_VERSION = 10,
    /// @brief Variation 244 - User-assigned owner name
    DNP3_STRING_ATTR_USER_ASSIGNED_OWNER_NAME = 11,
    /// @brief Variation 245 - User assigned location name
    DNP3_STRING_ATTR_USER_ASSIGNED_LOCATION = 12,
    /// @brief Variation 246 - User assigned ID code/number
    DNP3_STRING_ATTR_USER_ASSIGNED_ID = 13,
    /// @brief Variation 247 - User assigned device name
    DNP3_STRING_ATTR_USER_ASSIGNED_DEVICE_NAME = 14,
    /// @brief Variation 248 - Device serial number
    DNP3_STRING_ATTR_DEVICE_SERIAL_NUMBER = 15,
    /// @brief Variation 249 - DNP3 subset and conformance
    DNP3_STRING_ATTR_DEVICE_SUBSET_AND_CONFORMANCE = 16,
    /// @brief Variation 250 - Device manufacturer's product name and model
    DNP3_STRING_ATTR_PRODUCT_NAME_AND_MODEL = 17,
    /// @brief Variation 252 - Device manufacturer's name
    DNP3_STRING_ATTR_DEVICE_MANUFACTURERS_NAME = 18,
} dnp3_string_attr_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_string_attr_to_string(dnp3_string_attr_t value)
{
    switch (value)
    {
        case DNP3_STRING_ATTR_UNKNOWN: return "unknown";
        case DNP3_STRING_ATTR_CONFIG_ID: return "config_id";
        case DNP3_STRING_ATTR_CONFIG_VERSION: return "config_version";
        case DNP3_STRING_ATTR_CONFIG_DIGEST_ALGORITHM: return "config_digest_algorithm";
        case DNP3_STRING_ATTR_MASTER_RESOURCE_ID: return "master_resource_id";
        case DNP3_STRING_ATTR_USER_ASSIGNED_SECONDARY_OPERATOR_NAME: return "user_assigned_secondary_operator_name";
        case DNP3_STRING_ATTR_USER_ASSIGNED_PRIMARY_OPERATOR_NAME: return "user_assigned_primary_operator_name";
        case DNP3_STRING_ATTR_USER_ASSIGNED_SYSTEM_NAME: return "user_assigned_system_name";
        case DNP3_STRING_ATTR_USER_SPECIFIC_ATTRIBUTES: return "user_specific_attributes";
        case DNP3_STRING_ATTR_DEVICE_MANUFACTURER_SOFTWARE_VERSION: return "device_manufacturer_software_version";
        case DNP3_STRING_ATTR_DEVICE_MANUFACTURER_HARDWARE_VERSION: return "device_manufacturer_hardware_version";
        case DNP3_STRING_ATTR_USER_ASSIGNED_OWNER_NAME: return "user_assigned_owner_name";
        case DNP3_STRING_ATTR_USER_ASSIGNED_LOCATION: return "user_assigned_location";
        case DNP3_STRING_ATTR_USER_ASSIGNED_ID: return "user_assigned_id";
        case DNP3_STRING_ATTR_USER_ASSIGNED_DEVICE_NAME: return "user_assigned_device_name";
        case DNP3_STRING_ATTR_DEVICE_SERIAL_NUMBER: return "device_serial_number";
        case DNP3_STRING_ATTR_DEVICE_SUBSET_AND_CONFORMANCE: return "device_subset_and_conformance";
        case DNP3_STRING_ATTR_PRODUCT_NAME_AND_MODEL: return "product_name_and_model";
        case DNP3_STRING_ATTR_DEVICE_MANUFACTURERS_NAME: return "device_manufacturers_name";
        default: return "unknown string_attr value";
    }
}

/// @brief Enumeration of all the default uint attributes
typedef enum dnp3_uint_attr_t
{
    /// @brief The attribute variation is not defined or is not part of the default set
    DNP3_UINT_ATTR_UNKNOWN = 0,
    /// @brief Variation 209 - Secure authentication version
    DNP3_UINT_ATTR_SECURE_AUTH_VERSION = 1,
    /// @brief Variation 210 - Number of security statistics per association
    DNP3_UINT_ATTR_NUM_SECURITY_STATS_PER_ASSOC = 2,
    /// @brief Variation 212 - Number of master defined data-set prototypes
    DNP3_UINT_ATTR_NUM_MASTER_DEFINED_DATA_SET_PROTO = 3,
    /// @brief Variation 213 - Number of outstation defined data-set prototypes
    DNP3_UINT_ATTR_NUM_OUTSTATION_DEFINED_DATA_SET_PROTO = 4,
    /// @brief Variation 214 - Number of master defined data-sets
    DNP3_UINT_ATTR_NUM_MASTER_DEFINED_DATA_SETS = 5,
    /// @brief Variation 215 - Number of outstation defined data-sets
    DNP3_UINT_ATTR_NUM_OUTSTATION_DEFINED_DATA_SETS = 6,
    /// @brief Variation 216 - Maximum number of binary outputs per request
    DNP3_UINT_ATTR_MAX_BINARY_OUTPUT_PER_REQUEST = 7,
    /// @brief Variation 217 - Local timing accuracy (microseconds)
    DNP3_UINT_ATTR_LOCAL_TIMING_ACCURACY = 8,
    /// @brief Variation 218 - Duration of time accuracy (seconds)
    DNP3_UINT_ATTR_DURATION_OF_TIME_ACCURACY = 9,
    /// @brief Variation 220 - Maximum analog output index
    DNP3_UINT_ATTR_MAX_ANALOG_OUTPUT_INDEX = 10,
    /// @brief Variation 221 - Number of analog outputs
    DNP3_UINT_ATTR_NUM_ANALOG_OUTPUTS = 11,
    /// @brief Variation 223 - Maximum binary output index
    DNP3_UINT_ATTR_MAX_BINARY_OUTPUT_INDEX = 12,
    /// @brief Variation 224 - Number of binary outputs
    DNP3_UINT_ATTR_NUM_BINARY_OUTPUTS = 13,
    /// @brief Variation 228 - Maximum counter point index
    DNP3_UINT_ATTR_MAX_COUNTER_INDEX = 14,
    /// @brief Variation 229 - Number of counter points
    DNP3_UINT_ATTR_NUM_COUNTER = 15,
    /// @brief Variation 232 - Maximum analog input point index
    DNP3_UINT_ATTR_MAX_ANALOG_INPUT_INDEX = 16,
    /// @brief Variation 233 - Number of analog input points
    DNP3_UINT_ATTR_NUM_ANALOG_INPUT = 17,
    /// @brief Variation 235 - Maximum double-bit binary input point index
    DNP3_UINT_ATTR_MAX_DOUBLE_BIT_BINARY_INPUT_INDEX = 18,
    /// @brief Variation 236 - Number of double-bit binary input points
    DNP3_UINT_ATTR_NUM_DOUBLE_BIT_BINARY_INPUT = 19,
    /// @brief Variation 238 - Maximum binary input point index
    DNP3_UINT_ATTR_MAX_BINARY_INPUT_INDEX = 20,
    /// @brief Variation 239 - Number of binary input points
    DNP3_UINT_ATTR_NUM_BINARY_INPUT = 21,
    /// @brief Variation 240 - Maximum transmit fragment size
    DNP3_UINT_ATTR_MAX_TX_FRAGMENT_SIZE = 22,
    /// @brief Variation 241 - Maximum receive fragment size
    DNP3_UINT_ATTR_MAX_RX_FRAGMENT_SIZE = 23,
} dnp3_uint_attr_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_uint_attr_to_string(dnp3_uint_attr_t value)
{
    switch (value)
    {
        case DNP3_UINT_ATTR_UNKNOWN: return "unknown";
        case DNP3_UINT_ATTR_SECURE_AUTH_VERSION: return "secure_auth_version";
        case DNP3_UINT_ATTR_NUM_SECURITY_STATS_PER_ASSOC: return "num_security_stats_per_assoc";
        case DNP3_UINT_ATTR_NUM_MASTER_DEFINED_DATA_SET_PROTO: return "num_master_defined_data_set_proto";
        case DNP3_UINT_ATTR_NUM_OUTSTATION_DEFINED_DATA_SET_PROTO: return "num_outstation_defined_data_set_proto";
        case DNP3_UINT_ATTR_NUM_MASTER_DEFINED_DATA_SETS: return "num_master_defined_data_sets";
        case DNP3_UINT_ATTR_NUM_OUTSTATION_DEFINED_DATA_SETS: return "num_outstation_defined_data_sets";
        case DNP3_UINT_ATTR_MAX_BINARY_OUTPUT_PER_REQUEST: return "max_binary_output_per_request";
        case DNP3_UINT_ATTR_LOCAL_TIMING_ACCURACY: return "local_timing_accuracy";
        case DNP3_UINT_ATTR_DURATION_OF_TIME_ACCURACY: return "duration_of_time_accuracy";
        case DNP3_UINT_ATTR_MAX_ANALOG_OUTPUT_INDEX: return "max_analog_output_index";
        case DNP3_UINT_ATTR_NUM_ANALOG_OUTPUTS: return "num_analog_outputs";
        case DNP3_UINT_ATTR_MAX_BINARY_OUTPUT_INDEX: return "max_binary_output_index";
        case DNP3_UINT_ATTR_NUM_BINARY_OUTPUTS: return "num_binary_outputs";
        case DNP3_UINT_ATTR_MAX_COUNTER_INDEX: return "max_counter_index";
        case DNP3_UINT_ATTR_NUM_COUNTER: return "num_counter";
        case DNP3_UINT_ATTR_MAX_ANALOG_INPUT_INDEX: return "max_analog_input_index";
        case DNP3_UINT_ATTR_NUM_ANALOG_INPUT: return "num_analog_input";
        case DNP3_UINT_ATTR_MAX_DOUBLE_BIT_BINARY_INPUT_INDEX: return "max_double_bit_binary_input_index";
        case DNP3_UINT_ATTR_NUM_DOUBLE_BIT_BINARY_INPUT: return "num_double_bit_binary_input";
        case DNP3_UINT_ATTR_MAX_BINARY_INPUT_INDEX: return "max_binary_input_index";
        case DNP3_UINT_ATTR_NUM_BINARY_INPUT: return "num_binary_input";
        case DNP3_UINT_ATTR_MAX_TX_FRAGMENT_SIZE: return "max_tx_fragment_size";
        case DNP3_UINT_ATTR_MAX_RX_FRAGMENT_SIZE: return "max_rx_fragment_size";
        default: return "unknown uint_attr value";
    }
}

/// @brief Enumeration of all the default integer attributes
/// 
/// In 1815-2012 all integer attributes are mapped to boolean values
typedef enum dnp3_int_attr_t
{
    /// @brief The attribute variation is not defined or is not part of the default set
    DNP3_INT_ATTR_UNKNOWN = 0,
} dnp3_int_attr_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_int_attr_to_string(dnp3_int_attr_t value)
{
    switch (value)
    {
        case DNP3_INT_ATTR_UNKNOWN: return "unknown";
        default: return "unknown int_attr value";
    }
}

/// @brief Enumeration of all the known boolean attributes
/// 
/// Boolean attributes are actually just encoded as signed integer attributes where 1 == true
typedef enum dnp3_bool_attr_t
{
    /// @brief The attribute variation is not defined or is not part of the default set
    DNP3_BOOL_ATTR_UNKNOWN = 0,
    /// @brief Variation 219 - Supports analog output events
    DNP3_BOOL_ATTR_SUPPORTS_ANALOG_OUTPUT_EVENTS = 1,
    /// @brief Variation 222 - Supports binary output events
    DNP3_BOOL_ATTR_SUPPORTS_BINARY_OUTPUT_EVENTS = 2,
    /// @brief Variation 225 - Supports frozen counter events
    DNP3_BOOL_ATTR_SUPPORTS_FROZEN_COUNTER_EVENTS = 3,
    /// @brief Variation 226 - Supports frozen counters
    DNP3_BOOL_ATTR_SUPPORTS_FROZEN_COUNTERS = 4,
    /// @brief Variation 227 - Supports counter events
    DNP3_BOOL_ATTR_SUPPORTS_COUNTER_EVENTS = 5,
    /// @brief Variation 230 - Supports frozen analog input events
    DNP3_BOOL_ATTR_SUPPORTS_FROZEN_ANALOG_INPUTS = 6,
    /// @brief Variation 231 - Supports analog input events
    DNP3_BOOL_ATTR_SUPPORTS_ANALOG_INPUT_EVENTS = 7,
    /// @brief Variation 234 - Supports double-bit binary input events
    DNP3_BOOL_ATTR_SUPPORTS_DOUBLE_BIT_BINARY_INPUT_EVENTS = 8,
    /// @brief Variation 237 - Support binary input events
    DNP3_BOOL_ATTR_SUPPORTS_BINARY_INPUT_EVENTS = 9,
} dnp3_bool_attr_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_bool_attr_to_string(dnp3_bool_attr_t value)
{
    switch (value)
    {
        case DNP3_BOOL_ATTR_UNKNOWN: return "unknown";
        case DNP3_BOOL_ATTR_SUPPORTS_ANALOG_OUTPUT_EVENTS: return "supports_analog_output_events";
        case DNP3_BOOL_ATTR_SUPPORTS_BINARY_OUTPUT_EVENTS: return "supports_binary_output_events";
        case DNP3_BOOL_ATTR_SUPPORTS_FROZEN_COUNTER_EVENTS: return "supports_frozen_counter_events";
        case DNP3_BOOL_ATTR_SUPPORTS_FROZEN_COUNTERS: return "supports_frozen_counters";
        case DNP3_BOOL_ATTR_SUPPORTS_COUNTER_EVENTS: return "supports_counter_events";
        case DNP3_BOOL_ATTR_SUPPORTS_FROZEN_ANALOG_INPUTS: return "supports_frozen_analog_inputs";
        case DNP3_BOOL_ATTR_SUPPORTS_ANALOG_INPUT_EVENTS: return "supports_analog_input_events";
        case DNP3_BOOL_ATTR_SUPPORTS_DOUBLE_BIT_BINARY_INPUT_EVENTS: return "supports_double_bit_binary_input_events";
        case DNP3_BOOL_ATTR_SUPPORTS_BINARY_INPUT_EVENTS: return "supports_binary_input_events";
        default: return "unknown bool_attr value";
    }
}

/// @brief Enumeration of all the known DNP3 Time attributes
typedef enum dnp3_time_attr_t
{
    /// @brief The attribute variation is not defined or is not part of the default set
    DNP3_TIME_ATTR_UNKNOWN = 0,
    /// @brief Variation 198 - Time and date that the outstation's current configuration was built defined
    DNP3_TIME_ATTR_CONFIG_BUILD_DATE = 1,
    /// @brief Variation 199 - Time and date that the outstation's configuration was last modified
    DNP3_TIME_ATTR_CONFIG_LAST_CHANGE_DATE = 2,
} dnp3_time_attr_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_time_attr_to_string(dnp3_time_attr_t value)
{
    switch (value)
    {
        case DNP3_TIME_ATTR_UNKNOWN: return "unknown";
        case DNP3_TIME_ATTR_CONFIG_BUILD_DATE: return "config_build_date";
        case DNP3_TIME_ATTR_CONFIG_LAST_CHANGE_DATE: return "config_last_change_date";
        default: return "unknown time_attr value";
    }
}

/// @brief Enumeration of all known octet-string attributes
typedef enum dnp3_octet_string_attr_t
{
    /// @brief The attribute variation is not defined or is not part of the default set
    DNP3_OCTET_STRING_ATTR_UNKNOWN = 0,
    /// @brief Variation 200 - Digest (aka fingerprint) of the configuration using a CRC, HASH, MAC, or public key signature
    DNP3_OCTET_STRING_ATTR_CONFIG_DIGEST = 1,
} dnp3_octet_string_attr_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_octet_string_attr_to_string(dnp3_octet_string_attr_t value)
{
    switch (value)
    {
        case DNP3_OCTET_STRING_ATTR_UNKNOWN: return "unknown";
        case DNP3_OCTET_STRING_ATTR_CONFIG_DIGEST: return "config_digest";
        default: return "unknown octet_string_attr value";
    }
}

/// @brief Enumeration of all known bit-string attributes
typedef enum dnp3_bit_string_attr_t
{
    /// @brief The attribute variation is not defined or is not part of the default set
    DNP3_BIT_STRING_ATTR_UNKNOWN = 0,
} dnp3_bit_string_attr_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_bit_string_attr_to_string(dnp3_bit_string_attr_t value)
{
    switch (value)
    {
        case DNP3_BIT_STRING_ATTR_UNKNOWN: return "unknown";
        default: return "unknown bit_string_attr value";
    }
}

/// @brief Enumeration of all known float attributes
typedef enum dnp3_float_attr_t
{
    /// @brief The attribute variation is not defined or is not part of the default set
    DNP3_FLOAT_ATTR_UNKNOWN = 0,
    /// @brief Variation 203 - Altitude of the device
    DNP3_FLOAT_ATTR_DEVICE_LOCATION_ALTITUDE = 1,
    /// @brief Variation 204 - Longitude of the device from reference meridian (-180.0 to 180.0 deg)
    DNP3_FLOAT_ATTR_DEVICE_LOCATION_LONGITUDE = 2,
    /// @brief Variation 205 - Latitude of the device from the equator (90.0 to -90.0 deg)
    DNP3_FLOAT_ATTR_DEVICE_LOCATION_LATITUDE = 3,
} dnp3_float_attr_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_float_attr_to_string(dnp3_float_attr_t value)
{
    switch (value)
    {
        case DNP3_FLOAT_ATTR_UNKNOWN: return "unknown";
        case DNP3_FLOAT_ATTR_DEVICE_LOCATION_ALTITUDE: return "device_location_altitude";
        case DNP3_FLOAT_ATTR_DEVICE_LOCATION_LONGITUDE: return "device_location_longitude";
        case DNP3_FLOAT_ATTR_DEVICE_LOCATION_LATITUDE: return "device_location_latitude";
        default: return "unknown float_attr value";
    }
}

typedef struct dnp3_attr_prop_t dnp3_attr_prop_t;

/// @brief Attribute properties returned in Group0Var255
/// 
/// In 1815-2012 this only includes a field indicating if the property can be written
typedef struct dnp3_attr_prop_t
{
    /// @brief Indicate if the property can be used in a WRITE operation
    bool is_writable;
} dnp3_attr_prop_t;


typedef struct dnp3_attr_item_t dnp3_attr_item_t;

/// @brief An attribute variation and properties pair returned in Group0Var255
typedef struct dnp3_attr_item_t
{
    /// @brief Variation of the attribute
    uint8_t variation;
    /// @brief Properties of the attribute
    dnp3_attr_prop_t properties;
} dnp3_attr_item_t;


/// @brief Iterator of attr_item
typedef struct dnp3_attr_item_iter_t dnp3_attr_item_iter_t;

/// @brief returns a pointer to the next value or NULL
/// @param iter opaque iterator on which to retrieve the next value
/// @return next value or NULL
dnp3_attr_item_t* dnp3_attr_item_iter_next(dnp3_attr_item_iter_t* iter);


/// @brief A single value enum which is used as a placeholder for futures that don't return a value
typedef enum dnp3_nothing_t
{
    /// @brief The value type is meaningless
    DNP3_NOTHING_NOTHING = 0,
} dnp3_nothing_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_nothing_to_string(dnp3_nothing_t value)
{
    switch (value)
    {
        case DNP3_NOTHING_NOTHING: return "nothing";
        default: return "unknown nothing value";
    }
}

/// @brief Collection of byte_collection
typedef struct dnp3_byte_collection_t dnp3_byte_collection_t;

/// @brief Creates an instance of the collection
/// @param reserve_size preallocate a particular size
/// @return Allocated opaque collection instance
dnp3_byte_collection_t* dnp3_byte_collection_create(uint32_t reserve_size);

/// @brief Destroys an instance of the collection
/// @param instance instance to destroy
void dnp3_byte_collection_destroy(dnp3_byte_collection_t* instance);

/// @brief Add a value to the collection
/// @param instance instance to which to add the value
/// @param value value to add to the instance
void dnp3_byte_collection_add(dnp3_byte_collection_t* instance, uint8_t value);


/// @brief State of the serial port
typedef enum dnp3_port_state_t
{
    /// @brief Disabled until enabled
    DNP3_PORT_STATE_DISABLED = 0,
    /// @brief Waiting to perform an open retry
    DNP3_PORT_STATE_WAIT = 1,
    /// @brief Port is open
    DNP3_PORT_STATE_OPEN = 2,
    /// @brief Task has been shut down
    DNP3_PORT_STATE_SHUTDOWN = 3,
} dnp3_port_state_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_port_state_to_string(dnp3_port_state_t value)
{
    switch (value)
    {
        case DNP3_PORT_STATE_DISABLED: return "disabled";
        case DNP3_PORT_STATE_WAIT: return "wait";
        case DNP3_PORT_STATE_OPEN: return "open";
        case DNP3_PORT_STATE_SHUTDOWN: return "shutdown";
        default: return "unknown port_state value";
    }
}

/// @brief Callback interface for receiving updates about the state of a serial port
typedef struct dnp3_port_state_listener_t
{
    
    /// @brief Invoked when the serial port changes state
    /// @param state New state of the port
    /// @param ctx Context data
    void (*on_change)(dnp3_port_state_t, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_port_state_listener_t;

/// @brief State of the client connection.
/// 
/// Use by the @ref dnp3_client_state_listener_t.
typedef enum dnp3_client_state_t
{
    /// @brief Client is disabled and idle until enabled
    DNP3_CLIENT_STATE_DISABLED = 0,
    /// @brief Client is trying to establish a connection to the remote device
    DNP3_CLIENT_STATE_CONNECTING = 1,
    /// @brief Client is connected to the remote device
    DNP3_CLIENT_STATE_CONNECTED = 2,
    /// @brief Failed to establish a connection, waiting before retrying
    DNP3_CLIENT_STATE_WAIT_AFTER_FAILED_CONNECT = 3,
    /// @brief Client was disconnected, waiting before retrying
    DNP3_CLIENT_STATE_WAIT_AFTER_DISCONNECT = 4,
    /// @brief Client is shutting down
    DNP3_CLIENT_STATE_SHUTDOWN = 5,
} dnp3_client_state_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_client_state_to_string(dnp3_client_state_t value)
{
    switch (value)
    {
        case DNP3_CLIENT_STATE_DISABLED: return "disabled";
        case DNP3_CLIENT_STATE_CONNECTING: return "connecting";
        case DNP3_CLIENT_STATE_CONNECTED: return "connected";
        case DNP3_CLIENT_STATE_WAIT_AFTER_FAILED_CONNECT: return "wait_after_failed_connect";
        case DNP3_CLIENT_STATE_WAIT_AFTER_DISCONNECT: return "wait_after_disconnect";
        case DNP3_CLIENT_STATE_SHUTDOWN: return "shutdown";
        default: return "unknown client_state value";
    }
}

/// @brief Callback for monitoring the client TCP connection state
typedef struct dnp3_client_state_listener_t
{
    
    /// @brief Called when the client state changed
    /// @param state New state
    /// @param ctx Context data
    void (*on_change)(dnp3_client_state_t, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_client_state_listener_t;

/// @brief Group/Variation
typedef enum dnp3_variation_t
{
    /// @brief Device Attributes - Variations 0 to 253 and 255
    DNP3_VARIATION_GROUP0 = 0,
    /// @brief Device Attributes - Non-Specific All Attributes Request
    DNP3_VARIATION_GROUP0_VAR254 = 1,
    /// @brief Binary Input - Default variation
    DNP3_VARIATION_GROUP1_VAR0 = 2,
    /// @brief Binary Input - Packed format
    DNP3_VARIATION_GROUP1_VAR1 = 3,
    /// @brief Binary Input - With flags
    DNP3_VARIATION_GROUP1_VAR2 = 4,
    /// @brief Binary Input Event - Default variation
    DNP3_VARIATION_GROUP2_VAR0 = 5,
    /// @brief Binary Input Event - Without time
    DNP3_VARIATION_GROUP2_VAR1 = 6,
    /// @brief Binary Input Event - With absolute time
    DNP3_VARIATION_GROUP2_VAR2 = 7,
    /// @brief Binary Input Event - With relative time
    DNP3_VARIATION_GROUP2_VAR3 = 8,
    /// @brief Double-bit Binary Input - Default variation
    DNP3_VARIATION_GROUP3_VAR0 = 9,
    /// @brief Double-bit Binary Input - Packed format
    DNP3_VARIATION_GROUP3_VAR1 = 10,
    /// @brief Double-bit Binary Input - With flags
    DNP3_VARIATION_GROUP3_VAR2 = 11,
    /// @brief Double-bit Binary Input Event - Default variation
    DNP3_VARIATION_GROUP4_VAR0 = 12,
    /// @brief Double-bit Binary Input Event - Without time
    DNP3_VARIATION_GROUP4_VAR1 = 13,
    /// @brief Double-bit Binary Input Event - With absolute time
    DNP3_VARIATION_GROUP4_VAR2 = 14,
    /// @brief Double-bit Binary Input Event - With relative time
    DNP3_VARIATION_GROUP4_VAR3 = 15,
    /// @brief Binary Output - Default variation
    DNP3_VARIATION_GROUP10_VAR0 = 16,
    /// @brief Binary Output - Packed format
    DNP3_VARIATION_GROUP10_VAR1 = 17,
    /// @brief Binary Output - With flags
    DNP3_VARIATION_GROUP10_VAR2 = 18,
    /// @brief Binary Output Event - Default variation
    DNP3_VARIATION_GROUP11_VAR0 = 19,
    /// @brief Binary Output Event - Without time
    DNP3_VARIATION_GROUP11_VAR1 = 20,
    /// @brief Binary Output Event - With time
    DNP3_VARIATION_GROUP11_VAR2 = 21,
    /// @brief Binary Output Command - Control Relay Output Block
    DNP3_VARIATION_GROUP12_VAR1 = 22,
    /// @brief Binary Output Command Event - command status without time
    DNP3_VARIATION_GROUP13_VAR1 = 23,
    /// @brief Binary Output Command Event - command status with time
    DNP3_VARIATION_GROUP13_VAR2 = 24,
    /// @brief Counter - Default variation
    DNP3_VARIATION_GROUP20_VAR0 = 25,
    /// @brief Counter - 32-bit with flags
    DNP3_VARIATION_GROUP20_VAR1 = 26,
    /// @brief Counter - 16-bit with flags
    DNP3_VARIATION_GROUP20_VAR2 = 27,
    /// @brief Counter - 32-bit without flag
    DNP3_VARIATION_GROUP20_VAR5 = 28,
    /// @brief Counter - 16-bit without flag
    DNP3_VARIATION_GROUP20_VAR6 = 29,
    /// @brief Frozen Counter - Default variation
    DNP3_VARIATION_GROUP21_VAR0 = 30,
    /// @brief Frozen Counter - 32-bit with flags
    DNP3_VARIATION_GROUP21_VAR1 = 31,
    /// @brief Frozen Counter - 16-bit with flags
    DNP3_VARIATION_GROUP21_VAR2 = 32,
    /// @brief Frozen Counter - 32-bit with flags and time
    DNP3_VARIATION_GROUP21_VAR5 = 33,
    /// @brief Frozen Counter - 16-bit with flags and time
    DNP3_VARIATION_GROUP21_VAR6 = 34,
    /// @brief Frozen Counter - 32-bit without flag
    DNP3_VARIATION_GROUP21_VAR9 = 35,
    /// @brief Frozen Counter - 16-bit without flag
    DNP3_VARIATION_GROUP21_VAR10 = 36,
    /// @brief Counter Event - Default variation
    DNP3_VARIATION_GROUP22_VAR0 = 37,
    /// @brief Counter Event - 32-bit with flags
    DNP3_VARIATION_GROUP22_VAR1 = 38,
    /// @brief Counter Event - 16-bit with flags
    DNP3_VARIATION_GROUP22_VAR2 = 39,
    /// @brief Counter Event - 32-bit with flags and time
    DNP3_VARIATION_GROUP22_VAR5 = 40,
    /// @brief Counter Event - 16-bit with flags and time
    DNP3_VARIATION_GROUP22_VAR6 = 41,
    /// @brief Frozen Counter Event - Default variation
    DNP3_VARIATION_GROUP23_VAR0 = 42,
    /// @brief Frozen Counter Event - 32-bit with flags
    DNP3_VARIATION_GROUP23_VAR1 = 43,
    /// @brief Frozen Counter Event - 16-bit with flags
    DNP3_VARIATION_GROUP23_VAR2 = 44,
    /// @brief Frozen Counter Event - 32-bit with flags and time
    DNP3_VARIATION_GROUP23_VAR5 = 45,
    /// @brief Frozen Counter Event - 16-bit with flags and time
    DNP3_VARIATION_GROUP23_VAR6 = 46,
    /// @brief Analog Input - Default variation
    DNP3_VARIATION_GROUP30_VAR0 = 47,
    /// @brief Analog Input - 32-bit with flags
    DNP3_VARIATION_GROUP30_VAR1 = 48,
    /// @brief Analog Input - 16-bit with flags
    DNP3_VARIATION_GROUP30_VAR2 = 49,
    /// @brief Analog Input - 32-bit without flag
    DNP3_VARIATION_GROUP30_VAR3 = 50,
    /// @brief Analog Input - 16-bit without flag
    DNP3_VARIATION_GROUP30_VAR4 = 51,
    /// @brief Analog Input - Single-precision floating point with flags
    DNP3_VARIATION_GROUP30_VAR5 = 52,
    /// @brief Analog Input - Double-precision floating point with flags
    DNP3_VARIATION_GROUP30_VAR6 = 53,
    /// @brief Frozen Analog Input - Default variation
    DNP3_VARIATION_GROUP31_VAR0 = 54,
    /// @brief Frozen Analog Input - 32-bit with flags
    DNP3_VARIATION_GROUP31_VAR1 = 55,
    /// @brief Frozen Analog Input - 16-bit with flags
    DNP3_VARIATION_GROUP31_VAR2 = 56,
    /// @brief Frozen Analog Input - 32-bit with flags and time-of-freeze
    DNP3_VARIATION_GROUP31_VAR3 = 57,
    /// @brief Frozen Analog Input - 16-bit with flags and time-of-freeze
    DNP3_VARIATION_GROUP31_VAR4 = 58,
    /// @brief Frozen Analog Input - 32-bit without flags
    DNP3_VARIATION_GROUP31_VAR5 = 59,
    /// @brief Frozen Analog Input - 16-bit without flags
    DNP3_VARIATION_GROUP31_VAR6 = 60,
    /// @brief Frozen Analog Input - Single-precision floating point with flags
    DNP3_VARIATION_GROUP31_VAR7 = 61,
    /// @brief Frozen Analog Input - Double-precision floating point with flags
    DNP3_VARIATION_GROUP31_VAR8 = 62,
    /// @brief Analog Input Event - Default variation
    DNP3_VARIATION_GROUP32_VAR0 = 63,
    /// @brief Analog Input Event - 32-bit without time
    DNP3_VARIATION_GROUP32_VAR1 = 64,
    /// @brief Analog Input Event - 16-bit without time
    DNP3_VARIATION_GROUP32_VAR2 = 65,
    /// @brief Analog Input Event - 32-bit with time
    DNP3_VARIATION_GROUP32_VAR3 = 66,
    /// @brief Analog Input Event - 16-bit with time
    DNP3_VARIATION_GROUP32_VAR4 = 67,
    /// @brief Analog Input Event - Single-precision floating point without time
    DNP3_VARIATION_GROUP32_VAR5 = 68,
    /// @brief Analog Input Event - Double-precision floating point without time
    DNP3_VARIATION_GROUP32_VAR6 = 69,
    /// @brief Analog Input Event - Single-precision floating point with time
    DNP3_VARIATION_GROUP32_VAR7 = 70,
    /// @brief Analog Input Event - Double-precision floating point with time
    DNP3_VARIATION_GROUP32_VAR8 = 71,
    /// @brief Frozen Analog Input Event - Default variation
    DNP3_VARIATION_GROUP33_VAR0 = 72,
    /// @brief Frozen Analog Input Event - 32-bit without time
    DNP3_VARIATION_GROUP33_VAR1 = 73,
    /// @brief Frozen Analog Input Event - 16-bit without time
    DNP3_VARIATION_GROUP33_VAR2 = 74,
    /// @brief Frozen Analog Input Event - 32-bit with time
    DNP3_VARIATION_GROUP33_VAR3 = 75,
    /// @brief Frozen Analog Input Event - 16-bit with time
    DNP3_VARIATION_GROUP33_VAR4 = 76,
    /// @brief Frozen Analog Input Event - Single-precision floating point without time
    DNP3_VARIATION_GROUP33_VAR5 = 77,
    /// @brief Frozen Analog Input Event - Double-precision floating point without time
    DNP3_VARIATION_GROUP33_VAR6 = 78,
    /// @brief Frozen Analog Input Event - Single-precision floating point with time
    DNP3_VARIATION_GROUP33_VAR7 = 79,
    /// @brief Frozen Analog Input Event - Double-precision floating point with time
    DNP3_VARIATION_GROUP33_VAR8 = 80,
    /// @brief Analog Input Reporting Deadband - Default variation
    DNP3_VARIATION_GROUP34_VAR0 = 81,
    /// @brief Analog Input Reporting Deadband - 16-bit
    DNP3_VARIATION_GROUP34_VAR1 = 82,
    /// @brief Analog Input Reporting Deadband - 32-bit
    DNP3_VARIATION_GROUP34_VAR2 = 83,
    /// @brief Analog Input Reporting Deadband - Single-precision floating point
    DNP3_VARIATION_GROUP34_VAR3 = 84,
    /// @brief Analog Output Status - Default variation
    DNP3_VARIATION_GROUP40_VAR0 = 85,
    /// @brief Analog Output Status - 32-bit with flags
    DNP3_VARIATION_GROUP40_VAR1 = 86,
    /// @brief Analog Output Status - 16-bit with flags
    DNP3_VARIATION_GROUP40_VAR2 = 87,
    /// @brief Analog Output Status - Single-precision floating point with flags
    DNP3_VARIATION_GROUP40_VAR3 = 88,
    /// @brief Analog Output Status - Double-precision floating point with flags
    DNP3_VARIATION_GROUP40_VAR4 = 89,
    /// @brief Analog Output - 32-bit
    DNP3_VARIATION_GROUP41_VAR1 = 90,
    /// @brief Analog Output - 16-bit
    DNP3_VARIATION_GROUP41_VAR2 = 91,
    /// @brief Analog Output - Single-precision floating point
    DNP3_VARIATION_GROUP41_VAR3 = 92,
    /// @brief Analog Output - Double-precision floating point
    DNP3_VARIATION_GROUP41_VAR4 = 93,
    /// @brief Analog Output Event - Default variation
    DNP3_VARIATION_GROUP42_VAR0 = 94,
    /// @brief Analog Output Event - 32-bit without time
    DNP3_VARIATION_GROUP42_VAR1 = 95,
    /// @brief Analog Output Event - 16-bit without time
    DNP3_VARIATION_GROUP42_VAR2 = 96,
    /// @brief Analog Output Event - 32-bit with time
    DNP3_VARIATION_GROUP42_VAR3 = 97,
    /// @brief Analog Output Event - 16-bit with time
    DNP3_VARIATION_GROUP42_VAR4 = 98,
    /// @brief Analog Output Event - Single-precision floating point without time
    DNP3_VARIATION_GROUP42_VAR5 = 99,
    /// @brief Analog Output Event - Double-precision floating point without time
    DNP3_VARIATION_GROUP42_VAR6 = 100,
    /// @brief Analog Output Event - Single-precision floating point with time
    DNP3_VARIATION_GROUP42_VAR7 = 101,
    /// @brief Analog Output Event - Double-precision floating point with time
    DNP3_VARIATION_GROUP42_VAR8 = 102,
    /// @brief Analog Output Command Event - 32-bit without time
    DNP3_VARIATION_GROUP43_VAR1 = 103,
    /// @brief Analog Output Command Event - 16-bit without time
    DNP3_VARIATION_GROUP43_VAR2 = 104,
    /// @brief Analog Output Command Event - 32-bit with time
    DNP3_VARIATION_GROUP43_VAR3 = 105,
    /// @brief Analog Output Command Event - 16-bit with time
    DNP3_VARIATION_GROUP43_VAR4 = 106,
    /// @brief Analog Output Command Event - Single-precision floating point without time
    DNP3_VARIATION_GROUP43_VAR5 = 107,
    /// @brief Analog Output Command Event - Double-precision floating point without time
    DNP3_VARIATION_GROUP43_VAR6 = 108,
    /// @brief Analog Output Command Event - Single-precision floating point with time
    DNP3_VARIATION_GROUP43_VAR7 = 109,
    /// @brief Analog Output Command Event - Double-precision floating point with time
    DNP3_VARIATION_GROUP43_VAR8 = 110,
    /// @brief Time and Date - Absolute time
    DNP3_VARIATION_GROUP50_VAR1 = 111,
    /// @brief Time and Date - Absolute time and interval
    DNP3_VARIATION_GROUP50_VAR2 = 112,
    /// @brief Time and Date - Absolute time at last recorded time
    DNP3_VARIATION_GROUP50_VAR3 = 113,
    /// @brief Time and Date - Indexed absolute time and long interval
    DNP3_VARIATION_GROUP50_VAR4 = 114,
    /// @brief Time and date CTO - Absolute time, synchronized
    DNP3_VARIATION_GROUP51_VAR1 = 115,
    /// @brief Time and date CTO - Absolute time, unsynchronized
    DNP3_VARIATION_GROUP51_VAR2 = 116,
    /// @brief Time delay - Coarse
    DNP3_VARIATION_GROUP52_VAR1 = 117,
    /// @brief Time delay - Fine
    DNP3_VARIATION_GROUP52_VAR2 = 118,
    /// @brief Class objects - Class 0 data
    DNP3_VARIATION_GROUP60_VAR1 = 119,
    /// @brief Class objects - Class 1 data
    DNP3_VARIATION_GROUP60_VAR2 = 120,
    /// @brief Class objects - Class 2 data
    DNP3_VARIATION_GROUP60_VAR3 = 121,
    /// @brief Class objects - Class 3 data
    DNP3_VARIATION_GROUP60_VAR4 = 122,
    /// @brief File control - authentication
    DNP3_VARIATION_GROUP70_VAR2 = 123,
    /// @brief File control - file command
    DNP3_VARIATION_GROUP70_VAR3 = 124,
    /// @brief File control - file command status
    DNP3_VARIATION_GROUP70_VAR4 = 125,
    /// @brief File control - file transport
    DNP3_VARIATION_GROUP70_VAR5 = 126,
    /// @brief File control - file transport status
    DNP3_VARIATION_GROUP70_VAR6 = 127,
    /// @brief File control - file descriptor
    DNP3_VARIATION_GROUP70_VAR7 = 128,
    /// @brief File control - file specification string
    DNP3_VARIATION_GROUP70_VAR8 = 129,
    /// @brief Internal Indications - Packed format
    DNP3_VARIATION_GROUP80_VAR1 = 130,
    /// @brief Unsigned Integer - Default Variation
    DNP3_VARIATION_GROUP102_VAR0 = 131,
    /// @brief Unsigned Integer - 8-bit
    DNP3_VARIATION_GROUP102_VAR1 = 132,
    /// @brief Octet String
    DNP3_VARIATION_GROUP110 = 133,
    /// @brief Octet String Event
    DNP3_VARIATION_GROUP111 = 134,
} dnp3_variation_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_variation_to_string(dnp3_variation_t value)
{
    switch (value)
    {
        case DNP3_VARIATION_GROUP0: return "group0";
        case DNP3_VARIATION_GROUP0_VAR254: return "group0_var254";
        case DNP3_VARIATION_GROUP1_VAR0: return "group1_var0";
        case DNP3_VARIATION_GROUP1_VAR1: return "group1_var1";
        case DNP3_VARIATION_GROUP1_VAR2: return "group1_var2";
        case DNP3_VARIATION_GROUP2_VAR0: return "group2_var0";
        case DNP3_VARIATION_GROUP2_VAR1: return "group2_var1";
        case DNP3_VARIATION_GROUP2_VAR2: return "group2_var2";
        case DNP3_VARIATION_GROUP2_VAR3: return "group2_var3";
        case DNP3_VARIATION_GROUP3_VAR0: return "group3_var0";
        case DNP3_VARIATION_GROUP3_VAR1: return "group3_var1";
        case DNP3_VARIATION_GROUP3_VAR2: return "group3_var2";
        case DNP3_VARIATION_GROUP4_VAR0: return "group4_var0";
        case DNP3_VARIATION_GROUP4_VAR1: return "group4_var1";
        case DNP3_VARIATION_GROUP4_VAR2: return "group4_var2";
        case DNP3_VARIATION_GROUP4_VAR3: return "group4_var3";
        case DNP3_VARIATION_GROUP10_VAR0: return "group10_var0";
        case DNP3_VARIATION_GROUP10_VAR1: return "group10_var1";
        case DNP3_VARIATION_GROUP10_VAR2: return "group10_var2";
        case DNP3_VARIATION_GROUP11_VAR0: return "group11_var0";
        case DNP3_VARIATION_GROUP11_VAR1: return "group11_var1";
        case DNP3_VARIATION_GROUP11_VAR2: return "group11_var2";
        case DNP3_VARIATION_GROUP12_VAR1: return "group12_var1";
        case DNP3_VARIATION_GROUP13_VAR1: return "group13_var1";
        case DNP3_VARIATION_GROUP13_VAR2: return "group13_var2";
        case DNP3_VARIATION_GROUP20_VAR0: return "group20_var0";
        case DNP3_VARIATION_GROUP20_VAR1: return "group20_var1";
        case DNP3_VARIATION_GROUP20_VAR2: return "group20_var2";
        case DNP3_VARIATION_GROUP20_VAR5: return "group20_var5";
        case DNP3_VARIATION_GROUP20_VAR6: return "group20_var6";
        case DNP3_VARIATION_GROUP21_VAR0: return "group21_var0";
        case DNP3_VARIATION_GROUP21_VAR1: return "group21_var1";
        case DNP3_VARIATION_GROUP21_VAR2: return "group21_var2";
        case DNP3_VARIATION_GROUP21_VAR5: return "group21_var5";
        case DNP3_VARIATION_GROUP21_VAR6: return "group21_var6";
        case DNP3_VARIATION_GROUP21_VAR9: return "group21_var9";
        case DNP3_VARIATION_GROUP21_VAR10: return "group21_var10";
        case DNP3_VARIATION_GROUP22_VAR0: return "group22_var0";
        case DNP3_VARIATION_GROUP22_VAR1: return "group22_var1";
        case DNP3_VARIATION_GROUP22_VAR2: return "group22_var2";
        case DNP3_VARIATION_GROUP22_VAR5: return "group22_var5";
        case DNP3_VARIATION_GROUP22_VAR6: return "group22_var6";
        case DNP3_VARIATION_GROUP23_VAR0: return "group23_var0";
        case DNP3_VARIATION_GROUP23_VAR1: return "group23_var1";
        case DNP3_VARIATION_GROUP23_VAR2: return "group23_var2";
        case DNP3_VARIATION_GROUP23_VAR5: return "group23_var5";
        case DNP3_VARIATION_GROUP23_VAR6: return "group23_var6";
        case DNP3_VARIATION_GROUP30_VAR0: return "group30_var0";
        case DNP3_VARIATION_GROUP30_VAR1: return "group30_var1";
        case DNP3_VARIATION_GROUP30_VAR2: return "group30_var2";
        case DNP3_VARIATION_GROUP30_VAR3: return "group30_var3";
        case DNP3_VARIATION_GROUP30_VAR4: return "group30_var4";
        case DNP3_VARIATION_GROUP30_VAR5: return "group30_var5";
        case DNP3_VARIATION_GROUP30_VAR6: return "group30_var6";
        case DNP3_VARIATION_GROUP31_VAR0: return "group31_var0";
        case DNP3_VARIATION_GROUP31_VAR1: return "group31_var1";
        case DNP3_VARIATION_GROUP31_VAR2: return "group31_var2";
        case DNP3_VARIATION_GROUP31_VAR3: return "group31_var3";
        case DNP3_VARIATION_GROUP31_VAR4: return "group31_var4";
        case DNP3_VARIATION_GROUP31_VAR5: return "group31_var5";
        case DNP3_VARIATION_GROUP31_VAR6: return "group31_var6";
        case DNP3_VARIATION_GROUP31_VAR7: return "group31_var7";
        case DNP3_VARIATION_GROUP31_VAR8: return "group31_var8";
        case DNP3_VARIATION_GROUP32_VAR0: return "group32_var0";
        case DNP3_VARIATION_GROUP32_VAR1: return "group32_var1";
        case DNP3_VARIATION_GROUP32_VAR2: return "group32_var2";
        case DNP3_VARIATION_GROUP32_VAR3: return "group32_var3";
        case DNP3_VARIATION_GROUP32_VAR4: return "group32_var4";
        case DNP3_VARIATION_GROUP32_VAR5: return "group32_var5";
        case DNP3_VARIATION_GROUP32_VAR6: return "group32_var6";
        case DNP3_VARIATION_GROUP32_VAR7: return "group32_var7";
        case DNP3_VARIATION_GROUP32_VAR8: return "group32_var8";
        case DNP3_VARIATION_GROUP33_VAR0: return "group33_var0";
        case DNP3_VARIATION_GROUP33_VAR1: return "group33_var1";
        case DNP3_VARIATION_GROUP33_VAR2: return "group33_var2";
        case DNP3_VARIATION_GROUP33_VAR3: return "group33_var3";
        case DNP3_VARIATION_GROUP33_VAR4: return "group33_var4";
        case DNP3_VARIATION_GROUP33_VAR5: return "group33_var5";
        case DNP3_VARIATION_GROUP33_VAR6: return "group33_var6";
        case DNP3_VARIATION_GROUP33_VAR7: return "group33_var7";
        case DNP3_VARIATION_GROUP33_VAR8: return "group33_var8";
        case DNP3_VARIATION_GROUP34_VAR0: return "group34_var0";
        case DNP3_VARIATION_GROUP34_VAR1: return "group34_var1";
        case DNP3_VARIATION_GROUP34_VAR2: return "group34_var2";
        case DNP3_VARIATION_GROUP34_VAR3: return "group34_var3";
        case DNP3_VARIATION_GROUP40_VAR0: return "group40_var0";
        case DNP3_VARIATION_GROUP40_VAR1: return "group40_var1";
        case DNP3_VARIATION_GROUP40_VAR2: return "group40_var2";
        case DNP3_VARIATION_GROUP40_VAR3: return "group40_var3";
        case DNP3_VARIATION_GROUP40_VAR4: return "group40_var4";
        case DNP3_VARIATION_GROUP41_VAR1: return "group41_var1";
        case DNP3_VARIATION_GROUP41_VAR2: return "group41_var2";
        case DNP3_VARIATION_GROUP41_VAR3: return "group41_var3";
        case DNP3_VARIATION_GROUP41_VAR4: return "group41_var4";
        case DNP3_VARIATION_GROUP42_VAR0: return "group42_var0";
        case DNP3_VARIATION_GROUP42_VAR1: return "group42_var1";
        case DNP3_VARIATION_GROUP42_VAR2: return "group42_var2";
        case DNP3_VARIATION_GROUP42_VAR3: return "group42_var3";
        case DNP3_VARIATION_GROUP42_VAR4: return "group42_var4";
        case DNP3_VARIATION_GROUP42_VAR5: return "group42_var5";
        case DNP3_VARIATION_GROUP42_VAR6: return "group42_var6";
        case DNP3_VARIATION_GROUP42_VAR7: return "group42_var7";
        case DNP3_VARIATION_GROUP42_VAR8: return "group42_var8";
        case DNP3_VARIATION_GROUP43_VAR1: return "group43_var1";
        case DNP3_VARIATION_GROUP43_VAR2: return "group43_var2";
        case DNP3_VARIATION_GROUP43_VAR3: return "group43_var3";
        case DNP3_VARIATION_GROUP43_VAR4: return "group43_var4";
        case DNP3_VARIATION_GROUP43_VAR5: return "group43_var5";
        case DNP3_VARIATION_GROUP43_VAR6: return "group43_var6";
        case DNP3_VARIATION_GROUP43_VAR7: return "group43_var7";
        case DNP3_VARIATION_GROUP43_VAR8: return "group43_var8";
        case DNP3_VARIATION_GROUP50_VAR1: return "group50_var1";
        case DNP3_VARIATION_GROUP50_VAR2: return "group50_var2";
        case DNP3_VARIATION_GROUP50_VAR3: return "group50_var3";
        case DNP3_VARIATION_GROUP50_VAR4: return "group50_var4";
        case DNP3_VARIATION_GROUP51_VAR1: return "group51_var1";
        case DNP3_VARIATION_GROUP51_VAR2: return "group51_var2";
        case DNP3_VARIATION_GROUP52_VAR1: return "group52_var1";
        case DNP3_VARIATION_GROUP52_VAR2: return "group52_var2";
        case DNP3_VARIATION_GROUP60_VAR1: return "group60_var1";
        case DNP3_VARIATION_GROUP60_VAR2: return "group60_var2";
        case DNP3_VARIATION_GROUP60_VAR3: return "group60_var3";
        case DNP3_VARIATION_GROUP60_VAR4: return "group60_var4";
        case DNP3_VARIATION_GROUP70_VAR2: return "group70_var2";
        case DNP3_VARIATION_GROUP70_VAR3: return "group70_var3";
        case DNP3_VARIATION_GROUP70_VAR4: return "group70_var4";
        case DNP3_VARIATION_GROUP70_VAR5: return "group70_var5";
        case DNP3_VARIATION_GROUP70_VAR6: return "group70_var6";
        case DNP3_VARIATION_GROUP70_VAR7: return "group70_var7";
        case DNP3_VARIATION_GROUP70_VAR8: return "group70_var8";
        case DNP3_VARIATION_GROUP80_VAR1: return "group80_var1";
        case DNP3_VARIATION_GROUP102_VAR0: return "group102_var0";
        case DNP3_VARIATION_GROUP102_VAR1: return "group102_var1";
        case DNP3_VARIATION_GROUP110: return "group110";
        case DNP3_VARIATION_GROUP111: return "group111";
        default: return "unknown variation value";
    }
}

typedef struct dnp3_retry_strategy_t dnp3_retry_strategy_t;

/// @brief Retry strategy configuration.
/// 
/// The strategy uses an exponential back-off with a minimum and maximum value.
typedef struct dnp3_retry_strategy_t
{
    /// @brief Minimum delay between two retries
    /// @note The unit is milliseconds
    uint64_t min_delay;
    /// @brief Maximum delay between two retries
    /// @note The unit is milliseconds
    uint64_t max_delay;
} dnp3_retry_strategy_t;

/// @brief Initialize to defaults
/// 
/// @note Values are initialized to:
/// - @ref dnp3_retry_strategy_t.min_delay : 1000ms
/// - @ref dnp3_retry_strategy_t.max_delay : 10000ms
/// 
/// @returns New instance of @ref dnp3_retry_strategy_t
static dnp3_retry_strategy_t dnp3_retry_strategy_init()
{
    dnp3_retry_strategy_t _return_value = {
        1000,
        10000
    };
    return _return_value;
}


/// @brief Number of bits per character
typedef enum dnp3_data_bits_t
{
    /// @brief 5 bits per character
    DNP3_DATA_BITS_FIVE = 0,
    /// @brief 6 bits per character
    DNP3_DATA_BITS_SIX = 1,
    /// @brief 7 bits per character
    DNP3_DATA_BITS_SEVEN = 2,
    /// @brief 8 bits per character
    DNP3_DATA_BITS_EIGHT = 3,
} dnp3_data_bits_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_data_bits_to_string(dnp3_data_bits_t value)
{
    switch (value)
    {
        case DNP3_DATA_BITS_FIVE: return "five";
        case DNP3_DATA_BITS_SIX: return "six";
        case DNP3_DATA_BITS_SEVEN: return "seven";
        case DNP3_DATA_BITS_EIGHT: return "eight";
        default: return "unknown data_bits value";
    }
}

/// @brief Flow control modes
typedef enum dnp3_flow_control_t
{
    /// @brief No flow control
    DNP3_FLOW_CONTROL_NONE = 0,
    /// @brief Flow control using XON/XOFF bytes
    DNP3_FLOW_CONTROL_SOFTWARE = 1,
    /// @brief Flow control using RTS/CTS signals
    DNP3_FLOW_CONTROL_HARDWARE = 2,
} dnp3_flow_control_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_flow_control_to_string(dnp3_flow_control_t value)
{
    switch (value)
    {
        case DNP3_FLOW_CONTROL_NONE: return "none";
        case DNP3_FLOW_CONTROL_SOFTWARE: return "software";
        case DNP3_FLOW_CONTROL_HARDWARE: return "hardware";
        default: return "unknown flow_control value";
    }
}

/// @brief Parity checking modes
typedef enum dnp3_parity_t
{
    /// @brief No parity bit
    DNP3_PARITY_NONE = 0,
    /// @brief Parity bit sets odd number of 1 bits
    DNP3_PARITY_ODD = 1,
    /// @brief Parity bit sets even number of 1 bits
    DNP3_PARITY_EVEN = 2,
} dnp3_parity_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_parity_to_string(dnp3_parity_t value)
{
    switch (value)
    {
        case DNP3_PARITY_NONE: return "none";
        case DNP3_PARITY_ODD: return "odd";
        case DNP3_PARITY_EVEN: return "even";
        default: return "unknown parity value";
    }
}

/// @brief Number of stop bits
typedef enum dnp3_stop_bits_t
{
    /// @brief One stop bit
    DNP3_STOP_BITS_ONE = 0,
    /// @brief Two stop bits
    DNP3_STOP_BITS_TWO = 1,
} dnp3_stop_bits_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_stop_bits_to_string(dnp3_stop_bits_t value)
{
    switch (value)
    {
        case DNP3_STOP_BITS_ONE: return "one";
        case DNP3_STOP_BITS_TWO: return "two";
        default: return "unknown stop_bits value";
    }
}

typedef struct dnp3_serial_settings_t dnp3_serial_settings_t;

/// @brief Serial port settings
typedef struct dnp3_serial_settings_t
{
    /// @brief Baud rate (in symbols-per-second)
    uint32_t baud_rate;
    /// @brief Number of bits used to represent a character sent on the line
    dnp3_data_bits_t data_bits;
    /// @brief Type of signalling to use for controlling data transfer
    dnp3_flow_control_t flow_control;
    /// @brief Type of parity to use for error checking
    dnp3_parity_t parity;
    /// @brief Number of bits to use to signal the end of a character
    dnp3_stop_bits_t stop_bits;
} dnp3_serial_settings_t;

/// @brief Initialize to default values
/// 
/// @note Values are initialized to:
/// - @ref dnp3_serial_settings_t.baud_rate : 9600
/// - @ref dnp3_serial_settings_t.data_bits : @ref DNP3_DATA_BITS_EIGHT
/// - @ref dnp3_serial_settings_t.flow_control : @ref DNP3_FLOW_CONTROL_NONE
/// - @ref dnp3_serial_settings_t.parity : @ref DNP3_PARITY_NONE
/// - @ref dnp3_serial_settings_t.stop_bits : @ref DNP3_STOP_BITS_ONE
/// 
/// @returns New instance of @ref dnp3_serial_settings_t
static dnp3_serial_settings_t dnp3_serial_settings_init()
{
    dnp3_serial_settings_t _return_value = {
        9600,
        DNP3_DATA_BITS_EIGHT,
        DNP3_FLOW_CONTROL_NONE,
        DNP3_PARITY_NONE,
        DNP3_STOP_BITS_ONE
    };
    return _return_value;
}


/// @brief Controls how errors in parsed link-layer frames are handled. This behavior is configurable for physical layers with built-in error correction like TCP as the connection might be through a terminal server.
typedef enum dnp3_link_error_mode_t
{
    /// @brief Framing errors are discarded. The link-layer parser is reset on any error, and the parser begins scanning for 0x0564. This is always the behavior for serial ports.
    DNP3_LINK_ERROR_MODE_DISCARD = 0,
    /// @brief Framing errors are bubbled up to calling code, closing the session. Suitable for physical layers that provide error correction like TCP.
    DNP3_LINK_ERROR_MODE_CLOSE = 1,
} dnp3_link_error_mode_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_link_error_mode_to_string(dnp3_link_error_mode_t value)
{
    switch (value)
    {
        case DNP3_LINK_ERROR_MODE_DISCARD: return "discard";
        case DNP3_LINK_ERROR_MODE_CLOSE: return "close";
        default: return "unknown link_error_mode value";
    }
}

/// @brief Application layer function code
typedef enum dnp3_function_code_t
{
    /// @brief Master sends this to an outstation to confirm the receipt of an Application Layer fragment (value == 0)
    DNP3_FUNCTION_CODE_CONFIRM = 0,
    /// @brief Outstation shall return the data specified by the objects in the request (value == 1)
    DNP3_FUNCTION_CODE_READ = 1,
    /// @brief Outstation shall store the data specified by the objects in the request (value == 2)
    DNP3_FUNCTION_CODE_WRITE = 2,
    /// @brief Outstation shall select (or arm) the output points specified by the objects in the request in preparation for a subsequent operate command (value == 3)
    DNP3_FUNCTION_CODE_SELECT = 3,
    /// @brief Outstation shall activate the output points selected (or armed) by a previous select function code command (value == 4)
    DNP3_FUNCTION_CODE_OPERATE = 4,
    /// @brief Outstation shall immediately actuate the output points specified by the objects in the request (value == 5)
    DNP3_FUNCTION_CODE_DIRECT_OPERATE = 5,
    /// @brief Same as DirectOperate but outstation shall not send a response (value == 6)
    DNP3_FUNCTION_CODE_DIRECT_OPERATE_NO_RESPONSE = 6,
    /// @brief Outstation shall copy the point data values specified by the objects in the request to a separate freeze buffer (value == 7)
    DNP3_FUNCTION_CODE_IMMEDIATE_FREEZE = 7,
    /// @brief Same as ImmediateFreeze but outstation shall not send a response (value == 8)
    DNP3_FUNCTION_CODE_IMMEDIATE_FREEZE_NO_RESPONSE = 8,
    /// @brief Outstation shall copy the point data values specified by the objects in the request into a separate freeze buffer and then clear the values (value == 9)
    DNP3_FUNCTION_CODE_FREEZE_CLEAR = 9,
    /// @brief Same as FreezeClear but outstation shall not send a response (value == 10)
    DNP3_FUNCTION_CODE_FREEZE_CLEAR_NO_RESPONSE = 10,
    /// @brief Outstation shall copy the point data values specified by the objects in the request to a separate freeze buffer at the time and/or time intervals specified in a special time data information object (value == 11)
    DNP3_FUNCTION_CODE_FREEZE_AT_TIME = 11,
    /// @brief Same as FreezeAtTime but outstation shall not send a response (value == 12)
    DNP3_FUNCTION_CODE_FREEZE_AT_TIME_NO_RESPONSE = 12,
    /// @brief Outstation shall perform a complete reset of all hardware and software in the device (value == 13)
    DNP3_FUNCTION_CODE_COLD_RESTART = 13,
    /// @brief Outstation shall reset only portions of the device (value == 14)
    DNP3_FUNCTION_CODE_WARM_RESTART = 14,
    /// @brief Obsolete-Do not use for new designs (value == 15)
    DNP3_FUNCTION_CODE_INITIALIZE_DATA = 15,
    /// @brief Outstation shall place the applications specified by the objects in the request into the ready to run state (value == 16)
    DNP3_FUNCTION_CODE_INITIALIZE_APPLICATION = 16,
    /// @brief Outstation shall start running the applications specified by the objects in the request (value == 17)
    DNP3_FUNCTION_CODE_START_APPLICATION = 17,
    /// @brief Outstation shall stop running the applications specified by the objects in the request (value == 18)
    DNP3_FUNCTION_CODE_STOP_APPLICATION = 18,
    /// @brief This code is deprecated-Do not use for new designs (value == 19)
    DNP3_FUNCTION_CODE_SAVE_CONFIGURATION = 19,
    /// @brief Enables outstation to initiate unsolicited responses from points specified by the objects in the request (value == 20)
    DNP3_FUNCTION_CODE_ENABLE_UNSOLICITED = 20,
    /// @brief Prevents outstation from initiating unsolicited responses from points specified by the objects in the request (value == 21)
    DNP3_FUNCTION_CODE_DISABLE_UNSOLICITED = 21,
    /// @brief Outstation shall assign the events generated by the points specified by the objects in the request to one of the classes (value == 22)
    DNP3_FUNCTION_CODE_ASSIGN_CLASS = 22,
    /// @brief Outstation shall report the time it takes to process and initiate the transmission of its response (value == 23)
    DNP3_FUNCTION_CODE_DELAY_MEASURE = 23,
    /// @brief Outstation shall save the time when the last octet of this message is received (value == 24)
    DNP3_FUNCTION_CODE_RECORD_CURRENT_TIME = 24,
    /// @brief Outstation shall open a file (value == 25)
    DNP3_FUNCTION_CODE_OPEN_FILE = 25,
    /// @brief Outstation shall close a file (value == 26)
    DNP3_FUNCTION_CODE_CLOSE_FILE = 26,
    /// @brief Outstation shall delete a file (value == 27)
    DNP3_FUNCTION_CODE_DELETE_FILE = 27,
    /// @brief Outstation shall retrieve information about a file (value == 28)
    DNP3_FUNCTION_CODE_GET_FILE_INFO = 28,
    /// @brief Outstation shall return a file authentication key (value == 29)
    DNP3_FUNCTION_CODE_AUTHENTICATE_FILE = 29,
    /// @brief Outstation shall abort a file transfer operation (value == 30)
    DNP3_FUNCTION_CODE_ABORT_FILE = 30,
    /// @brief Master shall interpret this fragment as an Application Layer response to an ApplicationLayer request (value == 129)
    DNP3_FUNCTION_CODE_RESPONSE = 31,
    /// @brief Master shall interpret this fragment as an unsolicited response that was not prompted by an explicit request (value == 130)
    DNP3_FUNCTION_CODE_UNSOLICITED_RESPONSE = 32,
} dnp3_function_code_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_function_code_to_string(dnp3_function_code_t value)
{
    switch (value)
    {
        case DNP3_FUNCTION_CODE_CONFIRM: return "confirm";
        case DNP3_FUNCTION_CODE_READ: return "read";
        case DNP3_FUNCTION_CODE_WRITE: return "write";
        case DNP3_FUNCTION_CODE_SELECT: return "select";
        case DNP3_FUNCTION_CODE_OPERATE: return "operate";
        case DNP3_FUNCTION_CODE_DIRECT_OPERATE: return "direct_operate";
        case DNP3_FUNCTION_CODE_DIRECT_OPERATE_NO_RESPONSE: return "direct_operate_no_response";
        case DNP3_FUNCTION_CODE_IMMEDIATE_FREEZE: return "immediate_freeze";
        case DNP3_FUNCTION_CODE_IMMEDIATE_FREEZE_NO_RESPONSE: return "immediate_freeze_no_response";
        case DNP3_FUNCTION_CODE_FREEZE_CLEAR: return "freeze_clear";
        case DNP3_FUNCTION_CODE_FREEZE_CLEAR_NO_RESPONSE: return "freeze_clear_no_response";
        case DNP3_FUNCTION_CODE_FREEZE_AT_TIME: return "freeze_at_time";
        case DNP3_FUNCTION_CODE_FREEZE_AT_TIME_NO_RESPONSE: return "freeze_at_time_no_response";
        case DNP3_FUNCTION_CODE_COLD_RESTART: return "cold_restart";
        case DNP3_FUNCTION_CODE_WARM_RESTART: return "warm_restart";
        case DNP3_FUNCTION_CODE_INITIALIZE_DATA: return "initialize_data";
        case DNP3_FUNCTION_CODE_INITIALIZE_APPLICATION: return "initialize_application";
        case DNP3_FUNCTION_CODE_START_APPLICATION: return "start_application";
        case DNP3_FUNCTION_CODE_STOP_APPLICATION: return "stop_application";
        case DNP3_FUNCTION_CODE_SAVE_CONFIGURATION: return "save_configuration";
        case DNP3_FUNCTION_CODE_ENABLE_UNSOLICITED: return "enable_unsolicited";
        case DNP3_FUNCTION_CODE_DISABLE_UNSOLICITED: return "disable_unsolicited";
        case DNP3_FUNCTION_CODE_ASSIGN_CLASS: return "assign_class";
        case DNP3_FUNCTION_CODE_DELAY_MEASURE: return "delay_measure";
        case DNP3_FUNCTION_CODE_RECORD_CURRENT_TIME: return "record_current_time";
        case DNP3_FUNCTION_CODE_OPEN_FILE: return "open_file";
        case DNP3_FUNCTION_CODE_CLOSE_FILE: return "close_file";
        case DNP3_FUNCTION_CODE_DELETE_FILE: return "delete_file";
        case DNP3_FUNCTION_CODE_GET_FILE_INFO: return "get_file_info";
        case DNP3_FUNCTION_CODE_AUTHENTICATE_FILE: return "authenticate_file";
        case DNP3_FUNCTION_CODE_ABORT_FILE: return "abort_file";
        case DNP3_FUNCTION_CODE_RESPONSE: return "response";
        case DNP3_FUNCTION_CODE_UNSOLICITED_RESPONSE: return "unsolicited_response";
        default: return "unknown function_code value";
    }
}

/// @brief File type enumeration used in Group 70 objects
typedef enum dnp3_file_type_t
{
    /// @brief File is a directory
    DNP3_FILE_TYPE_DIRECTORY = 0,
    /// @brief File is a simple file type suitable for sequential file transfer
    DNP3_FILE_TYPE_SIMPLE = 1,
    /// @brief Some other unspecified value
    DNP3_FILE_TYPE_OTHER = 2,
} dnp3_file_type_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_file_type_to_string(dnp3_file_type_t value)
{
    switch (value)
    {
        case DNP3_FILE_TYPE_DIRECTORY: return "directory";
        case DNP3_FILE_TYPE_SIMPLE: return "simple";
        case DNP3_FILE_TYPE_OTHER: return "other";
        default: return "unknown file_type value";
    }
}

typedef struct dnp3_permission_set_t dnp3_permission_set_t;

/// @brief Defines read, write, execute permissions for particular group or user
typedef struct dnp3_permission_set_t
{
    /// @brief Permission to execute
    bool execute;
    /// @brief Permission to write
    bool write;
    /// @brief Permission to read
    bool read;
} dnp3_permission_set_t;

/// @brief Fully construct @ref dnp3_permission_set_t specifying the value of each field
/// @param execute Permission to execute
/// @param write Permission to write
/// @param read Permission to read
/// @returns New instance of @ref dnp3_permission_set_t
static dnp3_permission_set_t dnp3_permission_set_init(bool execute, bool write, bool read)
{
    dnp3_permission_set_t _return_value = {
        execute,
        write,
        read
    };
    return _return_value;
}

/// @brief Permission set with nothing enabled
/// 
/// @note Values are initialized to:
/// - @ref dnp3_permission_set_t.execute : @p false
/// - @ref dnp3_permission_set_t.write : @p false
/// - @ref dnp3_permission_set_t.read : @p false
/// 
/// @returns New instance of @ref dnp3_permission_set_t
static dnp3_permission_set_t dnp3_permission_set_none()
{
    dnp3_permission_set_t _return_value = {
        false,
        false,
        false
    };
    return _return_value;
}


typedef struct dnp3_permissions_t dnp3_permissions_t;

/// @brief Permissions for world, group, and owner
typedef struct dnp3_permissions_t
{
    /// @brief World permissions
    dnp3_permission_set_t world;
    /// @brief Group permissions
    dnp3_permission_set_t group;
    /// @brief Owner permissions
    dnp3_permission_set_t owner;
} dnp3_permissions_t;

/// @brief Fully construct @ref dnp3_permissions_t specifying the value of each field
/// @param world World permissions
/// @param group Group permissions
/// @param owner Owner permissions
/// @returns New instance of @ref dnp3_permissions_t
static dnp3_permissions_t dnp3_permissions_init(dnp3_permission_set_t world, dnp3_permission_set_t group, dnp3_permission_set_t owner)
{
    dnp3_permissions_t _return_value = {
        world,
        group,
        owner
    };
    return _return_value;
}

/// @brief Permissions with nothing enabled
/// 
/// @note Values are initialized to:
/// - @ref dnp3_permissions_t.world : Default @ref dnp3_permission_set_t
/// - @ref dnp3_permissions_t.group : Default @ref dnp3_permission_set_t
/// - @ref dnp3_permissions_t.owner : Default @ref dnp3_permission_set_t
/// 
/// @returns New instance of @ref dnp3_permissions_t
static dnp3_permissions_t dnp3_permissions_none()
{
    dnp3_permissions_t _return_value = {
        dnp3_permission_set_none(),
        dnp3_permission_set_none(),
        dnp3_permission_set_none()
    };
    return _return_value;
}


typedef struct dnp3_open_file_t dnp3_open_file_t;

/// @brief The result of opening a file on the outstation
typedef struct dnp3_open_file_t
{
    /// @brief The handle assigned to the file by the outstation
    /// 
    /// This must be used in subsequent requests to manipulate the file
    uint32_t file_handle;
    /// @brief Size of the file returned by the outstation
    uint32_t file_size;
    /// @brief Maximum block size returned by the outstation
    /// 
    /// The master must respect this value when writing data to a file or the transfer may not succeed
    uint16_t max_block_size;
} dnp3_open_file_t;


/// @brief Errors that can occur during file transfer
typedef enum dnp3_file_error_t
{
    /// @brief Success, i.e. no error occurred
    DNP3_FILE_ERROR_OK = 0,
    /// @brief Outstation returned an error status code
    DNP3_FILE_ERROR_BAD_STATUS = 1,
    /// @brief Outstation indicated no permission to access file
    DNP3_FILE_ERROR_NO_PERMISSION = 2,
    /// @brief Received an unexpected block number
    DNP3_FILE_ERROR_BAD_BLOCK_NUM = 3,
    /// @brief File transfer aborted by user
    DNP3_FILE_ERROR_ABORT_BY_USER = 4,
    /// @brief Exceeded the maximum length specified by the user
    DNP3_FILE_ERROR_MAX_LENGTH_EXCEEDED = 5,
    /// @brief File handle returned by the outstation did not match the request
    DNP3_FILE_ERROR_WRONG_HANDLE = 6,
    /// @brief too many user requests queued
    DNP3_FILE_ERROR_TOO_MANY_REQUESTS = 7,
    /// @brief outstation returned an IIN.2 error bit
    DNP3_FILE_ERROR_IIN_ERROR = 8,
    /// @brief response was malformed or contained object headers
    DNP3_FILE_ERROR_BAD_RESPONSE = 9,
    /// @brief timeout occurred before receiving a response
    DNP3_FILE_ERROR_RESPONSE_TIMEOUT = 10,
    /// @brief insufficient buffer space to serialize the request
    DNP3_FILE_ERROR_WRITE_ERROR = 11,
    /// @brief no connection
    DNP3_FILE_ERROR_NO_CONNECTION = 12,
    /// @brief master was shutdown
    DNP3_FILE_ERROR_SHUTDOWN = 13,
    /// @brief association was removed mid-task
    DNP3_FILE_ERROR_ASSOCIATION_REMOVED = 14,
    /// @brief request data could not be encoded
    DNP3_FILE_ERROR_BAD_ENCODING = 15,
} dnp3_file_error_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_file_error_to_string(dnp3_file_error_t value)
{
    switch (value)
    {
        case DNP3_FILE_ERROR_OK: return "ok";
        case DNP3_FILE_ERROR_BAD_STATUS: return "bad_status";
        case DNP3_FILE_ERROR_NO_PERMISSION: return "no_permission";
        case DNP3_FILE_ERROR_BAD_BLOCK_NUM: return "bad_block_num";
        case DNP3_FILE_ERROR_ABORT_BY_USER: return "abort_by_user";
        case DNP3_FILE_ERROR_MAX_LENGTH_EXCEEDED: return "max_length_exceeded";
        case DNP3_FILE_ERROR_WRONG_HANDLE: return "wrong_handle";
        case DNP3_FILE_ERROR_TOO_MANY_REQUESTS: return "too_many_requests";
        case DNP3_FILE_ERROR_IIN_ERROR: return "iin_error";
        case DNP3_FILE_ERROR_BAD_RESPONSE: return "bad_response";
        case DNP3_FILE_ERROR_RESPONSE_TIMEOUT: return "response_timeout";
        case DNP3_FILE_ERROR_WRITE_ERROR: return "write_error";
        case DNP3_FILE_ERROR_NO_CONNECTION: return "no_connection";
        case DNP3_FILE_ERROR_SHUTDOWN: return "shutdown";
        case DNP3_FILE_ERROR_ASSOCIATION_REMOVED: return "association_removed";
        case DNP3_FILE_ERROR_BAD_ENCODING: return "bad_encoding";
        default: return "unknown file_error value";
    }
}


typedef struct dnp3_file_info_t dnp3_file_info_t;

/// @brief Information about a file or directory returned from the outstation
/// 
/// This is a user-facing representation of Group 70 Variation 7
typedef struct dnp3_file_info_t
{
    /// @brief Name of the file or directory
    const char* file_name;
    /// @brief Simple file or directory
    dnp3_file_type_t file_type;
    /// @brief Size of the file in bytes
    /// 
    /// If a directory, this represents the number of files and directories contained within.
    uint32_t size;
    /// @brief DNP3 timestamp
    /// 
    /// Milliseconds since January 1st, 1970 UTC. Only the lower 48-bits are used
    uint64_t time_created;
    /// @brief Outstation permissions for the file
    dnp3_permissions_t permissions;
} dnp3_file_info_t;


/// @brief Callback interface used when opening a file
typedef struct dnp3_file_open_callback_t
{
    
    /// @brief Invoked when the asynchronous operation completes successfully
    /// @param result Value describing the open file
    /// @param ctx Context data
    void (*on_complete)(dnp3_open_file_t, void*);
    
    /// @brief Invoked when the asynchronous operation fails
    /// @param error Enumeration indicating which error occurred
    /// @param ctx Context data
    void (*on_failure)(dnp3_file_error_t, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_file_open_callback_t;

/// @brief Callback interface used when closing a file or writing a block of file data
typedef struct dnp3_file_operation_callback_t
{
    
    /// @brief Invoked when the asynchronous operation completes successfully
    /// @param result Indicates a successful write operation
    /// @param ctx Context data
    void (*on_complete)(dnp3_nothing_t, void*);
    
    /// @brief Invoked when the asynchronous operation fails
    /// @param error Enumeration indicating which error occurred
    /// @param ctx Context data
    void (*on_failure)(dnp3_file_error_t, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_file_operation_callback_t;

/// @brief Callback interface used when obtaining an authentication key
typedef struct dnp3_file_auth_callback_t
{
    
    /// @brief Invoked when the asynchronous operation completes successfully
    /// @param result File authentication key
    /// @param ctx Context data
    void (*on_complete)(uint32_t, void*);
    
    /// @brief Invoked when the asynchronous operation fails
    /// @param error Enumeration indicating which error occurred
    /// @param ctx Context data
    void (*on_failure)(dnp3_file_error_t, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_file_auth_callback_t;

/// @brief Different modes in which files may be opened
typedef enum dnp3_file_mode_t
{
    /// @brief Specifies that an existing file is to be opened for reading
    DNP3_FILE_MODE_READ = 0,
    /// @brief Specifies that the file is to be opened for writing, truncating any existing file to length 0
    DNP3_FILE_MODE_WRITE = 1,
    /// @brief Specifies that the file is to be opened for writing, appending to the end of the file
    DNP3_FILE_MODE_APPEND = 2,
} dnp3_file_mode_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_file_mode_to_string(dnp3_file_mode_t value)
{
    switch (value)
    {
        case DNP3_FILE_MODE_READ: return "read";
        case DNP3_FILE_MODE_WRITE: return "write";
        case DNP3_FILE_MODE_APPEND: return "append";
        default: return "unknown file_mode value";
    }
}

/// @brief Describes how the UDP socket reads and writes datagrams from remote endpoint(s)
typedef enum dnp3_udp_socket_mode_t
{
    /// @brief The UDP endpoint will only communicate with the specified remote endpoint
    DNP3_UDP_SOCKET_MODE_ONE_TO_ONE = 0,
    /// @brief The UDP endpoint will accept packets any remote endpoint.
    /// 
    /// When this mode is used with an outstation, the outstation will respond to the address from which the request was sent. It will use the supplied remote endpoint only for sending unsolicited responses.
    DNP3_UDP_SOCKET_MODE_ONE_TO_MANY = 1,
} dnp3_udp_socket_mode_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_udp_socket_mode_to_string(dnp3_udp_socket_mode_t value)
{
    switch (value)
    {
        case DNP3_UDP_SOCKET_MODE_ONE_TO_ONE: return "one_to_one";
        case DNP3_UDP_SOCKET_MODE_ONE_TO_MANY: return "one_to_many";
        default: return "unknown udp_socket_mode value";
    }
}

/// @brief Controls how the link-layer parser treats frames that span multiple calls to read of the physical layer.
/// 
/// UDP is unique in that the specification requires that link layer frames be wholly contained within datagrams, but this can be relaxed by configuration.
typedef enum dnp3_link_read_mode_t
{
    /// @brief Reading from a stream (TCP, serial, etc.) where link-layer frames MAY span separate calls to read
    DNP3_LINK_READ_MODE_STREAM = 0,
    /// @brief Reading datagrams (UDP) where link-layer frames MAY NOT span separate calls to read
    DNP3_LINK_READ_MODE_DATAGRAM = 1,
} dnp3_link_read_mode_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_link_read_mode_to_string(dnp3_link_read_mode_t value)
{
    switch (value)
    {
        case DNP3_LINK_READ_MODE_STREAM: return "stream";
        case DNP3_LINK_READ_MODE_DATAGRAM: return "datagram";
        default: return "unknown link_read_mode value";
    }
}

/// @brief Log level
/// 
/// Used in @ref dnp3_logger_t.on_message callback to identify the log level of a message.
typedef enum dnp3_log_level_t
{
    /// @brief Error log level
    DNP3_LOG_LEVEL_ERROR = 0,
    /// @brief Warning log level
    DNP3_LOG_LEVEL_WARN = 1,
    /// @brief Information log level
    DNP3_LOG_LEVEL_INFO = 2,
    /// @brief Debugging log level
    DNP3_LOG_LEVEL_DEBUG = 3,
    /// @brief Trace log level
    DNP3_LOG_LEVEL_TRACE = 4,
} dnp3_log_level_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_log_level_to_string(dnp3_log_level_t value)
{
    switch (value)
    {
        case DNP3_LOG_LEVEL_ERROR: return "error";
        case DNP3_LOG_LEVEL_WARN: return "warn";
        case DNP3_LOG_LEVEL_INFO: return "info";
        case DNP3_LOG_LEVEL_DEBUG: return "debug";
        case DNP3_LOG_LEVEL_TRACE: return "trace";
        default: return "unknown log_level value";
    }
}

typedef struct dnp3_logging_config_t dnp3_logging_config_t;

/// @brief Describes how each log event is formatted
typedef enum dnp3_log_output_format_t
{
    /// @brief A simple text-based format
    DNP3_LOG_OUTPUT_FORMAT_TEXT = 0,
    /// @brief Output formatted as JSON
    DNP3_LOG_OUTPUT_FORMAT_JSON = 1,
} dnp3_log_output_format_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_log_output_format_to_string(dnp3_log_output_format_t value)
{
    switch (value)
    {
        case DNP3_LOG_OUTPUT_FORMAT_TEXT: return "text";
        case DNP3_LOG_OUTPUT_FORMAT_JSON: return "json";
        default: return "unknown log_output_format value";
    }
}

/// @brief Describes if and how the time will be formatted in log messages
typedef enum dnp3_time_format_t
{
    /// @brief Don't format the timestamp as part of the message
    DNP3_TIME_FORMAT_NONE = 0,
    /// @brief Format the time using RFC 3339
    DNP3_TIME_FORMAT_RFC_3339 = 1,
    /// @brief Format the time in a human readable format e.g. 'Jun 25 14:27:12.955'
    DNP3_TIME_FORMAT_SYSTEM = 2,
} dnp3_time_format_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_time_format_to_string(dnp3_time_format_t value)
{
    switch (value)
    {
        case DNP3_TIME_FORMAT_NONE: return "none";
        case DNP3_TIME_FORMAT_RFC_3339: return "rfc_3339";
        case DNP3_TIME_FORMAT_SYSTEM: return "system";
        default: return "unknown time_format value";
    }
}

/// @brief Logging configuration options
typedef struct dnp3_logging_config_t
{
    /// @brief logging level
    dnp3_log_level_t level;
    /// @brief output formatting options
    dnp3_log_output_format_t output_format;
    /// @brief optional time format
    dnp3_time_format_t time_format;
    /// @brief optionally print the log level as part to the message string
    bool print_level;
    /// @brief optionally print the underlying Rust module information to the message string
    bool print_module_info;
} dnp3_logging_config_t;

/// @brief Initialize the configuration to default values
/// 
/// @note Values are initialized to:
/// - @ref dnp3_logging_config_t.level : @ref DNP3_LOG_LEVEL_INFO
/// - @ref dnp3_logging_config_t.output_format : @ref DNP3_LOG_OUTPUT_FORMAT_TEXT
/// - @ref dnp3_logging_config_t.time_format : @ref DNP3_TIME_FORMAT_SYSTEM
/// - @ref dnp3_logging_config_t.print_level : @p true
/// - @ref dnp3_logging_config_t.print_module_info : @p false
/// 
/// @returns New instance of @ref dnp3_logging_config_t
static dnp3_logging_config_t dnp3_logging_config_init()
{
    dnp3_logging_config_t _return_value = {
        DNP3_LOG_LEVEL_INFO,
        DNP3_LOG_OUTPUT_FORMAT_TEXT,
        DNP3_TIME_FORMAT_SYSTEM,
        true,
        false
    };
    return _return_value;
}


/// @brief Logging interface that receives the log messages and writes them somewhere.
typedef struct dnp3_logger_t
{
    
    /// @brief Called when a log message was received and should be logged
    /// @param level Level of the message
    /// @param message Actual formatted message
    /// @param ctx Context data
    void (*on_message)(dnp3_log_level_t, const char*, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_logger_t;

/// @brief Set the callback that will receive all the log messages
/// 
/// There is only a single globally allocated logger. Calling this method a second time will return an error.
/// 
/// If this method is never called, no logging will be performed.
/// @param config Configuration options for logging
/// @param logger Logger that will receive each logged message
/// @return Error code
dnp3_param_error_t dnp3_configure_logging(dnp3_logging_config_t config, dnp3_logger_t logger);


/// @brief Type of response
typedef enum dnp3_response_function_t
{
    /// @brief Solicited response
    DNP3_RESPONSE_FUNCTION_RESPONSE = 0,
    /// @brief Unsolicited response
    DNP3_RESPONSE_FUNCTION_UNSOLICITED_RESPONSE = 1,
} dnp3_response_function_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_response_function_to_string(dnp3_response_function_t value)
{
    switch (value)
    {
        case DNP3_RESPONSE_FUNCTION_RESPONSE: return "response";
        case DNP3_RESPONSE_FUNCTION_UNSOLICITED_RESPONSE: return "unsolicited_response";
        default: return "unknown response_function value";
    }
}

typedef struct dnp3_iin1_t dnp3_iin1_t;

/// @brief First IIN byte
typedef struct dnp3_iin1_t
{
    /// @brief Broadcast message was received
    bool broadcast;
    /// @brief Outstation has unreported Class 1 events
    bool class_1_events;
    /// @brief Outstation has unreported Class 2 events
    bool class_2_events;
    /// @brief Outstation has unreported Class 3 events
    bool class_3_events;
    /// @brief Outstation requires time synchronization
    bool need_time;
    /// @brief One or more of the outstation’s points are in local control mode
    bool local_control;
    /// @brief An abnormal, device-specific condition exists in the outstation
    bool device_trouble;
    /// @brief Outstation restarted
    bool device_restart;
} dnp3_iin1_t;


typedef struct dnp3_iin2_t dnp3_iin2_t;

/// @brief Second IIN byte
typedef struct dnp3_iin2_t
{
    /// @brief Outstation does not support this function code
    bool no_func_code_support;
    /// @brief Outstation does not support requested operation for objects in the request
    bool object_unknown;
    /// @brief Outstation does not support requested operation for objects in the request
    bool parameter_error;
    /// @brief An event buffer overflow condition exists in the outstation, and at least one unconfirmed event was lost
    bool event_buffer_overflow;
    /// @brief The operation requested is already executing (optional support)
    bool already_executing;
    /// @brief The outstation detected corrupt configuration (optional support)
    bool config_corrupt;
    /// @brief Reserved for future use - should always be set to 0
    bool reserved_2;
    /// @brief Reserved for future use - should always be set to 0
    bool reserved_1;
} dnp3_iin2_t;


typedef struct dnp3_iin_t dnp3_iin_t;

/// @brief Pair of IIN bytes
typedef struct dnp3_iin_t
{
    /// @brief First IIN byte
    dnp3_iin1_t iin1;
    /// @brief Second IIN byte
    dnp3_iin2_t iin2;
} dnp3_iin_t;


typedef struct dnp3_response_header_t dnp3_response_header_t;

/// @brief Response header information
typedef struct dnp3_response_header_t
{
    /// @brief Application control field
    dnp3_control_field_t control_field;
    /// @brief Response type
    dnp3_response_function_t func;
    /// @brief IIN bytes
    dnp3_iin_t iin;
} dnp3_response_header_t;


/// @brief Qualifier code used in the response
typedef enum dnp3_qualifier_code_t
{
    /// @brief 8-bit start stop (0x00)
    DNP3_QUALIFIER_CODE_RANGE8 = 0,
    /// @brief 16-bit start stop (0x01)
    DNP3_QUALIFIER_CODE_RANGE16 = 1,
    /// @brief All objects (0x06)
    DNP3_QUALIFIER_CODE_ALL_OBJECTS = 2,
    /// @brief 8-bit count (0x07)
    DNP3_QUALIFIER_CODE_COUNT8 = 3,
    /// @brief 16-bit count (0x08)
    DNP3_QUALIFIER_CODE_COUNT16 = 4,
    /// @brief 8-bit count and prefix (0x17)
    DNP3_QUALIFIER_CODE_COUNT_AND_PREFIX_8 = 5,
    /// @brief 16-bit count and prefix (0x28)
    DNP3_QUALIFIER_CODE_COUNT_AND_PREFIX_16 = 6,
    /// @brief 16-bit free format (0x5B)
    DNP3_QUALIFIER_CODE_FREE_FORMAT_16 = 7,
} dnp3_qualifier_code_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_qualifier_code_to_string(dnp3_qualifier_code_t value)
{
    switch (value)
    {
        case DNP3_QUALIFIER_CODE_RANGE8: return "range8";
        case DNP3_QUALIFIER_CODE_RANGE16: return "range16";
        case DNP3_QUALIFIER_CODE_ALL_OBJECTS: return "all_objects";
        case DNP3_QUALIFIER_CODE_COUNT8: return "count8";
        case DNP3_QUALIFIER_CODE_COUNT16: return "count16";
        case DNP3_QUALIFIER_CODE_COUNT_AND_PREFIX_8: return "count_and_prefix_8";
        case DNP3_QUALIFIER_CODE_COUNT_AND_PREFIX_16: return "count_and_prefix_16";
        case DNP3_QUALIFIER_CODE_FREE_FORMAT_16: return "free_format_16";
        default: return "unknown qualifier_code value";
    }
}

typedef struct dnp3_header_info_t dnp3_header_info_t;

/// @brief Information about the object header and specific variation
typedef struct dnp3_header_info_t
{
    /// @brief underlying variation in the response
    dnp3_variation_t variation;
    /// @brief Qualifier code used in the response
    dnp3_qualifier_code_t qualifier;
    /// @brief true if the received variation is an event type, false otherwise
    bool is_event;
    /// @brief true if a flags byte is present on the underlying variation, false otherwise
    bool has_flags;
} dnp3_header_info_t;


/// @brief Describes the source of a read event
typedef enum dnp3_read_type_t
{
    /// @brief Startup integrity poll
    DNP3_READ_TYPE_STARTUP_INTEGRITY = 0,
    /// @brief Unsolicited message
    DNP3_READ_TYPE_UNSOLICITED = 1,
    /// @brief Single poll requested by the user
    DNP3_READ_TYPE_SINGLE_POLL = 2,
    /// @brief Periodic poll configured by the user
    DNP3_READ_TYPE_PERIODIC_POLL = 3,
} dnp3_read_type_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_read_type_to_string(dnp3_read_type_t value)
{
    switch (value)
    {
        case DNP3_READ_TYPE_STARTUP_INTEGRITY: return "startup_integrity";
        case DNP3_READ_TYPE_UNSOLICITED: return "unsolicited";
        case DNP3_READ_TYPE_SINGLE_POLL: return "single_poll";
        case DNP3_READ_TYPE_PERIODIC_POLL: return "periodic_poll";
        default: return "unknown read_type value";
    }
}

/// @brief Callback interface used to received measurement values received from the outstation.
/// 
/// Methods are always invoked in the following order. @ref dnp3_read_handler_t.begin_fragment is called first, followed by one or more of the measurement handlers, and finally @ref dnp3_read_handler_t.end_fragment is called.
typedef struct dnp3_read_handler_t
{
    
    /// @brief Called when a valid response fragment is received, but before any measurements are processed
    /// @param read_type Describes what triggered the callback, e.g. response to a poll vs an unsolicited response
    /// @param header Header of the fragment
    /// @param ctx Context data
    void (*begin_fragment)(dnp3_read_type_t, dnp3_response_header_t, void*);
    
    /// @brief Called when all the data from a response fragment has been processed
    /// @param read_type Describes what triggered the read event
    /// @param header Header of the fragment
    /// @param ctx Context data
    void (*end_fragment)(dnp3_read_type_t, dnp3_response_header_t, void*);
    
    /// @brief Handle binary input data
    /// @param info Group/variation and qualifier information
    /// @param values iterator of values from the response
    /// @param ctx Context data
    void (*handle_binary_input)(dnp3_header_info_t, dnp3_binary_input_iterator_t*, void*);
    
    /// @brief Handle double-bit binary input data
    /// @param info Group/variation and qualifier information
    /// @param values iterator of values from the response
    /// @param ctx Context data
    void (*handle_double_bit_binary_input)(dnp3_header_info_t, dnp3_double_bit_binary_input_iterator_t*, void*);
    
    /// @brief Handle binary output status data
    /// @param info Group/variation and qualifier information
    /// @param values iterator of values from the response
    /// @param ctx Context data
    void (*handle_binary_output_status)(dnp3_header_info_t, dnp3_binary_output_status_iterator_t*, void*);
    
    /// @brief Handle counter data
    /// @param info Group/variation and qualifier information
    /// @param values iterator of values from the response
    /// @param ctx Context data
    void (*handle_counter)(dnp3_header_info_t, dnp3_counter_iterator_t*, void*);
    
    /// @brief Handle frozen counter input data
    /// @param info Group/variation and qualifier information
    /// @param values iterator of values from the response
    /// @param ctx Context data
    void (*handle_frozen_counter)(dnp3_header_info_t, dnp3_frozen_counter_iterator_t*, void*);
    
    /// @brief Handle analog input data
    /// @param info Group/variation and qualifier information
    /// @param values iterator of values from the response
    /// @param ctx Context data
    void (*handle_analog_input)(dnp3_header_info_t, dnp3_analog_input_iterator_t*, void*);
    
    /// @brief Handle frozen analog input data
    /// @param info Group/variation and qualifier information
    /// @param values iterator of values from the response
    /// @param ctx Context data
    void (*handle_frozen_analog_input)(dnp3_header_info_t, dnp3_frozen_analog_input_iterator_t*, void*);
    
    /// @brief Handle analog output status data
    /// @param info Group/variation and qualifier information
    /// @param values iterator of values from the response
    /// @param ctx Context data
    void (*handle_analog_output_status)(dnp3_header_info_t, dnp3_analog_output_status_iterator_t*, void*);
    
    /// @brief Handle binary output command events
    /// @param info Group/variation and qualifier information
    /// @param values iterator of values from the response
    /// @param ctx Context data
    void (*handle_binary_output_command_event)(dnp3_header_info_t, dnp3_binary_output_command_event_iterator_t*, void*);
    
    /// @brief Handle analog output command events
    /// @param info Group/variation and qualifier information
    /// @param values iterator of values from the response
    /// @param ctx Context data
    void (*handle_analog_output_command_event)(dnp3_header_info_t, dnp3_analog_output_command_event_iterator_t*, void*);
    
    /// @brief Handle unsigned integer values (g102)
    /// @param info Group/variation and qualifier information
    /// @param values iterator of values from the response
    /// @param ctx Context data
    void (*handle_unsigned_integer)(dnp3_header_info_t, dnp3_unsigned_integer_iterator_t*, void*);
    
    /// @brief Handle octet string data
    /// @param info Group/variation and qualifier information
    /// @param values iterator of values from the response
    /// @param ctx Context data
    void (*handle_octet_string)(dnp3_header_info_t, dnp3_octet_string_iterator_t*, void*);
    
    /// @brief Handle a known or unknown visible string device attribute
    /// @param info Group/variation and qualifier information
    /// @param attr Enumeration describing the attribute (possibly unknown) associated with the value
    /// @param set The set associated with this attribute. Examining this argument is only important if the attr argument is unknown.
    /// @param variation The variation associated with this attribute. Examining this argument is only important if the attr argument is unknown.
    /// @param value attribute value
    /// @param ctx Context data
    void (*handle_string_attr)(dnp3_header_info_t, dnp3_string_attr_t, uint8_t, uint8_t, const char*, void*);
    
    /// @brief Handle a known or unknown list of attribute variations
    /// @param info Group/variation and qualifier information
    /// @param attr Enumeration describing the attribute (possibly unknown) associated with the value
    /// @param set The set associated with this attribute. Examining this argument is only important if the attr argument is unknown.
    /// @param variation The variation associated with this attribute. Examining this argument is only important if the attr argument is unknown.
    /// @param value Iterator over a list of variation / properties pairs
    /// @param ctx Context data
    void (*handle_variation_list_attr)(dnp3_header_info_t, dnp3_variation_list_attr_t, uint8_t, uint8_t, dnp3_attr_item_iter_t*, void*);
    
    /// @brief Handle an unsigned integer device attribute
    /// @param info Group/variation and qualifier information
    /// @param attr Enumeration describing the attribute (possibly unknown) associated with the value
    /// @param set The set associated with this attribute. Examining this argument is only important if the attr argument is unknown.
    /// @param variation The variation associated with this attribute. Examining this argument is only important if the attr argument is unknown.
    /// @param value attribute value
    /// @param ctx Context data
    void (*handle_uint_attr)(dnp3_header_info_t, dnp3_uint_attr_t, uint8_t, uint8_t, uint32_t, void*);
    
    /// @brief Handle a boolean device attribute
    /// 
    /// These are actually signed integer values on the wire. This method is only called for known values
    /// @param info Group/variation and qualifier information
    /// @param attr Enumeration describing the attribute associated with the value
    /// @param set The set associated with this attribute. Examining this argument is only important if the attr argument is unknown.
    /// @param variation The variation associated with this attribute. Examining this argument is only important if the attr argument is unknown.
    /// @param value attribute value
    /// @param ctx Context data
    void (*handle_bool_attr)(dnp3_header_info_t, dnp3_bool_attr_t, uint8_t, uint8_t, bool, void*);
    
    /// @brief Handle a signed integer device attribute
    /// 
    /// There are no defined attributes for this type that aren't mapped to booleans so there is no enumeration
    /// @param info Group/variation and qualifier information
    /// @param attr Enumeration describing the attribute associated with the value
    /// @param set The set associated with this attribute. Examining this argument is only important if the attr argument is unknown.
    /// @param variation The variation associated with this attribute. Examining this argument is only important if the attr argument is unknown.
    /// @param value attribute value
    /// @param ctx Context data
    void (*handle_int_attr)(dnp3_header_info_t, dnp3_int_attr_t, uint8_t, uint8_t, int32_t, void*);
    
    /// @brief Handle a DNP3 time device attribute
    /// @param info Group/variation and qualifier information
    /// @param attr Enumeration describing the attribute associated with the value
    /// @param set The set associated with this attribute. Examining this argument is only important if the attr argument is unknown.
    /// @param variation The variation associated with this attribute. Examining this argument is only important if the attr argument is unknown.
    /// @param value 48-bit timestamp representing milliseconds since Unix epoch
    /// @param ctx Context data
    void (*handle_time_attr)(dnp3_header_info_t, dnp3_time_attr_t, uint8_t, uint8_t, uint64_t, void*);
    
    /// @brief Handle a floating point device attribute
    /// @param info Group/variation and qualifier information
    /// @param attr Enumeration describing the attribute associated with the value
    /// @param set The set associated with this attribute. Examining this argument is only important if the attr argument is unknown.
    /// @param variation The variation associated with this attribute. Examining this argument is only important if the attr argument is unknown.
    /// @param value Attribute value
    /// @param ctx Context data
    void (*handle_float_attr)(dnp3_header_info_t, dnp3_float_attr_t, uint8_t, uint8_t, double, void*);
    
    /// @brief Handle an octet string device attribute
    /// @param info Group/variation and qualifier information
    /// @param attr Enumeration describing the attribute associated with the value
    /// @param set The set associated with this attribute. Examining this argument is only important if the attr argument is unknown.
    /// @param variation The variation associated with this attribute. Examining this argument is only important if the attr argument is unknown.
    /// @param value Iterator over bytes in the octet-string
    /// @param ctx Context data
    void (*handle_octet_string_attr)(dnp3_header_info_t, dnp3_octet_string_attr_t, uint8_t, uint8_t, dnp3_byte_iterator_t*, void*);
    
    /// @brief Handle a bit string device attribute
    /// @param info Group/variation and qualifier information
    /// @param attr Enumeration describing the attribute associated with the value
    /// @param set The set associated with this attribute. Examining this argument is only important if the attr argument is unknown.
    /// @param variation The variation associated with this attribute. Examining this argument is only important if the attr argument is unknown.
    /// @param value Iterator over bytes in the bit-string
    /// @param ctx Context data
    void (*handle_bit_string_attr)(dnp3_header_info_t, dnp3_bit_string_attr_t, uint8_t, uint8_t, dnp3_byte_iterator_t*, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_read_handler_t;

typedef struct dnp3_master_channel_config_t dnp3_master_channel_config_t;

/// @brief Configuration for a MasterChannel that is independent of the physical layer
typedef struct dnp3_master_channel_config_t
{
    /// @brief Local DNP3 data-link address
    uint16_t address;
    /// @brief Decoding level for this master. You can modify this later on with @ref dnp3_master_channel_set_decode_level.
    dnp3_decode_level_t decode_level;
    /// @brief TX buffer size
    /// 
    /// Must be at least 249
    uint16_t tx_buffer_size;
    /// @brief RX buffer size
    /// 
    /// Must be at least 2048
    uint16_t rx_buffer_size;
} dnp3_master_channel_config_t;

/// @brief Initialize @ref dnp3_master_channel_config_t to default values
/// 
/// @note Values are initialized to:
/// - @ref dnp3_master_channel_config_t.decode_level : Default @ref dnp3_decode_level_t
/// - @ref dnp3_master_channel_config_t.tx_buffer_size : 2048
/// - @ref dnp3_master_channel_config_t.rx_buffer_size : 2048
/// 
/// @param address Local DNP3 data-link address
/// @returns New instance of @ref dnp3_master_channel_config_t
static dnp3_master_channel_config_t dnp3_master_channel_config_init(uint16_t address)
{
    dnp3_master_channel_config_t _return_value = {
        address,
        dnp3_decode_level_init(),
        2048,
        2048
    };
    return _return_value;
}


/// @brief Represents a communication channel for a master station
/// 
/// To communicate with a particular outstation, you need to add an association with @ref dnp3_master_channel_add_association.
/// 
/// @warning The class methods that return a value (e.g. as @ref dnp3_master_channel_add_association) cannot be called from within a callback.
typedef struct dnp3_master_channel_t dnp3_master_channel_t;

/// @brief Define a custom request to WRITE analog input dead-bands
typedef struct dnp3_write_dead_band_request_t dnp3_write_dead_band_request_t;

/// @brief A builder class to create one or more headers of analog input dead-bands
/// @return Instance of @ref dnp3_write_dead_band_request_t
dnp3_write_dead_band_request_t* dnp3_write_dead_band_request_create();

/// @brief Destroy a request created with @ref dnp3_write_dead_band_request_create
/// @param instance Instance of @ref dnp3_write_dead_band_request_t to destroy
void dnp3_write_dead_band_request_destroy(dnp3_write_dead_band_request_t* instance);

/// @brief Add a g34v1 (unsigned 16-bit) dead-band with 8-bit indexing  to the request
/// 
/// If this variation and index are the same as the current header, then it will be added to it. Otherwise, this call we create a new header of this type.d
/// @param instance Instance of @ref dnp3_write_dead_band_request_t
/// @param index Index of the analog input to which the dead-band applies
/// @param dead_band Value of the dead-band
void dnp3_write_dead_band_request_add_g34v1_u8(dnp3_write_dead_band_request_t* instance, uint8_t index, uint16_t dead_band);

/// @brief Add a g34v2 (unsigned 32-bit) dead-band with 8-bit indexing  to the request
/// 
/// If this variation and index are the same as the current header, then it will be added to it. Otherwise, this call we create a new header of this type.d
/// @param instance Instance of @ref dnp3_write_dead_band_request_t
/// @param index Index of the analog input to which the dead-band applies
/// @param dead_band Value of the dead-band
void dnp3_write_dead_band_request_add_g34v2_u8(dnp3_write_dead_band_request_t* instance, uint8_t index, uint32_t dead_band);

/// @brief Add a g34v3 (single-precision floating point) dead-band with 8-bit indexing  to the request
/// 
/// If this variation and index are the same as the current header, then it will be added to it. Otherwise, this call we create a new header of this type.d
/// @param instance Instance of @ref dnp3_write_dead_band_request_t
/// @param index Index of the analog input to which the dead-band applies
/// @param dead_band Value of the dead-band
void dnp3_write_dead_band_request_add_g34v3_u8(dnp3_write_dead_band_request_t* instance, uint8_t index, float dead_band);

/// @brief Add a g34v1 (unsigned 16-bit) dead-band with 16-bit indexing  to the request
/// 
/// If this variation and index are the same as the current header, then it will be added to it. Otherwise, this call we create a new header of this type.d
/// @param instance Instance of @ref dnp3_write_dead_band_request_t
/// @param index Index of the analog input to which the dead-band applies
/// @param dead_band Value of the dead-band
void dnp3_write_dead_band_request_add_g34v1_u16(dnp3_write_dead_band_request_t* instance, uint16_t index, uint16_t dead_band);

/// @brief Add a g34v2 (unsigned 32-bit) dead-band with 16-bit indexing  to the request
/// 
/// If this variation and index are the same as the current header, then it will be added to it. Otherwise, this call we create a new header of this type.d
/// @param instance Instance of @ref dnp3_write_dead_band_request_t
/// @param index Index of the analog input to which the dead-band applies
/// @param dead_band Value of the dead-band
void dnp3_write_dead_band_request_add_g34v2_u16(dnp3_write_dead_band_request_t* instance, uint16_t index, uint32_t dead_band);

/// @brief Add a g34v3 (single-precision floating point) dead-band with 16-bit indexing  to the request
/// 
/// If this variation and index are the same as the current header, then it will be added to it. Otherwise, this call we create a new header of this type.d
/// @param instance Instance of @ref dnp3_write_dead_band_request_t
/// @param index Index of the analog input to which the dead-band applies
/// @param dead_band Value of the dead-band
void dnp3_write_dead_band_request_add_g34v3_u16(dnp3_write_dead_band_request_t* instance, uint16_t index, float dead_band);

/// @brief If a header is currently being written, then this will complete the header so that no new objects may be added to it
/// 
/// This happens automatically if you change the type or index when adding dead-band values. This method allows you to fragment the same type across multiple object headers.
/// @param instance Instance of @ref dnp3_write_dead_band_request_t
void dnp3_write_dead_band_request_finish_header(dnp3_write_dead_band_request_t* instance);


/// @brief Errors that may occur when performing a request that expects a response with zero object headers
typedef enum dnp3_empty_response_error_t
{
    /// @brief Success, i.e. no error occurred
    DNP3_EMPTY_RESPONSE_ERROR_OK = 0,
    /// @brief IIN2 indicates request was not completely successful
    DNP3_EMPTY_RESPONSE_ERROR_REJECTED_BY_IIN2 = 1,
    /// @brief too many user requests queued
    DNP3_EMPTY_RESPONSE_ERROR_TOO_MANY_REQUESTS = 2,
    /// @brief outstation returned an IIN.2 error bit
    DNP3_EMPTY_RESPONSE_ERROR_IIN_ERROR = 3,
    /// @brief response was malformed or contained object headers
    DNP3_EMPTY_RESPONSE_ERROR_BAD_RESPONSE = 4,
    /// @brief timeout occurred before receiving a response
    DNP3_EMPTY_RESPONSE_ERROR_RESPONSE_TIMEOUT = 5,
    /// @brief insufficient buffer space to serialize the request
    DNP3_EMPTY_RESPONSE_ERROR_WRITE_ERROR = 6,
    /// @brief no connection
    DNP3_EMPTY_RESPONSE_ERROR_NO_CONNECTION = 7,
    /// @brief master was shutdown
    DNP3_EMPTY_RESPONSE_ERROR_SHUTDOWN = 8,
    /// @brief association was removed mid-task
    DNP3_EMPTY_RESPONSE_ERROR_ASSOCIATION_REMOVED = 9,
    /// @brief request data could not be encoded
    DNP3_EMPTY_RESPONSE_ERROR_BAD_ENCODING = 10,
} dnp3_empty_response_error_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_empty_response_error_to_string(dnp3_empty_response_error_t value)
{
    switch (value)
    {
        case DNP3_EMPTY_RESPONSE_ERROR_OK: return "ok";
        case DNP3_EMPTY_RESPONSE_ERROR_REJECTED_BY_IIN2: return "rejected_by_iin2";
        case DNP3_EMPTY_RESPONSE_ERROR_TOO_MANY_REQUESTS: return "too_many_requests";
        case DNP3_EMPTY_RESPONSE_ERROR_IIN_ERROR: return "iin_error";
        case DNP3_EMPTY_RESPONSE_ERROR_BAD_RESPONSE: return "bad_response";
        case DNP3_EMPTY_RESPONSE_ERROR_RESPONSE_TIMEOUT: return "response_timeout";
        case DNP3_EMPTY_RESPONSE_ERROR_WRITE_ERROR: return "write_error";
        case DNP3_EMPTY_RESPONSE_ERROR_NO_CONNECTION: return "no_connection";
        case DNP3_EMPTY_RESPONSE_ERROR_SHUTDOWN: return "shutdown";
        case DNP3_EMPTY_RESPONSE_ERROR_ASSOCIATION_REMOVED: return "association_removed";
        case DNP3_EMPTY_RESPONSE_ERROR_BAD_ENCODING: return "bad_encoding";
        default: return "unknown empty_response_error value";
    }
}


/// @brief Callback interface for any task that expects an empty response
typedef struct dnp3_empty_response_callback_t
{
    
    /// @brief Invoked when the asynchronous operation completes successfully
    /// @param result Result of operation
    /// @param ctx Context data
    void (*on_complete)(dnp3_nothing_t, void*);
    
    /// @brief Invoked when the asynchronous operation fails
    /// @param error Enumeration indicating which error occurred
    /// @param ctx Context data
    void (*on_failure)(dnp3_empty_response_error_t, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_empty_response_callback_t;

/// @brief Create a master channel that connects to the specified TCP endpoint(s)
/// @param runtime Runtime to use to drive asynchronous operations of the master
/// @param link_error_mode Controls how link errors are handled with respect to the TCP session
/// @param config Generic configuration for the channel
/// @param endpoints List of IP endpoints.
/// @param connect_strategy Controls the timing of (re)connection attempts
/// @param listener TCP connection listener used to receive updates on the status of the connection
/// @param out Handle to the master created, @p NULL if an error occurred
/// @return Error code
dnp3_param_error_t dnp3_master_channel_create_tcp(dnp3_runtime_t* runtime, dnp3_link_error_mode_t link_error_mode, dnp3_master_channel_config_t config, dnp3_endpoint_list_t* endpoints, dnp3_connect_strategy_t connect_strategy, dnp3_client_state_listener_t listener, dnp3_master_channel_t** out);

/// @brief Create a master channel that connects to the specified TCP endpoint(s)
/// 
/// This is just like @ref dnp3_master_channel_create_tcp_channel but adds the @ref dnp3_connect_options_t parameter
/// @param runtime Runtime to use to drive asynchronous operations of the master
/// @param link_error_mode Controls how link errors are handled with respect to the TCP session
/// @param config Generic configuration for the channel
/// @param endpoints List of IP endpoints.
/// @param connect_strategy Controls the timing of (re)connection attempts
/// @param connect_options Options that control the TCP connection process
/// @param listener TCP connection listener used to receive updates on the status of the connection
/// @param out Handle to the master created, @p NULL if an error occurred
/// @return Error code
dnp3_param_error_t dnp3_master_channel_create_tcp_2(dnp3_runtime_t* runtime, dnp3_link_error_mode_t link_error_mode, dnp3_master_channel_config_t config, dnp3_endpoint_list_t* endpoints, dnp3_connect_strategy_t connect_strategy, dnp3_connect_options_t* connect_options, dnp3_client_state_listener_t listener, dnp3_master_channel_t** out);

/// @brief Create a master channel on the specified serial port
/// 
/// The returned master must be gracefully shutdown with @ref dnp3_master_channel_destroy when done.
/// @param runtime Runtime to use to drive asynchronous operations of the master
/// @param config Generic configuration for the channel
/// @param path Path to the serial device. Generally /dev/tty0 on Linux and COM1 on Windows.
/// @param serial_params Serial port settings
/// @param open_retry_delay Delay between attempts to open the serial port (milliseconds)
/// @param listener Listener to receive updates on the status of the serial port
/// @param out Handle to the master created, @p NULL if an error occurred
/// @return Error code
dnp3_param_error_t dnp3_master_channel_create_serial(dnp3_runtime_t* runtime, dnp3_master_channel_config_t config, const char* path, dnp3_serial_settings_t serial_params, uint64_t open_retry_delay, dnp3_port_state_listener_t listener, dnp3_master_channel_t** out);

/// @brief Create a UDP master channel on the local endpoint
/// 
/// The returned master must be gracefully shutdown with @ref dnp3_master_channel_destroy when done.
/// @param runtime Runtime to use to drive asynchronous operations of the master
/// @param config Generic configuration for the channel
/// @param local_endpoint Local endpoint on which to bind the UDP socket
/// @param link_read_mode Determines how the link-layer parser treats frame that span datagrams. Typically set to @ref DNP3_LINK_READ_MODE_DATAGRAM
/// @param retry_delay Amount of time to wait after a failed attempt to bind the UDP socket (milliseconds)
/// @param out Handle to the master created, @p NULL if an error occurred
/// @return Error code
dnp3_param_error_t dnp3_master_channel_create_udp(dnp3_runtime_t* runtime, dnp3_master_channel_config_t config, const char* local_endpoint, dnp3_link_read_mode_t link_read_mode, uint64_t retry_delay, dnp3_master_channel_t** out);

/// @brief Shutdown a @ref dnp3_master_channel_t and release all resources
/// @param instance Instance of @ref dnp3_master_channel_t to destroy
void dnp3_master_channel_destroy(dnp3_master_channel_t* instance);

/// @brief Create a master channel that connects to the specified TCP endpoint(s) and establish a TLS session with the remote.
/// @param runtime Runtime to use to drive asynchronous operations of the master
/// @param link_error_mode Controls how link errors are handled with respect to the TCP session
/// @param config Generic configuration for the channel
/// @param endpoints List of IP endpoints.
/// @param connect_strategy Controls the timing of (re)connection attempts
/// @param listener TCP connection listener used to receive updates on the status of the connection
/// @param tls_config TLS client configuration
/// @param out Handle to the master created, @p NULL if an error occurred
/// @return Error code
dnp3_param_error_t dnp3_master_channel_create_tls(dnp3_runtime_t* runtime, dnp3_link_error_mode_t link_error_mode, dnp3_master_channel_config_t config, dnp3_endpoint_list_t* endpoints, dnp3_connect_strategy_t connect_strategy, dnp3_client_state_listener_t listener, dnp3_tls_client_config_t tls_config, dnp3_master_channel_t** out);

/// @brief Create a master channel that connects to the specified TCP endpoint(s) and establish a TLS session with the remote.
/// 
/// This is just like @ref dnp3_master_channel_create_tls_channel but adds the @ref dnp3_connect_options_t parameter
/// @param runtime Runtime to use to drive asynchronous operations of the master
/// @param link_error_mode Controls how link errors are handled with respect to the TCP session
/// @param config Generic configuration for the channel
/// @param endpoints List of IP endpoints.
/// @param connect_strategy Controls the timing of (re)connection attempts
/// @param connect_options Options that control the TCP connection process
/// @param listener TCP connection listener used to receive updates on the status of the connection
/// @param tls_config TLS client configuration
/// @param out Handle to the master created, @p NULL if an error occurred
/// @return Error code
dnp3_param_error_t dnp3_master_channel_create_tls_2(dnp3_runtime_t* runtime, dnp3_link_error_mode_t link_error_mode, dnp3_master_channel_config_t config, dnp3_endpoint_list_t* endpoints, dnp3_connect_strategy_t connect_strategy, dnp3_connect_options_t* connect_options, dnp3_client_state_listener_t listener, dnp3_tls_client_config_t tls_config, dnp3_master_channel_t** out);

/// @brief start communications
/// @param instance Instance of @ref dnp3_master_channel_t
/// @return Error code
dnp3_param_error_t dnp3_master_channel_enable(dnp3_master_channel_t* instance);

/// @brief stop communications
/// @param instance Instance of @ref dnp3_master_channel_t
/// @return Error code
dnp3_param_error_t dnp3_master_channel_disable(dnp3_master_channel_t* instance);

typedef struct dnp3_association_id_t dnp3_association_id_t;

/// @brief Association identifier
/// 
/// @warning This struct should never be initialized or modified by user code
typedef struct dnp3_association_id_t
{
    /// @brief Outstation address of the association
    uint16_t address;
} dnp3_association_id_t;

typedef struct dnp3_poll_id_t dnp3_poll_id_t;

/// @brief Poll identifier
/// 
/// @warning This struct should never be initialized or modified by user code
typedef struct dnp3_poll_id_t
{
    /// @brief Outstation address of the association
    uint16_t association_id;
    /// @brief Unique poll id assigned by the association
    uint64_t id;
} dnp3_poll_id_t;

typedef struct dnp3_event_classes_t dnp3_event_classes_t;

/// @brief Event classes
typedef struct dnp3_event_classes_t
{
    /// @brief Class 1 events
    bool class1;
    /// @brief Class 2 events
    bool class2;
    /// @brief Class 3 events
    bool class3;
} dnp3_event_classes_t;

/// @brief Fully construct @ref dnp3_event_classes_t specifying the value of each field
/// @param class1 Class 1 events
/// @param class2 Class 2 events
/// @param class3 Class 3 events
/// @returns New instance of @ref dnp3_event_classes_t
static dnp3_event_classes_t dnp3_event_classes_init(bool class1, bool class2, bool class3)
{
    dnp3_event_classes_t _return_value = {
        class1,
        class2,
        class3
    };
    return _return_value;
}

/// @brief Initialize all classes to true
/// 
/// @note Values are initialized to:
/// - @ref dnp3_event_classes_t.class1 : @p true
/// - @ref dnp3_event_classes_t.class2 : @p true
/// - @ref dnp3_event_classes_t.class3 : @p true
/// 
/// @returns New instance of @ref dnp3_event_classes_t
static dnp3_event_classes_t dnp3_event_classes_all()
{
    dnp3_event_classes_t _return_value = {
        true,
        true,
        true
    };
    return _return_value;
}

/// @brief Initialize all classes to false
/// 
/// @note Values are initialized to:
/// - @ref dnp3_event_classes_t.class1 : @p false
/// - @ref dnp3_event_classes_t.class2 : @p false
/// - @ref dnp3_event_classes_t.class3 : @p false
/// 
/// @returns New instance of @ref dnp3_event_classes_t
static dnp3_event_classes_t dnp3_event_classes_none()
{
    dnp3_event_classes_t _return_value = {
        false,
        false,
        false
    };
    return _return_value;
}


typedef struct dnp3_classes_t dnp3_classes_t;

/// @brief Class 0, 1, 2 and 3 config
typedef struct dnp3_classes_t
{
    /// @brief Class 0 (static data)
    bool class0;
    /// @brief Class 1 events
    bool class1;
    /// @brief Class 2 events
    bool class2;
    /// @brief Class 3 events
    bool class3;
} dnp3_classes_t;

/// @brief Fully construct @ref dnp3_classes_t specifying the value of each field
/// @param class0 Class 0 (static data)
/// @param class1 Class 1 events
/// @param class2 Class 2 events
/// @param class3 Class 3 events
/// @returns New instance of @ref dnp3_classes_t
static dnp3_classes_t dnp3_classes_init(bool class0, bool class1, bool class2, bool class3)
{
    dnp3_classes_t _return_value = {
        class0,
        class1,
        class2,
        class3
    };
    return _return_value;
}

/// @brief Initialize all classes to true
/// 
/// @note Values are initialized to:
/// - @ref dnp3_classes_t.class0 : @p true
/// - @ref dnp3_classes_t.class1 : @p true
/// - @ref dnp3_classes_t.class2 : @p true
/// - @ref dnp3_classes_t.class3 : @p true
/// 
/// @returns New instance of @ref dnp3_classes_t
static dnp3_classes_t dnp3_classes_all()
{
    dnp3_classes_t _return_value = {
        true,
        true,
        true,
        true
    };
    return _return_value;
}

/// @brief Initialize all classes to false
/// 
/// @note Values are initialized to:
/// - @ref dnp3_classes_t.class0 : @p false
/// - @ref dnp3_classes_t.class1 : @p false
/// - @ref dnp3_classes_t.class2 : @p false
/// - @ref dnp3_classes_t.class3 : @p false
/// 
/// @returns New instance of @ref dnp3_classes_t
static dnp3_classes_t dnp3_classes_none()
{
    dnp3_classes_t _return_value = {
        false,
        false,
        false,
        false
    };
    return _return_value;
}


/// @brief Automatic time synchronization configuration
typedef enum dnp3_auto_time_sync_t
{
    /// @brief Do not perform automatic time sync
    DNP3_AUTO_TIME_SYNC_NONE = 0,
    /// @brief Perform automatic time sync with Record Current Time (0x18) function code
    DNP3_AUTO_TIME_SYNC_LAN = 1,
    /// @brief Perform automatic time sync with Delay Measurement (0x17) function code
    DNP3_AUTO_TIME_SYNC_NON_LAN = 2,
} dnp3_auto_time_sync_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_auto_time_sync_to_string(dnp3_auto_time_sync_t value)
{
    switch (value)
    {
        case DNP3_AUTO_TIME_SYNC_NONE: return "none";
        case DNP3_AUTO_TIME_SYNC_LAN: return "lan";
        case DNP3_AUTO_TIME_SYNC_NON_LAN: return "non_lan";
        default: return "unknown auto_time_sync value";
    }
}

typedef struct dnp3_association_config_t dnp3_association_config_t;

/// @brief Association configuration
typedef struct dnp3_association_config_t
{
    /// @brief Timeout for receiving a response on this association
    /// @note The unit is milliseconds
    uint64_t response_timeout;
    /// @brief Classes to disable unsolicited responses at startup
    dnp3_event_classes_t disable_unsol_classes;
    /// @brief Classes to enable unsolicited responses at startup
    dnp3_event_classes_t enable_unsol_classes;
    /// @brief Startup integrity classes to ask on master startup and when an outstation restart is detected.
    /// 
    /// For conformance, this should be Class 1230.
    dnp3_classes_t startup_integrity_classes;
    /// @brief Automatic time synchronization configuration
    dnp3_auto_time_sync_t auto_time_sync;
    /// @brief Automatic tasks retry strategy
    dnp3_retry_strategy_t auto_tasks_retry_strategy;
    /// @brief Delay of inactivity before sending a REQUEST_LINK_STATUS to the outstation
    /// 
    /// A value of zero means no automatic keep-alive.
    /// @note The unit is seconds
    uint64_t keep_alive_timeout;
    /// @brief Automatic integrity scan when an EVENT_BUFFER_OVERFLOW is detected
    bool auto_integrity_scan_on_buffer_overflow;
    /// @brief Classes to automatically send reads when the IIN bit is asserted
    dnp3_event_classes_t event_scan_on_events_available;
    /// @brief maximum number of user requests (e.g. commands, adhoc reads, etc) that will be queued before back-pressure is applied by failing requests
    uint16_t max_queued_user_requests;
} dnp3_association_config_t;

/// @brief Initialize the configuration with the specified values
/// 
/// @note Values are initialized to:
/// - @ref dnp3_association_config_t.response_timeout : 5000ms
/// - @ref dnp3_association_config_t.auto_time_sync : @ref DNP3_AUTO_TIME_SYNC_NONE
/// - @ref dnp3_association_config_t.auto_tasks_retry_strategy : Default @ref dnp3_retry_strategy_t
/// - @ref dnp3_association_config_t.keep_alive_timeout : 60s
/// - @ref dnp3_association_config_t.auto_integrity_scan_on_buffer_overflow : @p true
/// - @ref dnp3_association_config_t.max_queued_user_requests : 16
/// 
/// @param disable_unsol_classes Classes to disable unsolicited responses at startup
/// @param enable_unsol_classes Classes to enable unsolicited responses at startup
/// @param startup_integrity_classes Startup integrity classes to ask on master startup and when an outstation restart is detected.
/// @param event_scan_on_events_available Classes to automatically send reads when the IIN bit is asserted
/// @returns New instance of @ref dnp3_association_config_t
static dnp3_association_config_t dnp3_association_config_init(dnp3_event_classes_t disable_unsol_classes, dnp3_event_classes_t enable_unsol_classes, dnp3_classes_t startup_integrity_classes, dnp3_event_classes_t event_scan_on_events_available)
{
    dnp3_association_config_t _return_value = {
        5000,
        disable_unsol_classes,
        enable_unsol_classes,
        startup_integrity_classes,
        DNP3_AUTO_TIME_SYNC_NONE,
        dnp3_retry_strategy_init(),
        60,
        true,
        event_scan_on_events_available,
        16
    };
    return _return_value;
}


typedef struct dnp3_utc_timestamp_t dnp3_utc_timestamp_t;

/// @brief Timestamp value returned by @ref dnp3_association_handler_t.get_current_time.
/// 
/// @ref dnp3_utc_timestamp_t.value is only valid if @ref dnp3_utc_timestamp_t.is_valid is true.
typedef struct dnp3_utc_timestamp_t
{
    /// @brief Count of milliseconds since UNIX epoch
    /// 
    /// @warning Only the lower 48-bits are used in DNP3 timestamps and time synchronization
    uint64_t value;
    /// @brief True if the timestamp is valid, false otherwise.
    bool is_valid;
} dnp3_utc_timestamp_t;

/// @brief Construct a valid @ref dnp3_utc_timestamp_t
/// 
/// @note Values are initialized to:
/// - @ref dnp3_utc_timestamp_t.is_valid : @p true
/// 
/// @param value Count of milliseconds since UNIX epoch
/// @returns New instance of @ref dnp3_utc_timestamp_t
static dnp3_utc_timestamp_t dnp3_utc_timestamp_valid(uint64_t value)
{
    dnp3_utc_timestamp_t _return_value = {
        value,
        true
    };
    return _return_value;
}

/// @brief Construct an invalid @ref dnp3_utc_timestamp_t
/// 
/// @note Values are initialized to:
/// - @ref dnp3_utc_timestamp_t.is_valid : @p false
/// - @ref dnp3_utc_timestamp_t.value : 0
/// 
/// @returns New instance of @ref dnp3_utc_timestamp_t
static dnp3_utc_timestamp_t dnp3_utc_timestamp_invalid()
{
    dnp3_utc_timestamp_t _return_value = {
        0,
        false
    };
    return _return_value;
}


/// @brief Callbacks for a particular outstation association
typedef struct dnp3_association_handler_t
{
    
    /// @brief Returns the current time or an invalid time if none is available
    /// 
    /// This callback is used when the master performs time synchronization for a particular outstation.
    /// 
    /// This could return the system clock or some other clock's time
    /// @param ctx Context data
    /// @return The current time
    dnp3_utc_timestamp_t (*get_current_time)(void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_association_handler_t;

/// @brief Task type used in @ref dnp3_association_information_t
typedef enum dnp3_task_type_t
{
    /// @brief User-defined read request
    DNP3_TASK_TYPE_USER_READ = 0,
    /// @brief Periodic poll task
    DNP3_TASK_TYPE_PERIODIC_POLL = 1,
    /// @brief Startup integrity scan
    DNP3_TASK_TYPE_STARTUP_INTEGRITY = 2,
    /// @brief Automatic event scan caused by RESTART IIN bit detection
    DNP3_TASK_TYPE_AUTO_EVENT_SCAN = 3,
    /// @brief Command request
    DNP3_TASK_TYPE_COMMAND = 4,
    /// @brief Clear RESTART IIN bit
    DNP3_TASK_TYPE_CLEAR_RESTART_BIT = 5,
    /// @brief Enable unsolicited startup request
    DNP3_TASK_TYPE_ENABLE_UNSOLICITED = 6,
    /// @brief Disable unsolicited startup request
    DNP3_TASK_TYPE_DISABLE_UNSOLICITED = 7,
    /// @brief Time synchronisation task
    DNP3_TASK_TYPE_TIME_SYNC = 8,
    /// @brief Cold or warm restart task
    DNP3_TASK_TYPE_RESTART = 9,
    /// @brief Write analog input dead-bands
    DNP3_TASK_TYPE_WRITE_DEAD_BANDS = 10,
    /// @brief Generic request that expects an empty response
    DNP3_TASK_TYPE_GENERIC_EMPTY_RESPONSE = 11,
    /// @brief Read a file from the outstation
    DNP3_TASK_TYPE_FILE_READ = 12,
    /// @brief Get information about a file
    DNP3_TASK_TYPE_GET_FILE_INFO = 13,
    /// @brief Send username and password and get back an auth key from the outstation
    DNP3_TASK_TYPE_FILE_AUTH = 14,
    /// @brief Open a file on the outstation
    DNP3_TASK_TYPE_FILE_OPEN = 15,
    /// @brief Write a file block to the outstation
    DNP3_TASK_TYPE_FILE_WRITE_BLOCK = 16,
    /// @brief Close a file on the outstation
    DNP3_TASK_TYPE_FILE_CLOSE = 17,
} dnp3_task_type_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_task_type_to_string(dnp3_task_type_t value)
{
    switch (value)
    {
        case DNP3_TASK_TYPE_USER_READ: return "user_read";
        case DNP3_TASK_TYPE_PERIODIC_POLL: return "periodic_poll";
        case DNP3_TASK_TYPE_STARTUP_INTEGRITY: return "startup_integrity";
        case DNP3_TASK_TYPE_AUTO_EVENT_SCAN: return "auto_event_scan";
        case DNP3_TASK_TYPE_COMMAND: return "command";
        case DNP3_TASK_TYPE_CLEAR_RESTART_BIT: return "clear_restart_bit";
        case DNP3_TASK_TYPE_ENABLE_UNSOLICITED: return "enable_unsolicited";
        case DNP3_TASK_TYPE_DISABLE_UNSOLICITED: return "disable_unsolicited";
        case DNP3_TASK_TYPE_TIME_SYNC: return "time_sync";
        case DNP3_TASK_TYPE_RESTART: return "restart";
        case DNP3_TASK_TYPE_WRITE_DEAD_BANDS: return "write_dead_bands";
        case DNP3_TASK_TYPE_GENERIC_EMPTY_RESPONSE: return "generic_empty_response";
        case DNP3_TASK_TYPE_FILE_READ: return "file_read";
        case DNP3_TASK_TYPE_GET_FILE_INFO: return "get_file_info";
        case DNP3_TASK_TYPE_FILE_AUTH: return "file_auth";
        case DNP3_TASK_TYPE_FILE_OPEN: return "file_open";
        case DNP3_TASK_TYPE_FILE_WRITE_BLOCK: return "file_write_block";
        case DNP3_TASK_TYPE_FILE_CLOSE: return "file_close";
        default: return "unknown task_type value";
    }
}

/// @brief Task error used in @ref dnp3_association_information_t
typedef enum dnp3_task_error_t
{
    /// @brief too many user requests queued
    DNP3_TASK_ERROR_TOO_MANY_REQUESTS = 0,
    /// @brief outstation returned an IIN.2 error bit
    DNP3_TASK_ERROR_IIN_ERROR = 1,
    /// @brief response was malformed or contained object headers
    DNP3_TASK_ERROR_BAD_RESPONSE = 2,
    /// @brief timeout occurred before receiving a response
    DNP3_TASK_ERROR_RESPONSE_TIMEOUT = 3,
    /// @brief insufficient buffer space to serialize the request
    DNP3_TASK_ERROR_WRITE_ERROR = 4,
    /// @brief no connection
    DNP3_TASK_ERROR_NO_CONNECTION = 5,
    /// @brief master was shutdown
    DNP3_TASK_ERROR_SHUTDOWN = 6,
    /// @brief association was removed mid-task
    DNP3_TASK_ERROR_ASSOCIATION_REMOVED = 7,
    /// @brief request data could not be encoded
    DNP3_TASK_ERROR_BAD_ENCODING = 8,
} dnp3_task_error_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_task_error_to_string(dnp3_task_error_t value)
{
    switch (value)
    {
        case DNP3_TASK_ERROR_TOO_MANY_REQUESTS: return "too_many_requests";
        case DNP3_TASK_ERROR_IIN_ERROR: return "iin_error";
        case DNP3_TASK_ERROR_BAD_RESPONSE: return "bad_response";
        case DNP3_TASK_ERROR_RESPONSE_TIMEOUT: return "response_timeout";
        case DNP3_TASK_ERROR_WRITE_ERROR: return "write_error";
        case DNP3_TASK_ERROR_NO_CONNECTION: return "no_connection";
        case DNP3_TASK_ERROR_SHUTDOWN: return "shutdown";
        case DNP3_TASK_ERROR_ASSOCIATION_REMOVED: return "association_removed";
        case DNP3_TASK_ERROR_BAD_ENCODING: return "bad_encoding";
        default: return "unknown task_error value";
    }
}

/// @brief Informational callbacks about the current state of an outstation association
typedef struct dnp3_association_information_t
{
    
    /// @brief Called when a new task is started
    /// @param task_type Type of task that was started
    /// @param function_code Function code used by the task
    /// @param seq Sequence number of the request
    /// @param ctx Context data
    void (*task_start)(dnp3_task_type_t, dnp3_function_code_t, uint8_t, void*);
    
    /// @brief Called when a task successfully completes
    /// @param task_type Type of task that was completed
    /// @param function_code Function code used by the task
    /// @param seq Sequence number of the response that completed the request. This will typically be the same as the seq number in the request, except for READ requests where the response is multi-fragmented.
    /// @param ctx Context data
    void (*task_success)(dnp3_task_type_t, dnp3_function_code_t, uint8_t, void*);
    
    /// @brief Called when a task fails
    /// @param task_type Type of task that was completed
    /// @param error Error that prevented 
    /// @param ctx Context data
    void (*task_fail)(dnp3_task_type_t, dnp3_task_error_t, void*);
    
    /// @brief Called when an unsolicited response is received
    /// @param is_duplicate Is the unsolicited response a duplicate response
    /// @param seq Sequence number of the response
    /// @param ctx Context data
    void (*unsolicited_response)(bool, uint8_t, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_association_information_t;

/// @brief Custom request
/// 
/// Whenever a method takes a request as a parameter, the request is internally copied. Therefore, it is possible to reuse the same requests over and over.
typedef struct dnp3_request_t dnp3_request_t;

/// @brief Create a new request
/// @return Instance of @ref dnp3_request_t
dnp3_request_t* dnp3_request_create();

/// @brief Create a new request asking for classes
/// 
/// An identical request can be created manually with @ref dnp3_request_add_all_objects_header and variations @ref DNP3_VARIATION_GROUP60_VAR1, @ref DNP3_VARIATION_GROUP60_VAR2, @ref DNP3_VARIATION_GROUP60_VAR3 and @ref DNP3_VARIATION_GROUP60_VAR4.
/// @param class0 Ask for class 0 (static data)
/// @param class1 Ask for class 1 events
/// @param class2 Ask for class 2 events
/// @param class3 Ask for class 3 events
/// @return Handle to the created request
dnp3_request_t* dnp3_request_new_class(bool class0, bool class1, bool class2, bool class3);

/// @brief Create a new request asking for all objects of a particular variation.
/// 
/// An identical request can be created manually with @ref dnp3_request_create and @ref dnp3_request_add_all_objects_header.
/// @param variation Variation to ask for
/// @return Handle to the created request
dnp3_request_t* dnp3_request_new_all_objects(dnp3_variation_t variation);

/// @brief Create a new request asking for range of objects (using 8-bit start/stop).
/// 
/// An identical request can be created manually with @ref dnp3_request_create and @ref dnp3_request_add_one_byte_range_header.
/// @param variation Variation to ask for
/// @param start Start index to ask
/// @param stop Stop index to ask (inclusive)
/// @return Handle to the created request
dnp3_request_t* dnp3_request_new_one_byte_range(dnp3_variation_t variation, uint8_t start, uint8_t stop);

/// @brief Create a new request asking for range of objects (using 16-bit start/stop).
/// 
/// An identical request can be created manually with @ref dnp3_request_create and @ref dnp3_request_add_two_byte_range_header.
/// @param variation Variation to ask for
/// @param start Start index to ask
/// @param stop Stop index to ask (inclusive)
/// @return Handle to the created request
dnp3_request_t* dnp3_request_new_two_byte_range(dnp3_variation_t variation, uint16_t start, uint16_t stop);

/// @brief Create a new request asking for a limited count of objects (using 8-bit start/stop).
/// 
/// An identical request can be created manually with @ref dnp3_request_create and @ref dnp3_request_add_one_byte_limited_count_header.
/// @param variation Variation to ask for
/// @param count Maximum number of events
/// @return Handle to the created request
dnp3_request_t* dnp3_request_new_one_byte_limited_count(dnp3_variation_t variation, uint8_t count);

/// @brief Create a new request asking for a limited count of objects (using 16-bit start/stop).
/// 
/// An identical request can be created manually with @ref dnp3_request_create and @ref dnp3_request_add_two_byte_limited_count_header.
/// @param variation Variation to ask for
/// @param count Maximum number of events
/// @return Handle to the created request
dnp3_request_t* dnp3_request_new_two_byte_limited_count(dnp3_variation_t variation, uint16_t count);

/// @brief Destroy a request created with @ref dnp3_request_create or @ref dnp3_request_class_request.
/// @param instance Instance of @ref dnp3_request_t to destroy
void dnp3_request_destroy(dnp3_request_t* instance);

/// @brief Add a one-byte start/stop header for use with a READ request
/// @param instance Instance of @ref dnp3_request_t
/// @param variation Variation of the device attribute
/// @param set The set (point) to which the attribute belongs
void dnp3_request_add_specific_attribute(dnp3_request_t* instance, uint8_t variation, uint8_t set);

/// @brief Add a one-byte start/stop header containing for use with a WRITE request
/// @param instance Instance of @ref dnp3_request_t
/// @param variation Variation of the attribute
/// @param set The set (point) to which the attribute belongs
/// @param value Value of the attribute
void dnp3_request_add_string_attribute(dnp3_request_t* instance, uint8_t variation, uint8_t set, const char* value);

/// @brief Add a one-byte start/stop header containing for use with a WRITE request
/// @param instance Instance of @ref dnp3_request_t
/// @param variation Variation of the attribute
/// @param set The set (point) to which the attribute belongs
/// @param value Value of the attribute
void dnp3_request_add_uint_attribute(dnp3_request_t* instance, uint8_t variation, uint8_t set, uint32_t value);

/// @brief Add a one-byte start/stop header for use with a READ request
/// @param instance Instance of @ref dnp3_request_t
/// @param variation Variation to ask for
/// @param start Start index to ask
/// @param stop Stop index to ask (inclusive)
void dnp3_request_add_one_byte_range_header(dnp3_request_t* instance, dnp3_variation_t variation, uint8_t start, uint8_t stop);

/// @brief Add a two-byte start/stop header for use with a READ request
/// @param instance Instance of @ref dnp3_request_t
/// @param variation Variation to ask for
/// @param start Start index to ask
/// @param stop Stop index to ask (inclusive)
void dnp3_request_add_two_byte_range_header(dnp3_request_t* instance, dnp3_variation_t variation, uint16_t start, uint16_t stop);

/// @brief Add an all objects variation request
/// @param instance Instance of @ref dnp3_request_t
/// @param variation Variation to ask for
void dnp3_request_add_all_objects_header(dnp3_request_t* instance, dnp3_variation_t variation);

/// @brief Add a one-byte limited count variation header for use with a READ request
/// @param instance Instance of @ref dnp3_request_t
/// @param variation Variation to ask for
/// @param count Maximum number of events
void dnp3_request_add_one_byte_limited_count_header(dnp3_request_t* instance, dnp3_variation_t variation, uint8_t count);

/// @brief Add a two-byte limited count variation header for use with a READ request
/// @param instance Instance of @ref dnp3_request_t
/// @param variation Variation to ask for
/// @param count Maximum number of events
void dnp3_request_add_two_byte_limited_count_header(dnp3_request_t* instance, dnp3_variation_t variation, uint16_t count);

/// @brief Add a single g51v1 time-and-interval
/// 
/// This is useful when constructing freeze-at-time requests
/// @param instance Instance of @ref dnp3_request_t
/// @param time DNP3 48-bit timestamp representing count of milliseconds since epoch UTC
/// @param interval_ms Interval expressed in milliseconds
void dnp3_request_add_time_and_interval(dnp3_request_t* instance, uint64_t time, uint32_t interval_ms);


/// @brief Add an association to the channel
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param address DNP3 data-link address of the remote outstation
/// @param config Association configuration
/// @param read_handler Interface uses to load measurement data
/// @param association_handler Association specific callbacks such as time synchronization
/// @param association_information Association information interface
/// @param out Id of the association
/// @return Error code
dnp3_param_error_t dnp3_master_channel_add_association(dnp3_master_channel_t* instance, uint16_t address, dnp3_association_config_t config, dnp3_read_handler_t read_handler, dnp3_association_handler_t association_handler, dnp3_association_information_t association_information, dnp3_association_id_t* out);

/// @brief Add a UDP association to the channel
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param address DNP3 data-link address of the remote outstation
/// @param destination IP address and port of the outstation
/// @param config Association configuration
/// @param read_handler Interface uses to load measurement data
/// @param association_handler Association specific callbacks such as time synchronization
/// @param association_information Association information interface
/// @param out Id of the association
/// @return Error code
dnp3_param_error_t dnp3_master_channel_add_udp_association(dnp3_master_channel_t* instance, uint16_t address, const char* destination, dnp3_association_config_t config, dnp3_read_handler_t read_handler, dnp3_association_handler_t association_handler, dnp3_association_information_t association_information, dnp3_association_id_t* out);

/// @brief Remove an association from the channel
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param id Id of the association
/// @return Error code
dnp3_param_error_t dnp3_master_channel_remove_association(dnp3_master_channel_t* instance, dnp3_association_id_t id);

/// @brief Add a periodic poll to an association
/// 
/// Each result of the poll will be sent to the @ref dnp3_read_handler_t of the association.
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param id Association on which to add the poll
/// @param request Request to perform
/// @param period Period to wait between each poll (in ms) (milliseconds)
/// @param out Id of the created poll
/// @return Error code
dnp3_param_error_t dnp3_master_channel_add_poll(dnp3_master_channel_t* instance, dnp3_association_id_t id, dnp3_request_t* request, uint64_t period, dnp3_poll_id_t* out);

/// @brief Add a periodic poll to an association
/// 
/// Each result of the poll will be sent to the @ref dnp3_read_handler_t of the association.
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param poll_id Id of the created poll
/// @return Error code
dnp3_param_error_t dnp3_master_channel_remove_poll(dnp3_master_channel_t* instance, dnp3_poll_id_t poll_id);

/// @brief Demand the immediate execution of a poll previously created with @ref dnp3_master_channel_add_poll.
/// 
/// This method returns immediately. The result will be sent to the registered @ref dnp3_read_handler_t.
/// 
/// This method resets the internal timer of the poll.
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param poll_id Id of the poll
/// @return Error code
dnp3_param_error_t dnp3_master_channel_demand_poll(dnp3_master_channel_t* instance, dnp3_poll_id_t poll_id);

/// @brief Set the decoding level for the channel
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param decode_level Decoding level
/// @return Error code
dnp3_param_error_t dnp3_master_channel_set_decode_level(dnp3_master_channel_t* instance, dnp3_decode_level_t decode_level);

/// @brief Get the decoding level for the channel
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param out Decode level
/// @return Error code
dnp3_param_error_t dnp3_master_channel_get_decode_level(dnp3_master_channel_t* instance, dnp3_decode_level_t* out);

/// @brief Errors that can occur during a read operation
typedef enum dnp3_read_error_t
{
    /// @brief Success, i.e. no error occurred
    DNP3_READ_ERROR_OK = 0,
    /// @brief too many user requests queued
    DNP3_READ_ERROR_TOO_MANY_REQUESTS = 1,
    /// @brief outstation returned an IIN.2 error bit
    DNP3_READ_ERROR_IIN_ERROR = 2,
    /// @brief response was malformed or contained object headers
    DNP3_READ_ERROR_BAD_RESPONSE = 3,
    /// @brief timeout occurred before receiving a response
    DNP3_READ_ERROR_RESPONSE_TIMEOUT = 4,
    /// @brief insufficient buffer space to serialize the request
    DNP3_READ_ERROR_WRITE_ERROR = 5,
    /// @brief no connection
    DNP3_READ_ERROR_NO_CONNECTION = 6,
    /// @brief master was shutdown
    DNP3_READ_ERROR_SHUTDOWN = 7,
    /// @brief association was removed mid-task
    DNP3_READ_ERROR_ASSOCIATION_REMOVED = 8,
    /// @brief request data could not be encoded
    DNP3_READ_ERROR_BAD_ENCODING = 9,
} dnp3_read_error_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_read_error_to_string(dnp3_read_error_t value)
{
    switch (value)
    {
        case DNP3_READ_ERROR_OK: return "ok";
        case DNP3_READ_ERROR_TOO_MANY_REQUESTS: return "too_many_requests";
        case DNP3_READ_ERROR_IIN_ERROR: return "iin_error";
        case DNP3_READ_ERROR_BAD_RESPONSE: return "bad_response";
        case DNP3_READ_ERROR_RESPONSE_TIMEOUT: return "response_timeout";
        case DNP3_READ_ERROR_WRITE_ERROR: return "write_error";
        case DNP3_READ_ERROR_NO_CONNECTION: return "no_connection";
        case DNP3_READ_ERROR_SHUTDOWN: return "shutdown";
        case DNP3_READ_ERROR_ASSOCIATION_REMOVED: return "association_removed";
        case DNP3_READ_ERROR_BAD_ENCODING: return "bad_encoding";
        default: return "unknown read_error value";
    }
}


/// @brief Handler for read tasks
typedef struct dnp3_read_task_callback_t
{
    
    /// @brief Invoked when the asynchronous operation completes successfully
    /// @param result Result of the read task
    /// @param ctx Context data
    void (*on_complete)(dnp3_nothing_t, void*);
    
    /// @brief Invoked when the asynchronous operation fails
    /// @param error Enumeration indicating which error occurred
    /// @param ctx Context data
    void (*on_failure)(dnp3_read_error_t, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_read_task_callback_t;

/// @brief Perform a read on the association.
/// 
/// The callback will be called once the read is completely received, but the actual values will be sent to the @ref dnp3_read_handler_t of the association.
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param association Association on which to perform the read
/// @param request Request to send
/// @param callback callback invoked when the operation completes
/// @return Error code
dnp3_param_error_t dnp3_master_channel_read(dnp3_master_channel_t* instance, dnp3_association_id_t association, dnp3_request_t* request, dnp3_read_task_callback_t callback);

/// @brief Perform a WRITE on the association using the supplied collection of dead-band headers
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param association Association on which to perform the WRITE
/// @param request Request containing headers of analog input dead-bands (group 34)
/// @param callback callback invoked when the operation completes
/// @return Error code
dnp3_param_error_t dnp3_master_channel_write_dead_bands(dnp3_master_channel_t* instance, dnp3_association_id_t association, dnp3_write_dead_band_request_t* request, dnp3_empty_response_callback_t callback);

/// @brief Send the specified request to the association using the supplied function and collection of request headers
/// 
/// This is useful for constructing various types of WRITE and FREEZE operations where an empty response is expected from the outstation, and the only indication of success are bits in IIN.2.
/// 
/// The request will fail if 1) The response contains object headers or 2) One of the error bits in IIN.2 is set.
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param association Association on which to perform the request
/// @param function Function code for the request
/// @param headers Headers that will be contained in the request
/// @param callback callback invoked when the operation completes
/// @return Error code
dnp3_param_error_t dnp3_master_channel_send_and_expect_empty_response(dnp3_master_channel_t* instance, dnp3_association_id_t association, dnp3_function_code_t function, dnp3_request_t* headers, dnp3_empty_response_callback_t callback);

/// @brief Perform a read on the association.
/// 
/// The callback will be called once the read is completely received, but the actual values will be sent to the @ref dnp3_read_handler_t passed as a parameter.
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param association Association on which to perform the read
/// @param request Request to send
/// @param handler Custom @ref dnp3_read_handler_t to send the data to
/// @param callback callback invoked when the operation completes
/// @return Error code
dnp3_param_error_t dnp3_master_channel_read_with_handler(dnp3_master_channel_t* instance, dnp3_association_id_t association, dnp3_request_t* request, dnp3_read_handler_t handler, dnp3_read_task_callback_t callback);

/// @brief Builder type used to construct command requests
typedef struct dnp3_command_set_t dnp3_command_set_t;

/// @brief Create a new set of commands
/// @return Instance of @ref dnp3_command_set_t
dnp3_command_set_t* dnp3_command_set_create();

/// @brief Destroy a set of commands
/// @param instance Instance of @ref dnp3_command_set_t to destroy
void dnp3_command_set_destroy(dnp3_command_set_t* instance);

/// @brief Finish any partially completed header. This allows for the construction of two headers with the same type and index
/// @param instance Instance of @ref dnp3_command_set_t
void dnp3_command_set_finish_header(dnp3_command_set_t* instance);

/// @brief Add a CROB with 1-byte prefix index
/// @param instance Instance of @ref dnp3_command_set_t
/// @param idx Index of the point to send the command to
/// @param header CROB data
void dnp3_command_set_add_g12_v1_u8(dnp3_command_set_t* instance, uint8_t idx, dnp3_group12_var1_t header);

/// @brief Add a CROB with 2-byte prefix index
/// @param instance Instance of @ref dnp3_command_set_t
/// @param idx Index of the point to send the command to
/// @param header CROB data
void dnp3_command_set_add_g12_v1_u16(dnp3_command_set_t* instance, uint16_t idx, dnp3_group12_var1_t header);

/// @brief Add a Analog Output command (signed 32-bit integer) with 1-byte prefix index
/// @param instance Instance of @ref dnp3_command_set_t
/// @param idx Index of the point to send the command to
/// @param value Value to set the analog output to
void dnp3_command_set_add_g41_v1_u8(dnp3_command_set_t* instance, uint8_t idx, int32_t value);

/// @brief Add a Analog Output command (signed 32-bit integer) with 2-byte prefix index
/// @param instance Instance of @ref dnp3_command_set_t
/// @param idx Index of the point to send the command to
/// @param value Value to set the analog output to
void dnp3_command_set_add_g41_v1_u16(dnp3_command_set_t* instance, uint16_t idx, int32_t value);

/// @brief Add a Analog Output command (signed 16-bit integer) with 1-byte prefix index
/// @param instance Instance of @ref dnp3_command_set_t
/// @param idx Index of the point to send the command to
/// @param value Value to set the analog output to
void dnp3_command_set_add_g41_v2_u8(dnp3_command_set_t* instance, uint8_t idx, int16_t value);

/// @brief Add a Analog Output command (signed 16-bit integer) with 2-byte prefix index
/// @param instance Instance of @ref dnp3_command_set_t
/// @param idx Index of the point to send the command to
/// @param value Value to set the analog output to
void dnp3_command_set_add_g41_v2_u16(dnp3_command_set_t* instance, uint16_t idx, int16_t value);

/// @brief Add a Analog Output command (single-precision float) with 1-byte prefix index
/// @param instance Instance of @ref dnp3_command_set_t
/// @param idx Index of the point to send the command to
/// @param value Value to set the analog output to
void dnp3_command_set_add_g41_v3_u8(dnp3_command_set_t* instance, uint8_t idx, float value);

/// @brief Add a Analog Output command (single-precision float) with 2-byte prefix index
/// @param instance Instance of @ref dnp3_command_set_t
/// @param idx Index of the point to send the command to
/// @param value Value to set the analog output to
void dnp3_command_set_add_g41_v3_u16(dnp3_command_set_t* instance, uint16_t idx, float value);

/// @brief Add a Analog Output command (double-precision float) with 1-byte prefix index
/// @param instance Instance of @ref dnp3_command_set_t
/// @param idx Index of the point to send the command to
/// @param value Value to set the analog output to
void dnp3_command_set_add_g41_v4_u8(dnp3_command_set_t* instance, uint8_t idx, double value);

/// @brief Add a Analog Output command (double-precision float) with 2-byte prefix index
/// @param instance Instance of @ref dnp3_command_set_t
/// @param idx Index of the point to send the command to
/// @param value Value to set the analog output to
void dnp3_command_set_add_g41_v4_u16(dnp3_command_set_t* instance, uint16_t idx, double value);


/// @brief Command operation mode
typedef enum dnp3_command_mode_t
{
    /// @brief Perform a Direct Operate (0x05)
    DNP3_COMMAND_MODE_DIRECT_OPERATE = 0,
    /// @brief Perform a Select and Operate (0x03 then 0x04)
    DNP3_COMMAND_MODE_SELECT_BEFORE_OPERATE = 1,
} dnp3_command_mode_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_command_mode_to_string(dnp3_command_mode_t value)
{
    switch (value)
    {
        case DNP3_COMMAND_MODE_DIRECT_OPERATE: return "direct_operate";
        case DNP3_COMMAND_MODE_SELECT_BEFORE_OPERATE: return "select_before_operate";
        default: return "unknown command_mode value";
    }
}

/// @brief Result of a command
typedef enum dnp3_command_error_t
{
    /// @brief Success, i.e. no error occurred
    DNP3_COMMAND_ERROR_OK = 0,
    /// @brief Outstation indicated that a command was not SUCCESS
    DNP3_COMMAND_ERROR_BAD_STATUS = 1,
    /// @brief Number of headers or objects in the response didn't match the number in the request
    DNP3_COMMAND_ERROR_HEADER_MISMATCH = 2,
    /// @brief too many user requests queued
    DNP3_COMMAND_ERROR_TOO_MANY_REQUESTS = 3,
    /// @brief outstation returned an IIN.2 error bit
    DNP3_COMMAND_ERROR_IIN_ERROR = 4,
    /// @brief response was malformed or contained object headers
    DNP3_COMMAND_ERROR_BAD_RESPONSE = 5,
    /// @brief timeout occurred before receiving a response
    DNP3_COMMAND_ERROR_RESPONSE_TIMEOUT = 6,
    /// @brief insufficient buffer space to serialize the request
    DNP3_COMMAND_ERROR_WRITE_ERROR = 7,
    /// @brief no connection
    DNP3_COMMAND_ERROR_NO_CONNECTION = 8,
    /// @brief master was shutdown
    DNP3_COMMAND_ERROR_SHUTDOWN = 9,
    /// @brief association was removed mid-task
    DNP3_COMMAND_ERROR_ASSOCIATION_REMOVED = 10,
    /// @brief request data could not be encoded
    DNP3_COMMAND_ERROR_BAD_ENCODING = 11,
} dnp3_command_error_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_command_error_to_string(dnp3_command_error_t value)
{
    switch (value)
    {
        case DNP3_COMMAND_ERROR_OK: return "ok";
        case DNP3_COMMAND_ERROR_BAD_STATUS: return "bad_status";
        case DNP3_COMMAND_ERROR_HEADER_MISMATCH: return "header_mismatch";
        case DNP3_COMMAND_ERROR_TOO_MANY_REQUESTS: return "too_many_requests";
        case DNP3_COMMAND_ERROR_IIN_ERROR: return "iin_error";
        case DNP3_COMMAND_ERROR_BAD_RESPONSE: return "bad_response";
        case DNP3_COMMAND_ERROR_RESPONSE_TIMEOUT: return "response_timeout";
        case DNP3_COMMAND_ERROR_WRITE_ERROR: return "write_error";
        case DNP3_COMMAND_ERROR_NO_CONNECTION: return "no_connection";
        case DNP3_COMMAND_ERROR_SHUTDOWN: return "shutdown";
        case DNP3_COMMAND_ERROR_ASSOCIATION_REMOVED: return "association_removed";
        case DNP3_COMMAND_ERROR_BAD_ENCODING: return "bad_encoding";
        default: return "unknown command_error value";
    }
}


/// @brief Handler for command tasks
typedef struct dnp3_command_task_callback_t
{
    
    /// @brief Invoked when the asynchronous operation completes successfully
    /// @param result Result of the command task
    /// @param ctx Context data
    void (*on_complete)(dnp3_nothing_t, void*);
    
    /// @brief Invoked when the asynchronous operation fails
    /// @param error Enumeration indicating which error occurred
    /// @param ctx Context data
    void (*on_failure)(dnp3_command_error_t, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_command_task_callback_t;

/// @brief Asynchronously perform a command operation on the association
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param association Id of the association
/// @param mode Operation mode
/// @param command Command to send
/// @param callback callback invoked when the operation completes
/// @return Error code
dnp3_param_error_t dnp3_master_channel_operate(dnp3_master_channel_t* instance, dnp3_association_id_t association, dnp3_command_mode_t mode, dnp3_command_set_t* command, dnp3_command_task_callback_t callback);

/// @brief Time synchronization mode
typedef enum dnp3_time_sync_mode_t
{
    /// @brief Perform a LAN time sync with Record Current Time (0x18) function code
    DNP3_TIME_SYNC_MODE_LAN = 0,
    /// @brief Perform a non-LAN time sync with Delay Measurement (0x17) function code
    DNP3_TIME_SYNC_MODE_NON_LAN = 1,
} dnp3_time_sync_mode_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_time_sync_mode_to_string(dnp3_time_sync_mode_t value)
{
    switch (value)
    {
        case DNP3_TIME_SYNC_MODE_LAN: return "lan";
        case DNP3_TIME_SYNC_MODE_NON_LAN: return "non_lan";
        default: return "unknown time_sync_mode value";
    }
}

/// @brief Possible errors that can occur during a time synchronization procedure
typedef enum dnp3_time_sync_error_t
{
    /// @brief Success, i.e. no error occurred
    DNP3_TIME_SYNC_ERROR_OK = 0,
    /// @brief Detected a clock rollback
    DNP3_TIME_SYNC_ERROR_CLOCK_ROLLBACK = 1,
    /// @brief The system time cannot be converted to a Unix timestamp
    DNP3_TIME_SYNC_ERROR_SYSTEM_TIME_NOT_UNIX = 2,
    /// @brief Outstation time delay exceeded the response delay
    DNP3_TIME_SYNC_ERROR_BAD_OUTSTATION_TIME_DELAY = 3,
    /// @brief Overflow in calculation
    DNP3_TIME_SYNC_ERROR_OVERFLOW = 4,
    /// @brief Outstation did not clear the NEED_TIME IIN bit
    DNP3_TIME_SYNC_ERROR_STILL_NEEDS_TIME = 5,
    /// @brief System time not available
    DNP3_TIME_SYNC_ERROR_SYSTEM_TIME_NOT_AVAILABLE = 6,
    /// @brief too many user requests queued
    DNP3_TIME_SYNC_ERROR_TOO_MANY_REQUESTS = 7,
    /// @brief outstation returned an IIN.2 error bit
    DNP3_TIME_SYNC_ERROR_IIN_ERROR = 8,
    /// @brief response was malformed or contained object headers
    DNP3_TIME_SYNC_ERROR_BAD_RESPONSE = 9,
    /// @brief timeout occurred before receiving a response
    DNP3_TIME_SYNC_ERROR_RESPONSE_TIMEOUT = 10,
    /// @brief insufficient buffer space to serialize the request
    DNP3_TIME_SYNC_ERROR_WRITE_ERROR = 11,
    /// @brief no connection
    DNP3_TIME_SYNC_ERROR_NO_CONNECTION = 12,
    /// @brief master was shutdown
    DNP3_TIME_SYNC_ERROR_SHUTDOWN = 13,
    /// @brief association was removed mid-task
    DNP3_TIME_SYNC_ERROR_ASSOCIATION_REMOVED = 14,
    /// @brief request data could not be encoded
    DNP3_TIME_SYNC_ERROR_BAD_ENCODING = 15,
} dnp3_time_sync_error_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_time_sync_error_to_string(dnp3_time_sync_error_t value)
{
    switch (value)
    {
        case DNP3_TIME_SYNC_ERROR_OK: return "ok";
        case DNP3_TIME_SYNC_ERROR_CLOCK_ROLLBACK: return "clock_rollback";
        case DNP3_TIME_SYNC_ERROR_SYSTEM_TIME_NOT_UNIX: return "system_time_not_unix";
        case DNP3_TIME_SYNC_ERROR_BAD_OUTSTATION_TIME_DELAY: return "bad_outstation_time_delay";
        case DNP3_TIME_SYNC_ERROR_OVERFLOW: return "overflow";
        case DNP3_TIME_SYNC_ERROR_STILL_NEEDS_TIME: return "still_needs_time";
        case DNP3_TIME_SYNC_ERROR_SYSTEM_TIME_NOT_AVAILABLE: return "system_time_not_available";
        case DNP3_TIME_SYNC_ERROR_TOO_MANY_REQUESTS: return "too_many_requests";
        case DNP3_TIME_SYNC_ERROR_IIN_ERROR: return "iin_error";
        case DNP3_TIME_SYNC_ERROR_BAD_RESPONSE: return "bad_response";
        case DNP3_TIME_SYNC_ERROR_RESPONSE_TIMEOUT: return "response_timeout";
        case DNP3_TIME_SYNC_ERROR_WRITE_ERROR: return "write_error";
        case DNP3_TIME_SYNC_ERROR_NO_CONNECTION: return "no_connection";
        case DNP3_TIME_SYNC_ERROR_SHUTDOWN: return "shutdown";
        case DNP3_TIME_SYNC_ERROR_ASSOCIATION_REMOVED: return "association_removed";
        case DNP3_TIME_SYNC_ERROR_BAD_ENCODING: return "bad_encoding";
        default: return "unknown time_sync_error value";
    }
}


/// @brief Handler for time synchronization tasks
typedef struct dnp3_time_sync_task_callback_t
{
    
    /// @brief Invoked when the asynchronous operation completes successfully
    /// @param result Result of the time synchronization task
    /// @param ctx Context data
    void (*on_complete)(dnp3_nothing_t, void*);
    
    /// @brief Invoked when the asynchronous operation fails
    /// @param error Enumeration indicating which error occurred
    /// @param ctx Context data
    void (*on_failure)(dnp3_time_sync_error_t, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_time_sync_task_callback_t;

/// @brief Asynchronously perform a time sync operation to the association
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param association Id of the association
/// @param mode Time sync mode
/// @param callback callback invoked when the operation completes
/// @return Error code
dnp3_param_error_t dnp3_master_channel_synchronize_time(dnp3_master_channel_t* instance, dnp3_association_id_t association, dnp3_time_sync_mode_t mode, dnp3_time_sync_task_callback_t callback);

/// @brief Errors that can occur during a cold/warm restart operation
typedef enum dnp3_restart_error_t
{
    /// @brief Success, i.e. no error occurred
    DNP3_RESTART_ERROR_OK = 0,
    /// @brief too many user requests queued
    DNP3_RESTART_ERROR_TOO_MANY_REQUESTS = 1,
    /// @brief outstation returned an IIN.2 error bit
    DNP3_RESTART_ERROR_IIN_ERROR = 2,
    /// @brief response was malformed or contained object headers
    DNP3_RESTART_ERROR_BAD_RESPONSE = 3,
    /// @brief timeout occurred before receiving a response
    DNP3_RESTART_ERROR_RESPONSE_TIMEOUT = 4,
    /// @brief insufficient buffer space to serialize the request
    DNP3_RESTART_ERROR_WRITE_ERROR = 5,
    /// @brief no connection
    DNP3_RESTART_ERROR_NO_CONNECTION = 6,
    /// @brief master was shutdown
    DNP3_RESTART_ERROR_SHUTDOWN = 7,
    /// @brief association was removed mid-task
    DNP3_RESTART_ERROR_ASSOCIATION_REMOVED = 8,
    /// @brief request data could not be encoded
    DNP3_RESTART_ERROR_BAD_ENCODING = 9,
} dnp3_restart_error_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_restart_error_to_string(dnp3_restart_error_t value)
{
    switch (value)
    {
        case DNP3_RESTART_ERROR_OK: return "ok";
        case DNP3_RESTART_ERROR_TOO_MANY_REQUESTS: return "too_many_requests";
        case DNP3_RESTART_ERROR_IIN_ERROR: return "iin_error";
        case DNP3_RESTART_ERROR_BAD_RESPONSE: return "bad_response";
        case DNP3_RESTART_ERROR_RESPONSE_TIMEOUT: return "response_timeout";
        case DNP3_RESTART_ERROR_WRITE_ERROR: return "write_error";
        case DNP3_RESTART_ERROR_NO_CONNECTION: return "no_connection";
        case DNP3_RESTART_ERROR_SHUTDOWN: return "shutdown";
        case DNP3_RESTART_ERROR_ASSOCIATION_REMOVED: return "association_removed";
        case DNP3_RESTART_ERROR_BAD_ENCODING: return "bad_encoding";
        default: return "unknown restart_error value";
    }
}


/// @brief Handler for restart tasks
typedef struct dnp3_restart_task_callback_t
{
    
    /// @brief Invoked when the asynchronous operation completes successfully
    /// @param result Result of the restart task
    /// @param ctx Context data
    void (*on_complete)(uint64_t, void*);
    
    /// @brief Invoked when the asynchronous operation fails
    /// @param error Enumeration indicating which error occurred
    /// @param ctx Context data
    void (*on_failure)(dnp3_restart_error_t, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_restart_task_callback_t;

/// @brief Asynchronously perform a cold restart operation to the association
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param association Id of the association
/// @param callback callback invoked when the operation completes
/// @return Error code
dnp3_param_error_t dnp3_master_channel_cold_restart(dnp3_master_channel_t* instance, dnp3_association_id_t association, dnp3_restart_task_callback_t callback);

/// @brief Asynchronously perform a warm restart operation to the association
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param association Id of the association
/// @param callback callback invoked when the operation completes
/// @return Error code
dnp3_param_error_t dnp3_master_channel_warm_restart(dnp3_master_channel_t* instance, dnp3_association_id_t association, dnp3_restart_task_callback_t callback);

/// @brief Callback interface for retrieving file info asynchronously
typedef struct dnp3_file_info_callback_t
{
    
    /// @brief Invoked when the asynchronous operation completes successfully
    /// @param result Information about the requested file
    /// @param ctx Context data
    void (*on_complete)(dnp3_file_info_t, void*);
    
    /// @brief Invoked when the asynchronous operation fails
    /// @param error Enumeration indicating which error occurred
    /// @param ctx Context data
    void (*on_failure)(dnp3_file_error_t, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_file_info_callback_t;

/// @brief  Callbacks for reading a file from the outstation asynchronously
typedef struct dnp3_file_reader_t
{
    
    /// @brief Called when the file is successfully opened
    /// 
    /// May optionally abort the operation by returning false
    /// @param size Size of the file returned by the outstation
    /// @param ctx Context data
    /// @return True to continue, false to abort
    bool (*opened)(uint32_t, void*);
    
    /// @brief Called when the next block is received
    /// 
    /// May optionally abort the transfer. This allows the application abort on internal errors like being or by user request.
    /// @param block_num The block number which increments as the transfer proceeds
    /// @param data iterator of bytes in the block
    /// @param ctx Context data
    /// @return True to continue, false to abort
    bool (*block_received)(uint32_t, dnp3_byte_iterator_t*, void*);
    
    /// @brief Called when the transfer is aborted before completion due to an error or user request
    /// @param error Error describing why the transfer aborted
    /// @param ctx Context data
    void (*aborted)(dnp3_file_error_t, void*);
    
    /// @brief Called when the transfer completes successfully
    /// @param ctx Context data
    void (*completed)(void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_file_reader_t;

typedef struct dnp3_file_read_config_t dnp3_file_read_config_t;

/// @brief Configuration related to reading a file
typedef struct dnp3_file_read_config_t
{
    /// @brief Maximum block size requested by the master during the file open
    uint16_t max_block_size;
    /// @brief Maximum file size accepted by the master
    uint32_t max_file_size;
} dnp3_file_read_config_t;

/// @brief Initialize the configuration to default values
/// 
/// @note Values are initialized to:
/// - @ref dnp3_file_read_config_t.max_block_size : 65535
/// - @ref dnp3_file_read_config_t.max_file_size : 4294967295
/// 
/// @returns New instance of @ref dnp3_file_read_config_t
static dnp3_file_read_config_t dnp3_file_read_config_defaults()
{
    dnp3_file_read_config_t _return_value = {
        65535,
        4294967295
    };
    return _return_value;
}


typedef struct dnp3_dir_read_config_t dnp3_dir_read_config_t;

/// @brief Configuration related to reading a file
typedef struct dnp3_dir_read_config_t
{
    /// @brief Maximum block size requested by the master during the file open
    uint16_t max_block_size;
    /// @brief Maximum number of bytes that may be accumulated while reading directory information
    uint32_t max_file_size;
} dnp3_dir_read_config_t;

/// @brief Initialize the configuration to default values
/// 
/// @note Values are initialized to:
/// - @ref dnp3_dir_read_config_t.max_block_size : 65535
/// - @ref dnp3_dir_read_config_t.max_file_size : 2048
/// 
/// @returns New instance of @ref dnp3_dir_read_config_t
static dnp3_dir_read_config_t dnp3_dir_read_config_defaults()
{
    dnp3_dir_read_config_t _return_value = {
        65535,
        2048
    };
    return _return_value;
}


/// @brief Start an operation to READ a file from the outstation using a @ref dnp3_file_reader_t to receive data
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param association Id of the association
/// @param remote_file_path Path of the remote file
/// @param config Configuration for the read operation
/// @param reader Interface used to receive file data
/// @return Error code
dnp3_param_error_t dnp3_master_channel_read_file(dnp3_master_channel_t* instance, dnp3_association_id_t association, const char* remote_file_path, dnp3_file_read_config_t config, dnp3_file_reader_t reader);

/// @brief Start an operation to READ a file from the outstation using a @ref dnp3_file_reader_t to receive data
/// 
/// This variant first requests an authentication key from the outstation using the supplied credentials
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param association Id of the association
/// @param remote_file_path Path of the remote file
/// @param config Configuration for the read operation
/// @param reader Interface used to receive file data
/// @param user_name User name sent to the outstation
/// @param password Password sent to the outstation
/// @return Error code
dnp3_param_error_t dnp3_master_channel_read_file_with_auth(dnp3_master_channel_t* instance, dnp3_association_id_t association, const char* remote_file_path, dnp3_file_read_config_t config, dnp3_file_reader_t reader, const char* user_name, const char* password);

/// @brief Obtain a file authentication key
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param association Id of the association
/// @param username User name
/// @param password Password
/// @param callback callback invoked when the operation completes
/// @return Error code
dnp3_param_error_t dnp3_master_channel_get_file_auth_key(dnp3_master_channel_t* instance, dnp3_association_id_t association, const char* username, const char* password, dnp3_file_auth_callback_t callback);

/// @brief Asynchronously open a file
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param association Id of the association
/// @param file_name Complete path to the remote file
/// @param auth_key Optional authentication key (0 == None)
/// @param permissions Permissions sent in the file open request
/// @param file_size File size sent in the request (zero when reading). When writing use the actual file size or 0xFFFFFFFF to indicate the size is unknown
/// @param file_mode Mode used to open the file
/// @param max_block_size Requested maximum block size
/// @param callback callback invoked when the operation completes
/// @return Error code
dnp3_param_error_t dnp3_master_channel_open_file(dnp3_master_channel_t* instance, dnp3_association_id_t association, const char* file_name, uint32_t auth_key, dnp3_permissions_t permissions, uint32_t file_size, dnp3_file_mode_t file_mode, uint16_t max_block_size, dnp3_file_open_callback_t callback);

/// @brief Asynchronously write a block of file data to the outstation
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param association Id of the association
/// @param handle Handle returned when the file was opened
/// @param block_number Sequential block number in the range [0 .. 0x7FFFFFFF]. The top bit is indicates the final block
/// @param final_block Indicate that this is the final block
/// @param block_data Collection of bytes. This must not be larger than the maximum block size returned by the outstation
/// @param callback callback invoked when the operation completes
/// @return Error code
dnp3_param_error_t dnp3_master_channel_write_file_block(dnp3_master_channel_t* instance, dnp3_association_id_t association, uint32_t handle, uint32_t block_number, bool final_block, dnp3_byte_collection_t* block_data, dnp3_file_operation_callback_t callback);

/// @brief Asynchronously close a file
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param association Id of the association
/// @param handle Handle returned when the file was opened
/// @param callback callback invoked when the operation completes
/// @return Error code
dnp3_param_error_t dnp3_master_channel_close_file(dnp3_master_channel_t* instance, dnp3_association_id_t association, uint32_t handle, dnp3_file_operation_callback_t callback);

/// @brief Asynchronously retrieve information on a particular file
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param association Id of the association
/// @param file_name Complete path to the remote file
/// @param callback callback invoked when the operation completes
/// @return Error code
dnp3_param_error_t dnp3_master_channel_get_file_info(dnp3_master_channel_t* instance, dnp3_association_id_t association, const char* file_name, dnp3_file_info_callback_t callback);

/// @brief Iterator of file_info
typedef struct dnp3_file_info_iterator_t dnp3_file_info_iterator_t;

/// @brief returns a pointer to the next value or NULL
/// @param iter opaque iterator on which to retrieve the next value
/// @return next value or NULL
dnp3_file_info_t* dnp3_file_info_iterator_next(dnp3_file_info_iterator_t* iter);


/// @brief Callback interface for retrieving a directory list
typedef struct dnp3_read_directory_callback_t
{
    
    /// @brief Invoked when the asynchronous operation completes successfully
    /// @param result iterator of @ref dnp3_file_info_t values
    /// @param ctx Context data
    void (*on_complete)(dnp3_file_info_iterator_t*, void*);
    
    /// @brief Invoked when the asynchronous operation fails
    /// @param error Enumeration indicating which error occurred
    /// @param ctx Context data
    void (*on_failure)(dnp3_file_error_t, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_read_directory_callback_t;

/// @brief Asynchronously retrieve a directory listing
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param association Id of the association
/// @param dir_path Complete path to the remote directory
/// @param config Configuration for the directory read operation
/// @param callback callback invoked when the operation completes
/// @return Error code
dnp3_param_error_t dnp3_master_channel_read_directory(dnp3_master_channel_t* instance, dnp3_association_id_t association, const char* dir_path, dnp3_dir_read_config_t config, dnp3_read_directory_callback_t callback);

/// @brief Asynchronously retrieve a directory listing by first obtaining an authentication key
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param association Id of the association
/// @param dir_path Complete path to the remote directory
/// @param config Configuration for the directory read operation
/// @param user_name User name sent to the outstation
/// @param password Password sent to the outstation
/// @param callback callback invoked when the operation completes
/// @return Error code
dnp3_param_error_t dnp3_master_channel_read_directory_with_auth(dnp3_master_channel_t* instance, dnp3_association_id_t association, const char* dir_path, dnp3_dir_read_config_t config, const char* user_name, const char* password, dnp3_read_directory_callback_t callback);

/// @brief Errors that can occur during a manually initiated link status check. See @ref dnp3_master_channel_check_link_status
typedef enum dnp3_link_status_error_t
{
    /// @brief Success, i.e. no error occurred
    DNP3_LINK_STATUS_ERROR_OK = 0,
    /// @brief There was activity on the link, but it wasn't a LINK_STATUS
    DNP3_LINK_STATUS_ERROR_UNEXPECTED_RESPONSE = 1,
    /// @brief too many user requests queued
    DNP3_LINK_STATUS_ERROR_TOO_MANY_REQUESTS = 2,
    /// @brief outstation returned an IIN.2 error bit
    DNP3_LINK_STATUS_ERROR_IIN_ERROR = 3,
    /// @brief response was malformed or contained object headers
    DNP3_LINK_STATUS_ERROR_BAD_RESPONSE = 4,
    /// @brief timeout occurred before receiving a response
    DNP3_LINK_STATUS_ERROR_RESPONSE_TIMEOUT = 5,
    /// @brief insufficient buffer space to serialize the request
    DNP3_LINK_STATUS_ERROR_WRITE_ERROR = 6,
    /// @brief no connection
    DNP3_LINK_STATUS_ERROR_NO_CONNECTION = 7,
    /// @brief master was shutdown
    DNP3_LINK_STATUS_ERROR_SHUTDOWN = 8,
    /// @brief association was removed mid-task
    DNP3_LINK_STATUS_ERROR_ASSOCIATION_REMOVED = 9,
    /// @brief request data could not be encoded
    DNP3_LINK_STATUS_ERROR_BAD_ENCODING = 10,
} dnp3_link_status_error_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_link_status_error_to_string(dnp3_link_status_error_t value)
{
    switch (value)
    {
        case DNP3_LINK_STATUS_ERROR_OK: return "ok";
        case DNP3_LINK_STATUS_ERROR_UNEXPECTED_RESPONSE: return "unexpected_response";
        case DNP3_LINK_STATUS_ERROR_TOO_MANY_REQUESTS: return "too_many_requests";
        case DNP3_LINK_STATUS_ERROR_IIN_ERROR: return "iin_error";
        case DNP3_LINK_STATUS_ERROR_BAD_RESPONSE: return "bad_response";
        case DNP3_LINK_STATUS_ERROR_RESPONSE_TIMEOUT: return "response_timeout";
        case DNP3_LINK_STATUS_ERROR_WRITE_ERROR: return "write_error";
        case DNP3_LINK_STATUS_ERROR_NO_CONNECTION: return "no_connection";
        case DNP3_LINK_STATUS_ERROR_SHUTDOWN: return "shutdown";
        case DNP3_LINK_STATUS_ERROR_ASSOCIATION_REMOVED: return "association_removed";
        case DNP3_LINK_STATUS_ERROR_BAD_ENCODING: return "bad_encoding";
        default: return "unknown link_status_error value";
    }
}


/// @brief Handler for link status check
typedef struct dnp3_link_status_callback_t
{
    
    /// @brief Invoked when the asynchronous operation completes successfully
    /// @param result Result of the link status
    /// @param ctx Context data
    void (*on_complete)(dnp3_nothing_t, void*);
    
    /// @brief Invoked when the asynchronous operation fails
    /// @param error Enumeration indicating which error occurred
    /// @param ctx Context data
    void (*on_failure)(dnp3_link_status_error_t, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_link_status_callback_t;

/// @brief Asynchronously perform a link status check
/// @param instance Instance of @ref dnp3_master_channel_t
/// @param association Id of the association
/// @param callback callback invoked when the operation completes
/// @return Error code
dnp3_param_error_t dnp3_master_channel_check_link_status(dnp3_master_channel_t* instance, dnp3_association_id_t association, dnp3_link_status_callback_t callback);


/// @brief Class used to accept a connection, reject it, or defer it to link identification
typedef struct dnp3_accept_handler_t dnp3_accept_handler_t;

/// @brief Accept the connection and create a master channel
/// @param instance Instance of @ref dnp3_accept_handler_t
/// @param error_mode Error mode to use for the link-layer. This should typically be @ref DNP3_LINK_ERROR_MODE_CLOSE
/// @param config Configuration of the channel
/// @return Enumeration describing the result of the operation
dnp3_param_error_t dnp3_accept_handler_accept(dnp3_accept_handler_t* instance, dnp3_link_error_mode_t error_mode, dnp3_master_channel_config_t config);

/// @brief Request that server attempt to identify the outstation by reading a link-layer header from the physical layer within a timeout.
/// 
/// This header is typically the beginning of an unsolicited fragment from the outstation.
/// @param instance Instance of @ref dnp3_accept_handler_t
/// @return Enumeration describing the result of the operation
dnp3_param_error_t dnp3_accept_handler_get_link_identity(dnp3_accept_handler_t* instance);


/// @brief Class used to accept a connection, reject it, or defer it to link identification
typedef struct dnp3_identified_link_handler_t dnp3_identified_link_handler_t;

/// @brief Accept the connection and create a master channel
/// @param instance Instance of @ref dnp3_identified_link_handler_t
/// @param error_mode Error mode to use for the link-layer. This should typically be @ref DNP3_LINK_ERROR_MODE_CLOSE
/// @param config Configuration of the channel
/// @return Enumeration describing the result of the operation
dnp3_param_error_t dnp3_identified_link_handler_accept(dnp3_identified_link_handler_t* instance, dnp3_link_error_mode_t error_mode, dnp3_master_channel_config_t config);


/// @brief Callbacks to user code that determine how the server processes connections
typedef struct dnp3_connection_handler_t
{
    
    /// @brief Filter the connection solely based on the remote address
    /// @param remote_addr Socket address of the remote outstation, e.g. 192.168.0.22:51532
    /// @param acceptor Class used to handle the accept
    /// @param ctx Context data
    void (*accept)(const char*, dnp3_accept_handler_t*, void*);
    
    /// @brief Start a communication session that was previously accepted using only the socket address
    /// 
    /// @warning You must add associations and/or enable the channel from a different thread than this callback as those methods cannot be called on the Tokio runtime
    /// @param remote_addr Socket address of the remote outstation, e.g. 192.168.0.22:51532
    /// @param channel Class used to control the channel
    /// @param ctx Context data
    void (*start)(const char*, dnp3_master_channel_t*, void*);
    
    /// @brief Filter the connection based on the source and destination of the first link-layer frame
    /// @param remote_addr Socket address of the remote outstation, e.g. 192.168.0.22:51532
    /// @param source Source address from the frame
    /// @param destination Destination address from the frame
    /// @param acceptor Class used to handle the accept
    /// @param ctx Context data
    void (*accept_with_link_id)(const char*, uint16_t, uint16_t, dnp3_identified_link_handler_t*, void*);
    
    /// @brief Start a communication session that was previously accepted using link identity information.
    /// 
    /// @warning You must add associations and/or enable the channel from a different thread than this callback as those methods cannot be called on the Tokio runtime
    /// @param remote_addr Socket address of the remote outstation, e.g. 192.168.0.22:51532
    /// @param source Source address from the frame
    /// @param destination Destination address from the frame
    /// @param channel Class used to control the channel
    /// @param ctx Context data
    void (*start_with_link_id)(const char*, uint16_t, uint16_t, dnp3_master_channel_t*, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_connection_handler_t;

/// @brief Class with methods used to spawn servers
typedef struct dnp3_master_server_t dnp3_master_server_t;

typedef struct dnp3_link_id_config_t dnp3_link_id_config_t;

/// @brief Configuration that controls how the server performs remote link identification
typedef struct dnp3_link_id_config_t
{
    /// @brief Set the maximum number of simultaneous tasks used to perform link identification
    uint16_t max_tasks;
    /// @brief Maximum time period to wait before receiving a link frame from the outstation
    /// @note The unit is milliseconds
    uint64_t timeout;
    /// @brief Set the decode level to use when reading the link header used for identification
    dnp3_phys_decode_level_t decode_level;
} dnp3_link_id_config_t;

/// @brief Initialize to default values
/// 
/// @note Values are initialized to:
/// - @ref dnp3_link_id_config_t.max_tasks : 16
/// - @ref dnp3_link_id_config_t.timeout : 5000ms
/// - @ref dnp3_link_id_config_t.decode_level : @ref DNP3_PHYS_DECODE_LEVEL_NOTHING
/// 
/// @returns New instance of @ref dnp3_link_id_config_t
static dnp3_link_id_config_t dnp3_link_id_config_init()
{
    dnp3_link_id_config_t _return_value = {
        16,
        5000,
        DNP3_PHYS_DECODE_LEVEL_NOTHING
    };
    return _return_value;
}


/// @brief Shutdown down the server
/// @param instance Instance of @ref dnp3_master_server_t to destroy
void dnp3_master_server_destroy(dnp3_master_server_t* instance);

/// @brief Spawn a TCP server that accepts connections from outstations
/// 
/// The behavior of each connection is controlled by callbacks to a user-defined implementation of a @ref dnp3_connection_handler_t.
/// @param runtime Runtime on which to spawn the server
/// @param local_addr Local address on which the server will accept connections
/// @param link_id_config Configuration used when identifying outstations based on received link-frames
/// @param connection_handler  Callbacks used to accept and start communication sessions
/// @param out Handle to the running server that allows it to be shut down
/// @return Error code
dnp3_param_error_t dnp3_create_master_tcp_server(dnp3_runtime_t* runtime, const char* local_addr, dnp3_link_id_config_t link_id_config, dnp3_connection_handler_t connection_handler, dnp3_master_server_t** out);


/// @brief Internal database access
/// 
/// @warning This object is only valid within a transaction
typedef struct dnp3_database_t dnp3_database_t;

/// @brief Event class
typedef enum dnp3_event_class_t
{
    /// @brief Does not generate events
    DNP3_EVENT_CLASS_NONE = 0,
    /// @brief Class 1 event
    DNP3_EVENT_CLASS_CLASS1 = 1,
    /// @brief Class 2 event
    DNP3_EVENT_CLASS_CLASS2 = 2,
    /// @brief Class 3 event
    DNP3_EVENT_CLASS_CLASS3 = 3,
} dnp3_event_class_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_event_class_to_string(dnp3_event_class_t value)
{
    switch (value)
    {
        case DNP3_EVENT_CLASS_NONE: return "none";
        case DNP3_EVENT_CLASS_CLASS1: return "class1";
        case DNP3_EVENT_CLASS_CLASS2: return "class2";
        case DNP3_EVENT_CLASS_CLASS3: return "class3";
        default: return "unknown event_class value";
    }
}

/// @brief Defines what occurred during an update operation and which fields of @ref dnp3_update_info_t are valid
typedef enum dnp3_update_result_t
{
    /// @brief No point exists for this type and index
    DNP3_UPDATE_RESULT_NO_POINT = 0,
    /// @brief The point exists, but the update did not create an event
    DNP3_UPDATE_RESULT_NO_EVENT = 1,
    /// @brief An event was created with the specified id
    DNP3_UPDATE_RESULT_CREATED = 2,
    /// @brief An event was created with the specified, but inserting it caused an event to be discarded
    DNP3_UPDATE_RESULT_OVERFLOW = 3,
} dnp3_update_result_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_update_result_to_string(dnp3_update_result_t value)
{
    switch (value)
    {
        case DNP3_UPDATE_RESULT_NO_POINT: return "no_point";
        case DNP3_UPDATE_RESULT_NO_EVENT: return "no_event";
        case DNP3_UPDATE_RESULT_CREATED: return "created";
        case DNP3_UPDATE_RESULT_OVERFLOW: return "overflow";
        default: return "unknown update_result value";
    }
}

typedef struct dnp3_update_info_t dnp3_update_info_t;

/// @brief Defines what occurred during an update operation. Only certain id fields are valid depending on the value of the enumeration
typedef struct dnp3_update_info_t
{
    /// @brief Defines what happened and which id fields are valid
    dnp3_update_result_t result;
    /// @brief The id of the created event if the result is @ref DNP3_UPDATE_RESULT_CREATED or @ref DNP3_UPDATE_RESULT_OVERFLOW
    uint64_t created;
    /// @brief The id of the discarded event if the result is @ref DNP3_UPDATE_RESULT_OVERFLOW
    uint64_t discarded;
} dnp3_update_info_t;


/// @brief Controls how events are processed when updating values in the database.
typedef enum dnp3_event_mode_t
{
    /// @brief Detect events in a type dependent fashion
    /// 
    /// This is the default mode that should be used.
    DNP3_EVENT_MODE_DETECT = 0,
    /// @brief Produce an event whether the value has changed or not
    DNP3_EVENT_MODE_FORCE = 1,
    /// @brief Never produce an event regardless of change
    DNP3_EVENT_MODE_SUPPRESS = 2,
} dnp3_event_mode_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_event_mode_to_string(dnp3_event_mode_t value)
{
    switch (value)
    {
        case DNP3_EVENT_MODE_DETECT: return "detect";
        case DNP3_EVENT_MODE_FORCE: return "force";
        case DNP3_EVENT_MODE_SUPPRESS: return "suppress";
        default: return "unknown event_mode value";
    }
}

typedef struct dnp3_update_options_t dnp3_update_options_t;

/// @brief Options that control how the update is performed.
/// 
/// 99% of the time, the default value should be used.
typedef struct dnp3_update_options_t
{
    /// @brief Optionally bypass updating the static database (the current value)
    bool update_static;
    /// @brief Determines how/if an event is produced
    dnp3_event_mode_t event_mode;
} dnp3_update_options_t;

/// @brief Default event detection mode. Updates the static value and automatically detects event.
/// 
/// @note Values are initialized to:
/// - @ref dnp3_update_options_t.update_static : @p true
/// - @ref dnp3_update_options_t.event_mode : @ref DNP3_EVENT_MODE_DETECT
/// 
/// @returns New instance of @ref dnp3_update_options_t
static dnp3_update_options_t dnp3_update_options_detect_event()
{
    dnp3_update_options_t _return_value = {
        true,
        DNP3_EVENT_MODE_DETECT
    };
    return _return_value;
}

/// @brief Only update the static value. Usefull during initialization of the database.
/// 
/// @note Values are initialized to:
/// - @ref dnp3_update_options_t.update_static : @p true
/// - @ref dnp3_update_options_t.event_mode : @ref DNP3_EVENT_MODE_SUPPRESS
/// 
/// @returns New instance of @ref dnp3_update_options_t
static dnp3_update_options_t dnp3_update_options_no_event()
{
    dnp3_update_options_t _return_value = {
        true,
        DNP3_EVENT_MODE_SUPPRESS
    };
    return _return_value;
}


/// @brief Point type on which to update the flags
typedef enum dnp3_update_flags_type_t
{
    /// @brief Binary input
    DNP3_UPDATE_FLAGS_TYPE_BINARY_INPUT = 0,
    /// @brief Doubl-bit binary input 
    DNP3_UPDATE_FLAGS_TYPE_DOUBLE_BIT_BINARY_INPUT = 1,
    /// @brief Binary output status
    DNP3_UPDATE_FLAGS_TYPE_BINARY_OUTPUT_STATUS = 2,
    /// @brief Counter
    DNP3_UPDATE_FLAGS_TYPE_COUNTER = 3,
    /// @brief Frozen counter
    DNP3_UPDATE_FLAGS_TYPE_FROZEN_COUNTER = 4,
    /// @brief Analog input
    DNP3_UPDATE_FLAGS_TYPE_ANALOG_INPUT = 5,
    /// @brief Analog output status
    DNP3_UPDATE_FLAGS_TYPE_ANALOG_OUTPUT_STATUS = 6,
} dnp3_update_flags_type_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_update_flags_type_to_string(dnp3_update_flags_type_t value)
{
    switch (value)
    {
        case DNP3_UPDATE_FLAGS_TYPE_BINARY_INPUT: return "binary_input";
        case DNP3_UPDATE_FLAGS_TYPE_DOUBLE_BIT_BINARY_INPUT: return "double_bit_binary_input";
        case DNP3_UPDATE_FLAGS_TYPE_BINARY_OUTPUT_STATUS: return "binary_output_status";
        case DNP3_UPDATE_FLAGS_TYPE_COUNTER: return "counter";
        case DNP3_UPDATE_FLAGS_TYPE_FROZEN_COUNTER: return "frozen_counter";
        case DNP3_UPDATE_FLAGS_TYPE_ANALOG_INPUT: return "analog_input";
        case DNP3_UPDATE_FLAGS_TYPE_ANALOG_OUTPUT_STATUS: return "analog_output_status";
        default: return "unknown update_flags_type value";
    }
}

/// @brief Update the flags for the specified point without changing the value
/// 
/// This is equivalent to getting the current value, changing the flags and the timestamp, then calling update
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index on which to perform the operation
/// @param flags_type Point type on which to perform the operation
/// @param flags New flags applied to the point
/// @param time New timestamp applied to the point
/// @param options Options that control how events and static values are handled
/// @return Provides detailed information about what occurred during the update operation
dnp3_update_info_t dnp3_database_update_flags(dnp3_database_t* instance, uint16_t index, dnp3_update_flags_type_t flags_type, dnp3_flags_t flags, dnp3_timestamp_t time, dnp3_update_options_t options);

/// @brief Static binary input variation
typedef enum dnp3_static_binary_input_variation_t
{
    /// @brief Binary input - packed format
    DNP3_STATIC_BINARY_INPUT_VARIATION_GROUP1_VAR1 = 0,
    /// @brief Binary input - with flags
    DNP3_STATIC_BINARY_INPUT_VARIATION_GROUP1_VAR2 = 1,
} dnp3_static_binary_input_variation_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_static_binary_input_variation_to_string(dnp3_static_binary_input_variation_t value)
{
    switch (value)
    {
        case DNP3_STATIC_BINARY_INPUT_VARIATION_GROUP1_VAR1: return "group1_var1";
        case DNP3_STATIC_BINARY_INPUT_VARIATION_GROUP1_VAR2: return "group1_var2";
        default: return "unknown static_binary_input_variation value";
    }
}

/// @brief Event binary input variation
typedef enum dnp3_event_binary_input_variation_t
{
    /// @brief Binary input event - without time
    DNP3_EVENT_BINARY_INPUT_VARIATION_GROUP2_VAR1 = 0,
    /// @brief Binary input event - with absolute time
    DNP3_EVENT_BINARY_INPUT_VARIATION_GROUP2_VAR2 = 1,
    /// @brief Binary input event - with relative time
    DNP3_EVENT_BINARY_INPUT_VARIATION_GROUP2_VAR3 = 2,
} dnp3_event_binary_input_variation_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_event_binary_input_variation_to_string(dnp3_event_binary_input_variation_t value)
{
    switch (value)
    {
        case DNP3_EVENT_BINARY_INPUT_VARIATION_GROUP2_VAR1: return "group2_var1";
        case DNP3_EVENT_BINARY_INPUT_VARIATION_GROUP2_VAR2: return "group2_var2";
        case DNP3_EVENT_BINARY_INPUT_VARIATION_GROUP2_VAR3: return "group2_var3";
        default: return "unknown event_binary_input_variation value";
    }
}

typedef struct dnp3_binary_input_config_t dnp3_binary_input_config_t;

/// @brief Binary Input configuration
typedef struct dnp3_binary_input_config_t
{
    /// @brief Default static variation
    dnp3_static_binary_input_variation_t static_variation;
    /// @brief Default event variation
    dnp3_event_binary_input_variation_t event_variation;
} dnp3_binary_input_config_t;

/// @brief Fully construct @ref dnp3_binary_input_config_t specifying the value of each field
/// @param static_variation Default static variation
/// @param event_variation Default event variation
/// @returns New instance of @ref dnp3_binary_input_config_t
static dnp3_binary_input_config_t dnp3_binary_input_config_create(dnp3_static_binary_input_variation_t static_variation, dnp3_event_binary_input_variation_t event_variation)
{
    dnp3_binary_input_config_t _return_value = {
        static_variation,
        event_variation
    };
    return _return_value;
}

/// @brief Initialize to defaults
/// 
/// @note Values are initialized to:
/// - @ref dnp3_binary_input_config_t.static_variation : @ref DNP3_STATIC_BINARY_INPUT_VARIATION_GROUP1_VAR1
/// - @ref dnp3_binary_input_config_t.event_variation : @ref DNP3_EVENT_BINARY_INPUT_VARIATION_GROUP2_VAR1
/// 
/// @returns New instance of @ref dnp3_binary_input_config_t
static dnp3_binary_input_config_t dnp3_binary_input_config_init()
{
    dnp3_binary_input_config_t _return_value = {
        DNP3_STATIC_BINARY_INPUT_VARIATION_GROUP1_VAR1,
        DNP3_EVENT_BINARY_INPUT_VARIATION_GROUP2_VAR1
    };
    return _return_value;
}


/// @brief Add a new BinaryInput point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point
/// @param point_class Event class
/// @param config Configuration
/// @return True if the point was successfully added, false otherwise
bool dnp3_database_add_binary_input(dnp3_database_t* instance, uint16_t index, dnp3_event_class_t point_class, dnp3_binary_input_config_t config);

/// @brief Remove a BinaryInput point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point
/// @return True if the point was successfully removed, false otherwise
bool dnp3_database_remove_binary_input(dnp3_database_t* instance, uint16_t index);

/// @brief Update a BinaryInput point
/// @param instance Instance of @ref dnp3_database_t
/// @param value New value of the point
/// @param options Update options
/// @return True if the point was successfully updated, false otherwise
bool dnp3_database_update_binary_input(dnp3_database_t* instance, dnp3_binary_input_t value, dnp3_update_options_t options);

/// @brief Update a BinaryInput point
/// @param instance Instance of @ref dnp3_database_t
/// @param value New value of the point
/// @param options Update options
/// @return Provides detailed information about what occurred during the update operation
dnp3_update_info_t dnp3_database_update_binary_input_2(dnp3_database_t* instance, dnp3_binary_input_t value, dnp3_update_options_t options);

/// @brief Get a BinaryInput point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point to get
/// @param out Binary Input point
/// @return Error code
dnp3_param_error_t dnp3_database_get_binary_input(dnp3_database_t* instance, uint16_t index, dnp3_binary_input_t* out);

/// @brief Static double-bit binary input variation
typedef enum dnp3_static_double_bit_binary_input_variation_t
{
    /// @brief Double-bit binary input - packed format
    DNP3_STATIC_DOUBLE_BIT_BINARY_INPUT_VARIATION_GROUP3_VAR1 = 0,
    /// @brief Double-bit binary input - with flags
    DNP3_STATIC_DOUBLE_BIT_BINARY_INPUT_VARIATION_GROUP3_VAR2 = 1,
} dnp3_static_double_bit_binary_input_variation_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_static_double_bit_binary_input_variation_to_string(dnp3_static_double_bit_binary_input_variation_t value)
{
    switch (value)
    {
        case DNP3_STATIC_DOUBLE_BIT_BINARY_INPUT_VARIATION_GROUP3_VAR1: return "group3_var1";
        case DNP3_STATIC_DOUBLE_BIT_BINARY_INPUT_VARIATION_GROUP3_VAR2: return "group3_var2";
        default: return "unknown static_double_bit_binary_input_variation value";
    }
}

/// @brief Event double-bit binary input variation
typedef enum dnp3_event_double_bit_binary_input_variation_t
{
    /// @brief Double-bit binary input event - without time
    DNP3_EVENT_DOUBLE_BIT_BINARY_INPUT_VARIATION_GROUP4_VAR1 = 0,
    /// @brief Double-bit binary input event - with absolute time
    DNP3_EVENT_DOUBLE_BIT_BINARY_INPUT_VARIATION_GROUP4_VAR2 = 1,
    /// @brief Double-bit binary input event - with relative time
    DNP3_EVENT_DOUBLE_BIT_BINARY_INPUT_VARIATION_GROUP4_VAR3 = 2,
} dnp3_event_double_bit_binary_input_variation_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_event_double_bit_binary_input_variation_to_string(dnp3_event_double_bit_binary_input_variation_t value)
{
    switch (value)
    {
        case DNP3_EVENT_DOUBLE_BIT_BINARY_INPUT_VARIATION_GROUP4_VAR1: return "group4_var1";
        case DNP3_EVENT_DOUBLE_BIT_BINARY_INPUT_VARIATION_GROUP4_VAR2: return "group4_var2";
        case DNP3_EVENT_DOUBLE_BIT_BINARY_INPUT_VARIATION_GROUP4_VAR3: return "group4_var3";
        default: return "unknown event_double_bit_binary_input_variation value";
    }
}

typedef struct dnp3_double_bit_binary_input_config_t dnp3_double_bit_binary_input_config_t;

/// @brief Double-Bit Binary Input configuration
typedef struct dnp3_double_bit_binary_input_config_t
{
    /// @brief Default static variation
    dnp3_static_double_bit_binary_input_variation_t static_variation;
    /// @brief Default event variation
    dnp3_event_double_bit_binary_input_variation_t event_variation;
} dnp3_double_bit_binary_input_config_t;

/// @brief Fully construct @ref dnp3_double_bit_binary_input_config_t specifying the value of each field
/// @param static_variation Default static variation
/// @param event_variation Default event variation
/// @returns New instance of @ref dnp3_double_bit_binary_input_config_t
static dnp3_double_bit_binary_input_config_t dnp3_double_bit_binary_input_config_create(dnp3_static_double_bit_binary_input_variation_t static_variation, dnp3_event_double_bit_binary_input_variation_t event_variation)
{
    dnp3_double_bit_binary_input_config_t _return_value = {
        static_variation,
        event_variation
    };
    return _return_value;
}

/// @brief Initialize to defaults
/// 
/// @note Values are initialized to:
/// - @ref dnp3_double_bit_binary_input_config_t.static_variation : @ref DNP3_STATIC_DOUBLE_BIT_BINARY_INPUT_VARIATION_GROUP3_VAR1
/// - @ref dnp3_double_bit_binary_input_config_t.event_variation : @ref DNP3_EVENT_DOUBLE_BIT_BINARY_INPUT_VARIATION_GROUP4_VAR1
/// 
/// @returns New instance of @ref dnp3_double_bit_binary_input_config_t
static dnp3_double_bit_binary_input_config_t dnp3_double_bit_binary_input_config_init()
{
    dnp3_double_bit_binary_input_config_t _return_value = {
        DNP3_STATIC_DOUBLE_BIT_BINARY_INPUT_VARIATION_GROUP3_VAR1,
        DNP3_EVENT_DOUBLE_BIT_BINARY_INPUT_VARIATION_GROUP4_VAR1
    };
    return _return_value;
}


/// @brief Add a new Double-Bit Binary Input point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point
/// @param point_class Event class
/// @param config Configuration
/// @return True if the point was successfully added, false otherwise
bool dnp3_database_add_double_bit_binary_input(dnp3_database_t* instance, uint16_t index, dnp3_event_class_t point_class, dnp3_double_bit_binary_input_config_t config);

/// @brief Remove a Double-Bit Binary Input point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point
/// @return True if the point was successfully removed, false otherwise
bool dnp3_database_remove_double_bit_binary_input(dnp3_database_t* instance, uint16_t index);

/// @brief Update a Double-Bit Binary Input point
/// @param instance Instance of @ref dnp3_database_t
/// @param value New value of the point
/// @param options Update options
/// @return True if the point was successfully updated, false otherwise
bool dnp3_database_update_double_bit_binary_input(dnp3_database_t* instance, dnp3_double_bit_binary_input_t value, dnp3_update_options_t options);

/// @brief Update a Double-Bit Binary Input point
/// @param instance Instance of @ref dnp3_database_t
/// @param value New value of the point
/// @param options Update options
/// @return Provides detailed information about what occurred during the update operation
dnp3_update_info_t dnp3_database_update_double_bit_binary_input_2(dnp3_database_t* instance, dnp3_double_bit_binary_input_t value, dnp3_update_options_t options);

/// @brief Get a Double-Bit Binary Input point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point to get
/// @param out Double-Bit Binary Input point
/// @return Error code
dnp3_param_error_t dnp3_database_get_double_bit_binary_input(dnp3_database_t* instance, uint16_t index, dnp3_double_bit_binary_input_t* out);

/// @brief Static binary output status variation
typedef enum dnp3_static_binary_output_status_variation_t
{
    /// @brief Binary output - packed format
    DNP3_STATIC_BINARY_OUTPUT_STATUS_VARIATION_GROUP10_VAR1 = 0,
    /// @brief Binary output - output status with flags
    DNP3_STATIC_BINARY_OUTPUT_STATUS_VARIATION_GROUP10_VAR2 = 1,
} dnp3_static_binary_output_status_variation_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_static_binary_output_status_variation_to_string(dnp3_static_binary_output_status_variation_t value)
{
    switch (value)
    {
        case DNP3_STATIC_BINARY_OUTPUT_STATUS_VARIATION_GROUP10_VAR1: return "group10_var1";
        case DNP3_STATIC_BINARY_OUTPUT_STATUS_VARIATION_GROUP10_VAR2: return "group10_var2";
        default: return "unknown static_binary_output_status_variation value";
    }
}

/// @brief Event binary output status variation
typedef enum dnp3_event_binary_output_status_variation_t
{
    /// @brief Binary output event - status without time
    DNP3_EVENT_BINARY_OUTPUT_STATUS_VARIATION_GROUP11_VAR1 = 0,
    /// @brief Binary output event - status with time
    DNP3_EVENT_BINARY_OUTPUT_STATUS_VARIATION_GROUP11_VAR2 = 1,
} dnp3_event_binary_output_status_variation_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_event_binary_output_status_variation_to_string(dnp3_event_binary_output_status_variation_t value)
{
    switch (value)
    {
        case DNP3_EVENT_BINARY_OUTPUT_STATUS_VARIATION_GROUP11_VAR1: return "group11_var1";
        case DNP3_EVENT_BINARY_OUTPUT_STATUS_VARIATION_GROUP11_VAR2: return "group11_var2";
        default: return "unknown event_binary_output_status_variation value";
    }
}

typedef struct dnp3_binary_output_status_config_t dnp3_binary_output_status_config_t;

/// @brief Binary Output Status configuration
typedef struct dnp3_binary_output_status_config_t
{
    /// @brief Default static variation
    dnp3_static_binary_output_status_variation_t static_variation;
    /// @brief Default event variation
    dnp3_event_binary_output_status_variation_t event_variation;
} dnp3_binary_output_status_config_t;

/// @brief Fully construct @ref dnp3_binary_output_status_config_t specifying the value of each field
/// @param static_variation Default static variation
/// @param event_variation Default event variation
/// @returns New instance of @ref dnp3_binary_output_status_config_t
static dnp3_binary_output_status_config_t dnp3_binary_output_status_config_create(dnp3_static_binary_output_status_variation_t static_variation, dnp3_event_binary_output_status_variation_t event_variation)
{
    dnp3_binary_output_status_config_t _return_value = {
        static_variation,
        event_variation
    };
    return _return_value;
}

/// @brief Initialize to defaults
/// 
/// @note Values are initialized to:
/// - @ref dnp3_binary_output_status_config_t.static_variation : @ref DNP3_STATIC_BINARY_OUTPUT_STATUS_VARIATION_GROUP10_VAR1
/// - @ref dnp3_binary_output_status_config_t.event_variation : @ref DNP3_EVENT_BINARY_OUTPUT_STATUS_VARIATION_GROUP11_VAR2
/// 
/// @returns New instance of @ref dnp3_binary_output_status_config_t
static dnp3_binary_output_status_config_t dnp3_binary_output_status_config_init()
{
    dnp3_binary_output_status_config_t _return_value = {
        DNP3_STATIC_BINARY_OUTPUT_STATUS_VARIATION_GROUP10_VAR1,
        DNP3_EVENT_BINARY_OUTPUT_STATUS_VARIATION_GROUP11_VAR2
    };
    return _return_value;
}


/// @brief Add a new Binary Output Status point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point
/// @param point_class Event class
/// @param config Configuration
/// @return True if the point was successfully added, false otherwise
bool dnp3_database_add_binary_output_status(dnp3_database_t* instance, uint16_t index, dnp3_event_class_t point_class, dnp3_binary_output_status_config_t config);

/// @brief Remove a Binary Output Status point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point
/// @return True if the point was successfully removed, false otherwise
bool dnp3_database_remove_binary_output_status(dnp3_database_t* instance, uint16_t index);

/// @brief Update a Binary Output Status point
/// @param instance Instance of @ref dnp3_database_t
/// @param value New value of the point
/// @param options Update options
/// @return True if the point was successfully updated, false otherwise
bool dnp3_database_update_binary_output_status(dnp3_database_t* instance, dnp3_binary_output_status_t value, dnp3_update_options_t options);

/// @brief Update a Binary Output Status point
/// @param instance Instance of @ref dnp3_database_t
/// @param value New value of the point
/// @param options Update options
/// @return Provides detailed information about what occurred during the update operation
dnp3_update_info_t dnp3_database_update_binary_output_status_2(dnp3_database_t* instance, dnp3_binary_output_status_t value, dnp3_update_options_t options);

/// @brief Get a Binary Output Status point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point to get
/// @param out Binary Output Status point
/// @return Error code
dnp3_param_error_t dnp3_database_get_binary_output_status(dnp3_database_t* instance, uint16_t index, dnp3_binary_output_status_t* out);

/// @brief Static counter variation
typedef enum dnp3_static_counter_variation_t
{
    /// @brief Counter - 32-bit with flag
    DNP3_STATIC_COUNTER_VARIATION_GROUP20_VAR1 = 0,
    /// @brief Counter - 16-bit with flag
    DNP3_STATIC_COUNTER_VARIATION_GROUP20_VAR2 = 1,
    /// @brief Counter - 32-bit without flag
    DNP3_STATIC_COUNTER_VARIATION_GROUP20_VAR5 = 2,
    /// @brief Counter - 16-bit without flag
    DNP3_STATIC_COUNTER_VARIATION_GROUP20_VAR6 = 3,
} dnp3_static_counter_variation_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_static_counter_variation_to_string(dnp3_static_counter_variation_t value)
{
    switch (value)
    {
        case DNP3_STATIC_COUNTER_VARIATION_GROUP20_VAR1: return "group20_var1";
        case DNP3_STATIC_COUNTER_VARIATION_GROUP20_VAR2: return "group20_var2";
        case DNP3_STATIC_COUNTER_VARIATION_GROUP20_VAR5: return "group20_var5";
        case DNP3_STATIC_COUNTER_VARIATION_GROUP20_VAR6: return "group20_var6";
        default: return "unknown static_counter_variation value";
    }
}

/// @brief Event counter variation
typedef enum dnp3_event_counter_variation_t
{
    /// @brief Counter event - 32-bit with flag
    DNP3_EVENT_COUNTER_VARIATION_GROUP22_VAR1 = 0,
    /// @brief Counter event - 16-bit with flag
    DNP3_EVENT_COUNTER_VARIATION_GROUP22_VAR2 = 1,
    /// @brief Counter event - 32-bit with flag and time
    DNP3_EVENT_COUNTER_VARIATION_GROUP22_VAR5 = 2,
    /// @brief Counter event - 16-bit with flag and time
    DNP3_EVENT_COUNTER_VARIATION_GROUP22_VAR6 = 3,
} dnp3_event_counter_variation_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_event_counter_variation_to_string(dnp3_event_counter_variation_t value)
{
    switch (value)
    {
        case DNP3_EVENT_COUNTER_VARIATION_GROUP22_VAR1: return "group22_var1";
        case DNP3_EVENT_COUNTER_VARIATION_GROUP22_VAR2: return "group22_var2";
        case DNP3_EVENT_COUNTER_VARIATION_GROUP22_VAR5: return "group22_var5";
        case DNP3_EVENT_COUNTER_VARIATION_GROUP22_VAR6: return "group22_var6";
        default: return "unknown event_counter_variation value";
    }
}

typedef struct dnp3_counter_config_t dnp3_counter_config_t;

/// @brief Counter configuration
typedef struct dnp3_counter_config_t
{
    /// @brief Default static variation
    dnp3_static_counter_variation_t static_variation;
    /// @brief Default event variation
    dnp3_event_counter_variation_t event_variation;
    /// @brief Deadband value
    uint32_t deadband;
} dnp3_counter_config_t;

/// @brief Fully construct @ref dnp3_counter_config_t specifying the value of each field
/// @param static_variation Default static variation
/// @param event_variation Default event variation
/// @param deadband Deadband value
/// @returns New instance of @ref dnp3_counter_config_t
static dnp3_counter_config_t dnp3_counter_config_create(dnp3_static_counter_variation_t static_variation, dnp3_event_counter_variation_t event_variation, uint32_t deadband)
{
    dnp3_counter_config_t _return_value = {
        static_variation,
        event_variation,
        deadband
    };
    return _return_value;
}

/// @brief Initialize to defaults
/// 
/// @note Values are initialized to:
/// - @ref dnp3_counter_config_t.static_variation : @ref DNP3_STATIC_COUNTER_VARIATION_GROUP20_VAR1
/// - @ref dnp3_counter_config_t.event_variation : @ref DNP3_EVENT_COUNTER_VARIATION_GROUP22_VAR1
/// - @ref dnp3_counter_config_t.deadband : 0
/// 
/// @returns New instance of @ref dnp3_counter_config_t
static dnp3_counter_config_t dnp3_counter_config_init()
{
    dnp3_counter_config_t _return_value = {
        DNP3_STATIC_COUNTER_VARIATION_GROUP20_VAR1,
        DNP3_EVENT_COUNTER_VARIATION_GROUP22_VAR1,
        0
    };
    return _return_value;
}


/// @brief Add a new Counter point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point
/// @param point_class Event class
/// @param config Configuration
/// @return True if the point was successfully added, false otherwise
bool dnp3_database_add_counter(dnp3_database_t* instance, uint16_t index, dnp3_event_class_t point_class, dnp3_counter_config_t config);

/// @brief Remove a Counter point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point
/// @return True if the point was successfully removed, false otherwise
bool dnp3_database_remove_counter(dnp3_database_t* instance, uint16_t index);

/// @brief Update a Counter point
/// @param instance Instance of @ref dnp3_database_t
/// @param value New value of the point
/// @param options Update options
/// @return True if the point was successfully updated, false otherwise
bool dnp3_database_update_counter(dnp3_database_t* instance, dnp3_counter_t value, dnp3_update_options_t options);

/// @brief Update a Counter point
/// @param instance Instance of @ref dnp3_database_t
/// @param value New value of the point
/// @param options Update options
/// @return Provides detailed information about what occurred during the update operation
dnp3_update_info_t dnp3_database_update_counter_2(dnp3_database_t* instance, dnp3_counter_t value, dnp3_update_options_t options);

/// @brief Get a Counter point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point to get
/// @param out Counter point
/// @return Error code
dnp3_param_error_t dnp3_database_get_counter(dnp3_database_t* instance, uint16_t index, dnp3_counter_t* out);

/// @brief Static frozen counter variation
typedef enum dnp3_static_frozen_counter_variation_t
{
    /// @brief Frozen Counter - 32-bit with flag
    DNP3_STATIC_FROZEN_COUNTER_VARIATION_GROUP21_VAR1 = 0,
    /// @brief Frozen Counter - 16-bit with flag
    DNP3_STATIC_FROZEN_COUNTER_VARIATION_GROUP21_VAR2 = 1,
    /// @brief Frozen Counter - 32-bit with flag and time
    DNP3_STATIC_FROZEN_COUNTER_VARIATION_GROUP21_VAR5 = 2,
    /// @brief Frozen Counter - 16-bit with flag and time
    DNP3_STATIC_FROZEN_COUNTER_VARIATION_GROUP21_VAR6 = 3,
    /// @brief Frozen Counter - 32-bit without flag
    DNP3_STATIC_FROZEN_COUNTER_VARIATION_GROUP21_VAR9 = 4,
    /// @brief Frozen Counter - 16-bit without flag
    DNP3_STATIC_FROZEN_COUNTER_VARIATION_GROUP21_VAR10 = 5,
} dnp3_static_frozen_counter_variation_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_static_frozen_counter_variation_to_string(dnp3_static_frozen_counter_variation_t value)
{
    switch (value)
    {
        case DNP3_STATIC_FROZEN_COUNTER_VARIATION_GROUP21_VAR1: return "group21_var1";
        case DNP3_STATIC_FROZEN_COUNTER_VARIATION_GROUP21_VAR2: return "group21_var2";
        case DNP3_STATIC_FROZEN_COUNTER_VARIATION_GROUP21_VAR5: return "group21_var5";
        case DNP3_STATIC_FROZEN_COUNTER_VARIATION_GROUP21_VAR6: return "group21_var6";
        case DNP3_STATIC_FROZEN_COUNTER_VARIATION_GROUP21_VAR9: return "group21_var9";
        case DNP3_STATIC_FROZEN_COUNTER_VARIATION_GROUP21_VAR10: return "group21_var10";
        default: return "unknown static_frozen_counter_variation value";
    }
}

/// @brief Event frozen counter variation
typedef enum dnp3_event_frozen_counter_variation_t
{
    /// @brief Frozen Counter event - 32-bit with flag
    DNP3_EVENT_FROZEN_COUNTER_VARIATION_GROUP23_VAR1 = 0,
    /// @brief Frozen Counter event - 16-bit with flag
    DNP3_EVENT_FROZEN_COUNTER_VARIATION_GROUP23_VAR2 = 1,
    /// @brief Frozen Counter event - 32-bit with flag and time
    DNP3_EVENT_FROZEN_COUNTER_VARIATION_GROUP23_VAR5 = 2,
    /// @brief Frozen Counter event - 16-bit with flag and time
    DNP3_EVENT_FROZEN_COUNTER_VARIATION_GROUP23_VAR6 = 3,
} dnp3_event_frozen_counter_variation_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_event_frozen_counter_variation_to_string(dnp3_event_frozen_counter_variation_t value)
{
    switch (value)
    {
        case DNP3_EVENT_FROZEN_COUNTER_VARIATION_GROUP23_VAR1: return "group23_var1";
        case DNP3_EVENT_FROZEN_COUNTER_VARIATION_GROUP23_VAR2: return "group23_var2";
        case DNP3_EVENT_FROZEN_COUNTER_VARIATION_GROUP23_VAR5: return "group23_var5";
        case DNP3_EVENT_FROZEN_COUNTER_VARIATION_GROUP23_VAR6: return "group23_var6";
        default: return "unknown event_frozen_counter_variation value";
    }
}

typedef struct dnp3_frozen_counter_config_t dnp3_frozen_counter_config_t;

/// @brief Frozen Counter configuration
typedef struct dnp3_frozen_counter_config_t
{
    /// @brief Default static variation
    dnp3_static_frozen_counter_variation_t static_variation;
    /// @brief Default event variation
    dnp3_event_frozen_counter_variation_t event_variation;
    /// @brief Deadband value
    uint32_t deadband;
} dnp3_frozen_counter_config_t;

/// @brief Fully construct @ref dnp3_frozen_counter_config_t specifying the value of each field
/// @param static_variation Default static variation
/// @param event_variation Default event variation
/// @param deadband Deadband value
/// @returns New instance of @ref dnp3_frozen_counter_config_t
static dnp3_frozen_counter_config_t dnp3_frozen_counter_config_create(dnp3_static_frozen_counter_variation_t static_variation, dnp3_event_frozen_counter_variation_t event_variation, uint32_t deadband)
{
    dnp3_frozen_counter_config_t _return_value = {
        static_variation,
        event_variation,
        deadband
    };
    return _return_value;
}

/// @brief Initialize to defaults
/// 
/// @note Values are initialized to:
/// - @ref dnp3_frozen_counter_config_t.static_variation : @ref DNP3_STATIC_FROZEN_COUNTER_VARIATION_GROUP21_VAR1
/// - @ref dnp3_frozen_counter_config_t.event_variation : @ref DNP3_EVENT_FROZEN_COUNTER_VARIATION_GROUP23_VAR1
/// - @ref dnp3_frozen_counter_config_t.deadband : 0
/// 
/// @returns New instance of @ref dnp3_frozen_counter_config_t
static dnp3_frozen_counter_config_t dnp3_frozen_counter_config_init()
{
    dnp3_frozen_counter_config_t _return_value = {
        DNP3_STATIC_FROZEN_COUNTER_VARIATION_GROUP21_VAR1,
        DNP3_EVENT_FROZEN_COUNTER_VARIATION_GROUP23_VAR1,
        0
    };
    return _return_value;
}


/// @brief Add a new Frozen Counter point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point
/// @param point_class Event class
/// @param config Configuration
/// @return True if the point was successfully added, false otherwise
bool dnp3_database_add_frozen_counter(dnp3_database_t* instance, uint16_t index, dnp3_event_class_t point_class, dnp3_frozen_counter_config_t config);

/// @brief Remove a Frozen Counter point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point
/// @return True if the point was successfully removed, false otherwise
bool dnp3_database_remove_frozen_counter(dnp3_database_t* instance, uint16_t index);

/// @brief Update an Frozen Counter point
/// @param instance Instance of @ref dnp3_database_t
/// @param value New value of the point
/// @param options Update options
/// @return True if the point was successfully updated, false otherwise
bool dnp3_database_update_frozen_counter(dnp3_database_t* instance, dnp3_frozen_counter_t value, dnp3_update_options_t options);

/// @brief Update an Frozen Counter point
/// @param instance Instance of @ref dnp3_database_t
/// @param value New value of the point
/// @param options Update options
/// @return Provides detailed information about what occurred during the update operation
dnp3_update_info_t dnp3_database_update_frozen_counter_2(dnp3_database_t* instance, dnp3_frozen_counter_t value, dnp3_update_options_t options);

/// @brief Get a Frozen Counter point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point to get
/// @param out Frozen Counter point
/// @return Error code
dnp3_param_error_t dnp3_database_get_frozen_counter(dnp3_database_t* instance, uint16_t index, dnp3_frozen_counter_t* out);

/// @brief Static analog variation
typedef enum dnp3_static_analog_input_variation_t
{
    /// @brief Analog input - 32-bit with flag
    DNP3_STATIC_ANALOG_INPUT_VARIATION_GROUP30_VAR1 = 0,
    /// @brief Analog input - 16-bit with flag
    DNP3_STATIC_ANALOG_INPUT_VARIATION_GROUP30_VAR2 = 1,
    /// @brief Analog input - 32-bit without flag
    DNP3_STATIC_ANALOG_INPUT_VARIATION_GROUP30_VAR3 = 2,
    /// @brief Analog input - 16-bit without flag
    DNP3_STATIC_ANALOG_INPUT_VARIATION_GROUP30_VAR4 = 3,
    /// @brief Analog input - single-precision, floating-point with flag
    DNP3_STATIC_ANALOG_INPUT_VARIATION_GROUP30_VAR5 = 4,
    /// @brief Analog input - double-precision, floating-point with flag
    DNP3_STATIC_ANALOG_INPUT_VARIATION_GROUP30_VAR6 = 5,
} dnp3_static_analog_input_variation_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_static_analog_input_variation_to_string(dnp3_static_analog_input_variation_t value)
{
    switch (value)
    {
        case DNP3_STATIC_ANALOG_INPUT_VARIATION_GROUP30_VAR1: return "group30_var1";
        case DNP3_STATIC_ANALOG_INPUT_VARIATION_GROUP30_VAR2: return "group30_var2";
        case DNP3_STATIC_ANALOG_INPUT_VARIATION_GROUP30_VAR3: return "group30_var3";
        case DNP3_STATIC_ANALOG_INPUT_VARIATION_GROUP30_VAR4: return "group30_var4";
        case DNP3_STATIC_ANALOG_INPUT_VARIATION_GROUP30_VAR5: return "group30_var5";
        case DNP3_STATIC_ANALOG_INPUT_VARIATION_GROUP30_VAR6: return "group30_var6";
        default: return "unknown static_analog_input_variation value";
    }
}

/// @brief Event analog variation
typedef enum dnp3_event_analog_input_variation_t
{
    /// @brief Analog input event - 32-bit without time
    DNP3_EVENT_ANALOG_INPUT_VARIATION_GROUP32_VAR1 = 0,
    /// @brief Analog input event - 16-bit without time
    DNP3_EVENT_ANALOG_INPUT_VARIATION_GROUP32_VAR2 = 1,
    /// @brief Analog input event - 32-bit with time
    DNP3_EVENT_ANALOG_INPUT_VARIATION_GROUP32_VAR3 = 2,
    /// @brief Analog input event - 16-bit with time
    DNP3_EVENT_ANALOG_INPUT_VARIATION_GROUP32_VAR4 = 3,
    /// @brief Analog input event - single-precision, floating-point without time
    DNP3_EVENT_ANALOG_INPUT_VARIATION_GROUP32_VAR5 = 4,
    /// @brief Analog input event - double-precision, floating-point without time
    DNP3_EVENT_ANALOG_INPUT_VARIATION_GROUP32_VAR6 = 5,
    /// @brief Analog input event - single-precision, floating-point with time
    DNP3_EVENT_ANALOG_INPUT_VARIATION_GROUP32_VAR7 = 6,
    /// @brief Analog input event - double-precision, floating-point with time
    DNP3_EVENT_ANALOG_INPUT_VARIATION_GROUP32_VAR8 = 7,
} dnp3_event_analog_input_variation_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_event_analog_input_variation_to_string(dnp3_event_analog_input_variation_t value)
{
    switch (value)
    {
        case DNP3_EVENT_ANALOG_INPUT_VARIATION_GROUP32_VAR1: return "group32_var1";
        case DNP3_EVENT_ANALOG_INPUT_VARIATION_GROUP32_VAR2: return "group32_var2";
        case DNP3_EVENT_ANALOG_INPUT_VARIATION_GROUP32_VAR3: return "group32_var3";
        case DNP3_EVENT_ANALOG_INPUT_VARIATION_GROUP32_VAR4: return "group32_var4";
        case DNP3_EVENT_ANALOG_INPUT_VARIATION_GROUP32_VAR5: return "group32_var5";
        case DNP3_EVENT_ANALOG_INPUT_VARIATION_GROUP32_VAR6: return "group32_var6";
        case DNP3_EVENT_ANALOG_INPUT_VARIATION_GROUP32_VAR7: return "group32_var7";
        case DNP3_EVENT_ANALOG_INPUT_VARIATION_GROUP32_VAR8: return "group32_var8";
        default: return "unknown event_analog_input_variation value";
    }
}

typedef struct dnp3_analog_input_config_t dnp3_analog_input_config_t;

/// @brief Analog configuration
typedef struct dnp3_analog_input_config_t
{
    /// @brief Default static variation
    dnp3_static_analog_input_variation_t static_variation;
    /// @brief Default event variation
    dnp3_event_analog_input_variation_t event_variation;
    /// @brief Deadband value
    double deadband;
} dnp3_analog_input_config_t;

/// @brief Fully construct @ref dnp3_analog_input_config_t specifying the value of each field
/// @param static_variation Default static variation
/// @param event_variation Default event variation
/// @param deadband Deadband value
/// @returns New instance of @ref dnp3_analog_input_config_t
static dnp3_analog_input_config_t dnp3_analog_input_config_create(dnp3_static_analog_input_variation_t static_variation, dnp3_event_analog_input_variation_t event_variation, double deadband)
{
    dnp3_analog_input_config_t _return_value = {
        static_variation,
        event_variation,
        deadband
    };
    return _return_value;
}

/// @brief Initialize to defaults
/// 
/// @note Values are initialized to:
/// - @ref dnp3_analog_input_config_t.static_variation : @ref DNP3_STATIC_ANALOG_INPUT_VARIATION_GROUP30_VAR1
/// - @ref dnp3_analog_input_config_t.event_variation : @ref DNP3_EVENT_ANALOG_INPUT_VARIATION_GROUP32_VAR1
/// - @ref dnp3_analog_input_config_t.deadband : 0
/// 
/// @returns New instance of @ref dnp3_analog_input_config_t
static dnp3_analog_input_config_t dnp3_analog_input_config_init()
{
    dnp3_analog_input_config_t _return_value = {
        DNP3_STATIC_ANALOG_INPUT_VARIATION_GROUP30_VAR1,
        DNP3_EVENT_ANALOG_INPUT_VARIATION_GROUP32_VAR1,
        0
    };
    return _return_value;
}


/// @brief Add a new AnalogInput point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point
/// @param point_class Event class
/// @param config Configuration
/// @return True if the point was successfully added, false otherwise
bool dnp3_database_add_analog_input(dnp3_database_t* instance, uint16_t index, dnp3_event_class_t point_class, dnp3_analog_input_config_t config);

/// @brief Remove an AnalogInput point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point
/// @return True if the point was successfully removed, false otherwise
bool dnp3_database_remove_analog_input(dnp3_database_t* instance, uint16_t index);

/// @brief Update a AnalogInput point
/// @param instance Instance of @ref dnp3_database_t
/// @param value New value of the point
/// @param options Update options
/// @return True if the point was successfully updated, false otherwise
bool dnp3_database_update_analog_input(dnp3_database_t* instance, dnp3_analog_input_t value, dnp3_update_options_t options);

/// @brief Update a AnalogInput point
/// @param instance Instance of @ref dnp3_database_t
/// @param value New value of the point
/// @param options Update options
/// @return Provides detailed information about what occurred during the update operation
dnp3_update_info_t dnp3_database_update_analog_input_2(dnp3_database_t* instance, dnp3_analog_input_t value, dnp3_update_options_t options);

/// @brief Get a AnalogInput point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point to get
/// @param out Analog point
/// @return Error code
dnp3_param_error_t dnp3_database_get_analog_input(dnp3_database_t* instance, uint16_t index, dnp3_analog_input_t* out);

/// @brief Static analog output status variation
typedef enum dnp3_static_analog_output_status_variation_t
{
    /// @brief Analog output status - 32-bit with flag
    DNP3_STATIC_ANALOG_OUTPUT_STATUS_VARIATION_GROUP40_VAR1 = 0,
    /// @brief Analog output status - 16-bit with flag
    DNP3_STATIC_ANALOG_OUTPUT_STATUS_VARIATION_GROUP40_VAR2 = 1,
    /// @brief Analog output status - single-precision, floating-point with flag
    DNP3_STATIC_ANALOG_OUTPUT_STATUS_VARIATION_GROUP40_VAR3 = 2,
    /// @brief Analog output status - double-precision, floating-point with flag
    DNP3_STATIC_ANALOG_OUTPUT_STATUS_VARIATION_GROUP40_VAR4 = 3,
} dnp3_static_analog_output_status_variation_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_static_analog_output_status_variation_to_string(dnp3_static_analog_output_status_variation_t value)
{
    switch (value)
    {
        case DNP3_STATIC_ANALOG_OUTPUT_STATUS_VARIATION_GROUP40_VAR1: return "group40_var1";
        case DNP3_STATIC_ANALOG_OUTPUT_STATUS_VARIATION_GROUP40_VAR2: return "group40_var2";
        case DNP3_STATIC_ANALOG_OUTPUT_STATUS_VARIATION_GROUP40_VAR3: return "group40_var3";
        case DNP3_STATIC_ANALOG_OUTPUT_STATUS_VARIATION_GROUP40_VAR4: return "group40_var4";
        default: return "unknown static_analog_output_status_variation value";
    }
}

/// @brief Event analog output status variation
typedef enum dnp3_event_analog_output_status_variation_t
{
    /// @brief Analog output event - 32-bit without time
    DNP3_EVENT_ANALOG_OUTPUT_STATUS_VARIATION_GROUP42_VAR1 = 0,
    /// @brief Analog output event - 16-bit without time
    DNP3_EVENT_ANALOG_OUTPUT_STATUS_VARIATION_GROUP42_VAR2 = 1,
    /// @brief Analog output event - 32-bit with time
    DNP3_EVENT_ANALOG_OUTPUT_STATUS_VARIATION_GROUP42_VAR3 = 2,
    /// @brief Analog output event - 16-bit with time
    DNP3_EVENT_ANALOG_OUTPUT_STATUS_VARIATION_GROUP42_VAR4 = 3,
    /// @brief Analog output event - single-precision, floating-point without time
    DNP3_EVENT_ANALOG_OUTPUT_STATUS_VARIATION_GROUP42_VAR5 = 4,
    /// @brief Analog output event - double-precision, floating-point without time
    DNP3_EVENT_ANALOG_OUTPUT_STATUS_VARIATION_GROUP42_VAR6 = 5,
    /// @brief Analog output event - single-precision, floating-point with time
    DNP3_EVENT_ANALOG_OUTPUT_STATUS_VARIATION_GROUP42_VAR7 = 6,
    /// @brief Analog output event - double-precision, floating-point with time
    DNP3_EVENT_ANALOG_OUTPUT_STATUS_VARIATION_GROUP42_VAR8 = 7,
} dnp3_event_analog_output_status_variation_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_event_analog_output_status_variation_to_string(dnp3_event_analog_output_status_variation_t value)
{
    switch (value)
    {
        case DNP3_EVENT_ANALOG_OUTPUT_STATUS_VARIATION_GROUP42_VAR1: return "group42_var1";
        case DNP3_EVENT_ANALOG_OUTPUT_STATUS_VARIATION_GROUP42_VAR2: return "group42_var2";
        case DNP3_EVENT_ANALOG_OUTPUT_STATUS_VARIATION_GROUP42_VAR3: return "group42_var3";
        case DNP3_EVENT_ANALOG_OUTPUT_STATUS_VARIATION_GROUP42_VAR4: return "group42_var4";
        case DNP3_EVENT_ANALOG_OUTPUT_STATUS_VARIATION_GROUP42_VAR5: return "group42_var5";
        case DNP3_EVENT_ANALOG_OUTPUT_STATUS_VARIATION_GROUP42_VAR6: return "group42_var6";
        case DNP3_EVENT_ANALOG_OUTPUT_STATUS_VARIATION_GROUP42_VAR7: return "group42_var7";
        case DNP3_EVENT_ANALOG_OUTPUT_STATUS_VARIATION_GROUP42_VAR8: return "group42_var8";
        default: return "unknown event_analog_output_status_variation value";
    }
}

typedef struct dnp3_analog_output_status_config_t dnp3_analog_output_status_config_t;

/// @brief Analog Output Status configuration
typedef struct dnp3_analog_output_status_config_t
{
    /// @brief Default static variation
    dnp3_static_analog_output_status_variation_t static_variation;
    /// @brief Default event variation
    dnp3_event_analog_output_status_variation_t event_variation;
    /// @brief Deadband value
    double deadband;
} dnp3_analog_output_status_config_t;

/// @brief Fully construct @ref dnp3_analog_output_status_config_t specifying the value of each field
/// @param static_variation Default static variation
/// @param event_variation Default event variation
/// @param deadband Deadband value
/// @returns New instance of @ref dnp3_analog_output_status_config_t
static dnp3_analog_output_status_config_t dnp3_analog_output_status_config_create(dnp3_static_analog_output_status_variation_t static_variation, dnp3_event_analog_output_status_variation_t event_variation, double deadband)
{
    dnp3_analog_output_status_config_t _return_value = {
        static_variation,
        event_variation,
        deadband
    };
    return _return_value;
}

/// @brief Initialize to defaults
/// 
/// @note Values are initialized to:
/// - @ref dnp3_analog_output_status_config_t.static_variation : @ref DNP3_STATIC_ANALOG_OUTPUT_STATUS_VARIATION_GROUP40_VAR1
/// - @ref dnp3_analog_output_status_config_t.event_variation : @ref DNP3_EVENT_ANALOG_OUTPUT_STATUS_VARIATION_GROUP42_VAR1
/// - @ref dnp3_analog_output_status_config_t.deadband : 0
/// 
/// @returns New instance of @ref dnp3_analog_output_status_config_t
static dnp3_analog_output_status_config_t dnp3_analog_output_status_config_init()
{
    dnp3_analog_output_status_config_t _return_value = {
        DNP3_STATIC_ANALOG_OUTPUT_STATUS_VARIATION_GROUP40_VAR1,
        DNP3_EVENT_ANALOG_OUTPUT_STATUS_VARIATION_GROUP42_VAR1,
        0
    };
    return _return_value;
}


/// @brief Add a new Analog Output Status point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point
/// @param point_class Event class
/// @param config Configuration
/// @return True if the point was successfully added, false otherwise
bool dnp3_database_add_analog_output_status(dnp3_database_t* instance, uint16_t index, dnp3_event_class_t point_class, dnp3_analog_output_status_config_t config);

/// @brief Remove an Analog Output Status point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point
/// @return True if the point was successfully removed, false otherwise
bool dnp3_database_remove_analog_output_status(dnp3_database_t* instance, uint16_t index);

/// @brief Update a Analog Output Status point
/// @param instance Instance of @ref dnp3_database_t
/// @param value New value of the point
/// @param options Update options
/// @return True if the point was successfully updated, false otherwise
bool dnp3_database_update_analog_output_status(dnp3_database_t* instance, dnp3_analog_output_status_t value, dnp3_update_options_t options);

/// @brief Update a Analog Output Status point
/// @param instance Instance of @ref dnp3_database_t
/// @param value New value of the point
/// @param options Update options
/// @return Provides detailed information about what occurred during the update operation
dnp3_update_info_t dnp3_database_update_analog_output_status_2(dnp3_database_t* instance, dnp3_analog_output_status_t value, dnp3_update_options_t options);

/// @brief Get a Analog Output Status point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point to get
/// @param out Analog Output Status point
/// @return Error code
dnp3_param_error_t dnp3_database_get_analog_output_status(dnp3_database_t* instance, uint16_t index, dnp3_analog_output_status_t* out);

/// @brief Collection of octet_string_value
typedef struct dnp3_octet_string_value_t dnp3_octet_string_value_t;

/// @brief Creates an instance of the collection
/// @return Allocated opaque collection instance
dnp3_octet_string_value_t* dnp3_octet_string_value_create();

/// @brief Destroys an instance of the collection
/// @param instance instance to destroy
void dnp3_octet_string_value_destroy(dnp3_octet_string_value_t* instance);

/// @brief Add a value to the collection
/// @param instance instance to which to add the value
/// @param value value to add to the instance
void dnp3_octet_string_value_add(dnp3_octet_string_value_t* instance, uint8_t value);


/// @brief Add a new Octet String point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point
/// @param point_class Event class
/// @return True if the point was successfully added, false otherwise
bool dnp3_database_add_octet_string(dnp3_database_t* instance, uint16_t index, dnp3_event_class_t point_class);

/// @brief Remove an Octet String point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the point
/// @return True if the point was successfully removed, false otherwise
bool dnp3_database_remove_octet_string(dnp3_database_t* instance, uint16_t index);

/// @brief Update an Octet String point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the octet string
/// @param value New value of the point
/// @param options Update options
/// @return True if the point was successfully updated, false otherwise
bool dnp3_database_update_octet_string(dnp3_database_t* instance, uint16_t index, dnp3_octet_string_value_t* value, dnp3_update_options_t options);

/// @brief Update an Octet String point
/// @param instance Instance of @ref dnp3_database_t
/// @param index Index of the octet string
/// @param value New value of the point
/// @param options Update options
/// @return Provides detailed information about what occurred during the update operation
dnp3_update_info_t dnp3_database_update_octet_string_2(dnp3_database_t* instance, uint16_t index, dnp3_octet_string_value_t* value, dnp3_update_options_t options);

/// @brief Errors that can occur when defining attributes
typedef enum dnp3_attr_def_error_t
{
    /// @brief attribute defined successfully
    DNP3_ATTR_DEF_ERROR_OK = 0,
    /// @brief Attribute has already been defined
    DNP3_ATTR_DEF_ERROR_ALREADY_DEFINED = 1,
    /// @brief Variation is reserved and cannot be defined
    DNP3_ATTR_DEF_ERROR_RESERVED_VARIATION = 2,
    /// @brief The type does not match the required type in set 0
    DNP3_ATTR_DEF_ERROR_BAD_TYPE = 3,
    /// @brief This attribute cannot be configured as writable
    DNP3_ATTR_DEF_ERROR_NOT_WRITABLE = 4,
} dnp3_attr_def_error_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_attr_def_error_to_string(dnp3_attr_def_error_t value)
{
    switch (value)
    {
        case DNP3_ATTR_DEF_ERROR_OK: return "ok";
        case DNP3_ATTR_DEF_ERROR_ALREADY_DEFINED: return "already_defined";
        case DNP3_ATTR_DEF_ERROR_RESERVED_VARIATION: return "reserved_variation";
        case DNP3_ATTR_DEF_ERROR_BAD_TYPE: return "bad_type";
        case DNP3_ATTR_DEF_ERROR_NOT_WRITABLE: return "not_writable";
        default: return "unknown attr_def_error value";
    }
}

/// @brief Define a string attribute
/// @param instance Instance of @ref dnp3_database_t
/// @param set The set to which the attribute belongs
/// @param writable True if the attribute may be written
/// @param variation The variation of the attribute
/// @param value The value of the attribute
/// @return Enumeration indicating if the operation was successful
dnp3_attr_def_error_t dnp3_database_define_string_attr(dnp3_database_t* instance, uint8_t set, bool writable, uint8_t variation, const char* value);

/// @brief Define a 32-bit floating point attribute
/// @param instance Instance of @ref dnp3_database_t
/// @param set The set to which the attribute belongs
/// @param writable True if the attribute may be written
/// @param variation The variation of the attribute
/// @param value The value of the attribute
/// @return Enumeration indicating if the operation was successful
dnp3_attr_def_error_t dnp3_database_define_float_attr(dnp3_database_t* instance, uint8_t set, bool writable, uint8_t variation, float value);

/// @brief Define a 64-bit floating point attribute
/// @param instance Instance of @ref dnp3_database_t
/// @param set The set to which the attribute belongs
/// @param writable True if the attribute may be written
/// @param variation The variation of the attribute
/// @param value The value of the attribute
/// @return Enumeration indicating if the operation was successful
dnp3_attr_def_error_t dnp3_database_define_double_attr(dnp3_database_t* instance, uint8_t set, bool writable, uint8_t variation, double value);

/// @brief Define an unsigned integer attribute
/// @param instance Instance of @ref dnp3_database_t
/// @param set The set to which the attribute belongs
/// @param writable True if the attribute may be written
/// @param variation The variation of the attribute
/// @param value The value of the attribute
/// @return Enumeration indicating if the operation was successful
dnp3_attr_def_error_t dnp3_database_define_uint_attr(dnp3_database_t* instance, uint8_t set, bool writable, uint8_t variation, uint32_t value);

/// @brief Define a signed integer attribute
/// @param instance Instance of @ref dnp3_database_t
/// @param set The set to which the attribute belongs
/// @param writable True if the attribute may be written
/// @param variation The variation of the attribute
/// @param value The value of the attribute
/// @return Enumeration indicating if the operation was successful
dnp3_attr_def_error_t dnp3_database_define_int_attr(dnp3_database_t* instance, uint8_t set, bool writable, uint8_t variation, int32_t value);

/// @brief Define a DNP3 time attribute
/// @param instance Instance of @ref dnp3_database_t
/// @param set The set to which the attribute belongs
/// @param writable True if the attribute may be written
/// @param variation The variation of the attribute
/// @param value The DNP3 timestamp value of the attribute. Only the lower 48-bits are used.
/// @return Enumeration indicating if the operation was successful
dnp3_attr_def_error_t dnp3_database_define_time_attr(dnp3_database_t* instance, uint8_t set, bool writable, uint8_t variation, uint64_t value);

/// @brief Define a boolean attribute which is mapped to an unsigned integer internally
/// @param instance Instance of @ref dnp3_database_t
/// @param set The set to which the attribute belongs
/// @param writable True if the attribute may be written
/// @param variation The variation of the attribute
/// @param value The value of the attribute
/// @return Enumeration indicating if the operation was successful
dnp3_attr_def_error_t dnp3_database_define_bool_attr(dnp3_database_t* instance, uint8_t set, bool writable, uint8_t variation, bool value);


/// @brief Database transaction interface
typedef struct dnp3_database_transaction_t
{
    
    /// @brief Execute a transaction on the provided database
    /// @param database Database
    /// @param ctx Context data
    void (*execute)(dnp3_database_t*, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_database_transaction_t;

/// @brief Handle typed used to perform transactions on the database inside of control and freeze callbacks
/// 
/// This type has the same transaction method as @ref dnp3_outstation_transaction but it is only used in these callbacks.
typedef struct dnp3_database_handle_t dnp3_database_handle_t;

/// @brief Acquire a mutex on the underlying database and apply a set of changes as a transaction
/// @param instance Instance of @ref dnp3_database_handle_t
/// @param callback callback interface
void dnp3_database_handle_transaction(dnp3_database_handle_t* instance, dnp3_database_transaction_t callback);


typedef struct dnp3_event_buffer_config_t dnp3_event_buffer_config_t;

/// @brief Maximum number of events for each type
/// 
/// A value of zero means that events will not be buffered for that type.
typedef struct dnp3_event_buffer_config_t
{
    /// @brief Maximum number of Binary Input events (g2)
    uint16_t max_binary;
    /// @brief Maximum number of Double-Bit Binary Input events (g4)
    uint16_t max_double_bit_binary;
    /// @brief Maximum number of Binary Output Status events (g11)
    uint16_t max_binary_output_status;
    /// @brief Maximum number of Counter events (g22)
    uint16_t max_counter;
    /// @brief Maximum number of Frozen Counter events (g23)
    uint16_t max_frozen_counter;
    /// @brief Maximum number of Analog Input events (g32)
    uint16_t max_analog;
    /// @brief Maximum number of Analog Output Status events (g42)
    uint16_t max_analog_output_status;
    /// @brief Maximum number of Octet String events (g111)
    uint16_t max_octet_string;
} dnp3_event_buffer_config_t;

/// @brief Fully construct @ref dnp3_event_buffer_config_t specifying the value of each field
/// @param max_binary Maximum number of Binary Input events (g2)
/// @param max_double_bit_binary Maximum number of Double-Bit Binary Input events (g4)
/// @param max_binary_output_status Maximum number of Binary Output Status events (g11)
/// @param max_counter Maximum number of Counter events (g22)
/// @param max_frozen_counter Maximum number of Frozen Counter events (g23)
/// @param max_analog Maximum number of Analog Input events (g32)
/// @param max_analog_output_status Maximum number of Analog Output Status events (g42)
/// @param max_octet_string Maximum number of Octet String events (g111)
/// @returns New instance of @ref dnp3_event_buffer_config_t
static dnp3_event_buffer_config_t dnp3_event_buffer_config_init(uint16_t max_binary, uint16_t max_double_bit_binary, uint16_t max_binary_output_status, uint16_t max_counter, uint16_t max_frozen_counter, uint16_t max_analog, uint16_t max_analog_output_status, uint16_t max_octet_string)
{
    dnp3_event_buffer_config_t _return_value = {
        max_binary,
        max_double_bit_binary,
        max_binary_output_status,
        max_counter,
        max_frozen_counter,
        max_analog,
        max_analog_output_status,
        max_octet_string
    };
    return _return_value;
}

/// @brief Create a configuration where no events are buffered.
/// 
/// @note Values are initialized to:
/// - @ref dnp3_event_buffer_config_t.max_binary : 0
/// - @ref dnp3_event_buffer_config_t.max_double_bit_binary : 0
/// - @ref dnp3_event_buffer_config_t.max_binary_output_status : 0
/// - @ref dnp3_event_buffer_config_t.max_counter : 0
/// - @ref dnp3_event_buffer_config_t.max_frozen_counter : 0
/// - @ref dnp3_event_buffer_config_t.max_analog : 0
/// - @ref dnp3_event_buffer_config_t.max_analog_output_status : 0
/// - @ref dnp3_event_buffer_config_t.max_octet_string : 0
/// 
/// @returns New instance of @ref dnp3_event_buffer_config_t
static dnp3_event_buffer_config_t dnp3_event_buffer_config_no_events()
{
    dnp3_event_buffer_config_t _return_value = {
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0
    };
    return _return_value;
}


typedef struct dnp3_class_zero_config_t dnp3_class_zero_config_t;

/// @brief Controls which types are reported during a Class 0 read.
typedef struct dnp3_class_zero_config_t
{
    /// @brief Include Binary Inputs in Class 0 reads
    bool binary;
    /// @brief Include Double-Bit Binary Inputs in Class 0 reads
    bool double_bit_binary;
    /// @brief Include Binary Output Status in Class 0 reads
    bool binary_output_status;
    /// @brief Include Counters in Class 0 reads
    bool counter;
    /// @brief Include Frozen Counters in Class 0 reads
    bool frozen_counter;
    /// @brief Include Analog Inputs in Class 0 reads
    bool analog;
    /// @brief Include Analog Output Status in Class 0 reads
    bool analog_output_status;
    /// @brief Include Octet Strings in Class 0 reads
    /// 
    /// @warning For conformance, this should be false.
    bool octet_string;
} dnp3_class_zero_config_t;

/// @brief Initialize to default values
/// 
/// @note Values are initialized to:
/// - @ref dnp3_class_zero_config_t.binary : @p true
/// - @ref dnp3_class_zero_config_t.double_bit_binary : @p true
/// - @ref dnp3_class_zero_config_t.binary_output_status : @p true
/// - @ref dnp3_class_zero_config_t.counter : @p true
/// - @ref dnp3_class_zero_config_t.frozen_counter : @p true
/// - @ref dnp3_class_zero_config_t.analog : @p true
/// - @ref dnp3_class_zero_config_t.analog_output_status : @p true
/// - @ref dnp3_class_zero_config_t.octet_string : @p false
/// 
/// @returns New instance of @ref dnp3_class_zero_config_t
static dnp3_class_zero_config_t dnp3_class_zero_config_init()
{
    dnp3_class_zero_config_t _return_value = {
        true,
        true,
        true,
        true,
        true,
        true,
        true,
        false
    };
    return _return_value;
}


typedef struct dnp3_outstation_features_t dnp3_outstation_features_t;

/// @brief Optional outstation features that can be enabled or disabled
typedef struct dnp3_outstation_features_t
{
    /// @brief Respond to the self address
    bool self_address;
    /// @brief Process valid broadcast messages
    bool broadcast;
    /// @brief Respond to enable/disable unsolicited response and produce unsolicited responses
    bool unsolicited;
    /// @brief Outstation will process every request as if it came from the configured master address
    /// 
    /// This feature is a hack that can make configuration of some systems easier/more flexible, but should not be used when unsolicited reporting is also required.
    bool respond_to_any_master;
} dnp3_outstation_features_t;

/// @brief Initialize to default values
/// 
/// @note Values are initialized to:
/// - @ref dnp3_outstation_features_t.self_address : @p false
/// - @ref dnp3_outstation_features_t.broadcast : @p true
/// - @ref dnp3_outstation_features_t.unsolicited : @p true
/// - @ref dnp3_outstation_features_t.respond_to_any_master : @p false
/// 
/// @returns New instance of @ref dnp3_outstation_features_t
static dnp3_outstation_features_t dnp3_outstation_features_init()
{
    dnp3_outstation_features_t _return_value = {
        false,
        true,
        true,
        false
    };
    return _return_value;
}


typedef struct dnp3_outstation_config_t dnp3_outstation_config_t;

/// @brief Outstation configuration
typedef struct dnp3_outstation_config_t
{
    /// @brief Link-layer outstation address
    uint16_t outstation_address;
    /// @brief Link-layer master address
    uint16_t master_address;
    /// @brief Event buffer sizes configuration
    dnp3_event_buffer_config_t event_buffer_config;
    /// @brief Solicited response buffer size
    /// 
    /// Must be at least 249 bytes
    uint16_t solicited_buffer_size;
    /// @brief Unsolicited response buffer size
    /// 
    /// Must be at least 249 bytes
    uint16_t unsolicited_buffer_size;
    /// @brief Receive buffer size
    /// 
    /// Must be at least 249 bytes
    uint16_t rx_buffer_size;
    /// @brief Decoding level
    dnp3_decode_level_t decode_level;
    /// @brief Confirmation timeout
    /// @note The unit is milliseconds
    uint64_t confirm_timeout;
    /// @brief Select timeout
    /// @note The unit is milliseconds
    uint64_t select_timeout;
    /// @brief Optional features
    dnp3_outstation_features_t features;
    /// @brief Maximum number of unsolicited retries
    uint32_t max_unsolicited_retries;
    /// @brief Delay to wait before retrying an unsolicited response
    /// @note The unit is milliseconds
    uint64_t unsolicited_retry_delay;
    /// @brief Delay of inactivity before sending a REQUEST_LINK_STATUS to the master
    /// 
    /// A value of zero means no automatic keep-alive will be sent.
    /// @note The unit is milliseconds
    uint64_t keep_alive_timeout;
    /// @brief Maximum number of headers that will be processed in a READ request.
    /// 
    /// Internally, this controls the size of a pre-allocated buffer used to process requests. A minimum value of `DEFAULT_READ_REQUEST_HEADERS` is always enforced. Requesting more than this number will result in the PARAMETER_ERROR IIN bit being set in the response.
    uint16_t max_read_request_headers;
    /// @brief Maximum number of controls in a single request.
    uint16_t max_controls_per_request;
    /// @brief Controls responses to Class 0 reads
    dnp3_class_zero_config_t class_zero;
} dnp3_outstation_config_t;

/// @brief Initialize to defaults
/// 
/// @note Values are initialized to:
/// - @ref dnp3_outstation_config_t.solicited_buffer_size : 2048
/// - @ref dnp3_outstation_config_t.unsolicited_buffer_size : 2048
/// - @ref dnp3_outstation_config_t.rx_buffer_size : 2048
/// - @ref dnp3_outstation_config_t.decode_level : Default @ref dnp3_decode_level_t
/// - @ref dnp3_outstation_config_t.confirm_timeout : 5000ms
/// - @ref dnp3_outstation_config_t.select_timeout : 5000ms
/// - @ref dnp3_outstation_config_t.features : Default @ref dnp3_outstation_features_t
/// - @ref dnp3_outstation_config_t.max_unsolicited_retries : 4294967295
/// - @ref dnp3_outstation_config_t.unsolicited_retry_delay : 5000ms
/// - @ref dnp3_outstation_config_t.keep_alive_timeout : 60000ms
/// - @ref dnp3_outstation_config_t.max_read_request_headers : 64
/// - @ref dnp3_outstation_config_t.max_controls_per_request : 65535
/// - @ref dnp3_outstation_config_t.class_zero : Default @ref dnp3_class_zero_config_t
/// 
/// @param outstation_address Link-layer outstation address
/// @param master_address Link-layer master address
/// @param event_buffer_config Event buffer sizes configuration
/// @returns New instance of @ref dnp3_outstation_config_t
static dnp3_outstation_config_t dnp3_outstation_config_init(uint16_t outstation_address, uint16_t master_address, dnp3_event_buffer_config_t event_buffer_config)
{
    dnp3_outstation_config_t _return_value = {
        outstation_address,
        master_address,
        event_buffer_config,
        2048,
        2048,
        2048,
        dnp3_decode_level_init(),
        5000,
        5000,
        dnp3_outstation_features_init(),
        4294967295,
        5000,
        60000,
        64,
        65535,
        dnp3_class_zero_config_init()
    };
    return _return_value;
}


/// @brief Type of restart delay value. Used by @ref dnp3_restart_delay_t.
typedef enum dnp3_restart_delay_type_t
{
    /// @brief Restart mode not supported
    DNP3_RESTART_DELAY_TYPE_NOT_SUPPORTED = 0,
    /// @brief Value is in seconds (corresponds to g51v1)
    DNP3_RESTART_DELAY_TYPE_SECONDS = 1,
    /// @brief Value is in milliseconds (corresponds to g51v2)
    DNP3_RESTART_DELAY_TYPE_MILLI_SECONDS = 2,
} dnp3_restart_delay_type_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_restart_delay_type_to_string(dnp3_restart_delay_type_t value)
{
    switch (value)
    {
        case DNP3_RESTART_DELAY_TYPE_NOT_SUPPORTED: return "not_supported";
        case DNP3_RESTART_DELAY_TYPE_SECONDS: return "seconds";
        case DNP3_RESTART_DELAY_TYPE_MILLI_SECONDS: return "milli_seconds";
        default: return "unknown restart_delay_type value";
    }
}

typedef struct dnp3_restart_delay_t dnp3_restart_delay_t;

/// @brief Restart delay used by @ref dnp3_outstation_application_t.cold_restart and @ref dnp3_outstation_application_t.warm_restart
/// 
/// If @ref dnp3_restart_delay_t.restart_type is not @ref DNP3_RESTART_DELAY_TYPE_NOT_SUPPORTED, then the @ref dnp3_restart_delay_t.value is valid. Otherwise, the outstation will return IIN2.0 NO_FUNC_CODE_SUPPORT.
typedef struct dnp3_restart_delay_t
{
    /// @brief Indicates what @ref dnp3_restart_delay_t.value is.
    dnp3_restart_delay_type_t restart_type;
    /// @brief Expected delay before the outstation comes back online.
    uint16_t value;
} dnp3_restart_delay_t;

/// @brief RestartDelay indicating that the request is not supported
/// 
/// @note Values are initialized to:
/// - @ref dnp3_restart_delay_t.restart_type : @ref DNP3_RESTART_DELAY_TYPE_NOT_SUPPORTED
/// - @ref dnp3_restart_delay_t.value : 0
/// 
/// @returns New instance of @ref dnp3_restart_delay_t
static dnp3_restart_delay_t dnp3_restart_delay_not_supported()
{
    dnp3_restart_delay_t _return_value = {
        DNP3_RESTART_DELAY_TYPE_NOT_SUPPORTED,
        0
    };
    return _return_value;
}

/// @brief RestartDelay with a count of seconds
/// 
/// @note Values are initialized to:
/// - @ref dnp3_restart_delay_t.restart_type : @ref DNP3_RESTART_DELAY_TYPE_SECONDS
/// 
/// @param value Expected delay before the outstation comes back online.
/// @returns New instance of @ref dnp3_restart_delay_t
static dnp3_restart_delay_t dnp3_restart_delay_seconds(uint16_t value)
{
    dnp3_restart_delay_t _return_value = {
        DNP3_RESTART_DELAY_TYPE_SECONDS,
        value
    };
    return _return_value;
}

/// @brief RestartDelay with a count of milliseconds
/// 
/// @note Values are initialized to:
/// - @ref dnp3_restart_delay_t.restart_type : @ref DNP3_RESTART_DELAY_TYPE_MILLI_SECONDS
/// 
/// @param value Expected delay before the outstation comes back online.
/// @returns New instance of @ref dnp3_restart_delay_t
static dnp3_restart_delay_t dnp3_restart_delay_milliseconds(uint16_t value)
{
    dnp3_restart_delay_t _return_value = {
        DNP3_RESTART_DELAY_TYPE_MILLI_SECONDS,
        value
    };
    return _return_value;
}


/// @brief Write time result used by @ref dnp3_outstation_application_t.write_absolute_time
typedef enum dnp3_write_time_result_t
{
    /// @brief The write time operation succeeded.
    DNP3_WRITE_TIME_RESULT_OK = 0,
    /// @brief The request parameters are nonsensical.
    DNP3_WRITE_TIME_RESULT_PARAMETER_ERROR = 1,
    /// @brief Writing time is not supported by this outstation (translated to NO_FUNC_CODE_SUPPORT).
    DNP3_WRITE_TIME_RESULT_NOT_SUPPORTED = 2,
} dnp3_write_time_result_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_write_time_result_to_string(dnp3_write_time_result_t value)
{
    switch (value)
    {
        case DNP3_WRITE_TIME_RESULT_OK: return "ok";
        case DNP3_WRITE_TIME_RESULT_PARAMETER_ERROR: return "parameter_error";
        case DNP3_WRITE_TIME_RESULT_NOT_SUPPORTED: return "not_supported";
        default: return "unknown write_time_result value";
    }
}

/// @brief Freeze operation type
typedef enum dnp3_freeze_type_t
{
    /// @brief Copy the current value of a counter to the associated point
    DNP3_FREEZE_TYPE_IMMEDIATE_FREEZE = 0,
    /// @brief Copy the current value of a counter to the associated point and clear the current value to 0.
    DNP3_FREEZE_TYPE_FREEZE_AND_CLEAR = 1,
} dnp3_freeze_type_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_freeze_type_to_string(dnp3_freeze_type_t value)
{
    switch (value)
    {
        case DNP3_FREEZE_TYPE_IMMEDIATE_FREEZE: return "immediate_freeze";
        case DNP3_FREEZE_TYPE_FREEZE_AND_CLEAR: return "freeze_and_clear";
        default: return "unknown freeze_type value";
    }
}

/// @brief Result of a freeze operation
typedef enum dnp3_freeze_result_t
{
    /// @brief Freeze operation was successful.
    DNP3_FREEZE_RESULT_OK = 0,
    /// @brief The request parameters are nonsensical.
    DNP3_FREEZE_RESULT_PARAMETER_ERROR = 1,
    /// @brief The demanded freeze operation is not supported by this device.
    DNP3_FREEZE_RESULT_NOT_SUPPORTED = 2,
} dnp3_freeze_result_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_freeze_result_to_string(dnp3_freeze_result_t value)
{
    switch (value)
    {
        case DNP3_FREEZE_RESULT_OK: return "ok";
        case DNP3_FREEZE_RESULT_PARAMETER_ERROR: return "parameter_error";
        case DNP3_FREEZE_RESULT_NOT_SUPPORTED: return "not_supported";
        default: return "unknown freeze_result value";
    }
}

typedef struct dnp3_application_iin_t dnp3_application_iin_t;

/// @brief Application-controlled IIN bits
typedef struct dnp3_application_iin_t
{
    /// @brief IIN1.4 - Time synchronization is required
    bool need_time;
    /// @brief IIN1.5 - Some output points are in local mode
    bool local_control;
    /// @brief IIN1.6 - Device trouble
    bool device_trouble;
    /// @brief IIN2.5 - Configuration corrupt
    bool config_corrupt;
} dnp3_application_iin_t;

/// @brief Initialize all fields in @ref dnp3_application_iin_t to false
/// 
/// @note Values are initialized to:
/// - @ref dnp3_application_iin_t.need_time : @p false
/// - @ref dnp3_application_iin_t.local_control : @p false
/// - @ref dnp3_application_iin_t.device_trouble : @p false
/// - @ref dnp3_application_iin_t.config_corrupt : @p false
/// 
/// @returns New instance of @ref dnp3_application_iin_t
static dnp3_application_iin_t dnp3_application_iin_init()
{
    dnp3_application_iin_t _return_value = {
        false,
        false,
        false,
        false
    };
    return _return_value;
}


typedef struct dnp3_class_count_t dnp3_class_count_t;

/// @brief Remaining number of events in the buffer after a confirm on a per-class basis
typedef struct dnp3_class_count_t
{
    /// @brief Number of class 1 events remaining in the buffer
    uint32_t num_class_1;
    /// @brief Number of class 2 events remaining in the buffer
    uint32_t num_class_2;
    /// @brief Number of class 3 events remaining in the buffer
    uint32_t num_class_3;
} dnp3_class_count_t;


typedef struct dnp3_type_count_t dnp3_type_count_t;

/// @brief Remaining number of events in the buffer after a confirm on a per-type basis
typedef struct dnp3_type_count_t
{
    /// @brief Number of binary input events remaining in the buffer
    uint32_t num_binary_input;
    /// @brief Number of double-bit binary input events remaining in the buffer
    uint32_t num_double_bit_binary_input;
    /// @brief Number of binary output status events remaining in the buffer
    uint32_t num_binary_output_status;
    /// @brief Number of counter events remaining in the buffer
    uint32_t num_counter;
    /// @brief Number of frozen counter events remaining in the buffer
    uint32_t num_frozen_counter;
    /// @brief Number of analog events remaining in the buffer
    uint32_t num_analog;
    /// @brief Number of analog output status events remaining in the buffer
    uint32_t num_analog_output_status;
    /// @brief Number octet string events remaining in the buffer
    uint32_t num_octet_string;
} dnp3_type_count_t;


typedef struct dnp3_buffer_state_t dnp3_buffer_state_t;

/// @brief Information about the state of buffer after a CONFIRM has been processed
typedef struct dnp3_buffer_state_t
{
    /// @brief Remaining number of events in the buffer on a per-class basis
    dnp3_class_count_t classes;
    /// @brief Remaining number of events in the buffer on a per-type basis
    dnp3_type_count_t types;
} dnp3_buffer_state_t;


/// @brief Dynamic information required by the outstation from the user application
typedef struct dnp3_outstation_application_t
{
    
    /// @brief Returns the DELAY_MEASUREMENT delay
    /// 
    /// The value returned by this method is used in conjunction with the DELAY_MEASUREMENT function code and returned in a g52v2 time delay object as part of a non-LAN time synchronization procedure.
    /// 
    /// It represents the processing delay from receiving the request to sending the response. This parameter should almost always use the default value of zero as only an RTOS or bare metal system would have access to this level of timing. Modern hardware can almost always respond in less than 1 millisecond anyway.
    /// 
    /// For more information, see IEEE-1815 2012, p. 64.
    /// @param ctx Context data
    /// @return Processing delay, in milliseconds
    uint16_t (*get_processing_delay_ms)(void*);
    
    /// @brief Handle a write of the absolute time during time synchronization procedures.
    /// @param time Received time in milliseconds since EPOCH (only 48 bits are used)
    /// @param ctx Context data
    /// @return Result of the write time operation
    dnp3_write_time_result_t (*write_absolute_time)(uint64_t, void*);
    
    /// @brief Returns the application-controlled IIN bits
    /// @param ctx Context data
    /// @return Application IIN bits
    dnp3_application_iin_t (*get_application_iin)(void*);
    
    /// @brief Request that the outstation perform a cold restart (IEEE-1815 2012, p. 58)
    /// 
    /// The outstation will not automatically restart. It is the responsibility of the user application to handle this request and take the appropriate action.
    /// @param ctx Context data
    /// @return The restart delay
    dnp3_restart_delay_t (*cold_restart)(void*);
    
    /// @brief Request that the outstation perform a warm restart (IEEE-1815 2012, p. 58)
    /// 
    /// The outstation will not automatically restart. It is the responsibility of the user application to handle this request and take the appropriate action.
    /// @param ctx Context data
    /// @return The restart delay
    dnp3_restart_delay_t (*warm_restart)(void*);
    
    /// @brief Freeze all the counters
    /// @param freeze_type Type of freeze operation
    /// @param database_handle Database handle
    /// @param ctx Context data
    /// @return Result of the freeze operation
    dnp3_freeze_result_t (*freeze_counters_all)(dnp3_freeze_type_t, dnp3_database_handle_t*, void*);
    
    /// @brief Freeze all the counters at a requested time and interval
    /// 
    /// Refer to the table on page 57 of IEEE 1815-2012 to interpret the time and interval parameters correctly
    /// @param database_handle Database handle
    /// @param time 48-bit DNP3 timestamp in milliseconds since epoch UTC
    /// @param interval Count of milliseconds representing the interval between freezes relative to the timestamp
    /// @param ctx Context data
    /// @return Result of the freeze operation
    dnp3_freeze_result_t (*freeze_counters_all_at_time)(dnp3_database_handle_t*, uint64_t, uint32_t, void*);
    
    /// @brief Freeze a range of counters
    /// @param start Start index to freeze (inclusive)
    /// @param stop Stop index to freeze (inclusive)
    /// @param freeze_type Type of freeze operation
    /// @param database_handle Database handle
    /// @param ctx Context data
    /// @return Result of the freeze operation
    dnp3_freeze_result_t (*freeze_counters_range)(uint16_t, uint16_t, dnp3_freeze_type_t, dnp3_database_handle_t*, void*);
    
    /// @brief Freeze a range of counters at a requested time and interval
    /// 
    /// Refer to the table on page 57 of IEEE 1815-2012 to interpret the time and interval parameters correctly
    /// @param start Start index to freeze (inclusive)
    /// @param stop Stop index to freeze (inclusive)
    /// @param database_handle Database handle
    /// @param time 48-bit DNP3 timestamp in milliseconds since epoch UTC
    /// @param interval Count of milliseconds representing the interval between freezes relative to the timestamp
    /// @param ctx Context data
    /// @return Result of the freeze operation
    dnp3_freeze_result_t (*freeze_counters_range_at_time)(uint16_t, uint16_t, dnp3_database_handle_t*, uint64_t, uint32_t, void*);
    
    /// @brief Controls outstation support for writing group 34, analog input dead-bands
    /// 
    /// Returning false, indicates that the writes to group34 should not be processed and requests to do so should be rejected with IIN2.NO_FUNC_CODE_SUPPORT
    /// 
    /// Returning true will allow the request to process the actual values with a sequence of calls:
    /// 
    /// 1) A single call to @ref dnp3_outstation_application_t.begin_write_analog_dead_bands
    /// 
    /// 2) Zero or more calls to @ref dnp3_outstation_application_t.write_analog_dead_band
    /// 
    /// 3) A single call to @ref dnp3_outstation_application_t.end_write_analog_dead_bands
    /// @param ctx Context data
    /// @return True if the outstation should process the request
    bool (*support_write_analog_dead_bands)(void*);
    
    /// @brief Called when the outstation begins processing a header to write analog dead-bands
    /// @param ctx Context data
    void (*begin_write_analog_dead_bands)(void*);
    
    /// @brief Called when the outstation begins processing a header to write analog dead-bands
    /// 
    /// Called for each analog dead-band in the write request where an analog input is defined at the specified index.
    /// 
    /// The dead-band is automatically updated in the database. This callback allows application code to persist the modified value to non-volatile memory if desired
    /// @param index Index of the analog input
    /// @param dead_band New dead-band value
    /// @param ctx Context data
    void (*write_analog_dead_band)(uint16_t, double, void*);
    
    /// @brief Called when the outstation completes processing a header to write analog dead-bands
    /// 
    /// Multiple dead-bands changes can be accumulated in calls to @ref dnp3_outstation_application_t.write_analog_dead_band and then be processed as a batch in this method.
    /// @param ctx Context data
    void (*end_write_analog_dead_bands)(void*);
    
    /// @brief Write a string attribute. This method is only called if the corresponding attribute has been configured as writable
    /// @param set Set to which the attribute belongs
    /// @param variation Variation of the attribute
    /// @param attr_type Enumeration describing which attribute it is, possibly unknown
    /// @param value Value of the attribute
    /// @param ctx Context data
    /// @return If true, the value will be modified in the in memory database and the outstation will return a successful response. If false, no change will be made and the outstation will return PARAM_ERROR
    bool (*write_string_attr)(uint8_t, uint8_t, dnp3_string_attr_t, const char*, void*);
    
    /// @brief Write a 32-bit floating point attribute. This method is only called if the corresponding attribute has been configured as writable
    /// @param set Set to which the attribute belongs
    /// @param variation Variation of the attribute
    /// @param attr_type Enumeration describing which attribute it is, possibly unknown
    /// @param value Value of the attribute
    /// @param ctx Context data
    /// @return If true, the value will be modified in the in memory database and the outstation will return a successful response. If false, no change will be made and the outstation will return PARAM_ERROR
    bool (*write_float_attr)(uint8_t, uint8_t, dnp3_float_attr_t, float, void*);
    
    /// @brief Write a 64-bit floating point attribute. This method is only called if the corresponding attribute has been configured as writable
    /// @param set Set to which the attribute belongs
    /// @param variation Variation of the attribute
    /// @param attr_type Enumeration describing which attribute it is, possibly unknown
    /// @param value Value of the attribute
    /// @param ctx Context data
    /// @return If true, the value will be modified in the in memory database and the outstation will return a successful response. If false, no change will be made and the outstation will return PARAM_ERROR
    bool (*write_double_attr)(uint8_t, uint8_t, dnp3_float_attr_t, double, void*);
    
    /// @brief Write an unsigned integer attribute. This method is only called if the corresponding attribute has been configured as writable
    /// @param set Set to which the attribute belongs
    /// @param variation Variation of the attribute
    /// @param attr_type Enumeration describing which attribute it is, possibly unknown
    /// @param value Value of the attribute
    /// @param ctx Context data
    /// @return If true, the value will be modified in the in memory database and the outstation will return a successful response. If false, no change will be made and the outstation will return PARAM_ERROR
    bool (*write_uint_attr)(uint8_t, uint8_t, dnp3_uint_attr_t, uint32_t, void*);
    
    /// @brief Write a signed integer attribute. This method is only called if the corresponding attribute has been configured as writable
    /// @param set Set to which the attribute belongs
    /// @param variation Variation of the attribute
    /// @param attr_type Enumeration describing which attribute it is, possibly unknown
    /// @param value Value of the attribute
    /// @param ctx Context data
    /// @return If true, the value will be modified in the in memory database and the outstation will return a successful response. If false, no change will be made and the outstation will return PARAM_ERROR
    bool (*write_int_attr)(uint8_t, uint8_t, dnp3_int_attr_t, int32_t, void*);
    
    /// @brief Write an octet-string attribute. This method is only called if the corresponding attribute has been configured as writable
    /// @param set Set to which the attribute belongs
    /// @param variation Variation of the attribute
    /// @param attr_type Enumeration describing which attribute it is, possibly unknown
    /// @param value Iterator over bytes of the value
    /// @param ctx Context data
    /// @return If true, the value will be modified in the in memory database and the outstation will return a successful response. If false, no change will be made and the outstation will return PARAM_ERROR
    bool (*write_octet_string_attr)(uint8_t, uint8_t, dnp3_octet_string_attr_t, dnp3_byte_iterator_t*, void*);
    
    /// @brief Write a bit-string attribute. This method is only called if the corresponding attribute has been configured as writable
    /// @param set Set to which the attribute belongs
    /// @param variation Variation of the attribute
    /// @param attr_type Enumeration describing which attribute it is, possibly unknown
    /// @param value Iterator over bytes of the value
    /// @param ctx Context data
    /// @return If true, the value will be modified in the in memory database and the outstation will return a successful response. If false, no change will be made and the outstation will return PARAM_ERROR
    bool (*write_bit_string_attr)(uint8_t, uint8_t, dnp3_bit_string_attr_t, dnp3_byte_iterator_t*, void*);
    
    /// @brief Write a DNP3 time attribute. This method is only called if the corresponding attribute has been configured as writable.
    /// @param set Set to which the attribute belongs
    /// @param variation Variation of the attribute
    /// @param attr_type Enumeration describing which attribute it is, possibly unknown
    /// @param value 48-bit DNP3 timestamp value
    /// @param ctx Context data
    /// @return If true, the value will be modified in the in memory database and the outstation will return a successful response. If false, no change will be made and the outstation will return PARAM_ERROR
    bool (*write_time_attr)(uint8_t, uint8_t, dnp3_time_attr_t, uint64_t, void*);
    
    /// @brief Called when a CONFIRM is received to a response or unsolicited response, but before any previously transmitted events are cleared from the buffer
    /// @param ctx Context data
    void (*begin_confirm)(void*);
    
    /// @brief Called when an event is cleared from the buffer due to master acknowledgement
    /// @param id Unique identifier previously assigned to the event by the database in an update method
    /// @param ctx Context data
    void (*event_cleared)(uint64_t, void*);
    
    /// @brief  Called when all relevant events have been cleared
    /// @param state information about the post-CONFIRM state of the buffer
    /// @param ctx Context data
    void (*end_confirm)(dnp3_buffer_state_t, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_outstation_application_t;

typedef struct dnp3_request_header_t dnp3_request_header_t;

/// @brief Application-layer header for requests
typedef struct dnp3_request_header_t
{
    /// @brief Control field
    dnp3_control_field_t control_field;
    /// @brief Function code
    dnp3_function_code_t function;
} dnp3_request_header_t;


/// @brief Enumeration describing how the outstation processed a broadcast request
typedef enum dnp3_broadcast_action_t
{
    /// @brief Outstation processed the broadcast
    DNP3_BROADCAST_ACTION_PROCESSED = 0,
    /// @brief Outstation ignored the broadcast message b/c it is disabled by configuration
    DNP3_BROADCAST_ACTION_IGNORED_BY_CONFIGURATION = 1,
    /// @brief Outstation was unable to parse the object headers and ignored the request
    DNP3_BROADCAST_ACTION_BAD_OBJECT_HEADERS = 2,
    /// @brief Outstation ignore the broadcast message b/c the function is not supported via Broadcast
    DNP3_BROADCAST_ACTION_UNSUPPORTED_FUNCTION = 3,
} dnp3_broadcast_action_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_broadcast_action_to_string(dnp3_broadcast_action_t value)
{
    switch (value)
    {
        case DNP3_BROADCAST_ACTION_PROCESSED: return "processed";
        case DNP3_BROADCAST_ACTION_IGNORED_BY_CONFIGURATION: return "ignored_by_configuration";
        case DNP3_BROADCAST_ACTION_BAD_OBJECT_HEADERS: return "bad_object_headers";
        case DNP3_BROADCAST_ACTION_UNSUPPORTED_FUNCTION: return "unsupported_function";
        default: return "unknown broadcast_action value";
    }
}

/// @brief Informational callbacks that the outstation doesn't rely on to function
/// 
/// It may be useful to certain applications to assess the health of the communication or to count statistics
typedef struct dnp3_outstation_information_t
{
    
    /// @brief Called when a request is processed from the IDLE state
    /// @param header Request header
    /// @param ctx Context data
    void (*process_request_from_idle)(dnp3_request_header_t, void*);
    
    /// @brief Called when a broadcast request is received by the outstation
    /// @param function_code Function code received
    /// @param action Broadcast action
    /// @param ctx Context data
    void (*broadcast_received)(dnp3_function_code_t, dnp3_broadcast_action_t, void*);
    
    /// @brief Outstation has begun waiting for a solicited confirm
    /// @param ecsn Expected sequence number
    /// @param ctx Context data
    void (*enter_solicited_confirm_wait)(uint8_t, void*);
    
    /// @brief Failed to receive a solicited confirm before the timeout occurred
    /// @param ecsn Expected sequence number
    /// @param ctx Context data
    void (*solicited_confirm_timeout)(uint8_t, void*);
    
    /// @brief Received the expected confirm
    /// @param ecsn Expected sequence number
    /// @param ctx Context data
    void (*solicited_confirm_received)(uint8_t, void*);
    
    /// @brief Received a new request while waiting for a solicited confirm, aborting the response series
    /// @param ctx Context data
    void (*solicited_confirm_wait_new_request)(void*);
    
    /// @brief Received a solicited confirm with the wrong sequence number
    /// @param ecsn Expected sequence number
    /// @param seq Received sequence number
    /// @param ctx Context data
    void (*wrong_solicited_confirm_seq)(uint8_t, uint8_t, void*);
    
    /// @brief Received a confirm when not expecting one
    /// @param unsolicited True if it's an unsolicited response confirm, false if it's a solicited response confirm
    /// @param seq Received sequence number
    /// @param ctx Context data
    void (*unexpected_confirm)(bool, uint8_t, void*);
    
    /// @brief Outstation has begun waiting for an unsolicited confirm
    /// @param ecsn Expected sequence number
    /// @param ctx Context data
    void (*enter_unsolicited_confirm_wait)(uint8_t, void*);
    
    /// @brief Failed to receive an unsolicited confirm before the timeout occurred
    /// @param ecsn Expected sequence number
    /// @param retry Is it a retry
    /// @param ctx Context data
    void (*unsolicited_confirm_timeout)(uint8_t, bool, void*);
    
    /// @brief Master confirmed an unsolicited message
    /// @param ecsn Expected sequence number
    /// @param ctx Context data
    void (*unsolicited_confirmed)(uint8_t, void*);
    
    /// @brief Master cleared the restart IIN bit
    /// @param ctx Context data
    void (*clear_restart_iin)(void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_outstation_information_t;

/// @brief Enumeration describing how the master requested the control operation
typedef enum dnp3_operate_type_t
{
    /// @brief control point was properly selected before the operate request
    DNP3_OPERATE_TYPE_SELECT_BEFORE_OPERATE = 0,
    /// @brief operate the control via a DirectOperate request
    DNP3_OPERATE_TYPE_DIRECT_OPERATE = 1,
    /// @brief operate the control via a DirectOperateNoAck request
    DNP3_OPERATE_TYPE_DIRECT_OPERATE_NO_ACK = 2,
} dnp3_operate_type_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_operate_type_to_string(dnp3_operate_type_t value)
{
    switch (value)
    {
        case DNP3_OPERATE_TYPE_SELECT_BEFORE_OPERATE: return "select_before_operate";
        case DNP3_OPERATE_TYPE_DIRECT_OPERATE: return "direct_operate";
        case DNP3_OPERATE_TYPE_DIRECT_OPERATE_NO_ACK: return "direct_operate_no_ack";
        default: return "unknown operate_type value";
    }
}

/// @brief Callbacks for handling controls
typedef struct dnp3_control_handler_t
{
    
    /// @brief Notifies the start of a command fragment
    /// @param ctx Context data
    void (*begin_fragment)(void*);
    
    /// @brief Notifies the end of a command fragment
    /// @param database Database handle
    /// @param ctx Context data
    void (*end_fragment)(dnp3_database_handle_t*, void*);
    
    /// @brief Select a CROB, but do not operate
    /// 
    /// Implementors can think of this function as asking the question "is this control supported"?
    /// 
    /// Most implementations should not alter the database in this method. It is only provided in the event that some event counters reflected via the API get updated on SELECT, but this would be highly abnormal.
    /// @param value Received CROB
    /// @param index Index of the point
    /// @param database_handle Database handle
    /// @param ctx Context data
    /// @return Command status
    dnp3_command_status_t (*select_g12v1)(dnp3_group12_var1_t, uint16_t, dnp3_database_handle_t*, void*);
    
    /// @brief Operate a control point
    /// @param value Received CROB
    /// @param index Index of the point
    /// @param op_type Operate type
    /// @param database_handle Database handle
    /// @param ctx Context data
    /// @return Command status
    dnp3_command_status_t (*operate_g12v1)(dnp3_group12_var1_t, uint16_t, dnp3_operate_type_t, dnp3_database_handle_t*, void*);
    
    /// @brief Select an analog output, but do not operate
    /// 
    /// Implementors can think of this function as asking the question "is this control supported"?
    /// 
    /// Most implementations should not alter the database in this method. It is only provided in the event that some event counters reflected via the API get updated on SELECT, but this would be highly abnormal.
    /// @param value Received analog output value
    /// @param index Index of the point
    /// @param database_handle Database handle
    /// @param ctx Context data
    /// @return Command status
    dnp3_command_status_t (*select_g41v1)(int32_t, uint16_t, dnp3_database_handle_t*, void*);
    
    /// @brief Operate a control point
    /// @param value Received analog output value
    /// @param index Index of the point
    /// @param op_type Operate type
    /// @param database_handle Database handle
    /// @param ctx Context data
    /// @return Command status
    dnp3_command_status_t (*operate_g41v1)(int32_t, uint16_t, dnp3_operate_type_t, dnp3_database_handle_t*, void*);
    
    /// @brief Select an analog output, but do not operate
    /// 
    /// Implementors can think of this function as asking the question "is this control supported"?
    /// 
    /// Most implementations should not alter the database in this method. It is only provided in the event that some event counters reflected via the API get updated on SELECT, but this would be highly abnormal.
    /// @param value Received analog output value
    /// @param index Index of the point
    /// @param database_handle Database handle
    /// @param ctx Context data
    /// @return Command status
    dnp3_command_status_t (*select_g41v2)(int16_t, uint16_t, dnp3_database_handle_t*, void*);
    
    /// @brief Operate a control point
    /// @param value Received analog output value
    /// @param index Index of the point
    /// @param op_type Operate type
    /// @param database_handle Database handle
    /// @param ctx Context data
    /// @return Command status
    dnp3_command_status_t (*operate_g41v2)(int16_t, uint16_t, dnp3_operate_type_t, dnp3_database_handle_t*, void*);
    
    /// @brief Select an analog output, but do not operate
    /// 
    /// Implementors can think of this function as asking the question "is this control supported"?
    /// 
    /// Most implementations should not alter the database in this method. It is only provided in the event that some event counters reflected via the API get updated on SELECT, but this would be highly abnormal.
    /// @param value Received analog output value
    /// @param index Index of the point
    /// @param database_handle Database handle
    /// @param ctx Context data
    /// @return Command status
    dnp3_command_status_t (*select_g41v3)(float, uint16_t, dnp3_database_handle_t*, void*);
    
    /// @brief Operate a control point
    /// @param value Received analog output value
    /// @param index Index of the point
    /// @param op_type Operate type
    /// @param database_handle Database handle
    /// @param ctx Context data
    /// @return Command status
    dnp3_command_status_t (*operate_g41v3)(float, uint16_t, dnp3_operate_type_t, dnp3_database_handle_t*, void*);
    
    /// @brief Select an analog output, but do not operate
    /// 
    /// Implementors can think of this function as asking the question "is this control supported"?
    /// 
    /// Most implementations should not alter the database in this method. It is only provided in the event that some event counters reflected via the API get updated on SELECT, but this would be highly abnormal.
    /// @param value Received analog output value
    /// @param index Index of the point
    /// @param database_handle Database handle
    /// @param ctx Context data
    /// @return Command status
    dnp3_command_status_t (*select_g41v4)(double, uint16_t, dnp3_database_handle_t*, void*);
    
    /// @brief Operate a control point
    /// @param value Received analog output value
    /// @param index Index of the point
    /// @param op_type Operate type
    /// @param database_handle Database handle
    /// @param ctx Context data
    /// @return Command status
    dnp3_command_status_t (*operate_g41v4)(double, uint16_t, dnp3_operate_type_t, dnp3_database_handle_t*, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_control_handler_t;

/// @brief Outstation connection state for connection-oriented transports, e.g. TCP
typedef enum dnp3_connection_state_t
{
    /// @brief Connected to the master
    DNP3_CONNECTION_STATE_CONNECTED = 0,
    /// @brief Disconnected from the master
    DNP3_CONNECTION_STATE_DISCONNECTED = 1,
} dnp3_connection_state_t;

/// @brief Converts the enum to a string
/// @param value Enum to convert
/// @returns String representation
static const char* dnp3_connection_state_to_string(dnp3_connection_state_t value)
{
    switch (value)
    {
        case DNP3_CONNECTION_STATE_CONNECTED: return "connected";
        case DNP3_CONNECTION_STATE_DISCONNECTED: return "disconnected";
        default: return "unknown connection_state value";
    }
}

/// @brief Callback interface for connection state events
typedef struct dnp3_connection_state_listener_t
{
    
    /// @brief Called when the connection state changes
    /// @param state New state of the connection
    /// @param ctx Context data
    void (*on_change)(dnp3_connection_state_t, void*);
    /// @brief Callback when the underlying owner doesn't need the interface anymore
    /// @param arg Context data
    void (*on_destroy)(void* arg);
    /// @brief Context data
    void* ctx;
} dnp3_connection_state_listener_t;

/// @brief Outstation handle
/// 
/// Use this handle to modify the internal database.
typedef struct dnp3_outstation_t dnp3_outstation_t;

/// @brief Create an outstation instance running as a TCP client
/// @param runtime runtime on which to spawn the outstation
/// @param link_error_mode Controls how link errors are handled with respect to the TCP session
/// @param endpoints List of IP endpoints.
/// @param connect_strategy Controls the timing of (re)connection attempts
/// @param connect_options Options that control the TCP connection process
/// @param config outstation configuration
/// @param application application interface
/// @param information informational events interface
/// @param control_handler control handler interface
/// @param listener Connection listener used to receive updates on the status of the connection
/// @param out Outstation instance
/// @return Error code
dnp3_param_error_t dnp3_outstation_create_tcp_client(dnp3_runtime_t* runtime, dnp3_link_error_mode_t link_error_mode, dnp3_endpoint_list_t* endpoints, dnp3_connect_strategy_t connect_strategy, dnp3_connect_options_t* connect_options, dnp3_outstation_config_t config, dnp3_outstation_application_t application, dnp3_outstation_information_t information, dnp3_control_handler_t control_handler, dnp3_client_state_listener_t listener, dnp3_outstation_t** out);

/// @brief Create an outstation instance running as a TLS client
/// @param runtime runtime on which to spawn the outstation
/// @param link_error_mode Controls how link errors are handled with respect to the TCP session
/// @param endpoints List of IP endpoints.
/// @param connect_strategy Controls the timing of (re)connection attempts
/// @param connect_options Options that control the TCP connection process
/// @param config outstation configuration
/// @param application application interface
/// @param information informational events interface
/// @param control_handler control handler interface
/// @param listener Connection listener used to receive updates on the status of the connection
/// @param tls_config TLS client configuration
/// @param out Outstation instance
/// @return Error code
dnp3_param_error_t dnp3_outstation_create_tls_client(dnp3_runtime_t* runtime, dnp3_link_error_mode_t link_error_mode, dnp3_endpoint_list_t* endpoints, dnp3_connect_strategy_t connect_strategy, dnp3_connect_options_t* connect_options, dnp3_outstation_config_t config, dnp3_outstation_application_t application, dnp3_outstation_information_t information, dnp3_control_handler_t control_handler, dnp3_client_state_listener_t listener, dnp3_tls_client_config_t tls_config, dnp3_outstation_t** out);

/// @brief Create an outstation instance running on a serial port
/// 
/// The port is opened immediately on the calling thread and fails if not successful
/// 
/// @warning Most users should prefer the fault tolerant version of the this method @ref dnp3_outstation_create_serial_session_fault_tolerant
/// @param runtime runtime on which to spawn the outstation
/// @param serial_path Path of the serial device
/// @param settings settings for the serial port
/// @param config outstation configuration
/// @param application application interface
/// @param information informational events interface
/// @param control_handler control handler interface
/// @param out Outstation instance or @p NULL if the port cannot be opened
/// @return Error code
dnp3_param_error_t dnp3_outstation_create_serial_session(dnp3_runtime_t* runtime, const char* serial_path, dnp3_serial_settings_t settings, dnp3_outstation_config_t config, dnp3_outstation_application_t application, dnp3_outstation_information_t information, dnp3_control_handler_t control_handler, dnp3_outstation_t** out);

/// @brief This method is implemented in terms of @ref dnp3_outstation_create_serial_session_2 but without a port listener
/// @param runtime runtime on which to spawn the outstation
/// @param serial_path Path of the serial device
/// @param settings settings for the serial port
/// @param open_retry_delay delay between attempts to open the serial port (milliseconds)
/// @param config outstation configuration
/// @param application application interface
/// @param information informational events interface
/// @param control_handler control handler interface
/// @param out Outstation instance or @p NULL if the port cannot be opened
/// @return Error code
dnp3_param_error_t dnp3_outstation_create_serial_session_fault_tolerant(dnp3_runtime_t* runtime, const char* serial_path, dnp3_serial_settings_t settings, uint64_t open_retry_delay, dnp3_outstation_config_t config, dnp3_outstation_application_t application, dnp3_outstation_information_t information, dnp3_control_handler_t control_handler, dnp3_outstation_t** out);

/// @brief Create an outstation instance running on a serial port which is tolerant to the serial port being added and removed
/// @param runtime runtime on which to spawn the outstation
/// @param serial_path Path of the serial device
/// @param settings settings for the serial port
/// @param open_retry_delay delay between attempts to open the serial port (milliseconds)
/// @param config outstation configuration
/// @param application application interface
/// @param information informational events interface
/// @param control_handler control handler interface
/// @param port_listener port state listener
/// @param out Outstation instance or @p NULL if the port cannot be opened
/// @return Error code
dnp3_param_error_t dnp3_outstation_create_serial_session_2(dnp3_runtime_t* runtime, const char* serial_path, dnp3_serial_settings_t settings, uint64_t open_retry_delay, dnp3_outstation_config_t config, dnp3_outstation_application_t application, dnp3_outstation_information_t information, dnp3_control_handler_t control_handler, dnp3_port_state_listener_t port_listener, dnp3_outstation_t** out);

typedef struct dnp3_outstation_udp_config_t dnp3_outstation_udp_config_t;

/// @brief UDP outstation configuration
typedef struct dnp3_outstation_udp_config_t
{
    /// @brief Local endpoint to which the UDP socket is bound. Must be a socket address (ip:port)
    const char* local_endpoint;
    /// @brief Remote endpoint where outbound requests are sent. Must be a socket address (ip:port)
    const char* remote_endpoint;
    /// @brief UDP socket mode to use
    dnp3_udp_socket_mode_t socket_mode;
    /// @brief Read mode to use, this should typically be set to @ref DNP3_LINK_READ_MODE_DATAGRAM
    dnp3_link_read_mode_t link_read_mode;
    /// @brief Period to wait before retrying after a failure to bind or connect the UDP socket
    /// @note The unit is milliseconds
    uint64_t retry_delay;
} dnp3_outstation_udp_config_t;

/// @brief Initialize the configuration with default settings for unspecified parameter
/// 
/// @note Values are initialized to:
/// - @ref dnp3_outstation_udp_config_t.retry_delay : 5000ms
/// - @ref dnp3_outstation_udp_config_t.link_read_mode : @ref DNP3_LINK_READ_MODE_DATAGRAM
/// - @ref dnp3_outstation_udp_config_t.socket_mode : @ref DNP3_UDP_SOCKET_MODE_ONE_TO_ONE
/// 
/// @param local_endpoint Local endpoint to which the UDP socket is bound. Must be a socket address (ip:port)
/// @param remote_endpoint Remote endpoint where outbound requests are sent. Must be a socket address (ip:port)
/// @returns New instance of @ref dnp3_outstation_udp_config_t
static dnp3_outstation_udp_config_t dnp3_outstation_udp_config_init(const char* local_endpoint, const char* remote_endpoint)
{
    dnp3_outstation_udp_config_t _return_value = {
        local_endpoint,
        remote_endpoint,
        DNP3_UDP_SOCKET_MODE_ONE_TO_ONE,
        DNP3_LINK_READ_MODE_DATAGRAM,
        5000
    };
    return _return_value;
}


/// @brief Create an outstation instance running on a serial port which is tolerant to the serial port being added and removed
/// @param runtime runtime on which to spawn the outstation
/// @param udp_config UDP configuration
/// @param config outstation configuration
/// @param application application interface
/// @param information informational events interface
/// @param control_handler control handler interface
/// @param out Outstation instance or @p NULL if the port cannot be opened
/// @return Error code
dnp3_param_error_t dnp3_outstation_create_udp(dnp3_runtime_t* runtime, dnp3_outstation_udp_config_t udp_config, dnp3_outstation_config_t config, dnp3_outstation_application_t application, dnp3_outstation_information_t information, dnp3_control_handler_t control_handler, dnp3_outstation_t** out);

/// @brief Free resources of the outstation.
/// 
/// @warning This does not shutdown the outstation. Only @ref dnp3_outstation_server_destroy will properly shutdown the outstation.
/// @param instance Instance of @ref dnp3_outstation_t to destroy
void dnp3_outstation_destroy(dnp3_outstation_t* instance);

/// @brief Acquire a mutex on the underlying database and apply a set of changes as a transaction
/// @param instance Instance of @ref dnp3_outstation_t
/// @param callback Interface on which to execute the transaction
void dnp3_outstation_transaction(dnp3_outstation_t* instance, dnp3_database_transaction_t callback);

/// @brief Set decoding log level
/// @param instance Instance of @ref dnp3_outstation_t
/// @param level Decode log
/// @return Error code
dnp3_param_error_t dnp3_outstation_set_decode_level(dnp3_outstation_t* instance, dnp3_decode_level_t level);

/// @brief enable communications
/// @param instance Instance of @ref dnp3_outstation_t
/// @return Error code
dnp3_param_error_t dnp3_outstation_enable(dnp3_outstation_t* instance);

/// @brief disable communications
/// @param instance Instance of @ref dnp3_outstation_t
/// @return Error code
dnp3_param_error_t dnp3_outstation_disable(dnp3_outstation_t* instance);


/// @brief Filters connecting client by their IP address to associate a connecting master with an outstation on the server
/// 
/// Address filters must be DISJOINT, i.e. two filters cannot accept the same IP address. The @ref dnp3_outstation_server_add_outstation method will fail if the filter conflicts with a previously added filter.
typedef struct dnp3_address_filter_t dnp3_address_filter_t;

/// @brief Create an address filter that accepts any IP address
/// @return Address filter
dnp3_address_filter_t* dnp3_address_filter_any();

/// @brief Create an address filter that matches a specific address or wildcards
/// 
/// Examples: 192.168.1.26, 192.168.0.*, *.*.*.*
/// 
/// Wildcards are only supported for IPv4 addresses
/// @param address IP address to accept
/// @param out Instance of @ref dnp3_address_filter_t
/// @return Error code
dnp3_param_error_t dnp3_address_filter_create(const char* address, dnp3_address_filter_t** out);

/// @brief Add an accepted IP address to the filter
/// @param instance Instance of @ref dnp3_address_filter_t
/// @param address IP address to add
/// @return Error code
dnp3_param_error_t dnp3_address_filter_add(dnp3_address_filter_t* instance, const char* address);

/// @brief Destroy an address filter
/// @param instance Instance of @ref dnp3_address_filter_t to destroy
void dnp3_address_filter_destroy(dnp3_address_filter_t* instance);


typedef struct dnp3_tls_server_config_t dnp3_tls_server_config_t;

/// @brief TLS server configuration
typedef struct dnp3_tls_server_config_t
{
    /// @brief Subject name which is verified in the presented client certificate, from the SAN extension or in the common name field.
    /// 
    /// @warning This argument is only used when used with @ref DNP3_CERTIFICATE_MODE_AUTHORITY_BASED 
    const char* dns_name;
    /// @brief Path to the PEM-encoded certificate of the peer
    const char* peer_cert_path;
    /// @brief Path to the PEM-encoded local certificate
    const char* local_cert_path;
    /// @brief Path to the PEM-encoded private key
    const char* private_key_path;
    /// @brief Optional password if the private key file is encrypted
    /// 
    /// Only PKCS#8 encrypted files are supported.
    /// 
    /// Pass empty string if the file is not encrypted.
    const char* password;
    /// @brief Minimum TLS version allowed
    dnp3_min_tls_version_t min_tls_version;
    /// @brief Certificate validation mode
    dnp3_certificate_mode_t certificate_mode;
    /// @brief If set to true, a '*' may be used for @ref dnp3_tls_server_config_t.dns_name to allow any authenticated client to connect
    bool allow_client_name_wildcard;
} dnp3_tls_server_config_t;

/// @brief construct the configuration with defaults
/// 
/// @note Values are initialized to:
/// - @ref dnp3_tls_server_config_t.min_tls_version : @ref DNP3_MIN_TLS_VERSION_V12
/// - @ref dnp3_tls_server_config_t.certificate_mode : @ref DNP3_CERTIFICATE_MODE_AUTHORITY_BASED
/// - @ref dnp3_tls_server_config_t.allow_client_name_wildcard : @p false
/// 
/// @param dns_name Subject name which is verified in the presented client certificate, from the SAN extension or in the common name field.
/// @param peer_cert_path Path to the PEM-encoded certificate of the peer
/// @param local_cert_path Path to the PEM-encoded local certificate
/// @param private_key_path Path to the PEM-encoded private key
/// @param password Optional password if the private key file is encrypted
/// @returns New instance of @ref dnp3_tls_server_config_t
static dnp3_tls_server_config_t dnp3_tls_server_config_init(const char* dns_name, const char* peer_cert_path, const char* local_cert_path, const char* private_key_path, const char* password)
{
    dnp3_tls_server_config_t _return_value = {
        dns_name,
        peer_cert_path,
        local_cert_path,
        private_key_path,
        password,
        DNP3_MIN_TLS_VERSION_V12,
        DNP3_CERTIFICATE_MODE_AUTHORITY_BASED,
        false
    };
    return _return_value;
}


/// @brief TCP server that listens for connections and routes the messages to outstations.
/// 
/// To add outstations to it, use @ref dnp3_outstation_server_add_outstation. Once all the outstations are added, the server can be started with @ref dnp3_outstation_server_bind.
/// 
/// @ref dnp3_outstation_server_destroy is used to gracefully shutdown all the outstations and the server.
typedef struct dnp3_outstation_server_t dnp3_outstation_server_t;

/// @brief Create a new TCP server.
/// 
/// To start it, use @ref dnp3_outstation_server_bind.
/// @param runtime Runtime to execute the server on
/// @param link_error_mode Controls how link errors are handled with respect to the TCP session
/// @param address Address to bind the server to e.g. 127.0.0.1:20000
/// @param out New TCP server instance
/// @return Error code
dnp3_param_error_t dnp3_outstation_server_create_tcp_server(dnp3_runtime_t* runtime, dnp3_link_error_mode_t link_error_mode, const char* address, dnp3_outstation_server_t** out);

/// @brief Create a new TLS server.
/// 
/// To start it, use @ref dnp3_outstation_server_bind.
/// @param runtime Runtime to execute the server on
/// @param link_error_mode Controls how link errors are handled with respect to the session
/// @param address Address to bind the server to e.g. 127.0.0.1:20000
/// @param tls_config TLS server configuration
/// @param out New TLS server instance
/// @return Error code
dnp3_param_error_t dnp3_outstation_server_create_tls_server(dnp3_runtime_t* runtime, dnp3_link_error_mode_t link_error_mode, const char* address, dnp3_tls_server_config_t tls_config, dnp3_outstation_server_t** out);

/// @brief Add an outstation to the server.
/// 
/// The returned @ref dnp3_outstation_t can be used to modify points of the outstation.
/// 
/// In order for the outstation to run, the TCP server must be running. Use @ref dnp3_outstation_server_bind to run it.
/// @param instance Instance of @ref dnp3_outstation_server_t
/// @param config Outstation configuration
/// @param application Outstation application callbacks
/// @param information Outstation information callbacks
/// @param control_handler Outstation control handler
/// @param listener Listener for the connection state
/// @param filter Address filter
/// @param out Outstation handle
/// @return Error code
dnp3_param_error_t dnp3_outstation_server_add_outstation(dnp3_outstation_server_t* instance, dnp3_outstation_config_t config, dnp3_outstation_application_t application, dnp3_outstation_information_t information, dnp3_control_handler_t control_handler, dnp3_connection_state_listener_t listener, dnp3_address_filter_t* filter, dnp3_outstation_t** out);

/// @brief Bind the server to the port and starts listening. Also starts all the outstations associated to it.
/// @param instance Instance of @ref dnp3_outstation_server_t
/// @return Error code
dnp3_param_error_t dnp3_outstation_server_bind(dnp3_outstation_server_t* instance);

/// @brief Gracefully shutdown all the outstations associated to this server, stops the server and release resources.
/// @param instance Instance of @ref dnp3_outstation_server_t to destroy
void dnp3_outstation_server_destroy(dnp3_outstation_server_t* instance);


/// @brief Get the version of the library as a string
/// @return Version number
const char* dnp3_version();

#ifdef __cplusplus
}
#endif
