package jts

// Io_InStream is an interface for classes providing an input stream of bytes.
// This interface is similar to Go's io.Reader, but with a narrower interface
// to make it easier to implement.
type Io_InStream interface {
	// Read reads buf's length bytes from the input stream and stores them in
	// the supplied buffer. Returns the number of bytes read, or -1 if at
	// end-of-file.
	Read(buf []byte) (int, error)
}
