package java

import "math"

// This file transliterates java.lang.FdLibm from JDK 21, the pure Java port
// of fdlibm 5.3 that backs java.lang.StrictMath. Results are bit-for-bit
// identical to StrictMath.
//
// Two Go rules need care to keep Java's IEEE 754 double semantics:
//
//   - Go may fuse a multiplication with a following addition or subtraction
//     (even across statements) into one fused multiply-add, which rounds once
//     where Java rounds twice. Every product is therefore wrapped in an
//     explicit float64 conversion, which the Go spec says prevents fusion.
//   - Go evaluates constant expressions exactly, whereas javac folds them with
//     a rounding after each operation. Every expression here that combines
//     only constants rounds at most once, so the two agree. Java's constant
//     expressions that overflow or divide by zero (such as huge*huge) are
//     written as their values.

// fdlibmHI returns the high 32 bits of x (fdlibm's __HI).
func fdlibmHI(x float64) int32 {
	return int32(math.Float64bits(x) >> 32)
}

// fdlibmLO returns the low 32 bits of x (fdlibm's __LO).
func fdlibmLO(x float64) int32 {
	return int32(uint32(math.Float64bits(x)))
}

// fdlibmSetHI returns x with its high 32 bits replaced by high.
func fdlibmSetHI(x float64, high int32) float64 {
	return math.Float64frombits(math.Float64bits(x)&0xffffffff | uint64(uint32(high))<<32)
}

// fdlibmSetLO returns x with its low 32 bits replaced by low.
func fdlibmSetLO(x float64, low int32) float64 {
	return math.Float64frombits(math.Float64bits(x)&0xffffffff00000000 | uint64(uint32(low)))
}

// fdlibmHILO returns the double with the given high and low 32 bits.
func fdlibmHILO(high, low int32) float64 {
	return math.Float64frombits(uint64(uint32(high))<<32 | uint64(uint32(low)))
}

const (
	fdlibmTwo24 = 0x1.0p24
	fdlibmTwo54 = 0x1.0p54
	fdlibmHuge  = 1.0e+300

	fdlibmSignBit       = -0x80000000 // 0x8000_0000 as a Java int
	fdlibmExpBits       = 0x7ff00000
	fdlibmExpSignifBits = 0x7fffffff
)

// fdlibmIsFinite is Double.isFinite.
func fdlibmIsFinite(x float64) bool {
	return math.Abs(x) <= math.MaxFloat64
}

// fdlibmNegZero returns -0.0, which Go cannot write as a constant.
func fdlibmNegZero() float64 {
	return math.Float64frombits(1 << 63)
}

// fdlibmSin is FdLibm.Sin.compute.
func fdlibmSin(x float64) float64 {
	ix := fdlibmHI(x) & fdlibmExpSignifBits
	if ix <= 0x3fe921fb {
		return fdlibmKernelSin(x, 0.0, 0)
	} else if ix >= fdlibmExpBits {
		return x - x
	}
	var y [2]float64
	n := fdlibmRemPio2(x, y[:])
	switch n & 3 {
	case 0:
		return fdlibmKernelSin(y[0], y[1], 1)
	case 1:
		return fdlibmKernelCos(y[0], y[1])
	case 2:
		return -fdlibmKernelSin(y[0], y[1], 1)
	default:
		return -fdlibmKernelCos(y[0], y[1])
	}
}

// fdlibmKernelSin is FdLibm.Sin.__kernel_sin.
func fdlibmKernelSin(x, y float64, iy int32) float64 {
	const (
		S1 = -0x1.5555555555549p-3
		S2 = 0x1.111111110f8a6p-7
		S3 = -0x1.a01a019c161d5p-13
		S4 = 0x1.71de357b1fe7dp-19
		S5 = -0x1.ae5e68a2b9cebp-26
		S6 = 0x1.5d93a5acfd57cp-33
	)
	ix := fdlibmHI(x) & fdlibmExpSignifBits
	if ix < 0x3e400000 {
		if F2I(x) == 0 {
			return x
		}
	}
	z := float64(x * x)
	v := float64(z * x)
	r := S2 + float64(z*(S3+float64(z*(S4+float64(z*(S5+float64(z*S6)))))))
	if iy == 0 {
		return x + float64(v*(S1+float64(z*r)))
	}
	return x - ((float64(z*(float64(0.5*y)-float64(v*r))) - y) - float64(v*S1))
}

// fdlibmCos is FdLibm.Cos.compute.
func fdlibmCos(x float64) float64 {
	ix := fdlibmHI(x) & fdlibmExpSignifBits
	if ix <= 0x3fe921fb {
		return fdlibmKernelCos(x, 0.0)
	} else if ix >= fdlibmExpBits {
		return x - x
	}
	var y [2]float64
	n := fdlibmRemPio2(x, y[:])
	switch n & 3 {
	case 0:
		return fdlibmKernelCos(y[0], y[1])
	case 1:
		return -fdlibmKernelSin(y[0], y[1], 1)
	case 2:
		return -fdlibmKernelCos(y[0], y[1])
	default:
		return fdlibmKernelSin(y[0], y[1], 1)
	}
}

// fdlibmKernelCos is FdLibm.Cos.__kernel_cos.
func fdlibmKernelCos(x, y float64) float64 {
	const (
		C1 = 0x1.555555555554cp-5
		C2 = -0x1.6c16c16c15177p-10
		C3 = 0x1.a01a019cb159p-16
		C4 = -0x1.27e4f809c52adp-22
		C5 = 0x1.1ee9ebdb4b1c4p-29
		C6 = -0x1.8fae9be8838d4p-37
	)
	ix := fdlibmHI(x) & fdlibmExpSignifBits
	if ix < 0x3e400000 {
		if F2I(x) == 0 {
			return 1.0
		}
	}
	z := float64(x * x)
	r := float64(z * (C1 + float64(z*(C2+float64(z*(C3+float64(z*(C4+float64(z*(C5+float64(z*C6)))))))))))
	if ix < 0x3fd33333 {
		return 1.0 - (float64(0.5*z) - (float64(z*r) - float64(x*y)))
	}
	var qx float64
	if ix > 0x3fe90000 {
		qx = 0.28125
	} else {
		qx = fdlibmHILO(ix-0x00200000, 0)
	}
	hz := float64(0.5*z) - qx
	a := 1.0 - qx
	return a - (hz - (float64(z*r) - float64(x*y)))
}

// fdlibmTan is FdLibm.Tan.compute.
func fdlibmTan(x float64) float64 {
	ix := fdlibmHI(x) & fdlibmExpSignifBits
	if ix <= 0x3fe921fb {
		return fdlibmKernelTan(x, 0.0, 1)
	} else if ix >= fdlibmExpBits {
		return x - x
	}
	var y [2]float64
	n := fdlibmRemPio2(x, y[:])
	return fdlibmKernelTan(y[0], y[1], 1-((n&1)<<1))
}

// fdlibmKernelTan is FdLibm.Tan.__kernel_tan.
func fdlibmKernelTan(x, y float64, iy int32) float64 {
	const (
		pio4   = 0x1.921fb54442d18p-1
		pio4lo = 0x1.1a62633145c07p-55
		T0     = 0x1.5555555555563p-2
		T1     = 0x1.111111110fe7ap-3
		T2     = 0x1.ba1ba1bb341fep-5
		T3     = 0x1.664f48406d637p-6
		T4     = 0x1.226e3e96e8493p-7
		T5     = 0x1.d6d22c9560328p-9
		T6     = 0x1.7dbc8fee08315p-10
		T7     = 0x1.344d8f2f26501p-11
		T8     = 0x1.026f71a8d1068p-12
		T9     = 0x1.47e88a03792a6p-14
		T10    = 0x1.2b80f32f0a7e9p-14
		T11    = -0x1.375cbdb605373p-16
		T12    = 0x1.b2a7074bf7ad4p-16
	)
	var z, r, v, w, s float64
	hx := fdlibmHI(x)
	ix := hx & fdlibmExpSignifBits
	if ix < 0x3e300000 {
		if F2I(x) == 0 {
			if ((ix | fdlibmLO(x)) | (iy + 1)) == 0 {
				return 1.0 / math.Abs(x)
			}
			if iy == 1 {
				return x
			}
			// Compute -1 / (x+y) carefully.
			w = x + y
			z = fdlibmSetLO(w, 0)
			v = y - (z - x)
			a := -1.0 / w
			t := fdlibmSetLO(a, 0)
			s = 1.0 + float64(t*z)
			return t + float64(a*(s+float64(t*v)))
		}
	}
	if ix >= 0x3fe59428 {
		if hx < 0 {
			x = -x
			y = -y
		}
		z = pio4 - x
		w = pio4lo - y
		x = z + w
		y = 0.0
	}
	z = float64(x * x)
	w = float64(z * z)
	r = T1 + float64(w*(T3+float64(w*(T5+float64(w*(T7+float64(w*(T9+float64(w*T11)))))))))
	v = float64(z * (T2 + float64(w*(T4+float64(w*(T6+float64(w*(T8+float64(w*(T10+float64(w*T12)))))))))))
	s = float64(z * x)
	r = y + float64(z*(float64(s*(r+v))+y))
	r += float64(T0 * s)
	w = x + r
	if ix >= 0x3fe59428 {
		v = float64(iy)
		return float64(float64(1-((hx>>30)&2)) * (v - float64(2.0*(x-(float64(w*w)/(w+v)-r)))))
	}
	if iy == 1 {
		return w
	}
	// Compute -1.0/(x + r) accurately.
	z = fdlibmSetLO(w, 0)
	v = r - (z - x)
	a := -1.0 / w
	t := fdlibmSetLO(a, 0)
	s = 1.0 + float64(t*z)
	return t + float64(a*(s+float64(t*v)))
}

