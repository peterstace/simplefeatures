package java

import (
	"errors"
	"io"
	"os"
	"sync"
	"syscall"
)

// ---------------------------------------------------------------------------
// InputStream

// InputStream is java.io.InputStream. See io_reader.go for how calls
// dispatch.
type InputStream struct {
	jself any
	jhash int32
}

// inputStreamImpl holds the InputStream methods that subclasses override.
type inputStreamImpl interface {
	read() int32
	readArr(b []int8, off, n int32) int32
	skip(n int64) int64
	available() int32
	close()
	mark(readlimit int32)
	reset()
	markSupported() bool
	toStr() *String
}

func (this *InputStream) JSelf() any                       { return this.jself }
func (this *InputStream) JSetSelf(self any)                { this.jself = self }
func (this *InputStream) JClass_InputStream() *InputStream { return this }
func (this *InputStream) JClassName() string               { return "java.io.InputStream" }
func (this *InputStream) JIdentityHash() *int32            { return &this.jhash }

func InputStream_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_InputStream() *InputStream })
	return ok
}

func InputStream_fromAny(o any) *InputStream {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_InputStream() *InputStream }](o).JClass_InputStream()
}

func (this *InputStream) InputStream_init() {}

func (this *InputStream) impl() inputStreamImpl {
	if v, ok := this.jself.(inputStreamImpl); ok {
		return v
	}
	return this
}

func (this *InputStream) Read() int32 { return this.impl().read() }

func (this *InputStream) Read_byteArr(b []int8) int32 {
	requireArray(b)
	return this.impl().readArr(b, 0, int32(len(b)))
}

func (this *InputStream) Read_byteArr_int_int(b []int8, off, n int32) int32 {
	return this.impl().readArr(b, off, n)
}

func (this *InputStream) Skip_long(n int64) int64  { return this.impl().skip(n) }
func (this *InputStream) Available() int32         { return this.impl().available() }
func (this *InputStream) Close()                   { this.impl().close() }
func (this *InputStream) Mark_int(readlimit int32) { this.impl().mark(readlimit) }
func (this *InputStream) Reset()                   { this.impl().reset() }
func (this *InputStream) MarkSupported() bool      { return this.impl().markSupported() }
func (this *InputStream) ToString() *String        { return this.impl().toStr() }

// read is abstract in InputStream; only a generated subclass reaches it.
func (this *InputStream) read() int32 {
	g, ok := this.jself.(interface{ Read() int32 })
	if !ok {
		panic(AbstractMethodError("java.io.InputStream.read()"))
	}
	return g.Read()
}

func (this *InputStream) readArr(b []int8, off, n int32) int32 {
	requireArray(b)
	checkFromIndexSize(off, n, int32(len(b)))
	if n == 0 {
		return 0
	}
	v := this.impl()
	c := v.read()
	if c == -1 {
		return -1
	}
	b[off] = int8(c)
	i := int32(1)
	func() {
		defer func() {
			if r := recover(); r != nil && !IOException_isInstance(r) {
				panic(r)
			}
		}()
		for ; i < n; i++ {
			c = v.read()
			if c == -1 {
				break
			}
			b[off+i] = int8(c)
		}
	}()
	return i
}

func (this *InputStream) skip(n int64) int64 {
	if n <= 0 {
		return 0
	}
	remaining := n
	size := int32(2048)
	if remaining < int64(size) {
		size = int32(remaining)
	}
	skipBuffer := make([]int8, size)
	for remaining > 0 {
		k := size
		if remaining < int64(k) {
			k = int32(remaining)
		}
		nr := this.impl().readArr(skipBuffer, 0, k)
		if nr < 0 {
			break
		}
		remaining -= int64(nr)
	}
	return n - remaining
}

func (this *InputStream) available() int32    { return 0 }
func (this *InputStream) close()              {}
func (this *InputStream) mark(int32)          {}
func (this *InputStream) markSupported() bool { return false }

func (this *InputStream) reset() {
	panic(IOException_new_String(Lit("mark/reset not supported")))
}

func (this *InputStream) toStr() *String { return ObjectImpl_ToString(Ref(this)) }

// ---------------------------------------------------------------------------
// ByteArrayInputStream

// ByteArrayInputStream is java.io.ByteArrayInputStream.
type ByteArrayInputStream struct {
	InputStream
	buf   []int8
	pos   int32
	mrk   int32
	count int32
}

