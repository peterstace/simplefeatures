package jts

// Io_OutStream is an interface for classes providing an output stream of bytes.
// This interface is similar to Go's io.Writer, but with a narrower interface
// to make it easier to implement.
type Io_OutStream interface {
	// Write writes len bytes from buf to the output stream.
	Write(buf []byte, length int) error
}
