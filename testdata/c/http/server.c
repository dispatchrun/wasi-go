#include "proxy.h"
#include <stdio.h>
#include <unistd.h>

char buffer[1024];
__attribute__((import_module("types"), import_name("log-it")))
void _wasm_log(int32_t, int32_t);

void send_log() {
    _wasm_log((int32_t) buffer, strlen(buffer));
}

void print_headers(types_headers_t header_handle) {
    types_list_tuple2_string_string_t header_list;
    types_fields_entries(header_handle, &header_list);

    for (int i = 0; i < header_list.len; i++) {
        char name[128];
        char value[128];
        strncpy(name, header_list.ptr[i].f0.ptr, header_list.ptr[i].f0.len);
        name[header_list.ptr[i].f0.len] = 0;
        strncpy(value, (const char*)header_list.ptr[i].f1.ptr, header_list.ptr[i].f1.len);
        value[header_list.ptr[i].f1.len] = 0;
        sprintf(buffer, "%s: %s\n", name, value);
        send_log();
    }
}

const char* str_for_method(types_method_t method) {
    switch (method.tag) {
        case TYPES_METHOD_GET: return "GET";
        case TYPES_METHOD_POST: return "POST";
        case TYPES_METHOD_PUT: return "PUT";
        case TYPES_METHOD_DELETE: return "DELETE";
        case TYPES_METHOD_PATCH: return "PATCH";
        case TYPES_METHOD_HEAD: return "HEAD";
        case TYPES_METHOD_OPTIONS: return "OPTIONS";
        case TYPES_METHOD_CONNECT: return "CONNECT";
        case TYPES_METHOD_TRACE: return "TRACE";
    }
    return "unknown";
}

static int count = 0;

void http_handle(uint32_t req, uint32_t res) {
    sprintf(buffer, "request: %d\n", req);
    send_log();
    proxy_string_t ret;
    types_incoming_request_authority(req, &ret);
    sprintf(buffer, "authority: %.*s\n", (int) ret.len, ret.ptr);
    send_log();

    proxy_string_t path;
    types_incoming_request_path(req, &path);
    sprintf(buffer, "path: %.*s\n", (int) path.len, path.ptr);
    send_log();

    types_method_t method;
    types_incoming_request_method(req, &method);
    sprintf(buffer, "method: %s\n", str_for_method(method));
    send_log();

    types_headers_t headers = types_incoming_request_headers(req);
    print_headers(headers);

    types_tuple2_string_string_t content_type[] = {{
        .f0 = { .ptr = "Server", .len = strlen("Server") },
        .f1 = { .ptr = "WASI-HTTP/0.0.1", .len = 15},
    },
    {
        .f0 = { .ptr = "Content-type", .len = 12 },
        .f1 = { .ptr = "text/plain", .len = strlen("text/plain")},
    }};
    types_list_tuple2_string_string_t headers_list = {
        .ptr = &content_type[0],
        .len = 2,
    };
    types_fields_t out_headers = types_new_fields(&headers_list);
    sprintf(buffer, "Headers are : %d\n", out_headers);
    send_log();

    types_outgoing_response_t response = types_new_outgoing_response(404, out_headers);

    sprintf(buffer, "Response is : %d\n", response);
    send_log();

    types_result_outgoing_response_error_t res_err = {
        .is_err = false,
        .val = {
            .ok = response,
        },
    };
    if (!types_set_response_outparam(res, &res_err)) {
        sprintf(buffer, "Failed to set response outparam: %d -> %d\n", res, response);
        send_log();
    }

    types_outgoing_stream_t stream;
    if (!types_outgoing_response_write(response, &stream)) {
        sprintf(buffer, "Failed to get response\n");
        send_log();
    }

    sprintf(buffer, "got response %d\n", stream);
    send_log();

    char buffer[64];
    snprintf(buffer, 64, "Hello from WASM! (%d)", count);
    count = count + 1;
    const char* body = buffer;
    streams_list_u8_t buf = {
        .ptr = (uint8_t *) body,
        .len = strlen(body),
    };
    uint64_t ret_val;
    streams_write(stream, &buf, &ret_val, NULL);

    types_drop_outgoing_response(res);
    types_drop_fields(out_headers);
}

int main() {
    return 0;
}