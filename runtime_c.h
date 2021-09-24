#ifndef _GOSUB_RUNTIME_C_H_
#define _GOSUB_RUNTIME_C_H_

typedef unsigned char t_bool;
typedef unsigned char t_uint1;
typedef signed char t_int1;
typedef unsigned int t_uint2;
typedef signed int t_int2;

struct String {
    char* base;
    int offset;
    int len;
};

struct Slice {
    char* base;
    int offset;
    int len;
};

void F_BUILTIN_println(int i);

#endif