func (this *ByteArrayInputStream) JClass_ByteArrayInputStream() *ByteArrayInputStream { return this }
func (this *ByteArrayInputStream) JClassName() string                                 { return "java.io.ByteArrayInputStream" }

func ByteArrayInputStream_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_ByteArrayInputStream() *ByteArrayInputStream })
	return ok
}

func ByteArrayInputStream_fromAny(o any) *ByteArrayInputStream {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_ByteArrayInputStream() *ByteArrayInputStream }](o).JClass_ByteArrayInputStream()
}

func ByteArrayInputStream_new_byteArr(buf []int8) *ByteArrayInputStream {
	this := &ByteArrayInputStream{}
	this.jself = this
	this.ByteArrayInputStream_init_byteArr(buf)
	return this
}

func ByteArrayInputStream_new_byteArr_int_int(buf []int8, offset, length int32) *ByteArrayInputStream {
	this := &ByteArrayInputStream{}
	this.jself = this
	this.ByteArrayInputStream_init_byteArr_int_int(buf, offset, length)
	return this
}

func (this *ByteArrayInputStream) ByteArrayInputStream_init_byteArr(buf []int8) {
	this.InputStream.InputStream_init()
	requireArray(buf)
	this.buf = buf
	this.count = int32(len(buf))
}

func (this *ByteArrayInputStream) ByteArrayInputStream_init_byteArr_int_int(buf []int8, offset, length int32) {
	this.InputStream.InputStream_init()
	this.buf = buf
	this.pos = offset
	requireArray(buf)
	this.count = lesserInt32(offset+length, int32(len(buf)))
	this.mrk = offset
}

func (this *ByteArrayInputStream) read() int32 {
	if this.pos < this.count {
		b := this.buf[this.pos]
		this.pos++
		return int32(b) & 0xff
	}
	return -1
}

func (this *ByteArrayInputStream) readArr(b []int8, off, n int32) int32 {
	requireArray(b)
	checkFromIndexSize(off, n, int32(len(b)))
	if this.pos >= this.count {
		return -1
	}
	if avail := this.count - this.pos; n > avail {
		n = avail
	}
	if n <= 0 {
		return 0
	}
	copy(b[off:off+n], this.buf[this.pos:this.pos+n])
	this.pos += n
	return n
}

func (this *ByteArrayInputStream) skip(n int64) int64 {
	k := int64(this.count - this.pos)
	if n < k {
		k = n
		if n < 0 {
			k = 0
		}
	}
	this.pos += int32(k)
	return k
}

func (this *ByteArrayInputStream) available() int32    { return this.count - this.pos }
func (this *ByteArrayInputStream) markSupported() bool { return true }
func (this *ByteArrayInputStream) mark(int32)          { this.mrk = this.pos }
func (this *ByteArrayInputStream) reset()              { this.pos = this.mrk }
func (this *ByteArrayInputStream) close()              {}

// ---------------------------------------------------------------------------
// FileInputStream

// FileInputStream is java.io.FileInputStream.
type FileInputStream struct {
	InputStream
	f      *os.File
	closed bool
}

func (this *FileInputStream) JClass_FileInputStream() *FileInputStream { return this }
func (this *FileInputStream) JClassName() string                       { return "java.io.FileInputStream" }

func FileInputStream_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_FileInputStream() *FileInputStream })
	return ok
}

func FileInputStream_fromAny(o any) *FileInputStream {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_FileInputStream() *FileInputStream }](o).JClass_FileInputStream()
}

func FileInputStream_new_String(name *String) *FileInputStream {
	this := &FileInputStream{}
	this.jself = this
	this.FileInputStream_init_String(name)
	return this
}

func FileInputStream_new_File(file *File) *FileInputStream {
	this := &FileInputStream{}
	this.jself = this
	this.FileInputStream_init_File(file)
	return this
}

func (this *FileInputStream) FileInputStream_init_String(name *String) {
	var file *File
	if name != nil {
		file = File_new_String(name)
	}
	this.FileInputStream_init_File(file)
}

func (this *FileInputStream) FileInputStream_init_File(file *File) {
	this.InputStream.InputStream_init()
	if file == nil {
		panic(NullPointerException_new())
	}
	if file.isInvalid() {
		panic(FileNotFoundException_new_String(Lit("Invalid file path")))
	}
	name := file.path.Go()
	f, err := os.Open(name)
	if err == nil {
		if st, serr := f.Stat(); serr == nil && st.IsDir() {
			f.Close()
			err = &os.PathError{Op: "open", Path: name, Err: syscall.EISDIR}
		}
	}
	if err != nil {
		panic(FileNotFoundException_new_String(FromGo(name + " (" + errnoText(err) + ")")))
	}
	this.f = f
}

