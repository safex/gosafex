package config

import (
	"fmt"
)

// Error is the configuraton
type Error struct {
	msg    string
	nested error
}

// NewError constructs a new config error
func NewError(msg string, err error) *Error {
	return &Error{msg, err}
}

func (e *Error) Error() string {
	if e.msg == "" && e.nested == nil {
		return "Config error: unknown"
	} else if e.nested == nil {
		return fmt.Sprintf("Config error: %s", e.msg)
	} else {
		return fmt.Sprintf("Config error: %s, error = %v", e.msg, e.nested)
	}
}
