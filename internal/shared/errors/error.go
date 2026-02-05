package errors

import (
	"context"
	"errors"
	"fmt"
	"maps"
)

type Error struct {
	Reason ErrorReason    `json:"reason"`
	Msg    string         `json:"msg"`
	Data   map[string]any `json:"data"`
	cause  error
}

func New(reason ErrorReason, message string, data map[string]any) *Error {
	// 如果未提供消息，使用映射表中的默认消息
	if message == "" {
		if msg, ok := defaultErrorMessages[reason]; ok {
			message = msg
		} else {
			// 提供回退消息，确保 Msg 不为空
			message = "未知错误"
		}
	}

	// 确保 data 不为 nil
	if data == nil {
		data = make(map[string]any)
	}

	return &Error{
		Reason: reason,
		Msg:    message,
		Data:   data,
	}
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("error: reason = %s msg = %s data = %v cause = %v", e.Reason, e.Msg, e.Data, e.cause)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func (e *Error) Is(err error) bool {
	if e == nil {
		return err == nil
	}
	if err == nil {
		return false
	}
	if se, ok := err.(*Error); ok {
		return se.Reason == e.Reason
	}
	var se *Error
	if errors.As(err, &se) {
		return se.Reason == e.Reason
	}
	return false
}

func (e *Error) WithCause(cause error) *Error {
	if e == nil {
		return nil
	}
	err := Clone(e)
	if err != nil {
		err.cause = cause
	}
	return err
}

// WithField 添加单个字段到错误上下文
func (e *Error) WithField(key string, value any) *Error {
	if e == nil {
		return nil
	}
	err := Clone(e)
	if err.Data == nil {
		err.Data = make(map[string]any)
	}
	err.Data[key] = value
	return err
}

func (e *Error) WithFields(md map[string]any) *Error {
	if e == nil {
		return nil
	}

	if len(md) == 0 {
		return e
	}

	err := Clone(e)
	maps.Copy(err.Data, md)
	return err
}

func (e *Error) Fields() map[string]any {
	if e == nil {
		return map[string]any{
			"reason": "ok",
			"msg":    "",
			"data":   map[string]any{},
		}
	}
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
	var reason ErrorReason = ReasonUnknown
	if errors.Is(err, context.Canceled) {
		reason = ReasonCanceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		reason = ReasonDeadlineExceeded
	}
	return &Error{
		Reason: reason,
		Msg:    defaultErrorMessages[reason],
		Data:   nil,
		cause:  err,
	}
}
