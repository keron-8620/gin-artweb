package errors

import (
	"context"
	"net/http"
	"strings"
)

var UnknowError = New(
	http.StatusInternalServerError,
	RsonUnknown.Reason,
	RsonUnknown.Msg,
	nil,
)

var ValidateError = New(
	http.StatusBadRequest,
	RsonValidation.Reason,
	RsonValidation.Msg,
	nil,
)

var (
	ErrCtxCancel = New(
		http.StatusBadRequest,
		RsonCtxCancel.Reason,
		RsonCtxCancel.Msg,
		nil,
	)
	ErrCtxDeadlineExceeded = New(
		http.StatusBadRequest,
		RsonCtxDeadlineExceeded.Reason,
		RsonCtxDeadlineExceeded.Msg,
		nil,
	)
	ErrCtxTimeOut = New(
		http.StatusBadRequest,
		RsonCtxTimeOut.Reason,
		RsonCtxTimeOut.Msg,
		nil,
	)
	ErrCtxInvalid = New(
		http.StatusBadRequest,
		RsonCtxInvalid.Reason,
		RsonCtxInvalid.Msg,
		nil,
	)
	ErrCtxMissing = New(
		http.StatusBadRequest,
		RsonCtxMissing.Reason,
		RsonCtxMissing.Msg,
		nil,
	)
	ErrCtxValueMissing = New(
		http.StatusBadRequest,
		RsonCtxValueMissing.Reason,
		RsonCtxValueMissing.Msg,
		nil,
	)
	ErrCtxTypeMismatch = New(
		http.StatusBadRequest,
		RsonCtxTypeMismatch.Reason,
		RsonCtxTypeMismatch.Msg,
		nil,
	)
	ErrCtxPropagationFailed = New(
		http.StatusBadRequest,
		RsonCtxPropagationFailed.Reason,
		RsonCtxPropagationFailed.Msg,
		nil,
	)
)

func FromCtxError(err error) *Error {
	if err == nil {
		return nil
	}

	switch err {
	case context.Canceled:
		return ErrCtxCancel.WithCause(err)
	case context.DeadlineExceeded:
		return ErrCtxDeadlineExceeded.WithCause(err)
	}

	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "timeout"):
		return ErrCtxTimeOut.WithCause(err)
	case strings.Contains(errStr, "invalid context") ||
		strings.Contains(errStr, "context invalid"):
		return ErrCtxInvalid.WithCause(err)
	default:
		return UnknowError.WithCause(err)
	}
}
