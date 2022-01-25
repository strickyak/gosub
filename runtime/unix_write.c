#include "___.defs.h"
#include "runtime/runt.h"

#ifdef USING_MODULE_unix
void unix__Write(P_int fd, P_uintptr buf, P_int size, P_int* count_out,
                 P_int* errno_out) {
  fprintf(stderr, "@@ unix_Write << %d, %lx, %d\n", fd, buf, size);
  int cc = write(fd, (char*)buf, size);
  int e = errno;
  fprintf(stderr, "@@ unix_Write >> cc=%d, e=%d\n", cc, e);
  *count_out = cc;
  *errno_out = e;
}
#endif
