package errors

func ErrorResponse(code int, err *Error) map[string]any {
	if err == nil {
		return nil
	}
	response := err.ToMap()
	response["code"] = code
	return response
}
