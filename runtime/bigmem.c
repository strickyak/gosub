#include "runt.h"

void oinit(word begin, word end, omarker fn) {
    printf("oinit: noop\n");
}

word oalloc(byte len, byte cls) {
    assert(len < INF);
    byte cap = (len+1) & 0xFE;
    size_t sz = sizeof(BigHeader) + cap + 8;
    BigHeader *h = malloc(sz);
    memset(h, 0, sz);
    h->guard1 = GUARD_ONE;
    h->cap = cap;
    h->cls = cls;
    h->guard2 = GUARD_TWO;
    return (word)(h+1);
}

void ozero(word begin, word len) {
}

void ofree(word addr) {
}

bool ovalidaddr(word addr) {
}

byte ocap(word addr) {
}

byte ocls(word addr) {
}

void osay(word addr) {
}

void omemcpy(word d, word s, byte n) {
}

int omemcmp(word pchar1, byte len1, word pchar2, byte len2) {
}


