#include "___.defs.h"

Slice FromStringToBytes(String a) {
    Slice z = {
        oalloc(a.len, C_Bytes), // base
        0, // offset
        a.len,
        };

    char* src = (char*)a.base + a.offset;
    char* dest = (char*)z.base;
    for (word i = 0; i < a.len; i++) {
        *dest++ = *src++;
    }
    return z;
}
