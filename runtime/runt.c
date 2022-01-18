#include "___.defs.h"

extern void main__main();
extern void initmods();
extern void initvars();

int main(int argc, const char* argv[]) {
  oinit(0, 0, 0);  // noop
  fprintf(stderr, "## Init Vars.\n");
  initvars();
  fprintf(stderr, "## Init Mods.\n");
  initmods();
  fprintf(stderr, "## Main.\n");
  main__main();
  fprintf(stderr, "## Exit.\n");
  return 0;
}

void panic_s(const char* why) {
  fprintf(stderr, "\nPANIC: %s\n", why);
  assert(0);
}

Slice NilSlice = {0};

String MakeStringFromC(const char* s) {
  int n = strlen(s);
  if (n >= INF - 1) panic_s("MakeStringFromC: too long");
  word p = oalloc(n + 1, 1);
  assert(p);
  strcpy((char*)p, s);
  String z = {p, 0, n};
  return z;
}
char* MakeCStrFromString(String s) {
  int n = s.len;
  assert(n < 254);
  char* p = (char*) oalloc(n+1, C_Bytes);
  assert(p);
  memcpy(p, STRING_START(s), n);
  return p;
}

Slice MakeSlice(const char* typecode, int len, int cap, int size) {
  // cap is ignored.
  if (!len) {
    Slice z0 = {0, 0, 0};
    return z0;
  }
  // TODO: use typecode to alloc correct kind.
  byte cls = C_Bytes;
  word p = oalloc(len * size, cls);
  assert(p);

  Slice z = {p, 0, len * size};
  return z;
}

#define INITIAL_CAP 100
#define MAX_CAP 254

Slice AppendSliceInt(Slice a, P_int x) {
  if (!a.base) {
    // Initial allocation.
    word p = oalloc(INITIAL_CAP, 1);
    assert(p);
    a.base = p;
    a.offset = 0;
    a.len = 0;
  }
  byte cap = ocap(a.base);
  if (a.offset + a.len + sizeof(P_int) > cap) {
    assert(cap < MAX_CAP);

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

Slice SliceAppend(const char* typecode, Slice a, void* new_elem_ptr,
                  int new_elem_size) {
  if (!a.base) {
    // Initial allocation.
    word p = oalloc(INITIAL_CAP, 1);
    assert(p);
    a.base = p;
    a.offset = 0;
    a.len = 0;
  }
  byte cap = ocap(a.base);
  if (a.offset + a.len + new_elem_size > cap) {
    assert(cap < MAX_CAP);

    word p = oalloc(MAX_CAP, 1);
    assert(p);
    omemcpy(p, a.base, cap);
    a.base = p;
  }
  assert(a.offset + a.len + new_elem_size <= cap);
  memcpy((char*)a.base + a.offset + a.len, new_elem_ptr, new_elem_size);
  a.len += new_elem_size;
  return a;
}
void SliceGet(Slice a, int size, int nth, void* value) {
  if (!a.base) panic_s("Get on nil slice");
  if (nth * size >= a.len) panic_s("Get slice index OOB");
  memcpy(value, (char*)a.base + a.offset + nth * size, size);
}
void SlicePut(Slice a, int size, int nth, void* value) {
  if (!a.base) panic_s("Put on nil slice");
  if (nth * size >= a.len) panic_s("Put slice index OOB");
  memcpy((char*)a.base + a.offset + nth * size, value, size);
}
int SliceLen(Slice a, int size) {
  if (!a.base) return 0;
  return a.len / size;
}

void builtin__println(Slice args) {
  if (args.base) {
    for (int i = 0; i * sizeof(P__any_) < args.len; i++) {
      if (i > 0) putchar(' ');

      P__any_* p = (P__any_*)(args.base + args.offset);
      switch (p[i].typecode[0]) {
        case 's':
          printf("%s", *(char**)(p[i].pointer));
          break;
        case 'z':
          printf("%s", *(P_bool*)(p[i].pointer) ? "true" : "false");
          break;
        case 'b':
          printf("%d", *(P_byte*)(p[i].pointer));
          break;
        case 'i':
          printf("%d", *(P_int*)(p[i].pointer));
          break;
        case 'u':
          printf("%u", *(P_uint*)(p[i].pointer));
          break;
        case 'p':
          printf("%lu", (unsigned long)*(P_uintptr*)(p[i].pointer));
          break;
        default:
          fprintf(stderr, "builtin__println: typecode `%s` not implemented.\n", p[i].typecode);
          exit(13);
          break;
      }
    }
  }
  printf("\n");
}
