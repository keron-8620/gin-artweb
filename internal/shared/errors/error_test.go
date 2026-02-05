package errors

import (
	"context"
	std_errors "errors"
	"strings"
	"testing"
)

func TestNewError(t *testing.T) {
	// 测试1: 基本错误创建
	err := New(ReasonUnknown, "", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Reason != ReasonUnknown {
		t.Errorf("expected reason %v, got %v", ReasonUnknown, err.Reason)
	}
	if err.Msg != "未知错误" {
		t.Errorf("expected message '未知错误', got '%s'", err.Msg)
	}
	if len(err.Data) != 0 {
		t.Errorf("expected empty data map, got %v", err.Data)
	}

	// 测试2: 带自定义消息的错误创建
	customMsg := "自定义错误消息"
	err = New(ReasonUnknown, customMsg, nil)
	if err.Msg != customMsg {
		t.Errorf("expected message '%s', got '%s'", customMsg, err.Msg)
	}

	// 测试3: 带数据的错误创建
	data := map[string]any{"key": "value"}
	err = New(ReasonUnknown, "", data)
	if len(err.Data) != 1 {
		t.Errorf("expected data map with 1 entry, got %v", err.Data)
	}
	if err.Data["key"] != "value" {
		t.Errorf("expected data['key'] = 'value', got '%v'", err.Data["key"])
	}

	// 测试4: 空数据的处理
	err = New(ReasonUnknown, "", nil)
	if err.Data == nil {
		t.Error("expected non-nil data map, got nil")
	}
	if len(err.Data) != 0 {
		t.Errorf("expected empty data map, got %v", err.Data)
	}
}

func TestErrorMethod(t *testing.T) {
	// 测试1: 正常错误的Error方法
	err := New(ReasonUnknown, "测试错误", nil)
	errorStr := err.Error()
	if errorStr == "" {
		t.Error("expected non-empty error string, got empty")
	}
	if !strings.Contains(errorStr, "测试错误") {
		t.Errorf("expected error string to contain '测试错误', got '%s'", errorStr)
	}

	// 测试2: nil错误的Error方法
	var nilErr *Error
	if nilErr.Error() != "" {
		t.Errorf("expected empty string for nil error, got '%s'", nilErr.Error())
	}
}

func TestUnwrapMethod(t *testing.T) {
	// 测试1: 带cause的错误
	cause := std_errors.New("cause error")
	err := New(ReasonUnknown, "", nil).WithCause(cause)
	if err.Unwrap() != cause {
		t.Errorf("expected cause %v, got %v", cause, err.Unwrap())
	}

	// 测试2: 不带cause的错误
	err = New(ReasonUnknown, "", nil)
	if err.Unwrap() != nil {
		t.Errorf("expected nil cause, got %v", err.Unwrap())
	}

	// 测试3: nil错误的Unwrap方法
	var nilErr *Error
	if nilErr.Unwrap() != nil {
		t.Errorf("expected nil for nil error, got %v", nilErr.Unwrap())
	}
}

func TestIsMethod(t *testing.T) {
	// 测试1: 相同原因的错误比较
	err1 := New(ReasonUnknown, "", nil)
	err2 := New(ReasonUnknown, "", nil)
	if !err1.Is(err2) {
		t.Error("expected err1.Is(err2) to be true for same reason")
	}

	// 测试2: 不同原因的错误比较
	err3 := New(ReasonCanceled, "", nil)
	if err1.Is(err3) {
		t.Error("expected err1.Is(err3) to be false for different reasons")
	}

	// 测试3: 与nil比较
	if err1.Is(nil) {
		t.Error("expected err1.Is(nil) to be false")
	}

	// 测试4: nil错误与nil比较
	var nilErr *Error
	if !nilErr.Is(nil) {
		t.Error("expected nilErr.Is(nil) to be true")
	}

	// 测试5: nil错误与非nil比较
	if nilErr.Is(err1) {
		t.Error("expected nilErr.Is(err1) to be false")
	}
}

func TestWithCauseMethod(t *testing.T) {
	// 测试1: 添加cause
	cause := std_errors.New("cause error")
	err := New(ReasonUnknown, "", nil)
	withCause := err.WithCause(cause)
	if withCause == err {
		t.Error("expected WithCause to return a new error instance")
	}
	if withCause.Unwrap() != cause {
		t.Errorf("expected cause %v, got %v", cause, withCause.Unwrap())
	}

	// 测试2: nil错误的WithCause方法
	var nilErr *Error
	if nilErr.WithCause(cause) != nil {
		t.Error("expected nil for nil error.WithCause")
	}
}

func TestWithFieldMethod(t *testing.T) {
	// 测试1: 添加单个字段
	err := New(ReasonUnknown, "", nil)
	withField := err.WithField("key", "value")
	if withField == err {
		t.Error("expected WithField to return a new error instance")
	}
	if withField.Data["key"] != "value" {
		t.Errorf("expected data['key'] = 'value', got '%v'", withField.Data["key"])
	}

	// 测试2: nil错误的WithField方法
	var nilErr *Error
	if nilErr.WithField("key", "value") != nil {
		t.Error("expected nil for nil error.WithField")
	}
}

func TestWithFieldsMethod(t *testing.T) {
	// 测试1: 添加多个字段
	err := New(ReasonUnknown, "", nil)
	fields := map[string]any{"key1": "value1", "key2": "value2"}
	withFields := err.WithFields(fields)
	if withFields == err {
		t.Error("expected WithFields to return a new error instance")
	}
	if withFields.Data["key1"] != "value1" {
		t.Errorf("expected data['key1'] = 'value1', got '%v'", withFields.Data["key1"])
	}
	if withFields.Data["key2"] != "value2" {
		t.Errorf("expected data['key2'] = 'value2', got '%v'", withFields.Data["key2"])
	}

	// 测试2: 空字段映射
	withEmptyFields := err.WithFields(nil)
	if withEmptyFields != err {
		t.Error("expected WithFields(nil) to return the same error instance")
	}

	// 测试3: nil错误的WithFields方法
	var nilErr *Error
	if nilErr.WithFields(fields) != nil {
		t.Error("expected nil for nil error.WithFields")
	}
}

func TestFieldsMethod(t *testing.T) {
	// 测试1: 正常错误的Fields方法
	err := New(ReasonUnknown, "测试错误", map[string]any{"key": "value"})
	fields := err.Fields()
	if fields["reason"] != ReasonUnknown {
		t.Errorf("expected reason %v, got %v", ReasonUnknown, fields["reason"])
	}
	if fields["msg"] != "测试错误" {
		t.Errorf("expected message '测试错误', got '%v'", fields["msg"])
	}
	data := fields["data"].(map[string]any)
	if data["key"] != "value" {
		t.Errorf("expected data['key'] = 'value', got '%v'", data["key"])
	}

	// 测试2: 带cause的错误的Fields方法
	cause := std_errors.New("cause error")
	err = err.WithCause(cause)
	fields = err.Fields()
	data = fields["data"].(map[string]any)
	if data["cause"] != "cause error" {
		t.Errorf("expected data['cause'] = 'cause error', got '%v'", data["cause"])
	}

	// 测试3: nil错误的Fields方法
	var nilErr *Error
	fields = nilErr.Fields()
	if fields["reason"] != "ok" {
		t.Errorf("expected reason 'ok' for nil error, got '%v'", fields["reason"])
	}
}

func TestCloneMethod(t *testing.T) {
	// 测试1: 克隆错误
	err := New(ReasonUnknown, "测试错误", map[string]any{"key": "value"})
	cloned := Clone(err)
	if cloned == err {
		t.Error("expected Clone to return a new error instance")
	}
	if cloned.Reason != err.Reason {
		t.Errorf("expected cloned reason %v, got %v", err.Reason, cloned.Reason)
	}
	if cloned.Msg != err.Msg {
		t.Errorf("expected cloned message '%s', got '%s'", err.Msg, cloned.Msg)
	}
	if len(cloned.Data) != len(err.Data) {
		t.Errorf("expected cloned data map with %d entries, got %v", len(err.Data), cloned.Data)
	}

	// 测试2: 克隆带cause的错误
	cause := std_errors.New("cause error")
	err = err.WithCause(cause)
	cloned = Clone(err)
	if cloned.Unwrap() != cause {
		t.Errorf("expected cloned cause %v, got %v", cause, cloned.Unwrap())
	}

	// 测试3: 克隆nil错误
	var nilErr *Error
	cloned = Clone(nilErr)
	if cloned != nil {
		t.Error("expected Clone(nil) to return nil")
	}
}

func TestFromError(t *testing.T) {
	// 测试1: 从标准错误创建
	stdErr := std_errors.New("standard error")
	err := FromError(stdErr)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Reason != ReasonUnknown {
		t.Errorf("expected reason %v, got %v", ReasonUnknown, err.Reason)
	}
	if err.Unwrap() != stdErr {
		t.Errorf("expected cause %v, got %v", stdErr, err.Unwrap())
	}

	// 测试2: 从context.Canceled创建
	canceledErr := context.Canceled
	err = FromError(canceledErr)
	if err.Reason != ReasonCanceled {
		t.Errorf("expected reason %v, got %v", ReasonCanceled, err.Reason)
	}

	// 测试3: 从context.DeadlineExceeded创建
	deadlineErr := context.DeadlineExceeded
	err = FromError(deadlineErr)
	if err.Reason != ReasonDeadlineExceeded {
		t.Errorf("expected reason %v, got %v", ReasonDeadlineExceeded, err.Reason)
	}

	// 测试4: 从自定义错误创建
	customErr := New(ReasonUnknown, "", nil)
	err = FromError(customErr)
	if err != customErr {
		t.Error("expected FromError to return the same custom error instance")
	}

	// 测试5: 从nil创建
	err = FromError(nil)
	if err != nil {
		t.Error("expected FromError(nil) to return nil")
	}
}

func TestErrorChain(t *testing.T) {
	// 测试错误链处理
	cause1 := std_errors.New("cause 1")
	cause2 := New(ReasonUnknown, "cause 2", nil).WithCause(cause1)
	err := New(ReasonUnknown, "top level error", nil).WithCause(cause2)

	// 测试错误链的展开
	currentErr := error(err)
	count := 0
	for currentErr != nil {
		count++
		currentErr = std_errors.Unwrap(currentErr)
	}
	if count != 3 {
		t.Errorf("expected error chain with 3 errors, got %d", count)
	}

	// 测试std_errors.Is
	if !std_errors.Is(err, cause1) {
		t.Error("expected std_errors.Is(err, cause1) to be true")
	}
	if !std_errors.Is(err, cause2) {
		t.Error("expected std_errors.Is(err, cause2) to be true")
	}
}
