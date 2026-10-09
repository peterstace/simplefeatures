package java

import (
	"fmt"
)

// The java.io class hierarchies (Reader, Writer, InputStream, OutputStream)
// follow one dispatch scheme. The abstract base class records the whole
// object in jself, as Throwable does, and each Java-named method of the base
// class dispatches through jself to an unexported method that each runtime
// subclass overrides. A call through a pointer to any class in the hierarchy
// therefore reaches the implementation of the object's most derived runtime
// class. A runtime subclass calls super by naming the embedded class's
// unexported method, and never redeclares a Java-named method of its
// superclass.
//
// A generated class that extends one of these classes implements the
// abstract methods under their Java names, and the unexported base
// implementation of an abstract method forwards to them. A generated
// override of a method that is not abstract is reached only through the
// generated class's own type.

// checkFromIndexSize implements Objects.checkFromIndexSize.
func checkFromIndexSize(off, size, length int32) {
	if length|off|size < 0 || size > length-off {
		panic(IndexOutOfBoundsException_new_String(FromGo(fmt.Sprintf(
			"Range [%d, %d + %d) out of bounds for length %d", off, off, size, length))))
	}
}

func requireArray[T any](a []T) {
	if a == nil {
		panic(NullPointerException_new())
	}
}

func lesserInt32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

// ---------------------------------------------------------------------------
// Reader

// Reader is java.io.Reader.
type Reader struct {
	jself      any
	jhash      int32
	skipBuffer []uint16
}

// readerImpl holds the Reader methods that subclasses override.
type readerImpl interface {
	read() int32
	readArr(cbuf []uint16, off, n int32) int32
	skip(n int64) int64
	ready() bool
	markSupported() bool
	mark(readAheadLimit int32)
	reset()
	close()
	toStr() *String
}

func (this *Reader) JSelf() any             { return this.jself }
func (this *Reader) JSetSelf(self any)      { this.jself = self }
func (this *Reader) JClass_Reader() *Reader { return this }
func (this *Reader) JClassName() string     { return "java.io.Reader" }
func (this *Reader) JIdentityHash() *int32  { return &this.jhash }

func Reader_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_Reader() *Reader })
	return ok
}

func Reader_fromAny(o any) *Reader {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_Reader() *Reader }](o).JClass_Reader()
}

func (this *Reader) Reader_init() {}

func (this *Reader) Reader_init_Object(lock any) {
	if lock == nil {
		panic(NullPointerException_new())
	}
}

func (this *Reader) impl() readerImpl {
	if v, ok := this.jself.(readerImpl); ok {
		return v
	}
	return this
}

func (this *Reader) Read() int32 { return this.impl().read() }

func (this *Reader) Read_charArr(cbuf []uint16) int32 {
	requireArray(cbuf)
	return this.impl().readArr(cbuf, 0, int32(len(cbuf)))
}

func (this *Reader) Read_charArr_int_int(cbuf []uint16, off, n int32) int32 {
	return this.impl().readArr(cbuf, off, n)
}

func (this *Reader) Skip_long(n int64) int64       { return this.impl().skip(n) }
func (this *Reader) Ready() bool                   { return this.impl().ready() }
func (this *Reader) MarkSupported() bool           { return this.impl().markSupported() }
func (this *Reader) Mark_int(readAheadLimit int32) { this.impl().mark(readAheadLimit) }
func (this *Reader) Reset()                        { this.impl().reset() }
func (this *Reader) Close()                        { this.impl().close() }
func (this *Reader) ToString() *String             { return this.impl().toStr() }

func (this *Reader) read() int32 {
	cb := make([]uint16, 1)
	if this.impl().readArr(cb, 0, 1) == -1 {
		return -1
	}
	return int32(cb[0])
}

// readArr is abstract in Reader; only a generated subclass reaches it.
func (this *Reader) readArr(cbuf []uint16, off, n int32) int32 {
	g, ok := this.jself.(interface {
		Read_charArr_int_int([]uint16, int32, int32) int32
	})
	if !ok {
		panic(AbstractMethodError("java.io.Reader.read(char[], int, int)"))
	}
	return g.Read_charArr_int_int(cbuf, off, n)
}

