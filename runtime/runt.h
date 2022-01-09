#ifndef _GOSUB_RUNTIME_RUNT_H_
#define _GOSUB_RUNTIME_RUNT_H_


#ifdef unix

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
typedef unsigned char bool;
typedef unsigned char byte;
typedef unsigned long word;

typedef void omarker();  // TODO: GC
#define true 1
#define false 0
#define INF 255       // sometimes this is INFinity, if type is byte
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

typedef const char* P__type_;
typedef void* P__any_;
typedef int P_int;
typedef int P__const_int_;
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

extern Slice MakeSlice();
extern Slice AppendSlice(Slice a, P_int x);
extern void builtin__println(Slice args);

#endif
