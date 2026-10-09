package java

import "fmt"

// Writer is java.io.Writer. See io_reader.go for how calls dispatch.
type Writer struct {
	jself any
	jhash int32
}

// writerImpl holds the Writer methods that subclasses override.
type writerImpl interface {
	writeChar(c int32)
	writeArr(cbuf []uint16, off, n int32)
	writeStr(s *String)
	writeStrRange(s *String, off, n int32)
	appendCsq(csq any) any
	appendCsqRange(csq any, start, end int32) any
	appendChar(c uint16) any
	flush()
	close()
	toStr() *String
}

func (this *Writer) JSelf() any             { return this.jself }
func (this *Writer) JSetSelf(self any)      { this.jself = self }
func (this *Writer) JClass_Writer() *Writer { return this }
func (this *Writer) JClassName() string     { return "java.io.Writer" }
func (this *Writer) JIdentityHash() *int32  { return &this.jhash }

func Writer_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_Writer() *Writer })
	return ok
}

func Writer_fromAny(o any) *Writer {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_Writer() *Writer }](o).JClass_Writer()
}

func (this *Writer) Writer_init() {}

func (this *Writer) Writer_init_Object(lock any) {
	if lock == nil {
		panic(NullPointerException_new())
	}
}

func (this *Writer) impl() writerImpl {
	if v, ok := this.jself.(writerImpl); ok {
		return v
	}
	return this
}

func (this *Writer) Write_int(c int32) { this.impl().writeChar(c) }

func (this *Writer) Write_charArr(cbuf []uint16) {
	requireArray(cbuf)
	this.impl().writeArr(cbuf, 0, int32(len(cbuf)))
}

func (this *Writer) Write_charArr_int_int(cbuf []uint16, off, n int32) {
	this.impl().writeArr(cbuf, off, n)
}

func (this *Writer) Write_String(s *String) { this.impl().writeStr(s) }

func (this *Writer) Write_String_int_int(s *String, off, n int32) {
	this.impl().writeStrRange(s, off, n)
}

// Append_CharSequence implements Writer.append(CharSequence). The result is
// the writer, typed as the root declaration Appendable.append returns.
func (this *Writer) Append_CharSequence(csq any) any { return this.impl().appendCsq(csq) }

func (this *Writer) Append_CharSequence_int_int(csq any, start, end int32) any {
	return this.impl().appendCsqRange(csq, start, end)
}

func (this *Writer) Append_char(c uint16) any { return this.impl().appendChar(c) }
func (this *Writer) Flush()                   { this.impl().flush() }
func (this *Writer) Close()                   { this.impl().close() }
func (this *Writer) ToString() *String        { return this.impl().toStr() }

func (this *Writer) writeChar(c int32) {
	this.impl().writeArr([]uint16{uint16(c)}, 0, 1)
}

// writeArr is abstract in Writer; only a generated subclass reaches it.
func (this *Writer) writeArr(cbuf []uint16, off, n int32) {
	g, ok := this.jself.(interface {
		Write_charArr_int_int([]uint16, int32, int32)
	})
	if !ok {
		panic(AbstractMethodError("java.io.Writer.write(char[], int, int)"))
	}
	g.Write_charArr_int_int(cbuf, off, n)
}

func (this *Writer) writeStr(s *String) {
	if s == nil {
		panic(NullPointerException_new())
	}
	this.impl().writeStrRange(s, 0, s.Length())
}

func (this *Writer) writeStrRange(s *String, off, n int32) {
	if s == nil {
		panic(NullPointerException_new())
	}
	// Java copies through a 1024-char buffer, which makes a negative length
	// fail in String.getChars rather than in an array allocation.
	size := n
	if size <= 1024 {
		size = 1024
	}
	cbuf := make([]uint16, size)
	getChars(s, off, off+n, cbuf, 0)
	this.impl().writeArr(cbuf, 0, n)
}

func (this *Writer) appendCsq(csq any) any {
	this.impl().writeStr(StrOfObject(csq))
	return Ref(this)
}

func (this *Writer) appendCsqRange(csq any, start, end int32) any {
	return this.impl().appendCsq(subSequence(csq, start, end))
}

func (this *Writer) appendChar(c uint16) any {
	this.impl().writeChar(int32(c))
	return Ref(this)
}

// flush is abstract in Writer; only a generated subclass reaches it.
func (this *Writer) flush() {
	g, ok := this.jself.(interface{ Flush() })
	if !ok {
		panic(AbstractMethodError("java.io.Writer.flush()"))
	}
	g.Flush()
}

