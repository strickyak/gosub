#ifndef _GOSUB_RUNTIME_RUNT_H_
#define _GOSUB_RUNTIME_RUNT_H_

#ifdef unix

#include <assert.h>
#include <errno.h>
#include <memory.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

// open(), creat()
#include <fcntl.h>
#include <sys/stat.h>
#include <sys/types.h>

typedef unsigned char bool;
typedef unsigned char byte;
typedef unsigned long word;

#define true 1
#define false 0
#define INF 0xFF  // sometimes this is INFinity, if type is byte
typedef void (*omarker)();  // TODO: GC

#include "runtime/unix_bigmem.h"

#else /* if not unix */

#include <cmoc.h>

// #include "../pythonine/v0.1/octet.h"
#include "octet.h"

//typedef unsigned char bool;
//typedef unsigned char byte;
//typedef unsigned int word;

#define fprintf(FD, S, ...) PrintError(S, __FILE__, __LINE__)
#define stderr 2

#endif /* unix */


#define P_true 1
#define P_false 0
#define NIL ((word)0)

#define Slice_(T) Slice
#define Map_(K, V) Map
#define Interface_(NAME) VoidStar
#define Struct_(NAME) word
#define Pointer_(NAME) VoidStar

typedef const char* P__type_;
typedef int P_int;
typedef int P__const_int_;
typedef unsigned int P_uint;
typedef word P_uintptr;
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

typedef struct String {
  word base;
  P_uint offset;
  P_uint len;
} String;
typedef String P_string;

typedef struct Slice {
  word base;
  P_uint offset;
  P_uint len;
} Slice;

typedef struct Any {
  void* pointer;  // for everything else
  const char* typecode;
} Any;
typedef struct Any P__any_;

extern Slice NilSlice;
extern void F_BUILTIN_println(int i);

#define STRING_START(S) ((char*)(S).base + (S).offset)

extern void panic_s(const char*);

// Strings
extern String MakeStringFromC(const char* s);
extern char* MakeCStrFromString(String s);
extern String StringAdd(String a, String b);
extern void StringGet(String a, int nth, P_byte* out);

// String & Slice
String FromBytesToString(Slice a);
Slice FromStrinToBytes(String a);

// Slices
extern Slice MakeSlice(const char* typecode, int len, int cap, int size);
extern Slice AppendSliceInt(Slice a, P_int x);
extern Slice SliceAppend(Slice a, void* new_elem_ptr, int new_elem_size, byte base_cls);
extern void SliceGet(Slice a, int size, int nth, void* value);
extern void SlicePut(Slice a, int size, int nth, void* value);
extern int SliceLen(Slice a, int size);
extern void builtin__println(Slice args);

#ifdef unix

#else

#include "os9_base.h"

#endif

#endif // _GOSUB_RUNTIME_RUNT_H_