func (this *Reader) skip(n int64) int64 {
	if n < 0 {
		panic(IllegalArgumentException_new_String(Lit("skip value is negative")))
	}
	nn := int32(8192)
	if n < int64(nn) {
		nn = int32(n)
	}
	if len(this.skipBuffer) < int(nn) {
		this.skipBuffer = make([]uint16, nn)
	}
	r := n
	for r > 0 {
		k := nn
		if r < int64(k) {
			k = int32(r)
		}
		nc := this.impl().readArr(this.skipBuffer, 0, k)
		if nc == -1 {
			break
		}
		r -= int64(nc)
	}
	return n - r
}

func (this *Reader) ready() bool         { return false }
func (this *Reader) markSupported() bool { return false }

func (this *Reader) mark(int32) {
	panic(IOException_new_String(Lit("mark() not supported")))
}

func (this *Reader) reset() {
	panic(IOException_new_String(Lit("reset() not supported")))
}

// close is abstract in Reader; only a generated subclass reaches it.
func (this *Reader) close() {
	g, ok := this.jself.(interface{ Close() })
	if !ok {
		panic(AbstractMethodError("java.io.Reader.close()"))
	}
	g.Close()
}

func (this *Reader) toStr() *String { return ObjectImpl_ToString(Ref(this)) }

// ---------------------------------------------------------------------------
// StringReader

// StringReader is java.io.StringReader.
type StringReader struct {
	Reader
	length int32
	str    *String
	next   int32
	mrk    int32
}

func (this *StringReader) JClass_StringReader() *StringReader { return this }
func (this *StringReader) JClassName() string                 { return "java.io.StringReader" }

func StringReader_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_StringReader() *StringReader })
	return ok
}

func StringReader_fromAny(o any) *StringReader {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_StringReader() *StringReader }](o).JClass_StringReader()
}

func StringReader_new(s *String) *StringReader {
	this := &StringReader{}
	this.jself = this
	this.StringReader_init(s)
	return this
}

func (this *StringReader) StringReader_init(s *String) {
	this.Reader.Reader_init()
	if s == nil {
		panic(NullPointerException_new())
	}
	this.length = s.Length()
	this.str = s
}

func (this *StringReader) ensureOpen() {
	if this.str == nil {
		panic(IOException_new_String(Lit("Stream closed")))
	}
}

func (this *StringReader) read() int32 {
	this.ensureOpen()
	if this.next >= this.length {
		return -1
	}
	c := this.str.units[this.next]
	this.next++
	return int32(c)
}

func (this *StringReader) readArr(cbuf []uint16, off, n int32) int32 {
	this.ensureOpen()
	requireArray(cbuf)
	checkFromIndexSize(off, n, int32(len(cbuf)))
	if n == 0 {
		return 0
	}
	if this.next >= this.length {
		return -1
	}
	k := lesserInt32(this.length-this.next, n)
	copy(cbuf[off:off+k], this.str.units[this.next:this.next+k])
	this.next += k
	return k
}

func (this *StringReader) skip(n int64) int64 {
	this.ensureOpen()
	if this.next >= this.length {
		return 0
	}
	r := int64(this.length - this.next)
	if n < r {
		r = n
	}
	if r < int64(-this.next) {
		r = int64(-this.next)
	}
	this.next += int32(r)
	return r
}

func (this *StringReader) ready() bool {
	this.ensureOpen()
	return true
}

func (this *StringReader) markSupported() bool { return true }

func (this *StringReader) mark(readAheadLimit int32) {
	if readAheadLimit < 0 {
		panic(IllegalArgumentException_new_String(Lit("Read-ahead limit < 0")))
	}
	this.ensureOpen()
	this.mrk = this.next
}

func (this *StringReader) reset() {
	this.ensureOpen()
	this.next = this.mrk
}

func (this *StringReader) close() { this.str = nil }

// ---------------------------------------------------------------------------
// BufferedReader

const (
	brInvalidated = -2
	brUnmarked    = -1
)

// BufferedReader is java.io.BufferedReader.
type BufferedReader struct {
	Reader
	in             *Reader
	cb             []uint16
	nChars         int32
	nextChar       int32
	markedChar     int32
	readAheadLimit int32
	skipLF         bool
	markedSkipLF   bool
}