// fdlibmTwoOverPi holds 2/pi in 24-bit chunks. It is a read-only lookup
// table: nothing writes to it, so sharing it is safe.
var fdlibmTwoOverPi = [...]int32{
	0xA2F983, 0x6E4E44, 0x1529FC, 0x2757D1, 0xF534DD, 0xC0DB62,
	0x95993C, 0x439041, 0xFE5163, 0xABDEBB, 0xC561B7, 0x246E3A,
	0x424DD2, 0xE00649, 0x2EEA09, 0xD1921C, 0xFE1DEB, 0x1CB129,
	0xA73EE8, 0x8235F5, 0x2EBB44, 0x84E99C, 0x7026B4, 0x5F7E41,
	0x3991D6, 0x398353, 0x39F49C, 0x845F8B, 0xBDF928, 0x3B1FF8,
	0x97FFDE, 0x05980F, 0xEF2F11, 0x8B5A0A, 0x6D1F6D, 0x367ECF,
	0x27CB09, 0xB74F46, 0x3F669E, 0x5FEA2D, 0x7527BA, 0xC7EBE5,
	0xF17B3D, 0x0739F7, 0x8A5292, 0xEA6BFB, 0x5FB11F, 0x8D5D08,
	0x560330, 0x46FC7B, 0x6BABF0, 0xCFBC20, 0x9AF436, 0x1DA9E3,
	0x91615E, 0xE61B08, 0x659985, 0x5F14A0, 0x68408D, 0xFFD880,
	0x4D7327, 0x310606, 0x1556CA, 0x73A8C9, 0x60E27B, 0xC08C6B,
}

// fdlibmNpio2HW holds the high words of n*pi/2 for n = 1..32. It is a
// read-only lookup table: nothing writes to it, so sharing it is safe.
var fdlibmNpio2HW = [...]int32{
	0x3FF921FB, 0x400921FB, 0x4012D97C, 0x401921FB, 0x401F6A7A, 0x4022D97C,
	0x4025FDBB, 0x402921FB, 0x402C463A, 0x402F6A7A, 0x4031475C, 0x4032D97C,
	0x40346B9C, 0x4035FDBB, 0x40378FDB, 0x403921FB, 0x403AB41B, 0x403C463A,
	0x403DD85A, 0x403F6A7A, 0x40407E4C, 0x4041475C, 0x4042106C, 0x4042D97C,
	0x4043A28C, 0x40446B9C, 0x404534AC, 0x4045FDBB, 0x4046C6CB, 0x40478FDB,
	0x404858EB, 0x404921FB,
}

// fdlibmRemPio2 is FdLibm.RemPio2.__ieee754_rem_pio2: it stores x rem pi/2
// in y[0]+y[1] and returns the quotient's low bits.
func fdlibmRemPio2(x float64, y []float64) int32 {
	const (
		invpio2 = 0x1.45f306dc9c883p-1
		pio2_1  = 0x1.921fb544p0
		pio2_1t = 0x1.0b4611a626331p-34
		pio2_2  = 0x1.0b4611a6p-34
		pio2_2t = 0x1.3198a2e037073p-69
		pio2_3  = 0x1.3198a2ep-69
		pio2_3t = 0x1.b839a252049c1p-104
	)
	hx := fdlibmHI(x)
	ix := hx & fdlibmExpSignifBits
	if ix <= 0x3fe921fb { // |x| ~<= pi/4, no need for reduction
		y[0] = x
		y[1] = 0
		return 0
	}
	if ix < 0x4002d97c { // |x| < 3pi/4, special case with n=+-1
		if hx > 0 {
			z := x - pio2_1
			if ix != 0x3ff921fb { // 33+53 bit pi is good enough
				y[0] = z - pio2_1t
				y[1] = (z - y[0]) - pio2_1t
			} else { // near pi/2, use 33+33+53 bit pi
				z -= pio2_2
				y[0] = z - pio2_2t
				y[1] = (z - y[0]) - pio2_2t
			}
			return 1
		}
		z := x + pio2_1
		if ix != 0x3ff921fb {
			y[0] = z + pio2_1t
			y[1] = (z - y[0]) + pio2_1t
		} else {
			z += pio2_2
			y[0] = z + pio2_2t
			y[1] = (z - y[0]) + pio2_2t
		}
		return -1
	}
	if ix <= 0x413921fb { // |x| ~<= 2^19*(pi/2), medium size
		t := math.Abs(x)
		n := F2I(float64(t*invpio2) + 0.5)
		fn := float64(n)
		r := t - float64(fn*pio2_1)
		w := float64(fn * pio2_1t) // 1st round good to 85 bit
		if n < 32 && ix != fdlibmNpio2HW[n-1] {
			y[0] = r - w // quick check no cancellation
		} else {
			j := ix >> 20
			y[0] = r - w
			i := j - ((fdlibmHI(y[0]) >> 20) & 0x7ff)
			if i > 16 { // 2nd iteration needed, good to 118
				t = r
				w = float64(fn * pio2_2)
				r = t - w
				w = float64(fn*pio2_2t) - ((t - r) - w)
				y[0] = r - w
				i = j - ((fdlibmHI(y[0]) >> 20) & 0x7ff)
				if i > 49 { // 3rd iteration need, 151 bits acc
					t = r
					w = float64(fn * pio2_3)
					r = t - w
					w = float64(fn*pio2_3t) - ((t - r) - w)
					y[0] = r - w
				}
			}
		}
		y[1] = (r - y[0]) - w
		if hx < 0 {
			y[0] = -y[0]
			y[1] = -y[1]
			return -n
		}
		return n
	}
	// All other (large) arguments.
	if ix >= fdlibmExpBits { // x is inf or NaN
		y[0] = x - x
		y[1] = y[0]
		return 0
	}
	// Set z = scalbn(|x|, ilogb(x)-23).
	z := fdlibmSetLO(0.0, fdlibmLO(x))
	e0 := (ix >> 20) - 1046 // e0 = ilogb(z) - 23
	z = fdlibmSetHI(z, ix-(e0<<20))
	var tx [3]float64
	for i := 0; i < 2; i++ {
		tx[i] = float64(F2I(z))
		z = float64((z - tx[i]) * fdlibmTwo24)
	}
	tx[2] = z
	nx := int32(3)
	for tx[nx-1] == 0.0 { // skip zero term
		nx--
	}
	n := fdlibmKernelRemPio2(tx[:], y, e0, nx, 2, fdlibmTwoOverPi[:])
	if hx < 0 {
		y[0] = -y[0]
		y[1] = -y[1]
		return -n
	}
	return n
}

// fdlibmInitJK is the initial value of jk for each precision. It is a
// read-only lookup table: nothing writes to it, so sharing it is safe.
var fdlibmInitJK = [...]int32{2, 3, 4, 6}

// fdlibmPIo2 holds pi/2 in 24-bit chunks. It is a read-only lookup table:
// nothing writes to it, so sharing it is safe.
var fdlibmPIo2 = [...]float64{
	0x1.921fb4p0,
	0x1.4442dp-24,
	0x1.846988p-48,
	0x1.8cc516p-72,
	0x1.01b838p-96,
	0x1.a25204p-120,
	0x1.382228p-145,
	0x1.9f31dp-169,
}