// close is abstract in Writer; only a generated subclass reaches it.
func (this *Writer) close() {
	g, ok := this.jself.(interface{ Close() })
	if !ok {
		panic(AbstractMethodError("java.io.Writer.close()"))
	}
	g.Close()
}

func (this *Writer) toStr() *String { return ObjectImpl_ToString(Ref(this)) }

// getChars implements String.getChars, with its bounds checks.
func getChars(s *String, begin, end int32, dst []uint16, dstBegin int32) {
	checkStringRange(begin, end, s.Length())
	if dstBegin|(end-begin)|int32(len(dst)) < 0 || end-begin > int32(len(dst))-dstBegin {
		panic(StringIndexOutOfBoundsException_new_String(FromGo(fmt.Sprintf(
			"Range [%d, %d + %d) out of bounds for length %d", dstBegin, dstBegin, end-begin, len(dst)))))
	}
	copy(dst[dstBegin:], s.units[begin:end])
}

// checkStringRange implements String's checkBoundsBeginEnd.
func checkStringRange(begin, end, length int32) {
	if begin < 0 || begin > end || end > length {
		panic(StringIndexOutOfBoundsException_new_String(FromGo(fmt.Sprintf(
			"Range [%d, %d) out of bounds for length %d", begin, end, length))))
	}
}

// subSequence implements CharSequence.subSequence for the CharSequence
// argument of Writer.append and PrintStream.append, which treat null as
// "null".
func subSequence(csq any, start, end int32) *String {
	s := Lit("null")
	if csq != nil {
		s = Object_ToString(csq)
	}
	checkStringRange(start, end, s.Length())
	return FromUnits(append([]uint16(nil), s.units[start:end]...))
}

// ---------------------------------------------------------------------------
// StringWriter

// StringWriter is java.io.StringWriter.
type StringWriter struct {
	Writer
	buf []uint16
}

func (this *StringWriter) JClass_StringWriter() *StringWriter { return this }
func (this *StringWriter) JClassName() string                 { return "java.io.StringWriter" }

func StringWriter_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_StringWriter() *StringWriter })
	return ok
}

func StringWriter_fromAny(o any) *StringWriter {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_StringWriter() *StringWriter }](o).JClass_StringWriter()
}

func StringWriter_new() *StringWriter {
	this := &StringWriter{}
	this.jself = this
	this.StringWriter_init()
	return this
}

func StringWriter_new_int(initialSize int32) *StringWriter {
	this := &StringWriter{}
	this.jself = this
	this.StringWriter_init_int(initialSize)
	return this
}

func (this *StringWriter) StringWriter_init() {
	this.Writer.Writer_init()
	this.buf = make([]uint16, 0, 16)
}

func (this *StringWriter) StringWriter_init_int(initialSize int32) {
	this.Writer.Writer_init()
	if initialSize < 0 {
		panic(IllegalArgumentException_new_String(Lit("Negative buffer size")))
	}
	this.buf = make([]uint16, 0, initialSize)
}

func (this *StringWriter) writeChar(c int32) {
	this.buf = append(this.buf, uint16(c))
}

func (this *StringWriter) writeArr(cbuf []uint16, off, n int32) {
	requireArray(cbuf)
	checkFromIndexSize(off, n, int32(len(cbuf)))
	this.buf = append(this.buf, cbuf[off:off+n]...)
}

// writeStr appends s, or "null" for a null s, as StringBuffer.append does.
func (this *StringWriter) writeStr(s *String) {
	if s == nil {
		s = Lit("null")
	}
	this.buf = append(this.buf, s.units...)
}

// writeStrRange appends s[off:off+n] with StringBuffer.append(CharSequence,
// int, int)'s null handling and bounds checks.
func (this *StringWriter) writeStrRange(s *String, off, n int32) {
	if s == nil {
		s = Lit("null")
	}
	end := off + n
	if off < 0 || off > end || end > s.Length() {
		panic(IndexOutOfBoundsException_new_String(FromGo(fmt.Sprintf(
			"Range [%d, %d) out of bounds for length %d", off, end, s.Length()))))
	}
	this.buf = append(this.buf, s.units[off:end]...)
}

func (this *StringWriter) appendCsq(csq any) any {
	this.writeStr(StrOfObject(csq))
	return Ref(this)
}

func (this *StringWriter) appendCsqRange(csq any, start, end int32) any {
	return this.impl().appendCsq(subSequence(csq, start, end))
}

func (this *StringWriter) appendChar(c uint16) any {
	this.impl().writeChar(int32(c))
	return Ref(this)
}

func (this *StringWriter) flush() {}
func (this *StringWriter) close() {}

func (this *StringWriter) toStr() *String {
	return FromUnits(append([]uint16(nil), this.buf...))
}