// bufferedReaderImpl holds the BufferedReader methods that subclasses
// override.
type bufferedReaderImpl interface {
	readLine() *String
}

func (this *BufferedReader) JClass_BufferedReader() *BufferedReader { return this }
func (this *BufferedReader) JClassName() string                     { return "java.io.BufferedReader" }

func BufferedReader_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_BufferedReader() *BufferedReader })
	return ok
}

func BufferedReader_fromAny(o any) *BufferedReader {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_BufferedReader() *BufferedReader }](o).JClass_BufferedReader()
}

func BufferedReader_new_Reader(in *Reader) *BufferedReader {
	this := &BufferedReader{}
	this.jself = this
	this.BufferedReader_init_Reader(in)
	return this
}

func BufferedReader_new_Reader_int(in *Reader, sz int32) *BufferedReader {
	this := &BufferedReader{}
	this.jself = this
	this.BufferedReader_init_Reader_int(in, sz)
	return this
}

func (this *BufferedReader) BufferedReader_init_Reader(in *Reader) {
	this.BufferedReader_init_Reader_int(in, 8192)
}

func (this *BufferedReader) BufferedReader_init_Reader_int(in *Reader, sz int32) {
	this.Reader.Reader_init()
	if in == nil {
		panic(NullPointerException_new())
	}
	if sz <= 0 {
		panic(IllegalArgumentException_new_String(Lit("Buffer size <= 0")))
	}
	this.in = in
	this.cb = make([]uint16, sz)
	this.markedChar = brUnmarked
}

func (this *BufferedReader) brImpl() bufferedReaderImpl {
	if v, ok := this.jself.(bufferedReaderImpl); ok {
		return v
	}
	return this
}

// ReadLine implements BufferedReader.readLine. A line ends at "\n", "\r",
// "\r\n" or the end of the input.
func (this *BufferedReader) ReadLine() *String { return this.brImpl().readLine() }

func (this *BufferedReader) ensureOpen() {
	if this.in == nil {
		panic(IOException_new_String(Lit("Stream closed")))
	}
}

func (this *BufferedReader) fill() {
	var dst int32
	if this.markedChar > brUnmarked {
		delta := this.nextChar - this.markedChar
		if delta >= this.readAheadLimit {
			this.markedChar = brInvalidated
			this.readAheadLimit = 0
			dst = 0
		} else {
			if this.readAheadLimit <= int32(len(this.cb)) {
				copy(this.cb[0:delta], this.cb[this.markedChar:this.markedChar+delta])
			} else {
				ncb := make([]uint16, this.readAheadLimit)
				copy(ncb[0:delta], this.cb[this.markedChar:this.markedChar+delta])
				this.cb = ncb
			}
			this.markedChar = 0
			dst = delta
			this.nextChar = delta
			this.nChars = delta
		}
	}
	var n int32
	for {
		n = this.in.Read_charArr_int_int(this.cb, dst, int32(len(this.cb))-dst)
		if n != 0 {
			break
		}
	}
	if n > 0 {
		this.nChars = dst + n
		this.nextChar = dst
	}
}

func (this *BufferedReader) read() int32 {
	this.ensureOpen()
	for {
		if this.nextChar >= this.nChars {
			this.fill()
			if this.nextChar >= this.nChars {
				return -1
			}
		}
		if this.skipLF {
			this.skipLF = false
			if this.cb[this.nextChar] == '\n' {
				this.nextChar++
				continue
			}
		}
		c := this.cb[this.nextChar]
		this.nextChar++
		return int32(c)
	}
}

func (this *BufferedReader) read1(cbuf []uint16, off, n int32) int32 {
	if this.nextChar >= this.nChars {
		if n >= int32(len(this.cb)) && this.markedChar <= brUnmarked && !this.skipLF {
			return this.in.Read_charArr_int_int(cbuf, off, n)
		}
		this.fill()
	}
	if this.nextChar >= this.nChars {
		return -1
	}
	if this.skipLF {
		this.skipLF = false
		if this.cb[this.nextChar] == '\n' {
			this.nextChar++
			if this.nextChar >= this.nChars {
				this.fill()
			}
			if this.nextChar >= this.nChars {
				return -1
			}
		}
	}
	k := lesserInt32(n, this.nChars-this.nextChar)
	copy(cbuf[off:off+k], this.cb[this.nextChar:this.nextChar+k])
	this.nextChar += k
	return k
}