// fdlibmKernelRemPio2 is FdLibm.KernelRemPio2.__kernel_rem_pio2.
func fdlibmKernelRemPio2(x, y []float64, e0, nx, prec int32, ipio2 []int32) int32 {
	const twon24 = 0x1.0p-24
	var iq [20]int32
	var f, fq, q [20]float64
	var z, fw float64
	var i, j, k, n, ih int32

	// Initialize jk.
	jk := fdlibmInitJK[prec]
	jp := jk

	// Determine jx, jv, q0, note that 3 > q0.
	jx := nx - 1
	jv := (e0 - 3) / 24
	if jv < 0 {
		jv = 0
	}
	q0 := e0 - 24*(jv+1)

	// Set up f[0] to f[jx+jk] where f[jx+jk] = ipio2[jv+jk].
	j = jv - jx
	m := jx + jk
	for i = 0; i <= m; i, j = i+1, j+1 {
		if j < 0 {
			f[i] = 0.0
		} else {
			f[i] = float64(ipio2[j])
		}
	}

	// Compute q[0],q[1],...q[jk].
	for i = 0; i <= jk; i++ {
		fw = 0.0
		for j = 0; j <= jx; j++ {
			fw += float64(x[j] * f[jx+i-j])
		}
		q[i] = fw
	}

	jz := jk
	for {
		// Distill q[] into iq[] reversingly.
		i = 0
		j = jz
		z = q[jz]
		for ; j > 0; i, j = i+1, j-1 {
			fw = float64(F2I(float64(twon24 * z)))
			iq[i] = F2I(z - float64(fdlibmTwo24*fw))
			z = q[j-1] + fw
		}

		// Compute n.
		z = Math_scalb_double_int(z, q0)                 // actual value of z
		z -= float64(8.0 * math.Floor(float64(z*0.125))) // trim off integer >= 8
		n = F2I(z)
		z -= float64(n)
		ih = 0
		switch {
		case q0 > 0: // need iq[jz-1] to determine n
			i = iq[jz-1] >> (24 - q0)
			n += i
			iq[jz-1] -= i << (24 - q0)
			ih = iq[jz-1] >> (23 - q0)
		case q0 == 0:
			ih = iq[jz-1] >> 23
		case z >= 0.5:
			ih = 2
		}

		if ih > 0 { // q > 0.5
			n++
			carry := int32(0)
			for i = 0; i < jz; i++ { // compute 1-q
				j = iq[i]
				if carry == 0 {
					if j != 0 {
						carry = 1
						iq[i] = 0x1000000 - j
					}
				} else {
					iq[i] = 0xffffff - j
				}
			}
			if q0 > 0 { // rare case: chance is 1 in 12
				switch q0 {
				case 1:
					iq[jz-1] &= 0x7fffff
				case 2:
					iq[jz-1] &= 0x3fffff
				}
			}
			if ih == 2 {
				z = 1.0 - z
				if carry != 0 {
					z -= Math_scalb_double_int(1.0, q0)
				}
			}
		}

		// Check if recomputation is needed.
		if z == 0.0 {
			j = 0
			for i = jz - 1; i >= jk; i-- {
				j |= iq[i]
			}
			if j == 0 { // need recomputation
				k = 1
				for iq[jk-k] == 0 { // k = no. of terms needed
					k++
				}
				for i = jz + 1; i <= jz+k; i++ { // add q[jz+1] to q[jz+k]
					f[jx+i] = float64(ipio2[jv+i])
					fw = 0.0
					for j = 0; j <= jx; j++ {
						fw += float64(x[j] * f[jx+i-j])
					}
					q[i] = fw
				}
				jz += k
				continue
			}
		}
		break
	}

	// Chop off zero terms.
	if z == 0.0 {
		jz--
		q0 -= 24
		for iq[jz] == 0 {
			jz--
			q0 -= 24
		}
	} else { // break z into 24-bit if necessary
		z = Math_scalb_double_int(z, -q0)
		if z >= fdlibmTwo24 {
			fw = float64(F2I(float64(twon24 * z)))
			iq[jz] = F2I(z - float64(fdlibmTwo24*fw))
			jz++
			q0 += 24
			iq[jz] = F2I(fw)
		} else {
			iq[jz] = F2I(z)
		}
	}

	// Convert integer "bit" chunk to floating-point value.
	fw = Math_scalb_double_int(1.0, q0)
	for i = jz; i >= 0; i-- {
		q[i] = float64(fw * float64(iq[i]))
		fw = float64(fw * twon24)
	}

	// Compute PIo2[0,...,jp]*q[jz,...,0].
	for i = jz; i >= 0; i-- {
		fw = 0.0
		for k = 0; k <= jp && k <= jz-i; k++ {
			fw += float64(fdlibmPIo2[k] * q[i+k])
		}
		fq[jz-i] = fw
	}

	// Compress fq[] into y[].
	switch prec {
	case 0:
		fw = 0.0
		for i = jz; i >= 0; i-- {
			fw += fq[i]
		}
		if ih == 0 {
			y[0] = fw
		} else {
			y[0] = -fw
		}
	case 1, 2:
		fw = 0.0
		for i = jz; i >= 0; i-- {
			fw += fq[i]
		}
		if ih == 0 {
			y[0] = fw
		} else {
			y[0] = -fw
		}
		fw = fq[0] - fw
		for i = 1; i <= jz; i++ {
			fw += fq[i]
		}
		if ih == 0 {
			y[1] = fw
		} else {
			y[1] = -fw
		}
	case 3:
		for i = jz; i > 0; i-- {
			fw = fq[i-1] + fq[i]
			fq[i] += fq[i-1] - fw
			fq[i-1] = fw
		}
		for i = jz; i > 1; i-- {
			fw = fq[i-1] + fq[i]
			fq[i] += fq[i-1] - fw
			fq[i-1] = fw
		}
		fw = 0.0
		for i = jz; i >= 2; i-- {
			fw += fq[i]
		}
		if ih == 0 {
			y[0] = fq[0]
			y[1] = fq[1]
			y[2] = fw
		} else {
			y[0] = -fq[0]
			y[1] = -fq[1]
			y[2] = -fw
		}
	}
	return n & 7
}

// fdlibmAsin is FdLibm.Asin.compute.
func fdlibmAsin(x float64) float64 {
	const (
		pio2_hi = 0x1.921fb54442d18p0
		pio2_lo = 0x1.1a62633145c07p-54
		pio4_hi = 0x1.921fb54442d18p-1
		pS0     = 0x1.5555555555555p-3
		pS1     = -0x1.4d61203eb6f7dp-2
		pS2     = 0x1.9c1550e884455p-3
		pS3     = -0x1.48228b5688f3bp-5
		pS4     = 0x1.9efe07501b288p-11
		pS5     = 0x1.23de10dfdf709p-15
		qS1     = -0x1.33a271c8a2d4bp1
		qS2     = 0x1.02ae59c598ac8p1
		qS3     = -0x1.6066c1b8d0159p-1
		qS4     = 0x1.3b8c5b12e9282p-4
	)
	var t, w, p, q, c, r, s float64
	hx := fdlibmHI(x)
	ix := hx & fdlibmExpSignifBits
	if ix >= 0x3ff00000 { // |x| >= 1
		if ((ix - 0x3ff00000) | fdlibmLO(x)) == 0 {
			// asin(1) = +-pi/2 with inexact
			return float64(x*pio2_hi) + float64(x*pio2_lo)
		}
		return (x - x) / (x - x) // asin(|x| > 1) is NaN
	} else if ix < 0x3fe00000 { // |x| < 0.5
		if ix < 0x3e400000 { // if |x| < 2**-27
			if fdlibmHuge+x > 1.0 { // return x with inexact if x != 0
				return x
			}
		} else {
			t = float64(x * x)
		}
		p = float64(t * (pS0 + float64(t*(pS1+float64(t*(pS2+float64(t*(pS3+float64(t*(pS4+float64(t*pS5)))))))))))
		q = 1.0 + float64(t*(qS1+float64(t*(qS2+float64(t*(qS3+float64(t*qS4)))))))
		w = p / q
		return x + float64(x*w)
	}
	// 1 > |x| >= 0.5
	w = 1.0 - math.Abs(x)
	t = float64(w * 0.5)
	p = float64(t * (pS0 + float64(t*(pS1+float64(t*(pS2+float64(t*(pS3+float64(t*(pS4+float64(t*pS5)))))))))))
	q = 1.0 + float64(t*(qS1+float64(t*(qS2+float64(t*(qS3+float64(t*qS4)))))))
	s = math.Sqrt(t)
	if ix >= 0x3fef3333 { // if |x| > 0.975
		w = p / q
		t = pio2_hi - (float64(2.0*(s+float64(s*w))) - pio2_lo)
	} else {
		w = fdlibmSetLO(s, 0)
		c = (t - float64(w*w)) / (s + w)
		r = p / q
		p = float64(float64(2.0*s)*r) - (pio2_lo - float64(2.0*c))
		q = pio4_hi - float64(2.0*w)
		t = pio4_hi - (p - q)
	}
	if hx > 0 {
		return t
	}
	return -t
}

// fdlibmAcos is FdLibm.Acos.compute.
func fdlibmAcos(x float64) float64 {
	const (
		pi      = 3.141592653589793
		pio2_hi = 0x1.921fb54442d18p0
		pio2_lo = 0x1.1a62633145c07p-54
		pS0     = 0x1.5555555555555p-3
		pS1     = -0x1.4d61203eb6f7dp-2
		pS2     = 0x1.9c1550e884455p-3
		pS3     = -0x1.48228b5688f3bp-5
		pS4     = 0x1.9efe07501b288p-11
		pS5     = 0x1.23de10dfdf709p-15
		qS1     = -0x1.33a271c8a2d4bp1
		qS2     = 0x1.02ae59c598ac8p1
		qS3     = -0x1.6066c1b8d0159p-1
		qS4     = 0x1.3b8c5b12e9282p-4
	)
	var z, p, q, r, w, s, c, df float64
	hx := fdlibmHI(x)
	ix := hx & fdlibmExpSignifBits
	if ix >= 0x3ff00000 { // |x| >= 1
		if ((ix - 0x3ff00000) | fdlibmLO(x)) == 0 { // |x| == 1
			if hx > 0 { // acos(1) = 0
				return 0.0
			}
			// acos(-1) = pi. The constant expression has one rounding, as
			// 2.0*pio2_lo is exact.
			return pi + 2.0*pio2_lo
		}
		return (x - x) / (x - x) // acos(|x| > 1) is NaN
	}
	if ix < 0x3fe00000 { // |x| < 0.5
		if ix <= 0x3c600000 { // if |x| < 2**-57
			return pio2_hi + pio2_lo
		}
		z = float64(x * x)
		p = float64(z * (pS0 + float64(z*(pS1+float64(z*(pS2+float64(z*(pS3+float64(z*(pS4+float64(z*pS5)))))))))))
		q = 1.0 + float64(z*(qS1+float64(z*(qS2+float64(z*(qS3+float64(z*qS4)))))))
		r = p / q
		return pio2_hi - (x - (pio2_lo - float64(x*r)))
	} else if hx < 0 { // x < -0.5
		z = float64((1.0 + x) * 0.5)
		p = float64(z * (pS0 + float64(z*(pS1+float64(z*(pS2+float64(z*(pS3+float64(z*(pS4+float64(z*pS5)))))))))))
		q = 1.0 + float64(z*(qS1+float64(z*(qS2+float64(z*(qS3+float64(z*qS4)))))))
		s = math.Sqrt(z)
		r = p / q
		w = float64(r*s) - pio2_lo
		return pi - float64(2.0*(s+w))
	}
	// x > 0.5
	z = float64((1.0 - x) * 0.5)
	s = math.Sqrt(z)
	df = fdlibmSetLO(s, 0)
	c = (z - float64(df*df)) / (s + df)
	p = float64(z * (pS0 + float64(z*(pS1+float64(z*(pS2+float64(z*(pS3+float64(z*(pS4+float64(z*pS5)))))))))))
	q = 1.0 + float64(z*(qS1+float64(z*(qS2+float64(z*(qS3+float64(z*qS4)))))))
	r = p / q
	w = float64(r*s) + c
	return float64(2.0 * (df + w))
}

