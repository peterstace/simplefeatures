#include "mul_add.hpp"

Mul::Mul(double x): m_x(x) {
}

Mul::~Mul() {
}

double Mul::multiply(double y) {
	return m_x * y;
}

Add::Add(double x): m_x(x) {
}

Add::~Add() {
}

double Add::sum(double y) {
	return m_x + y;
}