func (this *FileInputStream) ensureOpen() {
	if this.closed {
		panic(IOException_new_String(Lit("Stream Closed")))
	}
}

func (this *FileInputStream) read() int32 {
	this.ensureOpen()
	var b [1]byte
	n, err := this.f.Read(b[:])
	if n == 1 {
		return int32(b[0])
	}
	if err != nil && !errors.Is(err, io.EOF) {
		panic(IOException_new_String(FromGo(errnoText(err))))
	}
	return -1
}

func (this *FileInputStream) readArr(b []int8, off, n int32) int32 {
	requireArray(b)
	if off < 0 || n < 0 || int64(off)+int64(n) > int64(len(b)) {
		panic(IndexOutOfBoundsException_new())
	}
	if n == 0 {
		return 0
	}
	this.ensureOpen()
	buf := make([]byte, n)
	k, err := this.f.Read(buf)
	if k > 0 {
		for i := 0; i < k; i++ {
			b[int(off)+i] = int8(buf[i])
		}
		return int32(k)
	}
	if err != nil && !errors.Is(err, io.EOF) {
		panic(IOException_new_String(FromGo(errnoText(err))))
	}
	return -1
}

func (this *FileInputStream) skip(n int64) int64 {
	this.ensureOpen()
	cur, err := this.f.Seek(0, io.SeekCurrent)
	if err != nil {
		panic(IOException_new_String(FromGo(errnoText(err))))
	}
	end, err := this.f.Seek(n, io.SeekCurrent)
	if err != nil {
		panic(IOException_new_String(FromGo(errnoText(err))))
	}
	return end - cur
}

func (this *FileInputStream) available() int32 {
	this.ensureOpen()
	cur, err := this.f.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0
	}
	st, err := this.f.Stat()
	if err != nil || st.Size() <= cur {
		return 0
	}
	if st.Size()-cur > 1<<31-1 {
		return 1<<31 - 1
	}
	return int32(st.Size() - cur)
}

func (this *FileInputStream) close() {
	if this.closed {
		return
	}
	this.closed = true
	this.f.Close()
}

// errnoText returns the C library's message for the error behind err, which
// Java uses in its file exception messages.
func errnoText(err error) string {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		switch errno {
		case syscall.ENOENT:
			return "No such file or directory"
		case syscall.EACCES:
			return "Permission denied"
		case syscall.EISDIR:
			return "Is a directory"
		case syscall.ENOTDIR:
			return "Not a directory"
		case syscall.EEXIST:
			return "File exists"
		case syscall.ENOSPC:
			return "No space left on device"
		case syscall.EBADF:
			return "Bad file descriptor"
		}
		s := errno.Error()
		if s != "" && s[0] >= 'a' && s[0] <= 'z' {
			s = string(s[0]-'a'+'A') + s[1:]
		}
		return s
	}
	return err.Error()
}

// ---------------------------------------------------------------------------
// OutputStream

// OutputStream is java.io.OutputStream. See io_reader.go for how calls
// dispatch.
type OutputStream struct {
	jself any
	jhash int32
}

// outputStreamImpl holds the OutputStream methods that subclasses override.
type outputStreamImpl interface {
	writeByte(b int32)
	writeArr(b []int8, off, n int32)
	flush()
	close()
	toStr() *String
}

func (this *OutputStream) JSelf() any                         { return this.jself }
func (this *OutputStream) JSetSelf(self any)                  { this.jself = self }
func (this *OutputStream) JClass_OutputStream() *OutputStream { return this }
func (this *OutputStream) JClassName() string                 { return "java.io.OutputStream" }
func (this *OutputStream) JIdentityHash() *int32              { return &this.jhash }

func OutputStream_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_OutputStream() *OutputStream })
	return ok
}

func OutputStream_fromAny(o any) *OutputStream {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_OutputStream() *OutputStream }](o).JClass_OutputStream()
}

func (this *OutputStream) OutputStream_init() {}

func (this *OutputStream) impl() outputStreamImpl {
	if v, ok := this.jself.(outputStreamImpl); ok {
		return v
	}
	return this
}

