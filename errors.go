// Package errors is a simple error wrapping package that automatically
// adds source locations to errors. It has the same API as
// github.com/pkg/errors but is much lighter weight.
//
// If the envar "DEBUG=1" is true, any errors from this package that
// are printed will display <file>:<line> annotations at each wrapping
// location.
package errors

import (
	"errors" // nolint: depguard
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Debug is set by the envar "$DEBUG" but may be overridden.
var Debug = os.Getenv("DEBUG") != ""

type herr struct {
	cause error
	file  string
	line  int
	msg   string
}

func (h *herr) Error() string {
	return h.format(Debug)
}

func (h *herr) Unwrap() error {
	return h.cause
}

func (h *herr) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprint(s, h.format(true))
		} else {
			fmt.Fprint(s, h.Error())
		}
	case 's':
		fmt.Fprint(s, h.format(Debug))
	case 'q':
		fmt.Fprintf(s, "%q", h.format(Debug))
	}
}

func (h *herr) format(trace bool) string {
	var msg string
	if trace {
		msg += fmt.Sprintf("%s:%d", h.file, h.line)
		if h.msg != "" {
			msg += ": " + h.msg
		}
	} else {
		msg += h.msg
	}
	if h.cause != nil {
		if msg != "" {
			msg += ": "
		}
		if trace {
			msg += fmt.Sprintf("%+v", h.cause)
		} else {
			msg += h.cause.Error()
		}
	}
	return msg
}

var pkgPrefix = func() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(file)) + "/"
}()

func newErr(cause error, msg string) error {
	_, file, line, _ := runtime.Caller(2)
	file = strings.TrimPrefix(file, pkgPrefix)
	return &herr{cause: cause, file: file, line: line, msg: msg}
}

// New creates a new error.
func New(message string) error {
	return newErr(nil, message)
}

// Errorf creates a new error using fmt.Sprintf().
func Errorf(format string, args ...interface{}) error {
	return newErr(nil, fmt.Sprintf(format, args...))
}

// Wrap chains a new error to "err" if it is not nil.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return newErr(err, message)
}

// Wrapf chains a new fmt.Sprintf() formatted error to "err" if "err" is not nil.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return newErr(err, fmt.Sprintf(format, args...))
}

// Is mirrors the stdlib errors.Is function.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As mirrors the stdlib errors.As function.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Innermost returns true if err cannot be further unwrapped.
func Innermost(err error) bool {
	if err, ok := err.(interface{ Unwrap() []error }); ok && len(err.Unwrap()) > 0 {
		return false
	}
	if err, ok := err.(interface{ Unwrap() error }); ok && err.Unwrap() != nil {
		return false
	}
	return true
}

// Unwrap aliases the stdlib errors.Unwrap function.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// UnwrapAll recursively unwraps all errors in err, including all intermediate errors.
func UnwrapAll(err error) []error {
	out := []error{}
	if inner, ok := err.(interface{ Unwrap() []error }); ok {
		for _, e := range inner.Unwrap() {
			out = append(out, UnwrapAll(e)...)
		}
		return out
	}
	if inner, ok := err.(interface{ Unwrap() error }); ok && inner.Unwrap() != nil {
		out = append(out, UnwrapAll(inner.Unwrap())...)
	}
	out = append(out, err)
	return out
}

// WithStack chains source location information to an error if "err" is not nil.
func WithStack(err error) error {
	if err == nil {
		return nil
	}
	return newErr(err, "")
}
