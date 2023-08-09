#include "___.defs.h"
#include "os9_base.h"

void PrintError(const char* s, const char* filename, int lineno) {
 int n;
  Os9WritLn(2, s, strlen(s), &n);
  // CHECK? n == strlen(s)
  Os9WritLn(2, filename, strlen(filename), &n);
  // CHECK? n == strlen(filename)
}
