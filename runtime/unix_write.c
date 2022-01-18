#include "___.defs.h"
#include "runtime/runt.h"

#ifdef USING_MODULE_unix
void unix__Write(P_int fd, P_uintptr buf, P_int size, P_int* count_out,
                 P_int* errno_out) {
  int cc = write(fd, (char*)buf, size);
  *count_out = cc;
  *errno_out = errno;
}
#endif