// fdlibmAtanHi holds the high parts of atan(0.5), atan(1), atan(1.5) and
// atan(inf). It is a read-only lookup table: nothing writes to it, so sharing
// it is safe.
var fdlibmAtanHi = [...]float64{
	0x1.dac670561bb4fp-2,
	0x1.921fb54442d18p-1,
	0x1.f730bd281f69bp-1,
	0x1.921fb54442d18p0,
}

// fdlibmAtanLo holds the low parts matching fdlibmAtanHi. It is a read-only
// lookup table: nothing writes to it, so sharing it is safe.
var fdlibmAtanLo = [...]float64{
	0x1.a2b7f222f65e2p-56,
	0x1.1a62633145c07p-55,
	0x1.007887af0cbbdp-56,
	0x1.1a62633145c07p-54,
}

// fdlibmAtan is FdLibm.Atan.compute.
func fdlibmAtan(x float64) float64 {
	const (
		aT0  = 0x1.555555555550dp-2
		aT1  = -0x1.999999998ebc4p-3
		aT2  = 0x1.24924920083ffp-3
		aT3  = -0x1.c71c6fe231671p-4
		aT4  = 0x1.745cdc54c206ep-4
		aT5  = -0x1.3b0f2af749a6dp-4
		aT6  = 0x1.10d66a0d03d51p-4
		aT7  = -0x1.dde2d52defd9ap-5
		aT8  = 0x1.97b4b24760debp-5
		aT9  = -0x1.2b4442c6a6c2fp-5
		aT10 = 0x1.0ad3ae322da11p-6
	)
	var id int32
	hx := fdlibmHI(x)
	ix := hx & fdlibmExpSignifBits
	if ix >= 0x44100000 { // if |x| >= 2^66
		if ix > fdlibmExpBits || (ix == fdlibmExpBits && fdlibmLO(x) != 0) {
			return x + x // NaN
		}
		if hx > 0 {
			return fdlibmAtanHi[3] + fdlibmAtanLo[3]
		}
		return -fdlibmAtanHi[3] - fdlibmAtanLo[3]
	}
	if ix < 0x3fdc0000 { // |x| < 0.4375
		if ix < 0x3e200000 { // |x| < 2^-29
			if fdlibmHuge+x > 1.0 { // raise inexact
				return x
			}
		}
		id = -1
	} else {
		x = math.Abs(x)
		if ix < 0x3ff30000 { // |x| < 1.1875
			if ix < 0x3fe60000 { // 7/16 <= |x| < 11/16
				id = 0
				x = (float64(2.0*x) - 1.0) / (2.0 + x)
			} else { // 11/16 <= |x| < 19/16
				id = 1
				x = (x - 1.0) / (x + 1.0)
			}
		} else {
			if ix < 0x40038000 { // |x| < 2.4375
				id = 2
				x = (x - 1.5) / (1.0 + float64(1.5*x))
			} else { // 2.4375 <= |x| < 2^66
				id = 3
				x = -1.0 / x
			}
		}
	}
	// End of argument reduction.
	z := float64(x * x)
	w := float64(z * z)
	// Break sum from i=0 to 10 aT[i]z**(i+1) into odd and even poly.
	s1 := float64(z * (aT0 + float64(w*(aT2+float64(w*(aT4+float64(w*(aT6+float64(w*(aT8+float64(w*aT10)))))))))))
	s2 := float64(w * (aT1 + float64(w*(aT3+float64(w*(aT5+float64(w*(aT7+float64(w*aT9)))))))))
	if id < 0 {
		return x - float64(x*(s1+s2))
	}
	z = fdlibmAtanHi[id] - ((float64(x*(s1+s2)) - fdlibmAtanLo[id]) - x)
	if hx < 0 {
		return -z
	}
	return z
}

// fdlibmAtan2 is FdLibm.Atan2.compute.
func fdlibmAtan2(y, x float64) float64 {
	const (
		pi     = 3.141592653589793
		tiny   = 1.0e-300
		pi_o_4 = 0x1.921fb54442d18p-1
		pi_o_2 = 0x1.921fb54442d18p0
		pi_lo  = 0x1.1a62633145c07p-53
	)
	var z float64
	hx := fdlibmHI(x)
	ix := hx & fdlibmExpSignifBits
	lx := fdlibmLO(x)
	hy := fdlibmHI(y)
	iy := hy & fdlibmExpSignifBits
	ly := fdlibmLO(y)
	if x != x || y != y {
		return x + y
	}
	if ((hx - 0x3ff00000) | lx) == 0 { // x = 1.0
		return fdlibmAtan(y)
	}
	m := ((hy >> 31) & 1) | ((hx >> 30) & 2) // 2*sign(x) + sign(y)

	// The constant expressions below have one rounding each: 3.0*pi_o_4 is
	// exact, so Go's exact constant arithmetic agrees with javac's folding.

	// When y = 0.
	if (iy | ly) == 0 {
		switch m {
		case 0, 1:
			return y // atan(+/-0, +anything) = +/-0
		case 2:
			return pi + tiny // atan(+0, -anything) = pi
		default:
			return -pi - tiny // atan(-0, -anything) = -pi
		}
	}
	// When x = 0.
	if (ix | lx) == 0 {
		if hy < 0 {
			return -pi_o_2 - tiny
		}
		return pi_o_2 + tiny
	}

	// When x is INF.
	if ix == fdlibmExpBits {
		if iy == fdlibmExpBits {
			switch m {
			case 0:
				return pi_o_4 + tiny // atan(+INF, +INF)
			case 1:
				return -pi_o_4 - tiny // atan(-INF, +INF)
			case 2:
				return 3.0*pi_o_4 + tiny // atan(+INF, -INF)
			default:
				return -3.0*pi_o_4 - tiny // atan(-INF, -INF)
			}
		} else {
			switch m {
			case 0:
				return 0.0 // atan(+..., +INF)
			case 1:
				return fdlibmNegZero() // atan(-..., +INF)
			case 2:
				return pi + tiny // atan(+..., -INF)
			default:
				return -pi - tiny // atan(-..., -INF)
			}
		}
	}
	// When y is INF.
	if iy == fdlibmExpBits {
		if hy < 0 {
			return -pi_o_2 - tiny
		}
		return pi_o_2 + tiny
	}

	// Compute y/x.
	k := (iy - ix) >> 20
	switch {
	case k > 60: // |y/x| > 2**60
		z = pi_o_2 + 0.5*pi_lo
	case hx < 0 && k < -60: // |y|/x < -2**60
		z = 0.0
	default: // safe to do y/x
		z = fdlibmAtan(math.Abs(y / x))
	}
	switch m {
	case 0:
		return z // atan(+, +)
	case 1:
		return -z // atan(-, +)
	case 2:
		return pi - (z - pi_lo) // atan(+, -)
	default:
		return (z - pi_lo) - pi // atan(-, -)
	}
}

// fdlibmCbrt is FdLibm.Cbrt.compute.
func fdlibmCbrt(x float64) float64 {
	const (
		B1 = 715094163 // B1 = (682-0.03306235651)*2**20
		B2 = 696219795 // B2 = (664-0.03306235651)*2**20
		C  = 0x1.15f15f15f15f1p-1
		D  = -0x1.691de2532c834p-1
		E  = 0x1.6a0ea0ea0ea0fp0
		F  = 0x1.9b6db6db6db6ep0
		G  = 0x1.6db6db6db6db7p-2
	)
	t := 0.0
	if x == 0.0 || !fdlibmIsFinite(x) {
		return x // handles signed zeros properly
	}
	sign := 1.0
	if x < 0.0 {
		sign = -1.0
	}
	x = math.Abs(x)

	// Rough cbrt to 5 bits.
	if x < 0x1.0p-1022 { // subnormal number
		t = 0x1.0p54
		t = float64(t * x)
		t = fdlibmSetHI(t, fdlibmHI(t)/3+B2)
	} else {
		hx := fdlibmHI(x)
		t = fdlibmSetHI(t, hx/3+B1)
	}

	// New cbrt to 23 bits, may be implemented in single precision.
	r := float64(t*t) / x
	s := C + float64(r*t)
	t = float64(t * (G + F/(s+E+D/s)))

	// Chopped to 20 bits and make it larger than cbrt(x).
	t = fdlibmSetLO(t, 0)
	t = fdlibmSetHI(t, fdlibmHI(t)+0x00000001)

	// One step newton iteration to 53 bits with error less than 0.667 ulps.
	s = float64(t * t) // t*t is exact
	r = x / s
	w := t + t
	r = (r - t) / (w + r) // r-s is exact
	t += float64(t * r)

	// Restore the original sign bit.
	return float64(sign * t)
}