func (this *BufferedReader) readArr(cbuf []uint16, off, n int32) int32 {
	this.ensureOpen()
	requireArray(cbuf)
	checkFromIndexSize(off, n, int32(len(cbuf)))
	if n == 0 {
		return 0
	}
	k := this.read1(cbuf, off, n)
	if k <= 0 {
		return k
	}
	for k < n && this.in.Ready() {
		k1 := this.read1(cbuf, off+k, n-k)
		if k1 <= 0 {
			break
		}
		k += k1
	}
	return k
}

func (this *BufferedReader) readLine() *String {
	return this.readLineTerm(false, nil)
}

// readLineTerm is BufferedReader's package-private readLine(boolean,
// boolean[]): ignoreLF skips a leading '\n', and term[0] records whether the
// line ended with a terminator.
func (this *BufferedReader) readLineTerm(ignoreLF bool, term []bool) *String {
	var s []uint16
	haveS := false
	this.ensureOpen()
	omitLF := ignoreLF || this.skipLF
	if term != nil {
		term[0] = false
	}
	for {
		if this.nextChar >= this.nChars {
			this.fill()
		}
		if this.nextChar >= this.nChars {
			if haveS && len(s) > 0 {
				return FromUnits(s)
			}
			return nil
		}
		eol := false
		var c uint16
		if omitLF && this.cb[this.nextChar] == '\n' {
			this.nextChar++
		}
		this.skipLF = false
		omitLF = false
		i := this.nextChar
		for ; i < this.nChars; i++ {
			c = this.cb[i]
			if c == '\n' || c == '\r' {
				if term != nil {
					term[0] = true
				}
				eol = true
				break
			}
		}
		startChar := this.nextChar
		this.nextChar = i
		if eol {
			var str *String
			if !haveS {
				str = FromUnits(append([]uint16(nil), this.cb[startChar:i]...))
			} else {
				s = append(s, this.cb[startChar:i]...)
				str = FromUnits(s)
			}
			this.nextChar++
			if c == '\r' {
				this.skipLF = true
			}
			return str
		}
		if !haveS {
			haveS = true
			s = make([]uint16, 0, 80)
		}
		s = append(s, this.cb[startChar:i]...)
	}
}

func (this *BufferedReader) skip(n int64) int64 {
	if n < 0 {
		panic(IllegalArgumentException_new_String(Lit("skip value is negative")))
	}
	this.ensureOpen()
	r := n
	for r > 0 {
		if this.nextChar >= this.nChars {
			this.fill()
		}
		if this.nextChar >= this.nChars {
			break
		}
		if this.skipLF {
			this.skipLF = false
			if this.cb[this.nextChar] == '\n' {
				this.nextChar++
			}
		}
		d := int64(this.nChars - this.nextChar)
		if r <= d {
			this.nextChar += int32(r)
			r = 0
			break
		}
		r -= d
		this.nextChar = this.nChars
	}
	return n - r
}

func (this *BufferedReader) ready() bool {
	this.ensureOpen()
	if this.skipLF {
		if this.nextChar >= this.nChars && this.in.Ready() {
			this.fill()
		}
		if this.nextChar < this.nChars {
			if this.cb[this.nextChar] == '\n' {
				this.nextChar++
			}
			this.skipLF = false
		}
	}
	return this.nextChar < this.nChars || this.in.Ready()
}

func (this *BufferedReader) markSupported() bool { return true }

func (this *BufferedReader) mark(readAheadLimit int32) {
	if readAheadLimit < 0 {
		panic(IllegalArgumentException_new_String(Lit("Read-ahead limit < 0")))
	}
	this.ensureOpen()
	this.readAheadLimit = readAheadLimit
	this.markedChar = this.nextChar
	this.markedSkipLF = this.skipLF
}

func (this *BufferedReader) reset() {
	this.ensureOpen()
	if this.markedChar < 0 {
		if this.markedChar == brInvalidated {
			panic(IOException_new_String(Lit("Mark invalid")))
		}
		panic(IOException_new_String(Lit("Stream not marked")))
	}
	this.nextChar = this.markedChar
	this.skipLF = this.markedSkipLF
}

