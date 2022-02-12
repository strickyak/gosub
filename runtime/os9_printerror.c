#include "___.defs.h"
#include "os9_base.h"

void PrintError(const char* s, const char* filename, int lineno) {
 int n;
  Os9WritLn(2, s, strlen(s), &n);
  Os9WritLn(2, filename, strlen(s), &n);
}
