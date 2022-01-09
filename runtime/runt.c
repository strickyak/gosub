#include "runt.h"

extern void main__main();

int main(int argc, const char* argv[]) {
  oinit(0, 0, 0);  // noop
  main__main();
  return 0;
}

Slice MakeSlice() {
  Slice z = {0, 0, 0};
  return z;
}

Slice AppendSlice(Slice a, P_int x) {
  if (!a.base) {
    // Initial allocation.
#define INITIAL_CAP 100
    word p = oalloc(INITIAL_CAP, 1);
    assert(p);
    a.base = p;
    a.offset = 0;
    a.len = 0;
  }
  byte cap = ocap(a.base);
  if (a.offset + a.len + sizeof(P_int) > cap) {
#define MAX_CAP 254
    assert (cap < MAX_CAP);

    word p = oalloc(MAX_CAP, 1);
    assert(p);
    omemcpy(p, a.base, cap);
    a.base = p;
  }
  assert(a.offset + a.len + sizeof(P_int) <= cap);
  *(P_int*)(a.base + a.offset + a.len) = x;
  a.len += sizeof(P_int);
  return a;
}

void builtin__println(Slice args) {
  if (!args.base) return;

  fprintf(stderr, "## println: args{$%lx, $%x, $%x}\n", (long)args.base, args.offset, args.len);
  for (int i=0; i*sizeof(P_int)<args.len; i++) {
    P_int* p= (P_int*)(args.base + args.offset);
    printf("%d ", p[i]);
  }
  printf("\n");
}

#if 0
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
#endif
