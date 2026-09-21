package apperrors

import "net/http"

type BusinessError struct {
	Code    int
	Message string
	Status  int
}

func (e *BusinessError) Error() string { return e.Message }
func New(code int, message string, status int) *BusinessError {
	return &BusinessError{Code: code, Message: message, Status: status}
}

var (
	ErrNotFound     = New(40401, "资源不存在", http.StatusNotFound)
	ErrValidation   = New(40001, "请求参数不合法", http.StatusBadRequest)
	ErrUnauthorized = New(40101, "认证失败", http.StatusUnauthorized)
	ErrInternal     = New(50001, "服务器内部错误", http.StatusInternalServerError)
)
