#include "___.defs.h"
#include <unistd.h>

// ssize_t write(int fd, const void *buf, size_t count);

extern byte StaticBuffer[256];
extern byte* BufferP;
extern byte* BufferEnd;

void low__WriteStaticBuffer(P_int in_fd, P_int *out_count, P_int *out_errno) {
  *out_count = 0;
  *out_errno = 0;

  ssize_t len = BufferP - StaticBuffer;
  ssize_t cc = write(in_fd, StaticBuffer, len);
  if (cc != len) {
    *out_errno = errno;
  } else {
    *out_count = cc;
  }
}