func (this *OutputStream) Write_int(b int32) { this.impl().writeByte(b) }

func (this *OutputStream) Write_byteArr(b []int8) {
	requireArray(b)
	this.impl().writeArr(b, 0, int32(len(b)))
}

func (this *OutputStream) Write_byteArr_int_int(b []int8, off, n int32) {
	this.impl().writeArr(b, off, n)
}

func (this *OutputStream) Flush()            { this.impl().flush() }
func (this *OutputStream) Close()            { this.impl().close() }
func (this *OutputStream) ToString() *String { return this.impl().toStr() }

// writeByte is abstract in OutputStream; only a generated subclass reaches
// it.
func (this *OutputStream) writeByte(b int32) {
	g, ok := this.jself.(interface{ Write_int(int32) })
	if !ok {
		panic(AbstractMethodError("java.io.OutputStream.write(int)"))
	}
	g.Write_int(b)
}

func (this *OutputStream) writeArr(b []int8, off, n int32) {
	requireArray(b)
	checkFromIndexSize(off, n, int32(len(b)))
	v := this.impl()
	for i := int32(0); i < n; i++ {
		v.writeByte(int32(b[off+i]))
	}
}

func (this *OutputStream) flush()         {}
func (this *OutputStream) close()         {}
func (this *OutputStream) toStr() *String { return ObjectImpl_ToString(Ref(this)) }

// ---------------------------------------------------------------------------
// ByteArrayOutputStream

// ByteArrayOutputStream is java.io.ByteArrayOutputStream.
type ByteArrayOutputStream struct {
	OutputStream
	buf []int8
}

func (this *ByteArrayOutputStream) JClass_ByteArrayOutputStream() *ByteArrayOutputStream {
	return this
}
func (this *ByteArrayOutputStream) JClassName() string { return "java.io.ByteArrayOutputStream" }

func ByteArrayOutputStream_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_ByteArrayOutputStream() *ByteArrayOutputStream
	})
	return ok
}

func ByteArrayOutputStream_fromAny(o any) *ByteArrayOutputStream {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_ByteArrayOutputStream() *ByteArrayOutputStream
	}](o).JClass_ByteArrayOutputStream()
}

func ByteArrayOutputStream_new() *ByteArrayOutputStream {
	this := &ByteArrayOutputStream{}
	this.jself = this
	this.ByteArrayOutputStream_init()
	return this
}

func ByteArrayOutputStream_new_int(size int32) *ByteArrayOutputStream {
	this := &ByteArrayOutputStream{}
	this.jself = this
	this.ByteArrayOutputStream_init_int(size)
	return this
}

func (this *ByteArrayOutputStream) ByteArrayOutputStream_init() {
	this.ByteArrayOutputStream_init_int(32)
}

func (this *ByteArrayOutputStream) ByteArrayOutputStream_init_int(size int32) {
	this.OutputStream.OutputStream_init()
	if size < 0 {
		panic(IllegalArgumentException_new_String(Concat(Lit("Negative initial size: "), StrOfInt(size))))
	}
	this.buf = make([]int8, 0, size)
}

func (this *ByteArrayOutputStream) writeByte(b int32) {
	this.buf = append(this.buf, int8(b))
}

func (this *ByteArrayOutputStream) writeArr(b []int8, off, n int32) {
	requireArray(b)
	checkFromIndexSize(off, n, int32(len(b)))
	this.buf = append(this.buf, b[off:off+n]...)
}

func (this *ByteArrayOutputStream) close() {}

func (this *ByteArrayOutputStream) WriteTo_OutputStream(out *OutputStream) {
	out.Write_byteArr_int_int(this.buf, 0, int32(len(this.buf)))
}

func (this *ByteArrayOutputStream) Reset() { this.buf = this.buf[:0] }

func (this *ByteArrayOutputStream) ToByteArray() []int8 {
	return append(make([]int8, 0, len(this.buf)), this.buf...)
}

func (this *ByteArrayOutputStream) Size() int32 { return int32(len(this.buf)) }

// toStr implements ByteArrayOutputStream.toString, which decodes the bytes
// with the default charset, UTF-8.
func (this *ByteArrayOutputStream) toStr() *String {
	return FromUnits(decodeUTF8(this.buf))
}

