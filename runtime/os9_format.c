#include "___.defs.h"
#include "os9_base.h"

extern byte Buffer[256];
extern byte* BufferP;
extern byte* BufferEnd;

void low__WriteBuffer(P_int in_fd, P_int *out_count, P_int *out_errno) {
  // Change LF to CR for OS9.
  for (byte* p = Buffer; p < BufferP; p++) {
    if (*p == 10) *p = 13;
  }

  // Write to given fd.
  *out_errno = Os9Write(in_fd, Buffer, BufferP-Buffer, out_count);
}
