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

#define Slice_(T) Slice
#define Map_(K,V) Map
#define Interface_(NAME) Interface
#define Struct_(NAME) Struct
#define main__main main

typedef const char* P__type_;
typedef void* P__any_;
typedef int P_int;
typedef unsigned int P_uint;
typedef unsigned char P_byte;

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