// fdlibmHypot is FdLibm.Hypot.compute.
func fdlibmHypot(x, y float64) float64 {
	const (
		twoMinus600 = 0x1.0p-600
		twoPlus600  = 0x1.0p+600
	)
	a := math.Abs(x)
	b := math.Abs(y)

	if !fdlibmIsFinite(a) || !fdlibmIsFinite(b) {
		if math.IsInf(a, 1) || math.IsInf(b, 1) {
			return math.Inf(1)
		}
		return a + b // propagate NaN significand bits
	}

	if b > a {
		a, b = b, a
	}

	ha := fdlibmHI(a)
	hb := fdlibmHI(b)

	if (ha - hb) > 0x3c00000 {
		return a + b // x / y > 2**60
	}

	k := int32(0)
	if a > 0x1.00000ffffffffp500 { // a > ~2**500
		// Scale a and b by 2**-600.
		ha -= 0x25800000
		hb -= 0x25800000
		a = float64(a * twoMinus600)
		b = float64(b * twoMinus600)
		k += 600
	}
	var t1, t2 float64
	if b < 0x1.0p-500 { // b < 2**-500
		if b < 0x1.0p-1022 { // subnormal b or 0
			if b == 0.0 {
				return a
			}
			t1 = 0x1.0p1022 // t1 = 2^1022
			b = float64(b * t1)
			a = float64(a * t1)
			k -= 1022
		} else { // scale a and b by 2^600
			ha += 0x25800000 // a *= 2^600
			hb += 0x25800000 // b *= 2^600
			a = float64(a * twoPlus600)
			b = float64(b * twoPlus600)
			k -= 600
		}
	}
	// Medium size a and b.
	w := a - b
	if w > b {
		t1 = fdlibmSetHI(0.0, ha)
		t2 = a - t1
		w = math.Sqrt(float64(t1*t1) - (float64(b*(-b)) - float64(t2*(a+t1))))
	} else {
		a += a
		y1 := fdlibmSetHI(0.0, hb)
		y2 := b - y1
		t1 = fdlibmSetHI(0.0, ha+0x00100000)
		t2 = a - t1
		w = math.Sqrt(float64(t1*y1) - (float64(w*(-w)) - (float64(t1*y2) + float64(t2*b))))
	}
	if k != 0 {
		return float64(mathPowerOfTwoD(k) * w)
	}
	return w
}

// fdlibmPow is FdLibm.Pow.compute.
func fdlibmPow(x, y float64) float64 {
	var z, r, s, t, u, v, w float64
	var i, j, k, n int32

	// y == zero: x**0 = 1
	if y == 0.0 {
		return 1.0
	}

	// +/-NaN return x + y to propagate NaN significands.
	if x != x || y != y {
		return x + y
	}

	yAbs := math.Abs(y)
	xAbs := math.Abs(x)
	// Special values of y.
	switch {
	case y == 2.0:
		return float64(x * x)
	case y == 0.5:
		if x >= -math.MaxFloat64 { // handle x == -infinity later
			return math.Sqrt(x + 0.0) // add 0.0 to properly handle x == -0.0
		}
	case yAbs == 1.0: // y is +/-1
		if y == 1.0 {
			return x
		}
		return 1.0 / x
	case math.IsInf(yAbs, 1): // y is +/-infinity
		switch {
		case xAbs == 1.0:
			return y - y // inf**+/-1 is NaN
		case xAbs > 1.0: // (|x| > 1)**+/-inf = inf, 0
			if y >= 0 {
				return y
			}
			return 0.0
		default: // (|x| < 1)**-/+inf = inf, 0
			if y < 0 {
				return -y
			}
			return 0.0
		}
	}

	hx := fdlibmHI(x)
	ix := hx & fdlibmExpSignifBits

	// When x < 0, determine if y is an odd integer:
	// yIsInt = 0 ... y is not an integer
	// yIsInt = 1 ... y is an odd int
	// yIsInt = 2 ... y is an even int
	yIsInt := int32(0)
	if hx < 0 {
		if yAbs >= 0x1.0p53 { // |y| >= 2^53 = 9.007199254740992E15
			yIsInt = 2 // y is an even integer since ulp(2^53) = 2.0
		} else if yAbs >= 1.0 { // |y| >= 1.0
			yAbsAsLong := F2L(yAbs)
			if float64(yAbsAsLong) == yAbs {
				yIsInt = 2 - int32(yAbsAsLong&0x1)
			}
		}
	}

	// Special value of x.
	if xAbs == 0.0 || math.IsInf(xAbs, 1) || xAbs == 1.0 {
		z = xAbs // x is +/-0, +/-inf, +/-1
		if y < 0.0 {
			z = 1.0 / z // z = (1/|x|)
		}
		if hx < 0 {
			if ((ix - 0x3ff00000) | yIsInt) == 0 {
				z = (z - z) / (z - z) // (-1)**non-int is NaN
			} else if yIsInt == 1 {
				z = float64(-1.0 * z) // (x < 0)**odd = -(|x|**odd)
			}
		}
		return z
	}

	n = (hx >> 31) + 1

	// (x < 0)**(non-int) is NaN
	if (n | yIsInt) == 0 {
		return (x - x) / (x - x)
	}

	s = 1.0 // s (sign of result -ve**odd) = -1 else = 1
	if (n | (yIsInt - 1)) == 0 {
		s = -1.0 // (-ve)**(odd int)
	}

	var p_h, p_l, t1, t2 float64
	// |y| is huge
	if yAbs > 0x1.00000ffffffffp31 { // if |y| > ~2**31
		const (
			INV_LN2   = 0x1.71547652b82fep0   // 1/ln2
			INV_LN2_H = 0x1.715476p0          // 24 bits of 1/ln2
			INV_LN2_L = 0x1.4ae0bf85ddf44p-26 // 1/ln2 tail
		)

		// Over/underflow if x is not close to one.
		if xAbs < 0x1.fffff00000000p-1 { // |x| < ~0.9999995231628418
			if y < 0.0 {
				return float64(s * math.Inf(1))
			}
			return float64(s * 0.0)
		}
		if xAbs > 0x1.00000ffffffffp0 { // |x| > ~1.0
			if y > 0.0 {
				return float64(s * math.Inf(1))
			}
			return float64(s * 0.0)
		}
		// Now |1-x| is tiny <= 2**-20, sufficient to compute log(x) by
		// x - x^2/2 + x^3/3 - x^4/4.
		t = xAbs - 1.0 // t has 20 trailing zeros
		w = float64(float64(t*t) * (0.5 - float64(t*(0.3333333333333333333333-float64(t*0.25)))))
		u = float64(INV_LN2_H * t) // INV_LN2_H has 21 sig. bits
		v = float64(t*INV_LN2_L) - float64(w*INV_LN2)
		t1 = u + v
		t1 = fdlibmSetLO(t1, 0)
		t2 = v - (t1 - u)
	} else {
		const (
			CP   = 0x1.ec709dc3a03fdp-1   // 2/(3ln2)
			CP_H = 0x1.ec709ep-1          // (float)cp
			CP_L = -0x1.e2fe0145b01f5p-28 // tail of CP_H

			// Poly coefs for (3/2)*(log(x)-2s-2/3*s**3
			L1 = 0x1.3333_3333_33303p-1
			L2 = 0x1.b6db_6db6_fabffp-2
			L3 = 0x1.5555_5518_f264dp-2
			L4 = 0x1.1746_0a91_d4101p-2
			L5 = 0x1.d864_a93c_9db65p-3
			L6 = 0x1.a7e2_84a4_54eefp-3
		)
		var z_h, z_l, ss, s2, s_h, s_l, t_h, t_l float64
		n = 0
		// Take care of subnormal numbers.
		if ix < 0x00100000 {
			xAbs = float64(xAbs * 0x1.0p53) // 2^53 = 9007199254740992.0
			n -= 53
			ix = fdlibmHI(xAbs)
		}
		n += ((ix) >> 20) - 0x3ff
		j = ix & 0x000fffff
		// Determine interval.
		ix = j | 0x3ff00000 // normalize ix
		switch {
		case j <= 0x3988E:
			k = 0 // |x| <sqrt(3/2)
		case j < 0xBB67A:
			k = 1 // |x| <sqrt(3)
		default:
			k = 0
			n++
			ix -= 0x00100000
		}
		xAbs = fdlibmSetHI(xAbs, ix)

		// Compute ss = s_h + s_l = (x-1)/(x+1) or (x-1.5)/(x+1.5).

		bp := [2]float64{1.0, 1.5}
		dpH := [2]float64{0.0, 0x1.2b8034p-1}
		dpL := [2]float64{0.0, 0x1.cfdeb43cfd006p-27}

		u = xAbs - bp[k] // BP[0]=1.0, BP[1]=1.5
		v = 1.0 / (xAbs + bp[k])
		ss = float64(u * v)
		s_h = ss
		s_h = fdlibmSetLO(s_h, 0)
		// t_h=xAbs + BP[k] High
		t_h = 0.0
		t_h = fdlibmSetHI(t_h, ((ix>>1)|0x20000000)+0x00080000+(k<<18))
		t_l = xAbs - (t_h - bp[k])
		s_l = float64(v * ((u - float64(s_h*t_h)) - float64(s_h*t_l)))
		// Compute log(xAbs).
		s2 = float64(ss * ss)
		r = float64(float64(s2*s2) * (L1 + float64(s2*(L2+float64(s2*(L3+float64(s2*(L4+float64(s2*(L5+float64(s2*L6)))))))))))
		r += float64(s_l * (s_h + ss))
		s2 = float64(s_h * s_h)
		t_h = 3.0 + s2 + r
		t_h = fdlibmSetLO(t_h, 0)
		t_l = r - ((t_h - 3.0) - s2)
		// u+v = ss*(1+...)
		u = float64(s_h * t_h)
		v = float64(s_l*t_h) + float64(t_l*ss)
		// 2/(3log2)*(ss + ...)
		p_h = u + v
		p_h = fdlibmSetLO(p_h, 0)
		p_l = v - (p_h - u)
		z_h = float64(CP_H * p_h) // CP_H + CP_L = 2/(3*log2)
		z_l = float64(CP_L*p_h) + float64(p_l*CP) + dpL[k]
		// log2(xAbs) = (ss + ..)*2/(3*log2) = n + DP_H + z_h + z_l
		t = float64(n)
		t1 = ((z_h + z_l) + dpH[k]) + t
		t1 = fdlibmSetLO(t1, 0)
		t2 = z_l - (((t1 - t) - dpH[k]) - z_h)
	}

	// Split up y into (y1 + y2) and compute (y1 + y2) * (t1 + t2).
	y1 := y
	y1 = fdlibmSetLO(y1, 0)
	p_l = float64((y-y1)*t1) + float64(y*t2)
	p_h = float64(y1 * t1)
	z = p_l + p_h
	j = fdlibmHI(z)
	i = fdlibmLO(z)
	if j >= 0x40900000 { // z >= 1024
		if ((j - 0x40900000) | i) != 0 { // if z > 1024
			return float64(s * math.Inf(1)) // overflow
		}
		const OVT = 8.0085662595372944372e-0017 // -(1024-log2(ovfl+.5ulp))
		if p_l+OVT > z-p_h {
			return float64(s * math.Inf(1)) // overflow
		}
	} else if (j & fdlibmExpSignifBits) >= 0x4090cc00 { // z <= -1075
		const c090cc00 = -0x3f6f3400   // 0xc090cc00 as a Java int
		if ((j - c090cc00) | i) != 0 { // z < -1075
			return float64(s * 0.0) // underflow
		}
		if p_l <= z-p_h {
			return float64(s * 0.0) // underflow
		}
	}

	// Compute 2**(p_h+p_l).
	const (
		P1    = 0x1.555555555553ep-3
		P2    = -0x1.6c16c16bebd93p-9
		P3    = 0x1.1566aaf25de2cp-14
		P4    = -0x1.bbd41c5d26bf1p-20
		P5    = 0x1.6376972bea4d0p-25
		LG2   = 0x1.62e42fefa39efp-1
		LG2_H = 0x1.62e43p-1
		LG2_L = -0x1.05c610ca86c39p-29
	)
	i = j & fdlibmExpSignifBits
	k = (i >> 20) - 0x3ff
	n = 0
	if i > 0x3fe00000 { // if |z| > 0.5, set n = [z + 0.5]
		n = j + (0x00100000 >> (k + 1))
		k = ((n & fdlibmExpSignifBits) >> 20) - 0x3ff // new k for n
		t = 0.0
		t = fdlibmSetHI(t, n & ^(0x000fffff>>k))
		n = ((n & 0x000fffff) | 0x00100000) >> (20 - k)
		if j < 0 {
			n = -n
		}
		p_h -= t
	}
	t = p_l + p_h
	t = fdlibmSetLO(t, 0)
	u = float64(t * LG2_H)
	v = float64((p_l-(t-p_h))*LG2) + float64(t*LG2_L)
	z = u + v
	w = v - (z - u)
	t = float64(z * z)
	t1 = z - float64(t*(P1+float64(t*(P2+float64(t*(P3+float64(t*(P4+float64(t*P5)))))))))
	r = float64(z*t1)/(t1-2.0) - (w + float64(z*w))
	z = 1.0 - (r - z)
	j = fdlibmHI(z)
	j += n << 20
	if (j >> 20) <= 0 {
		z = Math_scalb_double_int(z, n) // subnormal output
	} else {
		zHi := fdlibmHI(z)
		zHi += n << 20
		z = fdlibmSetHI(z, zHi)
	}
	return float64(s * z)
}

