#include "runt.h"

extern void F_main__main();

word Q(int x) {
  unsigned u = (unsigned)x;
  return (u << 1) | 1;
}

int N(word x) {
  unsigned u = x>>1;
  return (int)u;
}

String s_new(byte n) {
    word guts = oalloc(n, C_Bytes);
    String z = {
      guts,
      Q(0),
      Q(n),
    };
    return z;
}
String s_from_c(const char* s) {
    int n = strlen(s);
    assert(n < INF);
    String z = s_new((byte)n);
    strcpy((char*)z.base, s);
    return z;
}

void F_BUILTIN_println(int i) {
    printf(" %d\n", i);
    fflush(stdout);
}

int main(int argc, const char* argv[]) {
  F_main__main();
  return 0;
}
