#pragma once
#ifdef __cplusplus
extern "C" {
#endif

#include "stddef.h"

char* sf_version();

void sf_union(
	unsigned char *g1b,
	size_t g1n,
	unsigned char
	*g2b, size_t g2n,
	unsigned char **g3b,
	size_t *g3n, char **estr
);

void sf_delete_uchar_buffer(unsigned char*);
void sf_delete_char_buffer(char*);

#ifdef __cplusplus
} // extern "C"
#endif
