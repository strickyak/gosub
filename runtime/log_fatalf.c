#include "___.defs.h"

#ifdef USING_MODULE_log

void log__Fatalf(P_string in_format, Slice_(P__any_) in_args) {
  fprintf(stderr, "TODO: log__Fatalf\n");
  fprintf(stderr, "%s\n", MakeCStrFromString(in_format));
  builtin__println(in_args);
  exit(3);
}

#endif
