#pragma once

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef uint64_t Dec64;

Dec64 parse64(const char* str);
char* string64(Dec64 wrapper);

Dec64 frombid64(uint64_t b);
uint64_t tobid64(Dec64 a);

int isNaN64(Dec64 a);

Dec64 add64(Dec64 a, Dec64 b);
Dec64 sub64(Dec64 a, Dec64 b);
Dec64 mul64(Dec64 a, Dec64 b);
Dec64 quo64(Dec64 a, Dec64 b);

#ifdef __cplusplus
}
#endif
