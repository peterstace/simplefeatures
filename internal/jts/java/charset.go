package java

// decodeUTF8 returns the UTF-16 decoding of src, with malformed input
// replaced by U+FFFD exactly as Java's UTF-8 decoder (and new String(bytes,
// UTF_8)) replaces it.
func decodeUTF8(src []int8) []uint16 {
	dst := make([]uint16, len(src))
	_, dp, _ := decodeUTF8Replace(src, 0, len(src), dst, 0, len(dst), true)
	return dst[:dp]
}

const replacementChar = 0xFFFD

func isNotContinuation(b int32) bool { return b&0xc0 != 0x80 }

// decodeUTF8Replace is CharsetDecoder.decode for UTF-8 with
// CodingErrorAction.REPLACE: it decodes src[sp:sl] into dst[dp:dl] and
// returns the new positions and whether it stopped because dst was full.
// Unless endOfInput is set, it stops before an incomplete sequence at the
// end of src, which more input may complete.
func decodeUTF8Replace(src []int8, sp, sl int, dst []uint16, dp, dl int, endOfInput bool) (int, int, bool) {
	for {
		var mal int
		var overflow bool
		sp, dp, mal, overflow = decodeUTF8Loop(src, sp, sl, dst, dp, dl)
		if overflow {
			return sp, dp, true
		}
		if mal == 0 {
			if !endOfInput || sp == sl {
				return sp, dp, false
			}
			mal = sl - sp
		}
		if dp >= dl {
			return sp, dp, true
		}
		dst[dp] = replacementChar
		dp++
		sp += mal
	}
}

// decodeUTF8Loop is sun.nio.cs.UTF_8.Decoder.decodeArrayLoop. It returns the
// new positions and either the length of a malformed sequence at sp, or
// whether it stopped for lack of room in dst, or neither (underflow).
func decodeUTF8Loop(src []int8, sp, sl int, dst []uint16, dp, dl int) (int, int, int, bool) {
	// xflow is UTF_8.Decoder.xflow: overflow unless the input lacks nb
	// bytes.
	xflow := func(nb int) (int, int, int, bool) {
		return sp, dp, 0, nb != 0 && sl-sp >= nb
	}
	for sp < sl {
		b1 := int32(src[sp])
		switch {
		case b1 >= 0:
			if dp >= dl {
				return xflow(1)
			}
			dst[dp] = uint16(b1)
			dp++
			sp++
		case b1>>5 == -2 && b1&0x1e != 0:
			if sl-sp < 2 || dp >= dl {
				return xflow(2)
			}
			b2 := int32(src[sp+1])
			if isNotContinuation(b2) {
				return sp, dp, 1, false
			}
			dst[dp] = uint16((b1 << 6) ^ b2 ^ (int32(int8(-0x40))<<6 ^ int32(int8(-0x80))))
			dp++
			sp += 2
		case b1>>4 == -2:
			if sl-sp < 3 || dp >= dl {
				if sl-sp > 1 && isMalformed3_2(b1, int32(src[sp+1])) {
					return sp, dp, 1, false
				}
				return xflow(3)
			}
			b2, b3 := int32(src[sp+1]), int32(src[sp+2])
			if isMalformed3(b1, b2, b3) {
				if b1 == int32(int8(-0x20)) && b2&0xe0 == 0x80 || isNotContinuation(b2) {
					return sp, dp, 1, false
				}
				return sp, dp, 2, false
			}
			c := uint16((b1 << 12) ^ (b2 << 6) ^ (b3 ^ (int32(int8(-0x20))<<12 ^ int32(int8(-0x80))<<6 ^ int32(int8(-0x80)))))
			if c >= 0xD800 && c <= 0xDFFF {
				return sp, dp, 3, false
			}
			dst[dp] = c
			dp++
			sp += 3
		case b1>>3 == -2:
			if sl-sp < 4 || dl-dp < 2 {
				u1 := b1 & 0xff
				if u1 > 0xf4 || sl-sp > 1 && isMalformed4_2(u1, int32(src[sp+1])&0xff) {
					return sp, dp, 1, false
				}
				if sl-sp > 2 && isNotContinuation(int32(src[sp+2])) {
					return sp, dp, 2, false
				}
				return xflow(4)
			}
			b2, b3, b4 := int32(src[sp+1]), int32(src[sp+2]), int32(src[sp+3])
			uc := (b1 << 18) ^ (b2 << 12) ^ (b3 << 6) ^
				(b4 ^ (int32(int8(-0x10))<<18 ^ int32(int8(-0x80))<<12 ^ int32(int8(-0x80))<<6 ^ int32(int8(-0x80))))
			if isNotContinuation(b2) || isNotContinuation(b3) || isNotContinuation(b4) ||
				uc < 0x10000 || uc >= 0x110000 {
				return sp, dp, malformed4Len(src[sp:]), false
			}
			uc -= 0x10000
			dst[dp] = uint16(0xD800 + uc>>10)
			dst[dp+1] = uint16(0xDC00 + uc&0x3ff)
			dp += 2
			sp += 4
		default:
			return sp, dp, 1, false
		}
	}
	return xflow(0)
}

