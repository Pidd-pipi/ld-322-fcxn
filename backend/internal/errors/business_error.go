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
	ErrNotFound       = New(40401, "资源不存在", http.StatusNotFound)
	ErrValidation     = New(40001, "请求参数不合法", http.StatusBadRequest)
	ErrUnauthorized   = New(40101, "认证失败", http.StatusUnauthorized)
	ErrAlertHandled   = New(40901, "该报警已确认，请勿重复操作", http.StatusConflict)
	ErrAlertRecovered = New(40902, "报警已恢复，无需确认", http.StatusConflict)
	ErrInternal       = New(50001, "服务器内部错误", http.StatusInternalServerError)
)
