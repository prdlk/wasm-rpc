package server

import "errors"

// Code is a coarse, transport-level status code, mirroring the subset of
// gRPC codes that make sense inside a browser-resident module.
type Code string

const (
	CodeInvalidArgument Code = "invalid_argument"
	CodeUnimplemented   Code = "unimplemented"
	CodeInternal        Code = "internal"
	CodeUnknown         Code = "unknown"
)

// Error is the standardized error payload used to reject the JS Promise.
// It crosses the boundary as a plain JS object:
//
//	{ code: string, message: string, method: string }
type Error struct {
	Code    Code
	Method  string
	Message string
	// detail carries the recovered stack trace for panics. It is logged
	// to the browser console but intentionally NOT included in the
	// rejection payload handed to application code.
	detail string
}

func (e *Error) Error() string {
	return string(e.Code) + ": " + e.Message
}

// Errorf builds a typed handler error that survives to the TS client
// with its code intact.
func Errorf(code Code, msg string) *Error {
	return &Error{Code: code, Message: msg}
}

func asError(err error, method string) *Error {
	var rpcErr *Error
	if errors.As(err, &rpcErr) {
		if rpcErr.Method == "" {
			rpcErr.Method = method
		}
		return rpcErr
	}
	return &Error{Code: CodeUnknown, Method: method, Message: err.Error()}
}
