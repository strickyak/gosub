#include "___.defs.h"
#include "os9_base.h"

extern byte StaticBuffer[256];
extern byte* BufferP;
extern byte* BufferEnd;

void low__WriteStaticBuffer(P_int in_fd, P_int *out_count, P_int *out_errno) {
  // Change LF to CR for OS9.
  for (byte* p = StaticBuffer; p < BufferP; p++) {
    if (*p == 10) *p = 13;
  }

  // Write to given fd.
  *out_errno = Os9Write(in_fd, (const char*)StaticBuffer, BufferP-StaticBuffer, out_count);
}

P_uintptr low__StaticBufferAddress() {
    return (P_uintptr) StaticBuffer;
}
