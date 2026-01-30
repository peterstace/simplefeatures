package jts

// Io_ParseException is thrown by a WKTReader when a parsing problem occurs.
type Io_ParseException struct {
	message string
	cause   error
}

// Io_NewParseException creates a ParseException with the given detail message.
func Io_NewParseException(message string) *Io_ParseException {
	return &Io_ParseException{message: message}
}

// Io_NewParseExceptionFromError creates a ParseException with the error's detail message.
func Io_NewParseExceptionFromError(e error) *Io_ParseException {
	return &Io_ParseException{message: e.Error(), cause: e}
}

// Io_NewParseExceptionWithCause creates a ParseException with a message and cause.
func Io_NewParseExceptionWithCause(message string, e error) *Io_ParseException {
	return &Io_ParseException{message: message, cause: e}
}

// Error implements the error interface.
func (p *Io_ParseException) Error() string {
	return p.message
}

// Unwrap returns the underlying cause for errors.Is/As support.
func (p *Io_ParseException) Unwrap() error {
	return p.cause
}
