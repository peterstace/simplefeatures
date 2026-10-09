package java

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// File is java.io.File on a Unix file system.
type File struct {
	jhash        int32
	path         *String
	prefixLength int32
}

func (this *File) JClass_File() *File    { return this }
func (this *File) JClassName() string    { return "java.io.File" }
func (this *File) JIdentityHash() *int32 { return &this.jhash }

func File_isInstance(o any) bool { _, ok := o.(*File); return ok }

func File_fromAny(o any) *File { return AnyTo[*File](o) }

// File_separatorChar is the static File.separatorChar.
var File_separatorChar uint16 = '/'

// File_separator is the static File.separator.
var File_separator = Lit("/")

// File_pathSeparatorChar is the static File.pathSeparatorChar.
var File_pathSeparatorChar uint16 = ':'

// File_pathSeparator is the static File.pathSeparator.
var File_pathSeparator = Lit(":")

func newFile(path string) *File {
	f := &File{path: FromGo(path)}
	if len(path) > 0 && path[0] == '/' {
		f.prefixLength = 1
	}
	return f
}

func File_new_String(pathname *String) *File {
	if pathname == nil {
		panic(NullPointerException_new())
	}
	return newFile(normalizePath(pathname.Go()))
}

func File_new_String_String(parent, child *String) *File {
	if child == nil {
		panic(NullPointerException_new())
	}
	switch {
	case parent == nil:
		return newFile(normalizePath(child.Go()))
	case parent.IsEmpty():
		return newFile(resolvePath("/", normalizePath(child.Go())))
	}
	return newFile(resolvePath(normalizePath(parent.Go()), normalizePath(child.Go())))
}

func File_new_File_String(parent *File, child *String) *File {
	if child == nil {
		panic(NullPointerException_new())
	}
	switch {
	case parent == nil:
		return newFile(normalizePath(child.Go()))
	case parent.path.IsEmpty():
		return newFile(resolvePath("/", normalizePath(child.Go())))
	}
	return newFile(resolvePath(parent.path.Go(), normalizePath(child.Go())))
}

// normalizePath is UnixFileSystem.normalize: it removes repeated and
// trailing slashes.
func normalizePath(p string) string {
	n := len(p)
	off := -1
	for i := 0; i+1 < n; i++ {
		if p[i] == '/' && p[i+1] == '/' {
			off = i
			break
		}
	}
	if off < 0 {
		if n > 0 && p[n-1] == '/' {
			off = n - 1
		} else {
			return p
		}
	}
	for n > off && p[n-1] == '/' {
		n--
	}
	if n == 0 {
		return "/"
	}
	if n == off {
		return p[:off]
	}
	out := make([]byte, 0, n)
	out = append(out, p[:off]...)
	var prev byte
	for i := off; i < n; i++ {
		c := p[i]
		if prev == '/' && c == '/' {
			continue
		}
		out = append(out, c)
		prev = c
	}
	return string(out)
}

// resolvePath is UnixFileSystem.resolve(String, String).
func resolvePath(parent, child string) string {
	trim := func(s string) string {
		if len(s) > 1 && s[len(s)-1] == '/' {
			return s[:len(s)-1]
		}
		return s
	}
	switch {
	case child == "":
		return parent
	case child[0] == '/':
		if parent == "/" {
			return child
		}
		return trim(parent + child)
	case parent == "/":
		return trim(parent + child)
	}
	return trim(parent + "/" + child)
}

func (this *File) isInvalid() bool {
	return this.path.IndexOf_int(0) >= 0
}

func (this *File) GetName() *String {
	index := this.path.LastIndexOf_int('/')
	if index < this.prefixLength {
		return this.path.Substring_int(this.prefixLength)
	}
	return this.path.Substring_int(index + 1)
}

func (this *File) GetParent() *String {
	index := this.path.LastIndexOf_int('/')
	if index < this.prefixLength {
		if this.prefixLength > 0 && this.path.Length() > this.prefixLength {
			return this.path.Substring_int_int(0, this.prefixLength)
		}
		return nil
	}
	return this.path.Substring_int_int(0, index)
}

func (this *File) GetParentFile() *File {
	p := this.GetParent()
	if p == nil {
		return nil
	}
	return newFile(p.Go())
}