// ToString_int implements the deprecated ByteArrayOutputStream.toString(int),
// which makes each byte a char with hibyte as its high byte.
func (this *ByteArrayOutputStream) ToString_int(hibyte int32) *String {
	u := make([]uint16, len(this.buf))
	for i, v := range this.buf {
		u[i] = uint16((hibyte&0xff)<<8 | int32(v)&0xff)
	}
	return FromUnits(u)
}

// ---------------------------------------------------------------------------
// FileOutputStream

// FileOutputStream is java.io.FileOutputStream.
type FileOutputStream struct {
	OutputStream
	f      *os.File
	closed bool
}

func (this *FileOutputStream) JClass_FileOutputStream() *FileOutputStream { return this }
func (this *FileOutputStream) JClassName() string                         { return "java.io.FileOutputStream" }

func FileOutputStream_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_FileOutputStream() *FileOutputStream })
	return ok
}

func FileOutputStream_fromAny(o any) *FileOutputStream {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_FileOutputStream() *FileOutputStream }](o).JClass_FileOutputStream()
}

func FileOutputStream_new_String(name *String) *FileOutputStream {
	return FileOutputStream_new_String_boolean(name, false)
}

func FileOutputStream_new_String_boolean(name *String, append_ bool) *FileOutputStream {
	this := &FileOutputStream{}
	this.jself = this
	this.FileOutputStream_init_String_boolean(name, append_)
	return this
}

func FileOutputStream_new_File(file *File) *FileOutputStream {
	return FileOutputStream_new_File_boolean(file, false)
}

func FileOutputStream_new_File_boolean(file *File, append_ bool) *FileOutputStream {
	this := &FileOutputStream{}
	this.jself = this
	this.FileOutputStream_init_File_boolean(file, append_)
	return this
}

func (this *FileOutputStream) FileOutputStream_init_String(name *String) {
	this.FileOutputStream_init_String_boolean(name, false)
}

func (this *FileOutputStream) FileOutputStream_init_String_boolean(name *String, append_ bool) {
	var file *File
	if name != nil {
		file = File_new_String(name)
	}
	this.FileOutputStream_init_File_boolean(file, append_)
}

func (this *FileOutputStream) FileOutputStream_init_File(file *File) {
	this.FileOutputStream_init_File_boolean(file, false)
}

func (this *FileOutputStream) FileOutputStream_init_File_boolean(file *File, append_ bool) {
	this.OutputStream.OutputStream_init()
	if file == nil {
		panic(NullPointerException_new())
	}
	if file.isInvalid() {
		panic(FileNotFoundException_new_String(Lit("Invalid file path")))
	}
	name := file.path.Go()
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if append_ {
		flags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	}
	f, err := os.OpenFile(name, flags, 0o666)
	if err != nil {
		panic(FileNotFoundException_new_String(FromGo(name + " (" + errnoText(err) + ")")))
	}
	this.f = f
}

func (this *FileOutputStream) ensureOpen() {
	if this.closed {
		panic(IOException_new_String(Lit("Stream Closed")))
	}
}

func (this *FileOutputStream) writeByte(b int32) {
	this.ensureOpen()
	if _, err := this.f.Write([]byte{byte(b)}); err != nil {
		panic(IOException_new_String(FromGo(errnoText(err))))
	}
}

func (this *FileOutputStream) writeArr(b []int8, off, n int32) {
	requireArray(b)
	if off < 0 || n < 0 || int64(off)+int64(n) > int64(len(b)) {
		panic(IndexOutOfBoundsException_new())
	}
	if n == 0 {
		return
	}
	this.ensureOpen()
	buf := make([]byte, n)
	for i := range buf {
		buf[i] = byte(b[int(off)+i])
	}
	if _, err := this.f.Write(buf); err != nil {
		panic(IOException_new_String(FromGo(errnoText(err))))
	}
}

func (this *FileOutputStream) close() {
	if this.closed {
		return
	}
	this.closed = true
	this.f.Close()
}

// ---------------------------------------------------------------------------
// FilterOutputStream

// FilterOutputStream is java.io.FilterOutputStream.
type FilterOutputStream struct {
	OutputStream
	out    *OutputStream
	closed bool
}

func (this *FilterOutputStream) JClass_FilterOutputStream() *FilterOutputStream { return this }
func (this *FilterOutputStream) JClassName() string                             { return "java.io.FilterOutputStream" }

func FilterOutputStream_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_FilterOutputStream() *FilterOutputStream })
	return ok
}

func FilterOutputStream_fromAny(o any) *FilterOutputStream {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_FilterOutputStream() *FilterOutputStream }](o).JClass_FilterOutputStream()
}