func (this *BufferedReader) close() {
	if this.in == nil {
		return
	}
	defer func() {
		this.in = nil
		this.cb = nil
	}()
	this.in.Close()
}

// ---------------------------------------------------------------------------
// LineNumberReader

const (
	lnrNone = 0
	lnrChar = 1
	lnrEOL  = 2
	lnrEOF  = 3
)

// LineNumberReader is java.io.LineNumberReader.
type LineNumberReader struct {
	BufferedReader
	prevChar         int32
	lineNumber       int32
	markedLineNumber int32
	lnSkipLF         bool
	lnMarkedSkipLF   bool
	lnSkipBuffer     []uint16
}

func (this *LineNumberReader) JClass_LineNumberReader() *LineNumberReader { return this }
func (this *LineNumberReader) JClassName() string                         { return "java.io.LineNumberReader" }

func LineNumberReader_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_LineNumberReader() *LineNumberReader })
	return ok
}

func LineNumberReader_fromAny(o any) *LineNumberReader {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_LineNumberReader() *LineNumberReader }](o).JClass_LineNumberReader()
}

func LineNumberReader_new_Reader(in *Reader) *LineNumberReader {
	this := &LineNumberReader{}
	this.jself = this
	this.LineNumberReader_init_Reader(in)
	return this
}

func LineNumberReader_new_Reader_int(in *Reader, sz int32) *LineNumberReader {
	this := &LineNumberReader{}
	this.jself = this
	this.LineNumberReader_init_Reader_int(in, sz)
	return this
}

func (this *LineNumberReader) LineNumberReader_init_Reader(in *Reader) {
	this.BufferedReader.BufferedReader_init_Reader(in)
}

func (this *LineNumberReader) LineNumberReader_init_Reader_int(in *Reader, sz int32) {
	this.BufferedReader.BufferedReader_init_Reader_int(in, sz)
}

func (this *LineNumberReader) SetLineNumber_int(lineNumber int32) { this.lineNumber = lineNumber }
func (this *LineNumberReader) GetLineNumber() int32               { return this.lineNumber }

func (this *LineNumberReader) read() int32 {
	c := this.BufferedReader.read()
	if this.lnSkipLF {
		if c == '\n' {
			c = this.BufferedReader.read()
		}
		this.lnSkipLF = false
	}
	switch c {
	case '\r':
		this.lnSkipLF = true
		fallthrough
	case '\n':
		this.lineNumber++
		this.prevChar = lnrEOL
		return '\n'
	case -1:
		if this.prevChar == lnrChar {
			this.lineNumber++
		}
		this.prevChar = lnrEOF
	default:
		this.prevChar = lnrChar
	}
	return c
}

func (this *LineNumberReader) readArr(cbuf []uint16, off, n int32) int32 {
	k := this.BufferedReader.readArr(cbuf, off, n)
	if k == -1 {
		if this.prevChar == lnrChar {
			this.lineNumber++
		}
		this.prevChar = lnrEOF
		return -1
	}
	for i := off; i < off+k; i++ {
		c := cbuf[i]
		if this.lnSkipLF {
			this.lnSkipLF = false
			if c == '\n' {
				continue
			}
		}
		switch c {
		case '\r':
			this.lnSkipLF = true
			this.lineNumber++
		case '\n':
			this.lineNumber++
		}
	}
	if k > 0 {
		switch cbuf[off+k-1] {
		case '\r', '\n':
			this.prevChar = lnrEOL
		default:
			this.prevChar = lnrChar
		}
	}
	return k
}

func (this *LineNumberReader) readLine() *String {
	term := make([]bool, 1)
	l := this.BufferedReader.readLineTerm(this.lnSkipLF, term)
	this.lnSkipLF = false
	if l != nil {
		this.lineNumber++
		if term[0] {
			this.prevChar = lnrEOL
		} else {
			this.prevChar = lnrEOF
		}
	} else {
		if this.prevChar == lnrChar {
			this.lineNumber++
		}
		this.prevChar = lnrEOF
	}
	return l
}

