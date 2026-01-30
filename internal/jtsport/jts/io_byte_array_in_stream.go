package jts

// Io_ByteArrayInStream allows an array of bytes to be used as an Io_InStream.
// To optimize memory usage, instances can be reused with different byte arrays.
type Io_ByteArrayInStream struct {
	// Implementation improvement suggested by Andrea Aime - Dec 15 2007.
	buffer   []byte
	position int
}

// Io_NewByteArrayInStream creates a new stream based on the given buffer.
func Io_NewByteArrayInStream(buffer []byte) *Io_ByteArrayInStream {
	s := &Io_ByteArrayInStream{}
	s.SetBytes(buffer)
	return s
}

// SetBytes sets this stream to read from the given buffer.
func (s *Io_ByteArrayInStream) SetBytes(buffer []byte) {
	s.buffer = buffer
	s.position = 0
}

// Read reads up to len(buf) bytes from the stream into the given byte buffer.
// Returns the number of bytes read.
func (s *Io_ByteArrayInStream) Read(buf []byte) (int, error) {
	numToRead := len(buf)
	// Don't try and copy past the end of the input.
	if s.position+numToRead > len(s.buffer) {
		numToRead = len(s.buffer) - s.position
		copy(buf, s.buffer[s.position:s.position+numToRead])
		// Zero out the unread bytes.
		for i := numToRead; i < len(buf); i++ {
			buf[i] = 0
		}
	} else {
		copy(buf, s.buffer[s.position:s.position+numToRead])
	}
	s.position += numToRead
	return numToRead, nil
}