func FilterOutputStream_new(out *OutputStream) *FilterOutputStream {
	this := &FilterOutputStream{}
	this.jself = this
	this.FilterOutputStream_init(out)
	return this
}

func (this *FilterOutputStream) FilterOutputStream_init(out *OutputStream) {
	this.OutputStream.OutputStream_init()
	this.out = out
}

func (this *FilterOutputStream) writeByte(b int32) { this.out.Write_int(b) }

func (this *FilterOutputStream) writeArr(b []int8, off, n int32) {
	requireArray(b)
	if off|n|(int32(len(b))-(n+off))|(off+n) < 0 {
		panic(IndexOutOfBoundsException_new())
	}
	v := this.impl()
	for i := int32(0); i < n; i++ {
		v.writeByte(int32(b[off+i]))
	}
}

func (this *FilterOutputStream) flush() { this.out.Flush() }

func (this *FilterOutputStream) close() {
	if this.closed {
		return
	}
	this.closed = true
	var flushErr any
	func() {
		defer func() { flushErr = recover() }()
		this.impl().flush()
	}()
	// As in Java, the stream is closed even if flushing failed, and an
	// exception from closing takes precedence over the one from flushing.
	this.out.Close()
	if flushErr != nil {
		panic(flushErr)
	}
}

// ---------------------------------------------------------------------------
// PrintStream

// PrintStream is java.io.PrintStream, encoding text as UTF-8. As in Java, an
// IOException from the underlying stream is swallowed and recorded for
// CheckError.
type PrintStream struct {
	FilterOutputStream
	mu        sync.Mutex
	autoFlush bool
	trouble   bool
	closing   bool
	enc       utf8Encoder
}

func (this *PrintStream) JClass_PrintStream() *PrintStream { return this }
func (this *PrintStream) JClassName() string               { return "java.io.PrintStream" }

func PrintStream_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_PrintStream() *PrintStream })
	return ok
}

func PrintStream_fromAny(o any) *PrintStream {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_PrintStream() *PrintStream }](o).JClass_PrintStream()
}

func PrintStream_new_OutputStream(out *OutputStream) *PrintStream {
	return PrintStream_new_OutputStream_boolean(out, false)
}

func PrintStream_new_OutputStream_boolean(out *OutputStream, autoFlush bool) *PrintStream {
	this := &PrintStream{}
	this.jself = this
	this.PrintStream_init_OutputStream_boolean(out, autoFlush)
	return this
}

func PrintStream_new_String(fileName *String) *PrintStream {
	this := &PrintStream{}
	this.jself = this
	this.PrintStream_init_String(fileName)
	return this
}

func PrintStream_new_File(file *File) *PrintStream {
	this := &PrintStream{}
	this.jself = this
	this.PrintStream_init_File(file)
	return this
}

func (this *PrintStream) PrintStream_init_OutputStream(out *OutputStream) {
	this.PrintStream_init_OutputStream_boolean(out, false)
}

func (this *PrintStream) PrintStream_init_OutputStream_boolean(out *OutputStream, autoFlush bool) {
	if out == nil {
		panic(NullPointerException_new_String(Lit("Null output stream")))
	}
	this.FilterOutputStream.FilterOutputStream_init(out)
	this.autoFlush = autoFlush
}

func (this *PrintStream) PrintStream_init_String(fileName *String) {
	this.FilterOutputStream.FilterOutputStream_init(&FileOutputStream_new_String(fileName).OutputStream)
}

func (this *PrintStream) PrintStream_init_File(file *File) {
	this.FilterOutputStream.FilterOutputStream_init(&FileOutputStream_new_File(file).OutputStream)
}

// guard runs f, recording an IOException as trouble instead of propagating
// it, as PrintStream's methods do.
func (this *PrintStream) guard(f func()) {
	defer func() {
		if r := recover(); r != nil {
			if !IOException_isInstance(r) {
				panic(r)
			}
			this.trouble = true
		}
	}()
	f()
}

func (this *PrintStream) ensureOpen() {
	if this.out == nil {
		panic(IOException_new_String(Lit("Stream closed")))
	}
}

func (this *PrintStream) flush() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.guard(func() {
		this.ensureOpen()
		this.out.Flush()
	})
}