func (this *LineNumberReader) skip(n int64) int64 {
	if n < 0 {
		panic(IllegalArgumentException_new_String(Lit("skip() value is negative")))
	}
	nn := int32(8192)
	if n < int64(nn) {
		nn = int32(n)
	}
	if len(this.lnSkipBuffer) < int(nn) {
		this.lnSkipBuffer = make([]uint16, nn)
	}
	r := n
	for r > 0 {
		k := nn
		if r < int64(k) {
			k = int32(r)
		}
		nc := this.impl().readArr(this.lnSkipBuffer, 0, k)
		if nc == -1 {
			break
		}
		r -= int64(nc)
	}
	if n-r > 0 {
		this.prevChar = lnrNone
	}
	return n - r
}

func (this *LineNumberReader) mark(readAheadLimit int32) {
	if this.lnSkipLF {
		readAheadLimit++
	}
	this.BufferedReader.mark(readAheadLimit)
	this.markedLineNumber = this.lineNumber
	this.lnMarkedSkipLF = this.lnSkipLF
}

func (this *LineNumberReader) reset() {
	this.BufferedReader.reset()
	this.lineNumber = this.markedLineNumber
	this.lnSkipLF = this.lnMarkedSkipLF
}

// ---------------------------------------------------------------------------
// InputStreamReader and FileReader

// InputStreamReader is java.io.InputStreamReader, decoding UTF-8 (the JDK
// 21 default charset) with malformed input replaced. It ports JDK 21's
// StreamDecoder, which determines how many characters each read returns.
type InputStreamReader struct {
	Reader
	in               *InputStream
	closed           bool
	bb               []int8
	bbPos            int
	bbLim            int
	haveLeftoverChar bool
	leftoverChar     uint16
}

func (this *InputStreamReader) JClass_InputStreamReader() *InputStreamReader { return this }
func (this *InputStreamReader) JClassName() string                           { return "java.io.InputStreamReader" }

func InputStreamReader_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_InputStreamReader() *InputStreamReader })
	return ok
}

func InputStreamReader_fromAny(o any) *InputStreamReader {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_InputStreamReader() *InputStreamReader }](o).JClass_InputStreamReader()
}

func InputStreamReader_new_InputStream(in *InputStream) *InputStreamReader {
	this := &InputStreamReader{}
	this.jself = this
	this.InputStreamReader_init_InputStream(in)
	return this
}

// InputStreamReader_new_InputStream_String implements
// InputStreamReader(InputStream, String) for the UTF-8 charset names.
func InputStreamReader_new_InputStream_String(in *InputStream, charsetName *String) *InputStreamReader {
	this := &InputStreamReader{}
	this.jself = this
	this.InputStreamReader_init_InputStream_String(in, charsetName)
	return this
}

func (this *InputStreamReader) InputStreamReader_init_InputStream(in *InputStream) {
	this.Reader.Reader_init()
	if in == nil {
		panic(NullPointerException_new())
	}
	this.in = in
	this.bb = make([]int8, 8192)
}

func (this *InputStreamReader) InputStreamReader_init_InputStream_String(in *InputStream, charsetName *String) {
	if in == nil {
		panic(NullPointerException_new())
	}
	if charsetName == nil {
		panic(NullPointerException_new_String(Lit("charsetName")))
	}
	if !isUTF8Name(charsetName.Go()) {
		panic(Unsupported("charset " + charsetName.Go()))
	}
	this.InputStreamReader_init_InputStream(in)
}

func isUTF8Name(n string) bool {
	switch n {
	case "UTF-8", "UTF8", "utf-8", "utf8", "unicode-1-1-utf-8":
		return true
	}
	return false
}

// GetEncoding implements InputStreamReader.getEncoding, which returns the
// historical name of the charset, or null once the stream is closed.
func (this *InputStreamReader) GetEncoding() *String {
	if this.closed {
		return nil
	}
	return Lit("UTF8")
}

func (this *InputStreamReader) ensureOpen() {
	if this.closed {
		panic(IOException_new_String(Lit("Stream closed")))
	}
}

func (this *InputStreamReader) read() int32 {
	if this.haveLeftoverChar {
		this.haveLeftoverChar = false
		return int32(this.leftoverChar)
	}
	cb := make([]uint16, 2)
	switch this.readArr(cb, 0, 2) {
	case -1:
		return -1
	case 2:
		this.leftoverChar = cb[1]
		this.haveLeftoverChar = true
	}
	return int32(cb[0])
}