// fdlibmExp is FdLibm.Exp.compute.
func fdlibmExp(x float64) float64 {
	const (
		twom1000    = 0x1.0p-1000
		o_threshold = 0x1.62e42fefa39efp9
		u_threshold = -0x1.74910d52d3051p9
		ln2HI0      = 0x1.62e42feep-1
		ln2LO0      = 0x1.a39ef35793c76p-33
		invln2      = 0x1.71547652b82fep0
		P1          = 0x1.555555555553ep-3
		P2          = -0x1.6c16c16bebd93p-9
		P3          = 0x1.1566aaf25de2cp-14
		P4          = -0x1.bbd41c5d26bf1p-20
		P5          = 0x1.6376972bea4d0p-25
	)
	half := [2]float64{0.5, -0.5}
	ln2HI := [2]float64{ln2HI0, -ln2HI0}
	ln2LO := [2]float64{ln2LO0, -ln2LO0}
	var y, c, t float64
	hi := 0.0
	lo := 0.0
	k := int32(0)

	hx := fdlibmHI(x)
	xsb := (hx >> 31) & 1 // sign bit of x
	hx &= fdlibmExpSignifBits

	// Filter out non-finite argument.
	if hx >= 0x40862E42 { // if |x| >= 709.78...
		if hx >= 0x7ff00000 {
			if ((hx & 0xfffff) | fdlibmLO(x)) != 0 {
				return x + x // NaN
			}
			if xsb == 0 { // exp(+-inf) = {inf, 0}
				return x
			}
			return 0.0
		}
		if x > o_threshold {
			return math.Inf(1) // overflow: huge * huge
		}
		if x < u_threshold {
			return 0.0 // underflow: twom1000 * twom1000
		}
	}

	// Argument reduction.
	switch {
	case hx > 0x3fd62e42: // if  |x| > 0.5 ln2
		if hx < 0x3FF0A2B2 { // and |x| < 1.5 ln2
			hi = x - ln2HI[xsb]
			lo = ln2LO[xsb]
			k = 1 - xsb - xsb
		} else {
			k = F2I(float64(invln2*x) + half[xsb])
			t = float64(k)
			hi = x - float64(t*ln2HI[0]) // t*ln2HI is exact here
			lo = float64(t * ln2LO[0])
		}
		x = hi - lo
	case hx < 0x3e300000: // when |x|<2**-28
		if fdlibmHuge+x > 1.0 {
			return 1.0 + x // trigger inexact
		}
	default:
		k = 0
	}

	// x is now in primary range.
	t = float64(x * x)
	c = x - float64(t*(P1+float64(t*(P2+float64(t*(P3+float64(t*(P4+float64(t*P5)))))))))
	if k == 0 {
		return 1.0 - (float64(x*c)/(c-2.0) - x)
	}
	y = 1.0 - ((lo - float64(x*c)/(2.0-c)) - hi)

	if k >= -1021 {
		y = fdlibmSetHI(y, fdlibmHI(y)+(k<<20)) // add k to y's exponent
		return y
	}
	y = fdlibmSetHI(y, fdlibmHI(y)+((k+1000)<<20)) // add k to y's exponent
	return float64(y * twom1000)
}

// fdlibmLog is FdLibm.Log.compute.
func fdlibmLog(x float64) float64 {
	const (
		ln2_hi = 0x1.62e42feep-1
		ln2_lo = 0x1.a39ef35793c76p-33
		Lg1    = 0x1.5555555555593p-1
		Lg2    = 0x1.999999997fa04p-2
		Lg3    = 0x1.2492494229359p-2
		Lg4    = 0x1.c71c51d8e78afp-3
		Lg5    = 0x1.7466496cb03dep-3
		Lg6    = 0x1.39a09d078c69fp-3
		Lg7    = 0x1.2f112df3e5244p-3
	)
	var hfsq, f, s, z, R, w, t1, t2, dk float64
	var k, i, j int32

	hx := fdlibmHI(x)
	lx := fdlibmLO(x)

	k = 0
	if hx < 0x00100000 { // x < 2**-1022
		if ((hx & fdlibmExpSignifBits) | lx) == 0 { // log(+-0) = -inf
			return math.Inf(-1)
		}
		if hx < 0 { // log(-#) = NaN
			return float64(x-x) / nanZero
		}
		k -= 54
		x = float64(x * fdlibmTwo54) // subnormal number, scale up x
		hx = fdlibmHI(x)
	}
	if hx >= fdlibmExpBits {
		return x + x
	}
	k += (hx >> 20) - 1023
	hx &= 0x000fffff
	i = (hx + 0x95f64) & 0x100000
	x = fdlibmSetHI(x, hx|(i^0x3ff00000)) // normalize x or x/2
	k += i >> 20
	f = x - 1.0
	if (0x000fffff & (2 + hx)) < 3 { // |f| < 2**-20
		if f == 0.0 {
			if k == 0 {
				return 0.0
			}
			dk = float64(k)
			return float64(dk*ln2_hi) + float64(dk*ln2_lo)
		}
		R = float64(float64(f*f) * (0.5 - float64(0.33333333333333333*f)))
		if k == 0 {
			return f - R
		}
		dk = float64(k)
		return float64(dk*ln2_hi) - ((R - float64(dk*ln2_lo)) - f)
	}
	s = f / (2.0 + f)
	dk = float64(k)
	z = float64(s * s)
	i = hx - 0x6147a
	w = float64(z * z)
	j = 0x6b851 - hx
	t1 = float64(w * (Lg2 + float64(w*(Lg4+float64(w*Lg6)))))
	t2 = float64(z * (Lg1 + float64(w*(Lg3+float64(w*(Lg5+float64(w*Lg7)))))))
	i |= j
	R = t2 + t1
	if i > 0 {
		hfsq = float64(float64(0.5*f) * f)
		if k == 0 {
			return f - (hfsq - float64(s*(hfsq+R)))
		}
		return float64(dk*ln2_hi) - ((hfsq - (float64(s*(hfsq+R)) + float64(dk*ln2_lo))) - f)
	}
	if k == 0 {
		return f - float64(s*(f-R))
	}
	return float64(dk*ln2_hi) - ((float64(s*(f-R)) - float64(dk*ln2_lo)) - f)
}

