#include "___.defs.h"
#include "os9_base.h"

// Hint: int Os9Read( int path, char* buf, int buflen, int* bytes_read)

void low__Read(P_int in_fd, P_uintptr in_buf, P_int in_size, P_int *out_count, P_int *out_errno) {
  *out_count=0;
  byte e = Os9Read(in_fd, in_buf, in_size, out_count);

  if (e == 211/*E$EOF*/) {
    // OS9 has error E$EOF, but golang and unix use count==0 and no error.
    e = 0;
    *out_count=0;
  } else {
    *out_errno = e;
  }
}
