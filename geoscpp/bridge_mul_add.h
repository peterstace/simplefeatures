#pragma once
#ifdef __cplusplus
extern "C" {
#endif

void* LIB_NewMul(double x);
void LIB_DestroyMul(void* mul);
double LIB_multiply(void* mul, double y);

void* LIB_NewAdd(double x);
void LIB_DestroyAdd(void* add);
double LIB_sum(void* add, double y);

double LIB_MulAdd(double a, double b, double c);

#ifdef __cplusplus
} // extern "C"
#endif
