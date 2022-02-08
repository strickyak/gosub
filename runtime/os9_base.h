#ifndef _GOSUB_RUNTIME_OS9_BASE_H_
#define _GOSUB_RUNTIME_OS9_BASE_H_

void PrintError(const char* s, const char* filename, int lineno);

void unix__Read(P_int in_fd, P_uintptr in_buf, P_int in_size, P_int *out_count, P_int *out_errno);

void unix__Write(P_int in_fd, P_uintptr in_buf, P_int in_size, P_int *out_count, P_int *out_errno);

#endif // _GOSUB_RUNTIME_OS9_BASE_H_
