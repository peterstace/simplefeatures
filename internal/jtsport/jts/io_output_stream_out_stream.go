package jts

import "io"

// Io_OutputStreamOutStream is an adapter to allow an io.Writer to be used as
// an Io_OutStream.
type Io_OutputStreamOutStream struct {
	os io.Writer
}

// Io_NewOutputStreamOutStream creates a new OutputStreamOutStream wrapping the
// given io.Writer.
func Io_NewOutputStreamOutStream(os io.Writer) *Io_OutputStreamOutStream {
	return &Io_OutputStreamOutStream{os: os}
}

// Write writes len bytes from buf to the output stream.
func (s *Io_OutputStreamOutStream) Write(buf []byte, length int) error {
	_, err := s.os.Write(buf[:length])
	return err
}