func (this *PrintStream) close() {
	this.mu.Lock()
	defer this.mu.Unlock()
	if this.closing {
		return
	}
	this.closing = true
	this.guard(func() {
		if b := this.enc.finish(nil); len(b) > 0 {
			this.writeBytes(b)
		}
		this.out.Close()
	})
	this.out = nil
}

// CheckError implements PrintStream.checkError.
func (this *PrintStream) CheckError() bool {
	if this.out != nil {
		this.Flush()
	}
	if ps := printStreamOf(this.out); ps != nil {
		return ps.CheckError()
	}
	return this.trouble
}

func printStreamOf(out *OutputStream) *PrintStream {
	if out == nil {
		return nil
	}
	ps, _ := out.jself.(interface{ JClass_PrintStream() *PrintStream })
	if ps == nil {
		return nil
	}
	return ps.JClass_PrintStream()
}

func (this *PrintStream) writeByte(b int32) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.guard(func() {
		this.ensureOpen()
		this.out.Write_int(b)
		if b == '\n' && this.autoFlush {
			this.out.Flush()
		}
	})
}

func (this *PrintStream) writeArr(buf []int8, off, n int32) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.guard(func() {
		this.ensureOpen()
		this.out.Write_byteArr_int_int(buf, off, n)
		if this.autoFlush {
			this.out.Flush()
		}
	})
}

// writeBytes writes encoded text to the underlying stream, flushing it if
// autoFlush is set, as PrintStream.write(byte[], int, int) does. The caller
// holds mu.
func (this *PrintStream) writeBytes(b []byte) {
	this.ensureOpen()
	this.out.Write_byteArr_int_int(bytesToJava(b), 0, int32(len(b)))
	if this.autoFlush {
		this.out.Flush()
	}
}

// writeText encodes s and writes it, followed by a line separator if
// newline is set. It flushes as Java's write(String) and writeln(String) do.
func (this *PrintStream) writeText(s []uint16, newline bool) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.guard(func() {
		this.ensureOpen()
		b := this.enc.encode(nil, s)
		if newline {
			b = this.enc.encode(b, []uint16{'\n'})
		}
		if len(b) > 0 {
			this.writeBytes(b)
		}
		if this.autoFlush && (newline || containsUnit(s, '\n')) {
			this.out.Flush()
		}
	})
}

func containsUnit(s []uint16, c uint16) bool {
	for _, u := range s {
		if u == c {
			return true
		}
	}
	return false
}

func bytesToJava(b []byte) []int8 {
	out := make([]int8, len(b))
	for i, v := range b {
		out[i] = int8(v)
	}
	return out
}

func (this *PrintStream) print(s *String) {
	this.writeText(stringValueOf(s).units, false)
}

func (this *PrintStream) println(s *String) {
	this.writeText(stringValueOf(s).units, true)
}

// stringValueOf implements String.valueOf for a String: null becomes "null".
func stringValueOf(s *String) *String {
	if s == nil {
		return Lit("null")
	}
	return s
}

func (this *PrintStream) Print_boolean(b bool)   { this.print(StrOfBool(b)) }
func (this *PrintStream) Print_char(c uint16)    { this.print(StrOfChar(c)) }
func (this *PrintStream) Print_int(i int32)      { this.print(StrOfInt(i)) }
func (this *PrintStream) Print_long(l int64)     { this.print(StrOfLong(l)) }
func (this *PrintStream) Print_float(f float32)  { this.print(StrOfFloat(f)) }
func (this *PrintStream) Print_double(d float64) { this.print(StrOfDouble(d)) }
func (this *PrintStream) Print_String(s *String) { this.print(s) }
func (this *PrintStream) Print_Object(o any)     { this.print(StrOfObject(o)) }

func (this *PrintStream) Print_charArr(s []uint16) {
	requireArray(s)
	this.writeText(s, false)
}

func (this *PrintStream) Println()                 { this.writeText(nil, true) }
func (this *PrintStream) Println_boolean(b bool)   { this.println(StrOfBool(b)) }
func (this *PrintStream) Println_char(c uint16)    { this.println(StrOfChar(c)) }
func (this *PrintStream) Println_int(i int32)      { this.println(StrOfInt(i)) }
func (this *PrintStream) Println_long(l int64)     { this.println(StrOfLong(l)) }
func (this *PrintStream) Println_float(f float32)  { this.println(StrOfFloat(f)) }
func (this *PrintStream) Println_double(d float64) { this.println(StrOfDouble(d)) }
func (this *PrintStream) Println_String(s *String) { this.println(s) }
func (this *PrintStream) Println_Object(o any)     { this.println(StrOfObject(o)) }

