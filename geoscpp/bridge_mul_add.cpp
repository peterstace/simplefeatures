#include "bridge_mul_add.h"
#include "mul_add.hpp"

Mul* AsMul(void* mul) {
	return reinterpret_cast<Mul*>(mul);
}

void* LIB_NewMul(double x) {
	return new Mul(x);
}

void LIB_DestroyMul(void* mul) {
	delete AsMul(mul);
}

double LIB_multiply(void* mul, double y) {
	return AsMul(mul)->multiply(y);
}

Add* AsAdd(void* mul) {
	return reinterpret_cast<Add*>(mul);
}

void* LIB_NewAdd(double x) {
	return new Add(x);
}

void LIB_DestroyAdd(void* add) {
	delete AsAdd(add);
}

double LIB_sum(void* add, double y) {
	return AsAdd(add)->sum(y);
}

double LIB_MulAdd(double a, double b, double c) {
	Mul mul(a);
	auto tmp = mul.multiply(b);
	Add add(tmp);
	return add.sum(c);
}