func (this *File) GetPath() *String { return this.path }

func (this *File) IsAbsolute() bool { return this.prefixLength != 0 }

func (this *File) GetAbsolutePath() *String {
	if this.IsAbsolute() {
		return this.path
	}
	wd, err := os.Getwd()
	if err != nil {
		panic(IOException_new_String(FromGo(err.Error())))
	}
	return FromGo(resolvePath(wd, this.path.Go()))
}

func (this *File) GetAbsoluteFile() *File {
	return newFile(this.GetAbsolutePath().Go())
}

// GetCanonicalPath implements File.getCanonicalPath: the absolute path with
// "." and ".." removed and symbolic links resolved.
func (this *File) GetCanonicalPath() *String {
	if this.isInvalid() {
		panic(IOException_new_String(Lit("Invalid file path")))
	}
	p := filepath.Clean(this.GetAbsolutePath().Go())
	if r, err := filepath.EvalSymlinks(p); err == nil {
		p = r
	}
	return FromGo(p)
}

func (this *File) GetCanonicalFile() *File {
	return newFile(this.GetCanonicalPath().Go())
}

func (this *File) stat() (fs.FileInfo, bool) {
	if this.isInvalid() {
		return nil, false
	}
	st, err := os.Stat(this.path.Go())
	return st, err == nil
}

func (this *File) Exists() bool {
	_, ok := this.stat()
	return ok
}

func (this *File) IsDirectory() bool {
	st, ok := this.stat()
	return ok && st.IsDir()
}

func (this *File) IsFile() bool {
	st, ok := this.stat()
	return ok && st.Mode().IsRegular()
}

func (this *File) LastModified() int64 {
	st, ok := this.stat()
	if !ok {
		return 0
	}
	return st.ModTime().UnixMilli()
}

func (this *File) Length() int64 {
	st, ok := this.stat()
	if !ok {
		return 0
	}
	return st.Size()
}

func (this *File) CreateNewFile() bool {
	if this.isInvalid() {
		panic(IOException_new_String(Lit("Invalid file path")))
	}
	f, err := os.OpenFile(this.path.Go(), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o666)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return false
		}
		panic(IOException_new_String(FromGo(errnoText(err))))
	}
	f.Close()
	return true
}

func (this *File) Delete() bool {
	return !this.isInvalid() && os.Remove(this.path.Go()) == nil
}

// List implements File.list. The names are in sorted order; Java returns
// them in the order the file system lists them.
func (this *File) List() []*String {
	if this.isInvalid() {
		return nil
	}
	entries, err := os.ReadDir(this.path.Go())
	if err != nil {
		return nil
	}
	out := make([]*String, len(entries))
	for i, e := range entries {
		out[i] = FromGo(e.Name())
	}
	return out
}

// ListFiles implements File.listFiles, in the order of List.
func (this *File) ListFiles() []*File {
	names := this.List()
	if names == nil {
		return nil
	}
	out := make([]*File, len(names))
	for i, n := range names {
		out[i] = newFile(resolvePath(this.path.Go(), n.Go()))
	}
	return out
}

func (this *File) Mkdir() bool {
	return !this.isInvalid() && os.Mkdir(this.path.Go(), 0o777) == nil
}

func (this *File) Mkdirs() bool {
	if this.Exists() {
		return false
	}
	if this.Mkdir() {
		return true
	}
	canon := this.GetCanonicalFile()
	parent := canon.GetParentFile()
	return parent != nil && (parent.Mkdirs() || parent.Exists()) && canon.Mkdir()
}

func (this *File) RenameTo_File(dest *File) bool {
	if dest == nil {
		panic(NullPointerException_new())
	}
	if this.isInvalid() || dest.isInvalid() {
		return false
	}
	return os.Rename(this.path.Go(), dest.path.Go()) == nil
}

func (this *File) CompareTo_Object(o any) int32 {
	return this.path.CompareTo_Object(AnyTo[*File](o).path)
}

func (this *File) Equals_Object(o any) bool {
	f, ok := o.(*File)
	return ok && f != nil && this.CompareTo_Object(f) == 0
}

func (this *File) HashCode() int32 { return this.path.HashCode() ^ 1234321 }

func (this *File) ToString() *String { return this.path }
