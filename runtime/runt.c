#include "___.defs.h"

extern void main__main();
extern void initmods();
extern void initvars();

struct Frame* CurrentFrame;
byte Heap[25000];

#ifndef unix
void mark_handle(word h);

void mark_with_shape(const char* s, word h) {
  if (!s) return;
  const byte* p = (const byte*)s;
  for (; *p; p++) {
    h += (*p);
    mark_handle(*(word*)h);
  }
}

void mark_handle(word h) {
  if (!h) return;
  byte cls = ocls(h);
  assert(cls < NUM_CLASSES);
  {
    P2 = Buffer2;
    PutS2("{");
    PutX2(h);
    PutS2(".");
    PutS2(ClassNames[cls]);
    PutS2("} ");
  }
  omark(h);
  mark_with_shape(ClassMarks[cls], h-1); // because first mark is relative to handle addr less one.
}
#endif

void mark_all() {
#ifndef unix
  // TODO: global vars
  
  // TODO: stack
  for (struct Frame* fr = CurrentFrame; fr; fr=fr->fr_prev) {
    word h = (word)fr;
    for (const byte* s = (const byte*)fr->fr_shape; *s; s++) {
      h += (*s);
      mark_handle(*(word*)h);
    }
  }
#endif
}

int main(int argc, const char* argv[]) {
  // printf("(\n");
  oinit((word)Heap, (word)Heap + sizeof Heap, mark_all);
  initvars();
  initmods();
  main__main();
  // printf(")\n");
  return 0;
}

void panic_s(const char* why) {
  fprintf(stderr, "\nPANIC: %s\n", why);
  assert(0);
}

Slice NilSlice = {0, 0, 0};

byte CheckLen(int i) {
  assert (i>=0);
  assert (i<INF);
  return (byte)i;
}

String MakeStringFromC(const char* s) {
  int n = strlen(s);
  if (n >= INF - 1) panic_s("MakeStringFromC: too long");
  word p = oalloc(CheckLen(n + 1), 1);
  assert(p);
  strcpy((char*)p, s);
  String z = {p, 0, n};
  return z;
}
char* MakeCStrFromString(String s) {
  int n = s.len;
  assert(n < 254);
  char* p = (char*) oalloc(CheckLen(n+1), C_Bytes);
  assert(p);
  memcpy(p, STRING_START(s), n);
  return p;
}

void StringGet(String a, int nth, P_byte* out) {
  if (!a.base) panic_s("Get on empty string");
  if (nth < 0) panic_s("string index negative");
  if (nth >= a.len) panic_s("string index OOB");
  *out =  *((char*)a.base + a.offset + nth);
}

String StringAdd(String a, String b) {
  int n = a.len + b.len;
  if (n >= INF - 1) panic_s("MakeStringFromC: too long");
  word p = oalloc(CheckLen(n + 1), 1);
  assert(p);
  strcpy((char*)p, STRING_START(a));
  strcpy((char*)p+a.len, STRING_START(b));
  String z = {p, 0, n};
  return z;
}

Slice MakeSlice(const char* typecode, int len, int cap, int size) {
  // cap is ignored.
  if (!len) {
    Slice z0 = {0, 0, 0};
    return z0;
  }
  // TODO: use typecode to alloc correct kind.
  byte cls = C_Bytes;
  word p = oalloc(CheckLen(len * size), cls);
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

Slice SliceAppend(Slice a, void* new_elem_ptr, int new_elem_size,
                  byte base_cls) {
  if (!a.base) {
    // Initial allocation.
    word p = oalloc(INITIAL_CAP, base_cls);
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
  if (nth < 0) panic_s("slice index negative");
  if (nth * size >= a.len) panic_s("slice index OOB");
  memcpy(value, (char*)a.base + a.offset + nth * size, size);
}
void SlicePut(Slice a, int size, int nth, void* value) {
  if (!a.base) panic_s("Put on nil slice");
  if (nth < 0) panic_s("slice index negative");
  if (nth * size >= a.len) panic_s("slice index OOB");
  memcpy((char*)a.base + a.offset + nth * size, value, size);
}
int SliceLen(Slice a, int size) {
  if (!a.base) return 0;
  return a.len / size;
}

void builtin__println(Slice args) {
  String fmt = MakeStringFromC("");
  P_uintptr n = args.len;
  bool once = true;
  while (n > 0) {
    fmt = StringAdd(fmt, MakeStringFromC(once? "%v" : " %v"));
    once = false;
    n -= sizeof(P__any_);
  }
  fmt = StringAdd(fmt, MakeStringFromC("\n"));

  low__FormatToBuffer(fmt, args);

  P_int count, errno;
	low__WriteBuffer(1, &count, &errno);
  if (errno) {
    byte berrno = (byte) errno;
    if (berrno==0) berrno=255;
    low__Exit(berrno);
  }
}

#if 1
char Buffer2[500];
char* P2;

void PUT2(int x) {
  *P2++ = (char)x;
}

void PUTHEX2(byte x) {
  assert(x < 16);
  if (x < 10) {
    PUT2('0' + x);
  } else {
    PUT2('a' + x - 10);
  }
}

void PutX2(P_uint x) {
  if (x > 15) {
    PutX2(x >> 4);
    PUTHEX2((byte)(x & 15));
  } else {
    PUTHEX2((byte)x);
  }
}

void PutS2(const char* s) {
  int n = strlen(s);
  memcpy(P2, s, n);
  P2 += n;
}

void Write2() {
  P_int count, errno;
	low__Write(2, (P_uintptr)Buffer2, strlen(Buffer2), &count, &errno);
}

// This can show the calling functions by Frames.
void Where() {
#if 0
  memset(Buffer2, 0, sizeof(Buffer2));
  P2 = Buffer2;

  PutS2("\nWHERE: ");
  for (const struct Frame* fp = CurrentFrame; fp; fp = fp->fr_prev) {
    PutS2(fp->fr_name);
#ifndef unix
    PutS2(": ");
    word mem = (word)fp; // beginning of frame
    for (const char* s=fp->fr_shape; *s; s++) {
      mem += (byte)*s;
      word handle = *(word*)mem;
      PutX2(handle);
      PutS2(" ");
    }
#endif
    PutS2(", ");
  }
  PutS2("\n");
  /*
  P_int count, errno;
	low__Write(1, (P_uintptr)Buffer2, strlen(Buffer2), &count, &errno);
  */
  Write2();
#endif
}
#endif