func (this *PrintStream) Println_charArr(s []uint16) {
	requireArray(s)
	this.writeText(s, true)
}

// Append_CharSequence implements PrintStream.append(CharSequence). The
// result is the stream, typed as the root declaration Appendable.append
// returns.
func (this *PrintStream) Append_CharSequence(csq any) any {
	this.print(StrOfObject(csq))
	return Ref(this)
}

func (this *PrintStream) Append_CharSequence_int_int(csq any, start, end int32) any {
	return this.Append_CharSequence(subSequence(csq, start, end))
}

func (this *PrintStream) Append_char(c uint16) any {
	this.Print_char(c)
	return Ref(this)
}

// Printf_String_ObjectArr implements PrintStream.printf(String, Object...).
// args is a Java Object[] of any element type.
func (this *PrintStream) Printf_String_ObjectArr(format *String, args any) *PrintStream {
	return this.Format_String_ObjectArr(format, args)
}

func (this *PrintStream) Printf_Locale_String_ObjectArr(l *Locale, format *String, args any) *PrintStream {
	return this.Format_Locale_String_ObjectArr(l, format, args)
}

// Format_String_ObjectArr implements PrintStream.format(String, Object...),
// formatting in the default locale.
func (this *PrintStream) Format_String_ObjectArr(format *String, args any) *PrintStream {
	return this.Format_Locale_String_ObjectArr(Locale_getDefault(), format, args)
}

// Format_Locale_String_ObjectArr implements PrintStream.format(Locale, String,
// Object...). Unlike Java, which writes the output of a format string as it
// goes, it writes nothing when an argument does not match its specifier.
func (this *PrintStream) Format_Locale_String_ObjectArr(l *Locale, format *String, args any) *PrintStream {
	if this.out == nil {
		this.trouble = true
		return this
	}
	this.print(Format(l, format, varargs(args)))
	return this
}

// ---------------------------------------------------------------------------
// System.out and System.err

// stdSink is the Go destination of System.out or System.err.
type stdSink struct {
	mu sync.Mutex
	w  io.Writer
}

func (s *stdSink) set(w io.Writer) io.Writer {
	s.mu.Lock()
	defer s.mu.Unlock()
	old := s.w
	s.w = w
	return old
}

func (s *stdSink) write(b []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.w.Write(b)
	return err
}

// sinkOutputStream is an OutputStream writing to a stdSink.
type sinkOutputStream struct {
	OutputStream
	sink *stdSink
}

func (this *sinkOutputStream) JClassName() string { return "java.io.BufferedOutputStream" }

func (this *sinkOutputStream) writeByte(b int32) {
	this.writeArr([]int8{int8(b)}, 0, 1)
}

func (this *sinkOutputStream) writeArr(b []int8, off, n int32) {
	requireArray(b)
	checkFromIndexSize(off, n, int32(len(b)))
	buf := make([]byte, n)
	for i := range buf {
		buf[i] = byte(b[int(off)+i])
	}
	if err := this.sink.write(buf); err != nil {
		panic(IOException_new_String(FromGo(err.Error())))
	}
}

func newSinkPrintStream(sink *stdSink) *PrintStream {
	out := &sinkOutputStream{sink: sink}
	out.jself = out
	return PrintStream_new_OutputStream_boolean(&out.OutputStream, true)
}

// stdoutSink and stderrSink receive the bytes written to System.out and
// System.err. Their writers are replaced only through SetStdoutWriter and
// SetStderrWriter, which lock them.
var (
	stdoutSink = &stdSink{w: os.Stdout}
	stderrSink = &stdSink{w: os.Stderr}
)

// System_out is the static System.out: an auto-flushing UTF-8 PrintStream
// writing to os.Stdout unless redirected.
var System_out = newSinkPrintStream(stdoutSink)

// System_err is the static System.err, writing to os.Stderr unless
// redirected.
var System_err = newSinkPrintStream(stderrSink)

// SetStdoutWriter makes the stream behind System.out write to w instead of
// its current writer, which it returns.
func SetStdoutWriter(w io.Writer) io.Writer { return stdoutSink.set(w) }

// SetStderrWriter makes the stream behind System.err write to w instead of
// its current writer, which it returns.
func SetStderrWriter(w io.Writer) io.Writer { return stderrSink.set(w) }
