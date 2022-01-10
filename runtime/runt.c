#include "runt.h"

extern void main__main();

int main(int argc, const char* argv[]) {
  oinit(0, 0, 0);  // noop
  main__main();
  fprintf(stderr, "## Exit.\n");
  return 0;
}

void panic_s(const char* why) {
  fprintf(stderr, "\nPANIC: %s\n", why);
  assert(0);
}

String MakeStringFromC(const char* s) {
  int n = strlen(s);
  if(n >= INF-1) panic_s("MakeStringFromC: too long");
  word p = oalloc(n+1, 1);
  assert(p);
  strcpy((char*)p, s);
  String z = {p, 0, n};
  return z;
}

Slice MakeSlice() {
  Slice z = {0, 0, 0};
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

Slice SliceAppend(Slice a, void* addr, int size) {
  if (!a.base) {
    // Initial allocation.
    word p = oalloc(INITIAL_CAP, 1);
    assert(p);
    a.base = p;
    a.offset = 0;
    a.len = 0;
  }
  byte cap = ocap(a.base);
  if (a.offset + a.len + size > cap) {
    assert (cap < MAX_CAP);

    word p = oalloc(MAX_CAP, 1);
    assert(p);
    omemcpy(p, a.base, cap);
    a.base = p;
  }
  assert(a.offset + a.len + size <= cap);
  memcpy((char*)a.base + a.offset + a.len, addr, size);
  a.len += size;
  return a;
}
void SliceGet(Slice a, int size, int nth, void* value) {
  if (!a.base) panic_s("Get on nil slice");
  if (nth*size >= a.len) panic_s("Get slice index OOB");
  memcpy(value, (char*)a.base + a.offset + nth*size, size);
}
void SlicePut(Slice a, int size, int nth, void* value) {
  if (!a.base) panic_s("Put on nil slice");
  if (nth*size >= a.len) panic_s("Put slice index OOB");
  memcpy((char*)a.base + a.offset + nth*size, value, size);
}
int SliceLen(Slice a, int size) {
  if (!a.base) return 0;
  return a.len / size;
}

void builtin__println(Slice args) {
  if (!args.base) return;

  fprintf(stderr, "## println: args{$%lx, $%x, $%x}\n", (long)args.base, args.offset, args.len);

  if (true) {

    for (int i=0; i*sizeof(P__any_)<args.len; i++) {
      P__any_* p= (P__any_*)(args.base + args.offset);
      printf("%d ", *(int*)(p[i].pointer));
    }

  } else if (true) {

    for (int i=0; i*sizeof(P_int)<args.len; i++) {
      P_int* p= (P_int*)(args.base + args.offset);
      printf("%d ", p[i]);
    }

  } else {

    for (int i=0; i<args.len; i++) {
      byte* p= (byte*)(args.base + args.offset);
      printf("$%02x ", p[i]);
    }

  }

  printf("\n");
}
