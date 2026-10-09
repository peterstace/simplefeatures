package java

import (
	"math"
	"sync"
)

// Random is java.util.Random: the 48-bit linear congruential generator, with
// every derived method computed exactly as JDK 21 computes it.
type Random struct {
	jhash                int32
	mu                   sync.Mutex
	seed                 int64
	nextNextGaussian     float64
	haveNextNextGaussian bool
}

const (
	randomMultiplier = 0x5DEECE66D
	randomAddend     = 0xB
	randomMask       = 1<<48 - 1
)

func (this *Random) JClass_Random() *Random { return this }
func (this *Random) JClassName() string     { return "java.util.Random" }
func (this *Random) JIdentityHash() *int32  { return &this.jhash }

func Random_isInstance(o any) bool { _, ok := o.(*Random); return ok }

func Random_fromAny(o any) *Random { return AnyTo[*Random](o) }

var (
	seedUniquifierMu sync.Mutex
	// seedUniquifier is Random's static seedUniquifier, advanced each time a
	// Random is created without a seed.
	seedUniquifier int64 = 8682522807148012
)

func nextSeedUniquifier() int64 {
	seedUniquifierMu.Lock()
	defer seedUniquifierMu.Unlock()
	seedUniquifier *= 1181783497276652981
	return seedUniquifier
}

// Random_new implements new Random(), seeded from the clock as Java does.
func Random_new() *Random {
	return Random_new_long(nextSeedUniquifier() ^ System_nanoTime())
}

func Random_new_long(seed int64) *Random {
	return &Random{seed: initialScramble(seed)}
}

func initialScramble(seed int64) int64 {
	return (seed ^ randomMultiplier) & randomMask
}

func (this *Random) SetSeed_long(seed int64) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.seed = initialScramble(seed)
	this.haveNextNextGaussian = false
}

// next advances the generator and returns its top bits. The caller holds mu.
func (this *Random) next(bits int32) int32 {
	this.seed = (this.seed*randomMultiplier + randomAddend) & randomMask
	return int32(uint64(this.seed) >> ((48 - bits) & 63))
}

// Next_int implements the protected Random.next(int).
func (this *Random) Next_int(bits int32) int32 {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.next(bits)
}

func (this *Random) NextBytes_byteArr(bytes []int8) {
	this.mu.Lock()
	defer this.mu.Unlock()
	for i, n := 0, len(bytes); i < n; {
		rnd := this.next(32)
		k := n - i
		if k > 4 {
			k = 4
		}
		for ; k > 0; k-- {
			bytes[i] = int8(rnd)
			i++
			rnd >>= 8
		}
	}
}

func (this *Random) NextInt() int32 {
	return this.Next_int(32)
}

func (this *Random) NextInt_int(bound int32) int32 {
	if bound <= 0 {
		panic(IllegalArgumentException_new_String(Lit("bound must be positive")))
	}
	this.mu.Lock()
	defer this.mu.Unlock()
	r := this.next(31)
	m := bound - 1
	if bound&m == 0 {
		return int32((int64(bound) * int64(r)) >> 31)
	}
	for u := r; ; u = this.next(31) {
		r = u % bound
		if u-r+m >= 0 {
			break
		}
	}
	return r
}

func (this *Random) NextLong() int64 {
	this.mu.Lock()
	defer this.mu.Unlock()
	hi := int64(this.next(32))
	return hi<<32 + int64(this.next(32))
}

func (this *Random) NextBoolean() bool {
	return this.Next_int(1) != 0
}

func (this *Random) NextFloat() float32 {
	return float32(this.Next_int(24)) * 0x1.0p-24
}

func (this *Random) NextDouble() float64 {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.nextDouble()
}

func (this *Random) nextDouble() float64 {
	hi := int64(this.next(26))
	return float64(hi<<27+int64(this.next(27))) * 0x1.0p-53
}

// NextGaussian implements Random.nextGaussian (the polar method), keeping the
// second value of each pair for the next call.
func (this *Random) NextGaussian() float64 {
	this.mu.Lock()
	defer this.mu.Unlock()
	if this.haveNextNextGaussian {
		this.haveNextNextGaussian = false
		return this.nextNextGaussian
	}
	var v1, v2, s float64
	for {
		v1 = float64(2*this.nextDouble()) - 1
		v2 = float64(2*this.nextDouble()) - 1
		s = float64(v1*v1) + float64(v2*v2)
		if s < 1 && s != 0 {
			break
		}
	}
	multiplier := math.Sqrt(float64(-2*StrictMath_log(s)) / s)
	this.nextNextGaussian = float64(v2 * multiplier)
	this.haveNextNextGaussian = true
	return float64(v1 * multiplier)
}