// fdlibmLog10 is FdLibm.Log10.compute.
func fdlibmLog10(x float64) float64 {
	const (
		ivln10    = 0x1.bcb7b1526e50ep-2
		log10_2hi = 0x1.34413509f6p-2
		log10_2lo = 0x1.9fef311f12b36p-42
	)
	hx := fdlibmHI(x)
	lx := fdlibmLO(x)

	k := int32(0)
	if hx < 0x00100000 { // x < 2**-1022
		if ((hx & fdlibmExpSignifBits) | lx) == 0 {
			return math.Inf(-1) // log(+-0)=-inf
		}
		if hx < 0 {
			return float64(x-x) / nanZero // log(-#) = NaN
		}
		k -= 54
		x = float64(x * fdlibmTwo54) // subnormal number, scale up x
		hx = fdlibmHI(x)
	}

	if hx >= fdlibmExpBits {
		return x + x
	}

	k += (hx >> 20) - 1023
	i := int32(uint32(k&fdlibmSignBit) >> 31)
	hx = (hx & 0x000fffff) | ((0x3ff - i) << 20)
	y := float64(k + i)
	x = fdlibmSetHI(x, hx) // replace high word of x with hx
	z := float64(y*log10_2lo) + float64(ivln10*fdlibmLog(x))
	return z + float64(y*log10_2hi)
}

// fdlibmLog1p is FdLibm.Log1p.compute.
func fdlibmLog1p(x float64) float64 {
	const (
		ln2_hi = 0x1.62e42feep-1
		ln2_lo = 0x1.a39ef35793c76p-33
		Lp1    = 0x1.5555555555593p-1
		Lp2    = 0x1.999999997fa04p-2
		Lp3    = 0x1.2492494229359p-2
		Lp4    = 0x1.c71c51d8e78afp-3
		Lp5    = 0x1.7466496cb03dep-3
		Lp6    = 0x1.39a09d078c69fp-3
		Lp7    = 0x1.2f112df3e5244p-3

		bfd2bec3 = -0x402d413d // 0xbfd2_bec3 as a Java int
	)
	var hfsq, f, c, s, z, R, u float64
	var k, hu int32

	hx := fdlibmHI(x)
	ax := hx & fdlibmExpSignifBits

	k = 1
	if hx < 0x3FDA827A { // x < 0.41422
		if ax >= 0x3ff00000 { // x <= -1.0
			if x == -1.0 { // log1p(-1)=-inf
				return math.Inf(-1)
			}
			return math.Float64frombits(0x7ff8000000000000) // log1p(x < -1) = NaN
		}

		if ax < 0x3e200000 { // |x| < 2**-29
			if fdlibmTwo54+x > 0.0 && // raise inexact
				ax < 0x3c900000 { // |x| < 2**-54
				return x
			}
			return x - float64(float64(x*x)*0.5)
		}

		if hx > 0 || hx <= bfd2bec3 { // -0.2929 < x < 0.41422
			k = 0
			f = x
			hu = 1
		}
	}

	if hx >= fdlibmExpBits {
		return x + x
	}

	if k != 0 {
		if hx < 0x43400000 {
			u = 1.0 + x
			hu = fdlibmHI(u)
			k = (hu >> 20) - 1023
			if k > 0 { // correction term
				c = 1.0 - (u - x)
			} else {
				c = x - (u - 1.0)
			}
			c /= u
		} else {
			u = x
			hu = fdlibmHI(u)
			k = (hu >> 20) - 1023
			c = 0
		}
		hu &= 0x000fffff
		if hu < 0x6a09e {
			u = fdlibmSetHI(u, hu|0x3ff00000) // normalize u
		} else {
			k++
			u = fdlibmSetHI(u, hu|0x3fe00000) // normalize u/2
			hu = (0x00100000 - hu) >> 2
		}
		f = u - 1.0
	}

	hfsq = float64(float64(0.5*f) * f)
	dk := float64(k)
	if hu == 0 { // |f| < 2**-20
		if f == 0.0 {
			if k == 0 {
				return 0.0
			}
			c += float64(dk * ln2_lo)
			return float64(dk*ln2_hi) + c
		}
		R = float64(hfsq * (1.0 - float64(0.66666666666666666*f)))
		if k == 0 {
			return f - R
		}
		return float64(dk*ln2_hi) - ((R - (float64(dk*ln2_lo) + c)) - f)
	}
	s = f / (2.0 + f)
	z = float64(s * s)
	R = float64(z * (Lp1 + float64(z*(Lp2+float64(z*(Lp3+float64(z*(Lp4+float64(z*(Lp5+float64(z*(Lp6+float64(z*Lp7)))))))))))))
	if k == 0 {
		return f - (hfsq - float64(s*(hfsq+R)))
	}
	return float64(dk*ln2_hi) - ((hfsq - (float64(s*(hfsq+R)) + (float64(dk*ln2_lo) + c))) - f)
}

// fdlibmExpm1 is FdLibm.Expm1.compute.
func fdlibmExpm1(x float64) float64 {
	const (
		tiny        = 1.0e-300
		o_threshold = 0x1.62e42fefa39efp9
		ln2_hi      = 0x1.62e42feep-1
		ln2_lo      = 0x1.a39ef35793c76p-33
		invln2      = 0x1.71547652b82fep0
		Q1          = -0x1.11111111110f4p-5
		Q2          = 0x1.a01a019fe5585p-10
		Q3          = -0x1.4ce199eaadbb7p-14
		Q4          = 0x1.0cfca86e65239p-18
		Q5          = -0x1.afdb76e09c32dp-23
	)
	var y, hi, lo, c, t, e, hxs, hfx, r1 float64
	var k int32

	hx := fdlibmHI(x)
	xsb := hx & fdlibmSignBit // sign bit of x
	hx &= fdlibmExpSignifBits // high word of |x|

	// Filter out huge and non-finite argument.
	if hx >= 0x4043687A { // if |x| >= 56*ln2
		if hx >= 0x40862E42 { // if |x| >= 709.78...
			if hx >= 0x7ff00000 {
				if ((hx & 0xfffff) | fdlibmLO(x)) != 0 {
					return x + x // NaN
				}
				if xsb == 0 { // exp(+-inf)={inf,-1}
					return x
				}
				return -1.0
			}
			if x > o_threshold {
				return math.Inf(1) // overflow: huge*huge
			}
		}
		if xsb != 0 { // x < -56*ln2, return -1.0 with inexact
			if x+tiny < 0.0 { // raise inexact
				return tiny - 1.0 // return -1
			}
		}
	}

	// Argument reduction.
	switch {
	case hx > 0x3fd62e42: // if  |x| > 0.5 ln2
		if hx < 0x3FF0A2B2 { // and |x| < 1.5 ln2
			if xsb == 0 {
				hi = x - ln2_hi
				lo = ln2_lo
				k = 1
			} else {
				hi = x + ln2_hi
				lo = -ln2_lo
				k = -1
			}
		} else {
			if xsb == 0 {
				k = F2I(float64(invln2*x) + 0.5)
			} else {
				k = F2I(float64(invln2*x) + -0.5)
			}
			t = float64(k)
			hi = x - float64(t*ln2_hi) // t*ln2_hi is exact here
			lo = float64(t * ln2_lo)
		}
		x = hi - lo
		c = (hi - x) - lo
	case hx < 0x3c900000: // when |x| < 2**-54, return x
		t = fdlibmHuge + x // return x with inexact flags when x != 0
		return x - (t - (fdlibmHuge + x))
	default:
		k = 0
	}

	// x is now in primary range.
	hfx = float64(0.5 * x)
	hxs = float64(x * hfx)
	r1 = 1.0 + float64(hxs*(Q1+float64(hxs*(Q2+float64(hxs*(Q3+float64(hxs*(Q4+float64(hxs*Q5)))))))))
	t = 3.0 - float64(r1*hfx)
	e = float64(hxs * ((r1 - t) / (6.0 - float64(x*t))))
	if k == 0 {
		return x - (float64(x*e) - hxs) // c is 0
	}
	e = float64(x*(e-c)) - c
	e -= hxs
	if k == -1 {
		return float64(0.5*(x-e)) - 0.5
	}
	if k == 1 {
		if x < -0.25 {
			return float64(-2.0 * (e - (x + 0.5)))
		}
		return 1.0 + float64(2.0*(x-e))
	}
	if k <= -2 || k > 56 { // suffice to return exp(x) - 1
		y = 1.0 - (e - x)
		y = fdlibmSetHI(y, fdlibmHI(y)+(k<<20)) // add k to y's exponent
		return y - 1.0
	}
	t = 1.0
	if k < 20 {
		t = fdlibmSetHI(t, 0x3ff00000-(0x200000>>k)) // t = 1-2^-k
		y = t - (e - x)
		y = fdlibmSetHI(y, fdlibmHI(y)+(k<<20)) // add k to y's exponent
	} else {
		t = fdlibmSetHI(t, (0x3ff-k)<<20) // 2^-k
		y = x - (e + t)
		y += 1.0
		y = fdlibmSetHI(y, fdlibmHI(y)+(k<<20)) // add k to y's exponent
	}
	return y
}

