#pragma once

class Mul {
public:
	Mul(double x);
	~Mul();
	double multiply(double y);
private:
	double m_x;
};

class Add {
public:
	Add(double x);
	~Add();
	double sum(double y);
private:
	double m_x;
};
