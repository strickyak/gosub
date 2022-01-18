#include "runtime/runt.h"
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
  if(n >= INF-1) panic_s("MakeStringFromC: too long");
  word p = oalloc(n+1, 1);
  assert(p);
  strcpy((char*)p, s);
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
  word p = oalloc(len*size, cls);
  assert(p);

  Slice z = {p, 0, len*size};
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

Slice SliceAppend(const char* typecode, Slice a, void* new_elem_ptr, int new_elem_size) {
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
    assert (cap < MAX_CAP);

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
  if (args.base) {
    for (int i=0; i*sizeof(P__any_)<args.len; i++) {
      if (i>0) putchar(' ');

      P__any_* p= (P__any_*)(args.base + args.offset);
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
        default:
          fprintf(stderr, "(typecode `%s` not implemented)", p[i].typecode);
          exit(13);
          break;
      }
    }
  }
  printf("\n");
}

#ifdef USING_MODULE_os
void XXX_os__File__Read(struct os__File *in_f, Slice_(P_byte) in_p, P_int *out_n,
                    Interface_(error) * out_err) {
  //- fprintf(stderr, "os__File__Read <== fd=%d. buflen=%d.\n", in_f->f_fd, in_p.len);

  int cc = read(in_f->f_fd, (char*)(in_p.base) + in_p.offset, in_p.len);
  *out_n = cc;

  int e = errno;
  if (e) {
    struct io__Error* io_error = (struct io__Error*) oalloc(sizeof(struct io__Error), CLASS_io__Error);

    char* c_str = strerror(e);
    char* o_str = (char*) oalloc(strlen(c_str)+1, C_Bytes);
    strcpy(o_str, c_str);
    String go_str = {(word)o_str, 0, strlen(c_str)};
    io_error->f_message = go_str;


    *(struct error**) out_err = (struct error*) io_error;
    return;
  }
  if (cc==0) {
    *(struct error**) out_err = (struct error*) io__EOF;
  } else {
    *(struct error**) out_err = (struct error*) 0;
  }

  //- fprintf(stderr, "os__File__Read ==> %d (%lx)\n", *out_n, (unsigned long)*out_err);
}
#endif

#ifdef USING_MODULE_log
void log__Fatalf(P_string in_format, Slice_(P__any_) in_args) {
  fprintf(stderr, "TODO: log__Fatalf\n");
  exit(13);
}
#endif

#ifdef XXX_USING_MODULE_unix
void unix__Open(P_string filename, P_uint flags, P_uint mode,
                P_int* fd_out, P_int* errno_out) {
  const char* s = STRING_START(&filename);
  int fd = open(s, flags, mode);
  *fd_out = fd;
  *errno_out = 0;
  if (fd<0) {
    *errno_out = errno;
  }
}
void unix__Creat(P_string filename, P_uint mode,
                P_int* fd_out, P_int* errno_out) {
  const char* s = STRING_START(&filename);
  int fd = creat(s, mode);
  *fd_out = fd;
  *errno_out = 0;
  if (fd<0) {
    *errno_out = errno;
  }
}
P_int unix__Close(P_int fd) {
  return close(fd);
}

void unix__Read(P_int fd, P_uintptr buf, P_int size,
                P_int* count_out, P_int* errno_out) {
  int cc = read(fd, (char*)buf, size);
  *count_out = cc;
  *errno_out = errno;
}
void unix__Write(P_int fd, P_uintptr buf, P_int size,
                P_int* count_out, P_int* errno_out) {
  int cc = write(fd, (char*)buf, size);
  *count_out = cc;
  *errno_out = errno;
}
#endif
