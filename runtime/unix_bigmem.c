#include "runtime/runt.h"

#ifdef unix

void oinit(word begin, word end, omarker fn) {
  fprintf(stderr, "## oinit: noop\n");
}

word oalloc(byte len, byte cls) {
  assert(len < INF);
  byte cap = (len + 1) & 0xFE;
  size_t sz = sizeof(BigHeader) + cap + 8;
  BigHeader* h = malloc(sz);
  memset(h, 0, sz);
  h->guard1 = GUARD_ONE;
  h->cap = cap;
  h->guard2 = GUARD_TWO;
  h->cls = cls;
  h->guard3 = GUARD_THREE;
  return (word)(h + 1);
}

void ozero(word begin, word len) { memset((char*)begin, 0, len); }

void ofree(word addr) {}

bool ovalidaddr(word addr) {
  if (!addr) return false;
  if (addr & 1) return false;
  BigHeader* bh = (BigHeader*)addr - 1;
  if (bh->guard1 != GUARD_ONE) return false;
  if (bh->guard2 != GUARD_TWO) return false;
  if (bh->guard3 != GUARD_THREE) return false;
  return true;
}

byte ocap(word addr) {
  assert(ovalidaddr(addr));
  BigHeader* bh = (BigHeader*)addr - 1;
  return bh->cap;
}

byte ocls(word addr) {
  assert(ovalidaddr(addr));
  BigHeader* bh = (BigHeader*)addr - 1;
  return bh->cls;
}

void osay(word addr) {
  assert(ovalidaddr(addr));
  BigHeader* bh = (BigHeader*)addr - 1;
  fprintf(stderr, " [[cls=%d cap=%d ", bh->cap, bh->cls);
  for (int i = 0; i < bh->cap; i++) {
    fprintf(stderr, "%02x ", ((char*)addr)[i]);
  }
  fprintf(stderr, "]]\n");
}

void omemcpy(word d, word s, byte n) { memcpy((char*)d, (char*)s, n); }

int omemcmp(word pchar1, byte len1, word pchar2, byte len2) {
  assert(0);  // memcpy((char*)d, (char*)s, n);
}

#endif // unix
