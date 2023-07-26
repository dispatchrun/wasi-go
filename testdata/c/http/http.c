#include "proxy.h"
#include <stdio.h>

void http_handle(uint32_t req, uint32_t resp_out) {
}

int request(uint8_t method_tag, uint8_t scheme_tag, const char * authority_str, const char* path_str, const char* query_str, const char* body) {
    types_tuple2_string_string_t content_type[] = {{
        .f0 = { .ptr = "User-agent", .len = 10 },
        .f1 = { .ptr = "WASI-HTTP/0.0.1", .len = 15},
    },
    {
        .f0 = { .ptr = "Content-type", .len = 12 },
        .f1 = { .ptr = "application/json", .len = 16},
    }};
    types_list_tuple2_string_string_t headers_list = {
        .ptr = &content_type[0],
        .len = 2,
    };
    types_fields_t headers = types_new_fields(&headers_list);
    types_method_t method = { .tag = method_tag };
    types_scheme_t scheme = { .tag = scheme_tag };
    proxy_string_t path, authority, query;
    proxy_string_set(&path, path_str);
    proxy_string_set(&authority, authority_str);
    proxy_string_set(&query, query_str);

    default_outgoing_http_outgoing_request_t req = types_new_outgoing_request(&method, &path, &query, &scheme, &authority, headers);
    default_outgoing_http_future_incoming_response_t res;

    if (req == 0) {
        printf("Error creating request\n");
        return 4;
    }
    if (body != NULL) {
        types_outgoing_stream_t ret;
        if (!types_outgoing_request_write(req, &ret)) {
            printf("Error getting output stream\n");
            return 7;
        }
        streams_list_u8_t buf = {
            .ptr = (uint8_t *) body,
            .len = strlen(body),
        };
        uint64_t ret_val;
        if (!streams_write(ret, &buf, &ret_val, NULL)) {
            printf("Error writing stream\n");
            return 8;
        }
    }

    res = default_outgoing_http_handle(req, NULL);
    if (res == 0) {
        printf("Error sending request\n");
        return 5;
    }
    
    types_result_incoming_response_error_t result;
    if (!types_future_incoming_response_get(res, &result)) {
        printf("failed to get value for incoming request\n");
        return 1;
    }

    if (result.is_err) {
        printf("response is error!\n");
        return 2;
    }
    // poll_drop_pollable(res);

    types_status_code_t code = types_incoming_response_status(result.val.ok);
    printf("STATUS: %d\n", code);

    types_headers_t header_handle = types_incoming_response_headers(result.val.ok);
    types_list_tuple2_string_string_t header_list;
    types_fields_entries(header_handle, &header_list);

    for (int i = 0; i < header_list.len; i++) {
        char name[128];
        char value[128];
        strncpy(name, header_list.ptr[i].f0.ptr, header_list.ptr[i].f0.len);
        name[header_list.ptr[i].f0.len] = 0;
        strncpy(value, header_list.ptr[i].f1.ptr, header_list.ptr[i].f1.len);
        value[header_list.ptr[i].f1.len] = 0;
        printf("%s: %s\n", name, value);
    }


    types_incoming_stream_t stream;
    if (!types_incoming_response_consume(result.val.ok, &stream)) {
        printf("stream is error!\n");
        return 3;
    }

    printf("Stream is %d\n", stream);

    int32_t len = 64 * 1024;
    streams_tuple2_list_u8_bool_t body_res;
    streams_stream_error_t err;
    if (!streams_read(stream, len, &body_res, &err)) {
        printf("BODY read is error!\n");
        return 6;
    }
    if (body_res.f0.len != strlen("Response")) {
        printf("Unexpected body length: %zu\n", body_res.f0.len);
        return 9;
    }
    printf("data from read: %.*s\n", (int) body_res.f0.ptr, body_res.f0.ptr);
    streams_tuple2_list_u8_bool_free(&body_res);


    types_drop_outgoing_request(req);
    streams_drop_input_stream(stream);
    types_drop_incoming_response(result.val.ok);

    return 0;
}

int main() {
    const char *authority = getenv("SERVER");
    int r = request(TYPES_METHOD_GET, TYPES_SCHEME_HTTP, authority, "/get", "?some=arg&goes=here", NULL);
    if (r != 0) {
        return r;
    }
    r = request(TYPES_METHOD_POST, TYPES_SCHEME_HTTP, authority, "/post", "", "{\"foo\": \"bar\"}");
    if (r != 0) {
        return r;
    }
    return 0;
}
