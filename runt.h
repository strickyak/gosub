#ifndef _GOSUB_RUNTIME_C_H_
#define _GOSUB_RUNTIME_C_H_


#ifdef unix

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
typedef unsigned char bool;
typedef unsigned char byte;
typedef unsigned long word;
typedef void omarker();
#define true 1
#define false 0
#define INF 255
#define NIL ((word)0)
#include "bigmem.h"

#else /* if not unix */

#include <cmoc.h>
#include "../pythonine/octet.h"

#endif /* unix */

typedef unsigned char t_bool;
typedef unsigned char t_uint1;
typedef signed char t_int1;
typedef unsigned int t_uint2;
typedef signed int t_int2;

enum ClsNum {
    C_Free = 0,
    C_Bytes = 1,
    C_Array = 2,
    C_String = 3,
    C_Slice = 4,
    C_Map = 5,
};

typedef struct _string {
    word base;
    word offset;
    word len;
} String;

typedef struct _slice {
    word base;
    word offset;
    word len;
} Slice;

typedef struct _interface {
    word handle;  // for structs
    word pointer;  // for everything else
    const char* type;
} Interface; 

void F_BUILTIN_println(int i);

#endif