// fdlibmSinh is FdLibm.Sinh.compute.
func fdlibmSinh(x float64) float64 {
	const shuge = 1.0e307
	var t, w, h float64

	jx := fdlibmHI(x)
	ix := jx & fdlibmExpSignifBits

	// x is INF or NaN.
	if ix >= fdlibmExpBits {
		return x + x
	}

	h = 0.5
	if jx < 0 {
		h = -h
	}
	// |x| in [0,22], return sign(x)*0.5*(E+E/(E+1)))
	if ix < 0x40360000 { // |x| < 22
		if ix < 0x3e300000 { // |x| < 2**-28
			if shuge+x > 1.0 { // sinh(tiny) = tiny with inexact
				return x
			}
		}
		t = fdlibmExpm1(math.Abs(x))
		if ix < 0x3ff00000 {
			return float64(h * (float64(2.0*t) - float64(t*t)/(t+1.0)))
		}
		return float64(h * (t + t/(t+1.0)))
	}

	// |x| in [22, log(maxdouble)] return 0.5*exp(|x|)
	if ix < 0x40862E42 {
		return float64(h * fdlibmExp(math.Abs(x)))
	}

	// |x| in [log(maxdouble), overflowthreshold]
	lx := fdlibmLO(x)
	if ix < 0x408633CE || (ix == 0x408633ce && uint32(lx) <= 0x8fb9f87d) {
		w = fdlibmExp(float64(0.5 * math.Abs(x)))
		t = float64(h * w)
		return float64(t * w)
	}

	// |x| > overflowthreshold, sinh(x) overflow.
	return float64(x * shuge)
}

// fdlibmCosh is FdLibm.Cosh.compute.
func fdlibmCosh(x float64) float64 {
	var t, w float64

	ix := fdlibmHI(x) & fdlibmExpSignifBits

	// x is INF or NaN.
	if ix >= fdlibmExpBits {
		return float64(x * x)
	}

	// |x| in [0,0.5*ln2], return 1+expm1(|x|)^2/(2*exp(|x|))
	if ix < 0x3fd62e43 {
		t = fdlibmExpm1(math.Abs(x))
		w = 1.0 + t
		if ix < 0x3c800000 { // cosh(tiny) = 1
			return w
		}
		return 1.0 + float64(t*t)/(w+w)
	}

	// |x| in [0.5*ln2, 22], return (exp(|x|) + 1/exp(|x|)/2
	if ix < 0x40360000 {
		t = fdlibmExp(math.Abs(x))
		return float64(0.5*t) + 0.5/t
	}

	// |x| in [22, log(maxdouble)] return 0.5*exp(|x|)
	if ix < 0x40862E42 {
		return float64(0.5 * fdlibmExp(math.Abs(x)))
	}

	// |x| in [log(maxdouble), overflowthreshold]
	lx := fdlibmLO(x)
	if ix < 0x408633CE || (ix == 0x408633ce && uint32(lx) <= 0x8fb9f87d) {
		w = fdlibmExp(float64(0.5 * math.Abs(x)))
		t = float64(0.5 * w)
		return float64(t * w)
	}

	// |x| > overflowthreshold, cosh(x) overflow.
	return math.Inf(1)
}

// fdlibmTanh is FdLibm.Tanh.compute.
func fdlibmTanh(x float64) float64 {
	const tiny = 1.0e-300
	var t, z float64

	jx := fdlibmHI(x)
	ix := jx & fdlibmExpSignifBits

	// x is INF or NaN.
	if ix >= fdlibmExpBits {
		if jx >= 0 { // tanh(+-inf)=+-1
			return 1.0/x + 1.0
		}
		return 1.0/x - 1.0 // tanh(NaN) = NaN
	}

	if ix < 0x40360000 { // |x| < 22
		if ix < 0x3c800000 { // |x| < 2**-55
			return float64(x * (1.0 + x)) // tanh(small) = small
		}
		if ix >= 0x3ff00000 { // |x| >= 1
			t = fdlibmExpm1(float64(2.0 * math.Abs(x)))
			z = 1.0 - 2.0/(t+2.0)
		} else {
			t = fdlibmExpm1(float64(-2.0 * math.Abs(x)))
			z = -t / (t + 2.0)
		}
	} else { // |x| > 22, return +-1
		z = 1.0 - tiny // raised inexact flag
	}
	if jx >= 0 {
		return z
	}
	return -z
}

// fdlibmIEEEremainder is FdLibm.IEEEremainder.compute.
func fdlibmIEEEremainder(x, p float64) float64 {
	hx := fdlibmHI(x)
	lx := fdlibmLO(x)
	hp := fdlibmHI(p)
	lp := fdlibmLO(p)
	sx := hx & fdlibmSignBit
	hp &= fdlibmExpSignifBits
	hx &= fdlibmExpSignifBits

	// Purge off exception values.
	if (hp | lp) == 0 { // p = 0
		return float64(x*p) / float64(x*p)
	}
	if (hx >= fdlibmExpBits) || // not finite
		((hp >= fdlibmExpBits) && // p is NaN
			(((hp - fdlibmExpBits) | lp) != 0)) {
		return float64(x*p) / float64(x*p)
	}

	if hp <= 0x7fdfffff { // now x < 2p
		x = fdlibmFmod(x, p+p)
	}
	if ((hx - hp) | (lx - lp)) == 0 {
		return float64(0.0 * x)
	}
	x = math.Abs(x)
	p = math.Abs(p)
	if hp < 0x00200000 {
		if x+x > p {
			x -= p
			if x+x >= p {
				x -= p
			}
		}
	} else {
		pHalf := float64(0.5 * p)
		if x > pHalf {
			x -= p
			if x >= pHalf {
				x -= p
			}
		}
	}
	return fdlibmSetHI(x, fdlibmHI(x)^sx)
}

// fdlibmFmod is FdLibm.IEEEremainder.__ieee754_fmod.
func fdlibmFmod(x, y float64) float64 {
	var n, hz, ix, iy, lz int32

	hx := fdlibmHI(x)
	lx := fdlibmLO(x)
	hy := fdlibmHI(y)
	ly := fdlibmLO(y)
	sx := hx & fdlibmSignBit // sign of x
	hx ^= sx                 // |x|
	hy &= fdlibmExpSignifBits

	// Purge off exception values.
	if (hy|ly) == 0 || (hx >= fdlibmExpBits) || // y = 0, or x not finite
		((hy | int32(uint32(ly|-ly)>>31)) > fdlibmExpBits) { // or y is NaN
		return float64(x*y) / float64(x*y)
	}
	if hx <= hy {
		if (hx < hy) || (uint32(lx) < uint32(ly)) { // |x| < |y| return x
			return x
		}
		if lx == ly {
			return fdlibmSignedZero(sx) // |x| = |y| return x*0
		}
	}

	ix = fdlibmIlogb(hx, lx)
	iy = fdlibmIlogb(hy, ly)

	// Set up {hx, lx}, {hy, ly} and align y to x.
	if ix >= -1022 {
		hx = 0x00100000 | (0x000fffff & hx)
	} else { // subnormal x, shift x to normal
		n = -1022 - ix
		if n <= 31 {
			hx = (hx << n) | int32(uint32(lx)>>(32-n))
			lx <<= n
		} else {
			hx = lx << (n - 32)
			lx = 0
		}
	}
	if iy >= -1022 {
		hy = 0x00100000 | (0x000fffff & hy)
	} else { // subnormal y, shift y to normal
		n = -1022 - iy
		if n <= 31 {
			hy = (hy << n) | int32(uint32(ly)>>(32-n))
			ly <<= n
		} else {
			hy = ly << (n - 32)
			ly = 0
		}
	}

	// Fix point fmod.
	n = ix - iy
	for ; n != 0; n-- {
		hz = hx - hy
		lz = lx - ly
		if uint32(lx) < uint32(ly) {
			hz--
		}
		if hz < 0 {
			hx = hx + hx + int32(uint32(lx)>>31)
			lx += lx
		} else {
			if (hz | lz) == 0 { // return sign(x)*0
				return fdlibmSignedZero(sx)
			}
			hx = hz + hz + int32(uint32(lz)>>31)
			lx = lz + lz
		}
	}
	hz = hx - hy
	lz = lx - ly
	if uint32(lx) < uint32(ly) {
		hz--
	}
	if hz >= 0 {
		hx = hz
		lx = lz
	}

	// Convert back to floating value and restore the sign.
	if (hx | lx) == 0 { // return sign(x)*0
		return fdlibmSignedZero(sx)
	}
	for hx < 0x00100000 { // normalize x
		hx = hx + hx + int32(uint32(lx)>>31)
		lx += lx
		iy--
	}
	if iy >= -1022 { // normalize output
		hx = (hx - 0x00100000) | ((iy + 1023) << 20)
		x = fdlibmHILO(hx|sx, lx)
	} else { // subnormal output
		n = -1022 - iy
		switch {
		case n <= 20:
			lx = int32(uint32(lx)>>n) | (hx << (32 - n))
			hx >>= n
		case n <= 31:
			lx = (hx << (32 - n)) | int32(uint32(lx)>>n)
			hx = sx
		default:
			lx = hx >> (n - 32)
			hx = sx
		}
		x = fdlibmHILO(hx|sx, lx)
		x = float64(x * 1.0) // create necessary signal
	}
	return x // exact output
}

// fdlibmSignedZero returns a zero with the sign of the int argument.
func fdlibmSignedZero(sign int32) float64 {
	return float64(0.0 * float64(sign))
}

// fdlibmIlogb returns the binary exponent of the double with high word hz and
// low word lz.
func fdlibmIlogb(hz, lz int32) int32 {
	var iz, i int32
	if hz < 0x00100000 { // subnormal z
		if hz == 0 {
			for iz, i = -1043, lz; i > 0; i <<= 1 {
				iz--
			}
		} else {
			for iz, i = -1022, hz<<11; i > 0; i <<= 1 {
				iz--
			}
		}
	} else {
		iz = (hz >> 20) - 1023
	}
	return iz
}

// nanZero is a zero divisor that the compiler cannot fold. FdLibm computes NaN
// as (x-x)/zero at run time, giving the hardware's default NaN, whose bits
// differ from Go's math.NaN(). It is never written.
var nanZero float64
