#ifndef _BIGMEM_H_
#define _BIGMEM_H_

#include "runt.h"

// typedef unsigned char bool;
// typedef unsigned char byte;
// typedef unsigned long word;
// #define true 1
// #define false 0

// #ifndef INF
// #define INF 0xFF /* infinity, not a valid byte index */
// #define NIL ((word)0)
// #endif

#ifndef GUARD
#define GUARD 1 /* must be 0 or 1 */
#endif

#define GUARD_ONE 0xAA
#define GUARD_TWO 0xBB

typedef struct BigHeader {
    byte guard1;
    byte cap;
    byte cls;
    byte guard2;
} BigHeader;

void oinit(word begin, word end, omarker fn);
word oalloc(byte len, byte cls);
void ozero(word begin, word len);
void ofree(word addr);  // Unsafe.
bool ovalidaddr(word addr);
byte ocap(word addr);  // capacity in bytes.
byte ocls(word addr);
void osay(word addr);
void omemcpy(word d, word s, byte n);
int omemcmp(word pchar1, byte len1, word pchar2, byte len2);

#endif