func isMalformed3(b1, b2, b3 int32) bool {
	return b1 == int32(int8(-0x20)) && b2&0xe0 == 0x80 || b2&0xc0 != 0x80 || b3&0xc0 != 0x80
}

func isMalformed3_2(b1, b2 int32) bool {
	return b1 == int32(int8(-0x20)) && b2&0xe0 == 0x80 || b2&0xc0 != 0x80
}

// isMalformed4_2 takes unsigned bytes.
func isMalformed4_2(b1, b2 int32) bool {
	return b1 == 0xf0 && (b2 < 0x90 || b2 > 0xbf) || b1 == 0xf4 && b2&0xf0 != 0x80 || b2&0xc0 != 0x80
}

func malformed4Len(s []int8) int {
	b1, b2 := int32(s[0])&0xff, int32(s[1])&0xff
	if b1 > 0xf4 || b1 == 0xf0 && (b2 < 0x90 || b2 > 0xbf) || b1 == 0xf4 && b2&0xf0 != 0x80 ||
		isNotContinuation(b2) {
		return 1
	}
	if isNotContinuation(int32(s[2])) {
		return 2
	}
	return 3
}

// utf8Encoder encodes UTF-16 to UTF-8 as Java's StreamEncoder does with
// malformed input replaced by '?'. A high surrogate at the end of one write
// is held until the next write shows whether a low surrogate follows.
type utf8Encoder struct {
	pendingHigh uint16
	hasPending  bool
}

func (e *utf8Encoder) encode(dst []byte, units []uint16) []byte {
	for i := 0; i < len(units); i++ {
		c := units[i]
		if e.hasPending {
			e.hasPending = false
			if c >= 0xDC00 && c <= 0xDFFF {
				dst = appendRune4(dst, e.pendingHigh, c)
				continue
			}
			dst = append(dst, '?')
		}
		switch {
		case c < 0x80:
			dst = append(dst, byte(c))
		case c < 0x800:
			dst = append(dst, byte(0xc0|c>>6), byte(0x80|c&0x3f))
		case c >= 0xD800 && c <= 0xDBFF:
			if i+1 == len(units) {
				e.pendingHigh, e.hasPending = c, true
			} else if d := units[i+1]; d >= 0xDC00 && d <= 0xDFFF {
				dst = appendRune4(dst, c, d)
				i++
			} else {
				dst = append(dst, '?')
			}
		case c >= 0xDC00 && c <= 0xDFFF:
			dst = append(dst, '?')
		default:
			dst = append(dst, byte(0xe0|c>>12), byte(0x80|c>>6&0x3f), byte(0x80|c&0x3f))
		}
	}
	return dst
}

// finish encodes a held high surrogate, which is malformed at the end of the
// input.
func (e *utf8Encoder) finish(dst []byte) []byte {
	if e.hasPending {
		e.hasPending = false
		dst = append(dst, '?')
	}
	return dst
}

func appendRune4(dst []byte, hi, lo uint16) []byte {
	uc := (int32(hi)-0xD800)<<10 + (int32(lo) - 0xDC00) + 0x10000
	return append(dst, byte(0xf0|uc>>18), byte(0x80|uc>>12&0x3f), byte(0x80|uc>>6&0x3f), byte(0x80|uc&0x3f))
}
