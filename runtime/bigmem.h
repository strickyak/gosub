#ifndef GOSUB_BIGMEM_H_
#define GOSUB_BIGMEM_H_

#include "runt.h"

#ifndef GUARD
#define GUARD 1 /* must be 0 or 1 */
#endif

#define GUARD_ONE 0xAA
#define GUARD_TWO 0xBB
#define GUARD_THREE 0xCC

typedef struct BigHeader {
    byte guard0;
    byte guard1;
    byte cap;
    byte guard2;
    byte cls;
    byte guard3;
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
