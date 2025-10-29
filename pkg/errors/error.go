package errors

import (
	"errors"
	"fmt"
	"maps"
)

type Error struct {
	Code   int            `json:"code"`
	Reason string         `json:"reason"`
	Msg    string         `json:"msg"`
	Data   map[string]any `json:"data"`
	cause  error
}

func New(code int, reason, message string, data map[string]any) *Error {
	return &Error{
		Code:   code,
		Reason: reason,
		Msg:    message,
		Data:   data,
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("error: code=%d reason = %s msg = %s data = %v cause = %v", e.Code, e.Reason, e.Msg, e.Data, e.cause)
}

func (e *Error) Unwrap() error { return e.cause }

func (e *Error) Is(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == e.Code && se.Reason == e.Reason
	}
	return false
}

func (e *Error) WithCause(cause error) *Error {
	err := Clone(e)
	err.cause = cause
	return err
}

func (e *Error) WithData(md map[string]any) *Error {
	err := Clone(e)
	err.Data = md
	return err
}

func (e *Error) Reply() map[string]any {
	data := e.Data
	if e.cause != nil {
		data = make(map[string]any, len(e.Data)+1)
		maps.Copy(data, e.Data)
		data["cause"] = e.cause.Error()
	}
	if data == nil {
		data = map[string]any{}
	}
	return map[string]any{
		"code":   e.Code,
		"reason": e.Reason,
		"msg":    e.Msg,
		"data":   data,
	}
}

func Clone(err *Error) *Error {
	if err == nil {
		return nil
	}
	metadata := make(map[string]any, len(err.Data))
	maps.Copy(metadata, err.Data)
	return &Error{
		Code:   err.Code,
		Reason: err.Reason,
		Msg:    err.Msg,
		Data:   metadata,
		cause:  err.cause,
	}
}

func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	if se := new(Error); errors.As(err, &se) {
		return se
	}
	return &Error{
		Code:   500,
		Reason: RsonUnknown.Reason,
		Msg:    RsonUnknown.Msg,
		Data:   nil,
		cause:  err,
	}
}