func (this *InputStreamReader) readArr(cbuf []uint16, off, n int32) int32 {
	this.ensureOpen()
	requireArray(cbuf)
	if off < 0 || off > int32(len(cbuf)) || n < 0 || off+n > int32(len(cbuf)) || off+n < 0 {
		panic(IndexOutOfBoundsException_new())
	}
	if n == 0 {
		return 0
	}
	var k int32
	if this.haveLeftoverChar {
		cbuf[off] = this.leftoverChar
		off++
		n--
		this.haveLeftoverChar = false
		k = 1
		if n == 0 || !this.implReady() {
			return k
		}
	}
	if n == 1 {
		c := this.read()
		if c == -1 {
			if k == 0 {
				return -1
			}
			return k
		}
		cbuf[off] = uint16(c)
		return k + 1
	}
	nr := this.implRead(cbuf, int(off), int(off+n))
	if nr < 0 {
		if k == 1 {
			return 1
		}
		return nr
	}
	return k + nr
}

// readBytes reads more bytes into bb after the undecoded ones, returning
// the number of undecoded bytes or -1 at the end of the stream.
func (this *InputStreamReader) readBytes() int32 {
	copy(this.bb, this.bb[this.bbPos:this.bbLim])
	pos := this.bbLim - this.bbPos
	this.bbPos, this.bbLim = 0, pos
	n := this.in.Read_byteArr_int_int(this.bb, int32(pos), int32(len(this.bb)-pos))
	if n < 0 {
		return n
	}
	if n == 0 {
		panic(IOException_new_String(Lit("Underlying input stream returned zero bytes")))
	}
	this.bbLim = pos + int(n)
	return int32(this.bbLim)
}

func (this *InputStreamReader) implRead(cbuf []uint16, off, end int) int32 {
	cp := off
	eof := false
	for {
		var overflow bool
		this.bbPos, cp, overflow = decodeUTF8Replace(this.bb, this.bbPos, this.bbLim, cbuf, cp, end, eof)
		if overflow {
			break
		}
		if eof || cp == end || cp > off && !this.inReady() {
			break
		}
		if this.readBytes() < 0 {
			eof = true
			if cp == off && this.bbPos == this.bbLim {
				break
			}
		}
	}
	if cp == off && eof {
		return -1
	}
	return int32(cp - off)
}

func (this *InputStreamReader) inReady() bool {
	ready := false
	func() {
		defer func() {
			if r := recover(); r != nil && !IOException_isInstance(r) {
				panic(r)
			}
		}()
		ready = this.in.Available() > 0
	}()
	return ready
}

func (this *InputStreamReader) implReady() bool {
	return this.bbPos < this.bbLim || this.inReady()
}

func (this *InputStreamReader) ready() bool {
	this.ensureOpen()
	return this.haveLeftoverChar || this.implReady()
}

func (this *InputStreamReader) close() {
	if this.closed {
		return
	}
	defer func() { this.closed = true }()
	this.in.Close()
}

// FileReader is java.io.FileReader.
type FileReader struct {
	InputStreamReader
}

func (this *FileReader) JClass_FileReader() *FileReader { return this }
func (this *FileReader) JClassName() string             { return "java.io.FileReader" }

func FileReader_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_FileReader() *FileReader })
	return ok
}

func FileReader_fromAny(o any) *FileReader {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_FileReader() *FileReader }](o).JClass_FileReader()
}

func FileReader_new_String(fileName *String) *FileReader {
	this := &FileReader{}
	this.jself = this
	this.FileReader_init_String(fileName)
	return this
}

func FileReader_new_File(file *File) *FileReader {
	this := &FileReader{}
	this.jself = this
	this.FileReader_init_File(file)
	return this
}

func (this *FileReader) FileReader_init_String(fileName *String) {
	this.InputStreamReader.InputStreamReader_init_InputStream(&FileInputStream_new_String(fileName).InputStream)
}

func (this *FileReader) FileReader_init_File(file *File) {
	this.InputStreamReader.InputStreamReader_init_InputStream(&FileInputStream_new_File(file).InputStream)
}
