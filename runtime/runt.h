#ifndef _GOSUB_RUNTIME_RUNT_H_
#define _GOSUB_RUNTIME_RUNT_H_


#ifdef unix

#include <assert.h>
#include <memory.h>
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
typedef unsigned char bool;
typedef unsigned char byte;
typedef unsigned long word;

typedef void omarker();  // TODO: GC
#define true 1
#define false 0
#define P_true 1
#define P_false 0
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
#define Struct_(NAME) word
#define Pointer_(NAME) VoidStar

typedef const char* P__type_;
typedef int P_int;
typedef int P__const_int_;
typedef unsigned int P_uint;
typedef unsigned char P_byte;
typedef unsigned char P_bool;
typedef void* VoidStar;

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
    P_uint offset;
    P_uint len;
} String;
typedef String P_string;

typedef struct _slice {
    word base;
    P_uint offset;
    P_uint len;
} Slice;

typedef struct _interface {
    // TODO // word handle;  // for structs
    void* pointer;  // for everything else
    const char* typecode;
} Interface; 
typedef Interface P__any_;

extern void F_BUILTIN_println(int i);

extern void panic_s(const char*);
extern String MakeStringFromC(const char* s);
extern Slice MakeSlice();
extern Slice AppendSliceInt(Slice a, P_int x);
extern Slice SliceAppend(Slice a, void* addr, int size);
extern void SliceGet(Slice a, int size, int nth, void* value);
extern void SlicePut(Slice a, int size, int nth, void* value);
extern int SliceLen(Slice a, int size);
extern void builtin__println(Slice args);

#endif
