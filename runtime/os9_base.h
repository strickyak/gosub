#ifndef _GOSUB_RUNTIME_OS9_BASE_H_
#define _GOSUB_RUNTIME_OS9_BASE_H_

#define F_Exit    0x06     // Terminate Process
#define F_Sleep   0x0A     // Suspend Process
#define I_Write   0x8A     // Write Data
#define I_ReadLn  0x8B     // Read Line of ASCII Data
#define I_WritLn  0x8C     // Write Line of ASCII Data

// From os9.c:
asm void stkcheck();
asm void pc_trace(int mark, char* ptr);
asm void exit(int status);
asm char* gets(char* buf);
asm void puts(const char* s);

asm int Os9Create(char* path, int mode, int attrs, int* fd);
asm int Os9Open(char* path, int mode, int* fd);
asm int Os9Delete(char* path);
asm int Os9ChgDir(char* path, int mode);
asm int Os9MakDir(char* path, int mode);
asm int Os9GetStt(int path, int func, int* dOut, int* xOut, int* uOut);
asm int Os9Read(int path, char* buf, int buflen, int* bytes_read);
asm int Os9ReadLn(int path, char* buf, int buflen, int* bytes_read);
asm int Os9Write(int path, const char* buf, int max, int* bytes_written);
asm int Os9WritLn(int path, const char* buf, int max, int* bytes_written);
asm int Os9Dup(int path, int* new_path);
asm int Os9Close(int path);
asm int Os9Sleep(int secs);
asm int Os9Wait(int* child_id_and_exit_status);
asm int Os9Fork(const char* program, const char* params, int paramlen, int lang_type, int mem_size, int* child_id);
asm int Os9Chain(const char* program, const char* params, int paramlen, int lang_type, int mem_size);
asm int Os9Send(int process_id, int signal_code);
asm void Os9Exit(int status);

void PrintError(const char* s, const char* filename, int lineno);

void low__Read(P_int in_fd, P_uintptr in_buf, P_int in_size, P_int *out_count, P_int *out_errno);

void low__Write(P_int in_fd, P_uintptr in_buf, P_int in_size, P_int *out_count, P_int *out_errno);

#endif // _GOSUB_RUNTIME_OS9_BASE_H_
