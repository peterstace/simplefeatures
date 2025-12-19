package jts

// Io_ByteOrderDataInStream allows reading a stream of Go primitive datatypes
// from an underlying Io_InStream, with the representation being in either
// common byte ordering.
type Io_ByteOrderDataInStream struct {
	byteOrder int
	stream    Io_InStream
	// Buffers to hold primitive datatypes.
	buf1    []byte
	buf4    []byte
	buf8    []byte
	bufLast []byte

	count int64
}

// Io_NewByteOrderDataInStream creates a new ByteOrderDataInStream with no
// underlying stream.
func Io_NewByteOrderDataInStream() *Io_ByteOrderDataInStream {
	return &Io_ByteOrderDataInStream{
		byteOrder: Io_ByteOrderValues_BIG_ENDIAN,
		stream:    nil,
		buf1:      make([]byte, 1),
		buf4:      make([]byte, 4),
		buf8:      make([]byte, 8),
	}
}

// Io_NewByteOrderDataInStreamFromStream creates a new ByteOrderDataInStream
// with the given underlying stream.
func Io_NewByteOrderDataInStreamFromStream(stream Io_InStream) *Io_ByteOrderDataInStream {
	return &Io_ByteOrderDataInStream{
		byteOrder: Io_ByteOrderValues_BIG_ENDIAN,
		stream:    stream,
		buf1:      make([]byte, 1),
		buf4:      make([]byte, 4),
		buf8:      make([]byte, 8),
	}
}

// SetInStream allows a single ByteOrderDataInStream to be reused on multiple
// InStreams.
func (s *Io_ByteOrderDataInStream) SetInStream(stream Io_InStream) {
	s.stream = stream
}

// SetOrder sets the byte ordering using the codes in Io_ByteOrderValues.
func (s *Io_ByteOrderDataInStream) SetOrder(byteOrder int) {
	s.byteOrder = byteOrder
}

// GetCount returns the number of bytes read from the stream.
func (s *Io_ByteOrderDataInStream) GetCount() int64 {
	return s.count
}

// GetData returns the data item that was last read from the stream.
func (s *Io_ByteOrderDataInStream) GetData() []byte {
	return s.bufLast
}

// ReadByte reads a byte value.
func (s *Io_ByteOrderDataInStream) ReadByte() (byte, error) {
	if err := s.read(s.buf1); err != nil {
		return 0, err
	}
	return s.buf1[0], nil
}

// ReadInt reads an int32 value.
func (s *Io_ByteOrderDataInStream) ReadInt() (int32, error) {
	if err := s.read(s.buf4); err != nil {
		return 0, err
	}
	return Io_ByteOrderValues_GetInt(s.buf4, s.byteOrder), nil
}

// ReadLong reads an int64 value.
func (s *Io_ByteOrderDataInStream) ReadLong() (int64, error) {
	if err := s.read(s.buf8); err != nil {
		return 0, err
	}
	return Io_ByteOrderValues_GetLong(s.buf8, s.byteOrder), nil
}

// ReadDouble reads a float64 value.
func (s *Io_ByteOrderDataInStream) ReadDouble() (float64, error) {
	if err := s.read(s.buf8); err != nil {
		return 0, err
	}
	return Io_ByteOrderValues_GetDouble(s.buf8, s.byteOrder), nil
}

func (s *Io_ByteOrderDataInStream) read(buf []byte) error {
	num, err := s.stream.Read(buf)
	if err != nil {
		return err
	}
	if num < len(buf) {
		return Io_NewParseException("Attempt to read past end of input")
	}
	s.bufLast = buf
	s.count += int64(num)
	return nil
}
